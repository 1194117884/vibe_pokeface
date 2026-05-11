package doudizhu

import (
	"fmt"
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
		Phase:         PhaseBidding,
		Players:       playerHands,
		CurrentSeat:   0,
		LandlordCards: remaining,
		RoundNum:      1,
	}
	return state, nil
}

// ExecuteAction processes a player action and transitions the game state.
// Returns an error if the action is invalid or it's not the player's turn.
func (e *Engine) ExecuteAction(state game.GameState, action game.PlayerAction) (game.GameState, error) {
	gs, ok := state.(*GameState)
	if !ok {
		return nil, fmt.Errorf("invalid state type")
	}
	// Find player by UserID
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

	if gs.Phase == PhaseBidding {
		return e.handleBid(gs, seat, action)
	}
	if gs.Phase == PhasePlaying {
		return e.handlePlay(gs, seat, action)
	}
	return nil, fmt.Errorf("game already ended")
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
	if gs.Phase == PhaseBidding {
		return action.Action == "bid_call" || action.Action == "bid_pass"
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
		return cardsInHand(gs.Players[seat].Hand, cards)
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

// CalculateScore computes the scores for each player based on the game result.
// Landlord earns 2 points for winning, farmers earn 1 each.
// Losers lose the same amount.
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

	for i, p := range gs.Players {
		var score int
		if p.IsLandlord {
			if winnerIsLandlord {
				score = 2
			} else {
				score = -2
			}
		} else {
			if winnerIsLandlord {
				score = -1
			} else {
				score = 1
			}
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
	for _, p := range gs.Players {
		role := "farmer"
		if p.IsLandlord {
			role = "landlord"
		}
		sb.WriteString(fmt.Sprintf("Player %d (%s): %d cards\n", p.Seat, role, len(p.Hand)))
	}
	return sb.String()
}

// handleBid processes a bidding action (bid_call or bid_pass).
func (e *Engine) handleBid(state *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action != "bid_call" && action.Action != "bid_pass" {
		return nil, fmt.Errorf("invalid bid action: %s", action.Action)
	}

	state.BidHistory = append(state.BidHistory, BidRecord{
		Seat:   seat,
		Called: action.Action == "bid_call",
	})

	if action.Action == "bid_call" {
		state.LandlordSeat = seat
		state.Players[seat].IsLandlord = true
		state.Players[seat].Hand = append(state.Players[seat].Hand, state.LandlordCards...)
		SortCards(state.Players[seat].Hand)
		state.Phase = PhasePlaying
		state.CurrentSeat = seat
		return state, nil
	}

	// Check if all 3 passed
	passed := 0
	for _, b := range state.BidHistory {
		if !b.Called {
			passed++
		}
	}
	if passed >= 3 {
		state.Phase = PhaseEnded
		return state, nil
	}

	state.CurrentSeat = (seat + 1) % 3
	return state, nil
}

// handlePlay processes a card play action during the playing phase.
func (e *Engine) handlePlay(state *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	// Handle pass action
	if action.Action == "pass" {
		state.ConsecutivePasses++
		if state.ConsecutivePasses >= 2 {
			// Two passes after a play — clear last play, current player leads
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
