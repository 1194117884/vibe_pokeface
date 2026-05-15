package doudizhu

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

// Engine implements the Dou Di Zhu game logic.
type Engine struct{}

// Init creates a new game state by shuffling and dealing a 54-card deck
// to the 3 players. Returns an error if not exactly 3 players are provided.
func (e *Engine) Init(players []game.PlayerInfo) (game.GameState, error) {
	if len(players) != 3 {
		return nil, fmt.Errorf("doudizhu requires exactly 3 players, got %d", len(players))
	}

	deck := NewDeck()
	Shuffle(deck)
	h1, h2, h3, remaining := Deal(deck)

	SortCards(h1)
	SortCards(h2)
	SortCards(h3)

	playerHands := make([]PlayerHand, 3)
	for i, info := range players {
		var hand []Card
		switch i {
		case 0:
			hand = h1
		case 1:
			hand = h2
		case 2:
			hand = h3
		}
		playerHands[i] = PlayerHand{
			UserID: info.ID,
			Seat:   info.Seat,
			Hand:   hand,
		}
	}

	state := &GameState{
		Phase:         PhaseCalling,
		Players:       playerHands,
		CurrentSeat:   rand.Intn(3),
		LandlordCards: remaining,
		LandlordSeat:  -1,
		RoundNum:      1,
		Multiplier:    1,
		HasPassed:     make(map[int]bool),
		Revealed:      make(map[int]bool),
		Doubled:       make(map[int]bool),
	}
	return state, nil
}

// ExecuteAction processes a player action and transitions the game state.
func (e *Engine) ExecuteAction(state game.GameState, action game.PlayerAction) (game.GameState, error) {
	gs, ok := state.(*GameState)
	if !ok {
		return nil, fmt.Errorf("invalid state type")
	}
	seat := -1
	for i, p := range gs.Players {
		if p.UserID == action.PlayerID {
			seat = i
			break
		}
	}
	if seat == -1 || seat != gs.CurrentSeat {
		return nil, fmt.Errorf("not your turn")
	}

	switch gs.Phase {
	case PhaseCalling:
		return e.handleCallBid(gs, seat, action)
	case PhaseSnatching:
		return e.handleSnatchBid(gs, seat, action)
	case PhaseRevealing:
		return e.handleRevealBid(gs, seat, action)
	case PhaseDoubling:
		return e.handleDoubleBid(gs, seat, action)
	case PhasePlaying:
		return e.handlePlay(gs, seat, action)
	default:
		return nil, fmt.Errorf("game already ended")
	}
}

// ValidateAction checks whether a player action is valid without mutating state.
func (e *Engine) ValidateAction(state game.GameState, action game.PlayerAction) bool {
	gs, ok := state.(*GameState)
	if !ok {
		return false
	}
	seat := -1
	for i, p := range gs.Players {
		if p.UserID == action.PlayerID {
			seat = i
			break
		}
	}
	if seat == -1 || seat != gs.CurrentSeat {
		return false
	}
	if gs.Phase == PhaseCalling || gs.Phase == PhaseSnatching {
		return action.Action == "bid_call" || action.Action == "bid_pass"
	}
	if gs.Phase == PhaseRevealing {
		return action.Action == "reveal_all" || action.Action == "pass"
	}
	if gs.Phase == PhaseDoubling {
		return action.Action == "double" || action.Action == "no_double"
	}
	if gs.Phase == PhasePlaying {
		if action.Action == "pass" {
			return gs.LastPlay != nil && gs.LastPlay.Seat != seat
		}
		if action.Action != "play" {
			return false
		}
		cards := make([]Card, len(action.Cards))
		for i, id := range action.Cards {
			cards[i] = Card{ID: id}
		}
		play := ParsePlay(cards)
		if play.Type == PlayInvalid && len(cards) > 0 {
			return false
		}
		if !cardsInHand(gs.Players[seat].Hand, cards) {
			return false
		}
		if gs.LastPlay != nil && !CanBeat(play, gs.LastPlay.Play) {
			return false
		}
		return true
	}
	return false
}

// IsRoundEnd returns true if the game has ended.
func (e *Engine) IsRoundEnd(state game.GameState) bool {
	gs, ok := state.(*GameState)
	if !ok {
		return false
	}
	return gs.Phase == PhaseEnded
}

// CalculateScore computes the scores for each player based on the game result
// multiplied by the game multiplier (from 抢地主 snatches).
func (e *Engine) CalculateScore(state game.GameState) ([]game.PlayerScore, error) {
	gs, ok := state.(*GameState)
	if !ok {
		return nil, fmt.Errorf("invalid state type")
	}
	if gs.WinnerSeat == nil {
		return nil, fmt.Errorf("no winner yet")
	}

	scores := make([]game.PlayerScore, 3)
	winnerIsLandlord := gs.Players[*gs.WinnerSeat].IsLandlord
	mult := gs.Multiplier
	if mult < 1 {
		mult = 1
	}

	for i, p := range gs.Players {
		var base int
		if p.IsLandlord {
			if winnerIsLandlord {
				base = 2
			} else {
				base = -2
			}
		} else {
			if winnerIsLandlord {
				base = -1
			} else {
				base = 1
			}
		}
		score := base * mult
		if gs.Doubled[p.Seat] {
			score *= 2
		}
		scores[i] = game.PlayerScore{
			PlayerID: p.UserID,
			Score:    score,
		}
	}
	return scores, nil
}

// SerializeForAI returns a simple text representation of the game state for AI consumption.
func (e *Engine) SerializeForAI(state game.GameState) string {
	gs, ok := state.(*GameState)
	if !ok {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Phase: %d\n", gs.Phase))
	sb.WriteString(fmt.Sprintf("Landlord: seat %d\n", gs.LandlordSeat))
	sb.WriteString(fmt.Sprintf("Current turn: seat %d\n", gs.CurrentSeat))
	sb.WriteString(fmt.Sprintf("Multiplier: %d\n", gs.Multiplier))
	for _, p := range gs.Players {
		role := "farmer"
		if p.IsLandlord {
			role = "landlord"
		}
		sb.WriteString(fmt.Sprintf("Player %d (%s): %d cards\n", p.Seat, role, len(p.Hand)))
	}
	return sb.String()
}

// reDeal shuffles a new deck, re-deals to all players, and resets the
// bidding state (PhaseCalling, random first bidder, cleared history).
func (e *Engine) reDeal(state *GameState) {
	deck := NewDeck()
	Shuffle(deck)
	h1, h2, h3, remaining := Deal(deck)

	SortCards(h1)
	SortCards(h2)
	SortCards(h3)

	for i := range state.Players {
		var hand []Card
		switch i {
		case 0:
			hand = h1
		case 1:
			hand = h2
		case 2:
			hand = h3
		}
		state.Players[i].Hand = hand
		state.Players[i].IsLandlord = false
	}

	state.LandlordCards = remaining
	state.LandlordSeat = -1
	state.Phase = PhaseCalling
	state.CurrentSeat = rand.Intn(3)
	state.BidHistory = nil
	state.HasPassed = make(map[int]bool)
	state.Multiplier = 1
	state.SnatchCount = 0
	state.Revealed = make(map[int]bool)
	state.Doubled = make(map[int]bool)
	state.RevealCount = 0
	state.DoubleCount = 0
	state.LastPlay = nil
	state.ConsecutivePasses = 0
	state.RoundNum++
}

// handleCallBid processes the 叫地主 phase.
// First player to bid_call becomes landlord nominee; bidding immediately
// transitions to PhaseSnatching. Players who passed before the caller are
// marked in HasPassed and cannot participate in snatching.
// If all 3 pass, the round ends with no winner.
func (e *Engine) handleCallBid(state *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action != "bid_call" && action.Action != "bid_pass" {
		return nil, fmt.Errorf("invalid bid action: %s", action.Action)
	}

	state.BidHistory = append(state.BidHistory, BidRecord{
		Seat:   seat,
		Called: action.Action == "bid_call",
	})

	if action.Action == "bid_pass" {
		state.HasPassed[seat] = true
		// All 3 passed — re-deal and restart
		if len(state.BidHistory) >= 3 {
			e.reDeal(state)
			return state, nil
		}
		state.CurrentSeat = (seat + 1) % 3
		return state, nil
	}

	// bid_call — immediately becomes landlord nominee, enter snatching
	state.LandlordSeat = seat
	state.Phase = PhaseSnatching
	state.SnatchCount = 0
	state.CurrentSeat = (seat + 1) % 3
	// Auto-advance past players who already passed during 叫地主
	for state.SnatchCount < 3 && state.HasPassed[state.CurrentSeat] {
		state.SnatchCount++
		state.CurrentSeat = (state.CurrentSeat + 1) % 3
	}
	return state, nil
}

// handleSnatchBid processes the 抢地主 phase.
// Each player gets one chance to snatch (bid_call) or pass.
// Players marked in HasPassed are auto-skipped.
// Each snatch doubles the multiplier and updates the landlord nominee.
// After all 3 players have had their turn, the final nominee becomes landlord.
func (e *Engine) handleSnatchBid(state *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action != "bid_call" && action.Action != "bid_pass" {
		return nil, fmt.Errorf("invalid snatch action: %s", action.Action)
	}

	// Auto-skip players who passed during 叫地主
	if state.HasPassed[seat] {
		state.SnatchCount++
	} else {
		state.BidHistory = append(state.BidHistory, BidRecord{
			Seat:   seat,
			Called: action.Action == "bid_call",
		})

		if action.Action == "bid_call" {
			state.LandlordSeat = seat
			state.Multiplier *= 2
		}
		state.SnatchCount++
	}

	// All 3 have had their turn — finalize landlord
	if state.SnatchCount >= 3 {
		if state.LandlordSeat >= 0 {
			state.Players[state.LandlordSeat].IsLandlord = true
			state.Players[state.LandlordSeat].Hand = append(
				state.Players[state.LandlordSeat].Hand, state.LandlordCards...)
			SortCards(state.Players[state.LandlordSeat].Hand)
			state.Phase = PhaseRevealing
			state.RevealCount = 0
			state.CurrentSeat = state.LandlordSeat
		} else {
			state.Phase = PhaseEnded
		}
		return state, nil
	}

	state.CurrentSeat = (seat + 1) % 3
	// Auto-advance past HasPassed players
	for state.SnatchCount < 3 && state.HasPassed[state.CurrentSeat] {
		state.SnatchCount++
		state.CurrentSeat = (state.CurrentSeat + 1) % 3
	}
	return state, nil
}

// handleRevealBid processes the 明牌 phase.
// Each player can reveal_all (multiplier ×2) or pass. After all 3
// decide, proceed to PhaseDoubling.
func (e *Engine) handleRevealBid(state *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action != "reveal_all" && action.Action != "pass" {
		return nil, fmt.Errorf("invalid reveal action: %s", action.Action)
	}

	if action.Action == "reveal_all" {
		state.Revealed[seat] = true
		state.Multiplier *= 2
	}
	state.RevealCount++

	if state.RevealCount >= 3 {
		state.Phase = PhaseDoubling
		state.DoubleCount = 0
		state.CurrentSeat = state.LandlordSeat
		return state, nil
	}

	state.CurrentSeat = (seat + 1) % 3
	return state, nil
}

// handleDoubleBid processes the 加倍 phase.
// Each player can double or no_double. Choice is recorded in Doubled
// and affects scoring in CalculateScore. After all 3, proceed to PhasePlaying.
func (e *Engine) handleDoubleBid(state *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action != "double" && action.Action != "no_double" {
		return nil, fmt.Errorf("invalid double action: %s", action.Action)
	}

	if action.Action == "double" {
		state.Doubled[seat] = true
	}
	state.DoubleCount++

	if state.DoubleCount >= 3 {
		state.Phase = PhasePlaying
		state.CurrentSeat = state.LandlordSeat
		return state, nil
	}

	state.CurrentSeat = (seat + 1) % 3
	return state, nil
}

// handlePlay processes a card play action during the playing phase.
func (e *Engine) handlePlay(state *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action == "pass" {
		state.ConsecutivePasses++
		if state.ConsecutivePasses >= 2 {
			state.LastPlay = nil
			state.ConsecutivePasses = 0
		}
		state.CurrentSeat = (seat + 1) % 3
		return state, nil
	}

	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	play := ParsePlay(cards)
	if play.Type == PlayInvalid {
		return nil, fmt.Errorf("invalid card combination")
	}

	lastPlay := state.LastPlay
	if lastPlay != nil {
		if !CanBeat(play, lastPlay.Play) {
			return nil, fmt.Errorf("cannot beat last play")
		}
	}

	newHand := removeCardsFromHand(state.Players[seat].Hand, cards)
	state.Players[seat].Hand = newHand

	record := &PlayRecord{
		Seat:  seat,
		Play:  play,
		Cards: cards,
	}
	state.LastPlay = record
	state.LastPlaySeat = seat
	state.ConsecutivePasses = 0

	if len(newHand) == 0 {
		state.Phase = PhaseEnded
		state.WinnerSeat = &seat
		return state, nil
	}

	state.CurrentSeat = (seat + 1) % 3
	return state, nil
}

// removeCardsFromHand removes the specified cards from a hand, matching by ID.
func removeCardsFromHand(hand, cards []Card) []Card {
	idSet := make(map[int]int)
	for _, c := range cards {
		idSet[c.ID]++
	}
	var result []Card
	for _, c := range hand {
		if idSet[c.ID] > 0 {
			idSet[c.ID]--
		} else {
			result = append(result, c)
		}
	}
	return result
}

// cardsInHand checks whether all specified cards exist in the hand, matching by ID.
func cardsInHand(hand, cards []Card) bool {
	idSet := make(map[int]int)
	for _, c := range hand {
		idSet[c.ID]++
	}
	for _, c := range cards {
		if idSet[c.ID] <= 0 {
			return false
		}
		idSet[c.ID]--
	}
	return true
}
