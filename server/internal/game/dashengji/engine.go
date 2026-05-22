package dashengji

import (
	"fmt"
	"math/rand"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

// Engine implements the Dashengji (打升级) game logic.
type Engine struct{}

// Init creates a new game state for 4 players with a 162-card deck.
func (e *Engine) Init(players []game.PlayerInfo) (game.GameState, error) {
	if len(players) != 4 {
		return nil, fmt.Errorf("dashengji requires exactly 4 players, got %d", len(players))
	}

	deck := NewDeck()
	Shuffle(deck)
	h0, h1, h2, h3, bottom := Deal(deck)

	SortCards(h0, -1, -1)
	SortCards(h1, -1, -1)
	SortCards(h2, -1, -1)
	SortCards(h3, -1, -1)

	playerHands := make([]PlayerHand, 4)
	for i, info := range players {
		var hand []Card
		switch i {
		case 0:
			hand = h0
		case 1:
			hand = h1
		case 2:
			hand = h2
		case 3:
			hand = h3
		}
		playerHands[i] = PlayerHand{UserID: info.ID, Seat: info.Seat, Hand: hand}
	}

	var dealerSeats [2]int
	if rand.Intn(2) == 0 {
		dealerSeats = [2]int{0, 2}
	} else {
		dealerSeats = [2]int{1, 3}
	}

	state := &GameState{
		Phase:            PhaseSetTrump,
		Players:          playerHands,
		CurrentSeat:      dealerSeats[0],
		DealerSeats:      dealerSeats,
		CurrentLevel:     3,
		LevelRank:        3,
		TrumpSuit:        -1,
		BottomCards:      bottom,
		RoundNum:         1,
		HasPassedTrump:   make(map[int]bool),
		HasPassedCounter: make(map[int]bool),
		DealerHistory:    make([]int, 0),
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
	if seat == -1 {
		return nil, fmt.Errorf("player %d not found", action.PlayerID)
	}
	if seat != gs.CurrentSeat {
		return nil, &game.GameError{Code: game.ErrNotYourTurn}
	}

	allowed := phaseActions[gs.Phase]
	valid := false
	for _, a := range allowed {
		if a == action.Action {
			valid = true
			break
		}
	}
	if !valid {
		return nil, &game.GameError{Code: game.ErrPhaseMismatch, Phase: gs.Phase.String(), Action: action.Action}
	}

	switch gs.Phase {
	case PhaseSetTrump:
		return e.handleSetTrump(gs, seat, action)
	case PhaseCounterTrump:
		return e.handleCounterTrump(gs, seat, action)
	case PhaseTakeBottom:
		return e.handleTakeBottom(gs, seat, action)
	case PhaseDiscardBottom:
		return e.handleDiscardBottom(gs, seat, action)
	case PhasePlaying:
		return e.handlePlay(gs, seat, action)
	default:
		return nil, fmt.Errorf("unknown phase: %v", gs.Phase)
	}
}

func (e *Engine) handleSetTrump(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action == "pass_trump" {
		gs.HasPassedTrump[seat] = true
		nextSeat := e.nextDealerSeat(gs, seat)
		if nextSeat == -1 {
			return e.swapDealerAndReDeal(gs)
		}
		gs.CurrentSeat = nextSeat
		return gs, nil
	}

	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	if !CheckSetTrump(cards, gs.LevelRank, true) {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	gs.TrumpSuit = GetTrumpSuit(cards, gs.LevelRank)
	gs.TrumpCards = cards
	gs.TrumpRevealed = true

	for i := range gs.Players {
		SortCards(gs.Players[i].Hand, gs.TrumpSuit, gs.LevelRank)
	}

	if IsDeadTrump(cards, gs.LevelRank) {
		gs.IsDeadTrump = true
		gs.Phase = PhaseTakeBottom
		gs.CurrentSeat = gs.nextTakeBottomSeat()
	} else {
		gs.Phase = PhaseCounterTrump
		gs.CurrentSeat = gs.DealerSeats[0] ^ 1
	}
	return gs, nil
}

func (e *Engine) handleCounterTrump(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action == "pass_counter" {
		gs.HasPassedCounter[seat] = true
		nextSeat := e.nextNonDealerSeat(gs, seat)
		if nextSeat == -1 {
			gs.Phase = PhaseTakeBottom
			gs.CurrentSeat = gs.nextTakeBottomSeat()
			return gs, nil
		}
		gs.CurrentSeat = nextSeat
		return gs, nil
	}

	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	if !CheckCounterTrump(cards, gs.LevelRank) {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	gs.TrumpSuit = GetTrumpSuit(cards, gs.LevelRank)
	gs.TrumpCards = cards

	for i := range gs.Players {
		SortCards(gs.Players[i].Hand, gs.TrumpSuit, gs.LevelRank)
	}

	gs.Phase = PhaseTakeBottom
	gs.CurrentSeat = gs.nextTakeBottomSeat()
	return gs, nil
}

func (e *Engine) handleTakeBottom(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	playerIdx := -1
	for i, p := range gs.Players {
		if p.Seat == seat {
			playerIdx = i
			break
		}
	}
	if playerIdx == -1 {
		return nil, fmt.Errorf("player at seat %d not found", seat)
	}

	gs.Players[playerIdx].Hand = append(gs.Players[playerIdx].Hand, gs.BottomCards...)
	SortCards(gs.Players[playerIdx].Hand, gs.TrumpSuit, gs.LevelRank)
	gs.BottomTaken = true
	gs.TakeBottomSeat = seat
	gs.DealerHistory = append(gs.DealerHistory, seat)
	gs.BottomRevealed = true

	gs.Phase = PhaseDiscardBottom
	return gs, nil
}

func (e *Engine) handleDiscardBottom(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if len(action.Cards) != 6 {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	playerIdx := -1
	for i, p := range gs.Players {
		if p.Seat == seat {
			playerIdx = i
			break
		}
	}
	if playerIdx == -1 {
		return nil, fmt.Errorf("player at seat %d not found", seat)
	}

	discardSet := make(map[int]bool)
	for _, id := range action.Cards {
		discardSet[id] = true
	}

	newHand := make([]Card, 0, len(gs.Players[playerIdx].Hand)-6)
	discarded := make([]Card, 0, 6)
	for _, c := range gs.Players[playerIdx].Hand {
		if discardSet[c.ID] {
			discarded = append(discarded, c)
		} else {
			newHand = append(newHand, c)
		}
	}
	gs.Players[playerIdx].Hand = newHand
	gs.DiscardedCards = discarded
	gs.BottomRevealed = true

	gs.Phase = PhasePlaying
	gs.CurrentSeat = seat

	return gs, nil
}

func (e *Engine) handlePlay(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	playerIdx := -1
	for i, p := range gs.Players {
		if p.Seat == seat {
			playerIdx = i
			break
		}
	}
	if playerIdx == -1 {
		return nil, fmt.Errorf("player at seat %d not found", seat)
	}

	if action.Action == "pass" {
		if gs.LastPlay == nil || gs.LastPlay.Seat == seat {
			return nil, &game.GameError{Code: game.ErrCannotPass}
		}
		e.advanceSeat(gs)
		return gs, nil
	}

	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	play := ParsePlayWithContext(cards, gs.TrumpSuit, gs.LevelRank)
	if play.Type == PlayInvalid {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	if gs.LastPlay != nil && gs.LastPlay.Seat != seat {
		if err := validateFollow(cards, gs.Players[playerIdx].Hand, gs.LastPlay.Play,
			gs.LastPlay.Cards, gs.TrumpSuit, gs.LevelRank); err != nil {
			return nil, err
		}

		if !CanBeat(play, gs.LastPlay.Play) {
			return nil, &game.GameError{Code: game.ErrCannotBeat}
		}
	}

	cardSet := make(map[int]bool)
	for _, id := range action.Cards {
		cardSet[id] = true
	}
	newHand := make([]Card, 0, len(gs.Players[playerIdx].Hand)-len(cards))
	for _, c := range gs.Players[playerIdx].Hand {
		if !cardSet[c.ID] {
			newHand = append(newHand, c)
		}
	}
	gs.Players[playerIdx].Hand = newHand

	record := PlayRecord{Seat: seat, Play: play, Cards: cards}
	gs.LastPlay = &record
	gs.PlayHistory = append(gs.PlayHistory, record)

	// Count points from this trick
	gs.RoundPoints += CountRoundPoints(cards)

	if len(newHand) == 0 {
		gs.Phase = PhaseEnded
		gs.WinnerSeat = &seat
		return gs, nil
	}

	e.advanceSeat(gs)
	return gs, nil
}

func (e *Engine) advanceSeat(gs *GameState) {
	gs.CurrentSeat = (gs.CurrentSeat + 1) % 4
}

func (e *Engine) nextDealerSeat(gs *GameState, current int) int {
	for _, s := range gs.DealerSeats {
		if s != current && !gs.HasPassedTrump[s] {
			return s
		}
	}
	return -1
}

func (e *Engine) nextNonDealerSeat(gs *GameState, current int) int {
	nonDealer0 := gs.DealerSeats[0] ^ 1
	nonDealer1 := nonDealer0 ^ 2
	for _, s := range []int{nonDealer0, nonDealer1} {
		if s != current && !gs.HasPassedCounter[s] {
			return s
		}
	}
	return -1
}

func (gs *GameState) nextTakeBottomSeat() int {
	if len(gs.DealerHistory) == 0 {
		return gs.DealerSeats[0]
	}
	lastTaker := gs.DealerHistory[len(gs.DealerHistory)-1]
	if lastTaker == gs.DealerSeats[0] {
		return gs.DealerSeats[1]
	}
	return gs.DealerSeats[0]
}

func (e *Engine) swapDealerAndReDeal(gs *GameState) (*GameState, error) {
	if gs.DealerSeats[0] == 0 {
		gs.DealerSeats = [2]int{1, 3}
	} else {
		gs.DealerSeats = [2]int{0, 2}
	}

	deck := NewDeck()
	Shuffle(deck)
	h0, h1, h2, h3, bottom := Deal(deck)
	SortCards(h0, -1, -1)
	SortCards(h1, -1, -1)
	SortCards(h2, -1, -1)
	SortCards(h3, -1, -1)

	for i := range gs.Players {
		var hand []Card
		switch i {
		case 0:
			hand = h0
		case 1:
			hand = h1
		case 2:
			hand = h2
		case 3:
			hand = h3
		}
		gs.Players[i].Hand = hand
	}

	gs.Phase = PhaseSetTrump
	gs.CurrentSeat = gs.DealerSeats[0]
	gs.TrumpSuit = -1
	gs.IsDeadTrump = false
	gs.TrumpCards = nil
	gs.TrumpRevealed = false
	gs.BottomCards = bottom
	gs.BottomTaken = false
	gs.HasPassedTrump = make(map[int]bool)
	gs.HasPassedCounter = make(map[int]bool)
	gs.RoundNum++

	return gs, nil
}

// ValidateAction checks if an action is valid without modifying state.
func (e *Engine) ValidateAction(state game.GameState, action game.PlayerAction) error {
	gs, ok := state.(*GameState)
	if !ok {
		return fmt.Errorf("invalid state type")
	}
	seat := -1
	for i, p := range gs.Players {
		if p.UserID == action.PlayerID {
			seat = i
			break
		}
	}
	if seat != gs.CurrentSeat {
		return &game.GameError{Code: game.ErrNotYourTurn}
	}
	allowed := phaseActions[gs.Phase]
	for _, a := range allowed {
		if a == action.Action {
			return nil
		}
	}
	return &game.GameError{Code: game.ErrPhaseMismatch, Phase: gs.Phase.String(), Action: action.Action}
}

// IsRoundEnd returns true if the phase is PhaseEnded.
func (e *Engine) IsRoundEnd(state game.GameState) bool {
	gs, ok := state.(*GameState)
	if !ok {
		return false
	}
	return gs.Phase == PhaseEnded
}

// CalculateScore computes round scores for all players.
func (e *Engine) CalculateScore(state game.GameState) ([]game.PlayerScore, error) {
	gs, ok := state.(*GameState)
	if !ok {
		return nil, fmt.Errorf("invalid state type")
	}
	if gs.WinnerSeat == nil {
		return nil, fmt.Errorf("round not ended")
	}

	dealerWon := IsDealerTeam(*gs.WinnerSeat)
	change := CalculateLevelChange(dealerWon, gs.RoundPoints)

	scores := make([]game.PlayerScore, 4)
	for i, p := range gs.Players {
		scores[i] = game.PlayerScore{PlayerID: p.UserID}
		if IsDealerTeam(i) {
			scores[i].Score = change
		} else {
			scores[i].Score = -change
		}
	}
	return scores, nil
}

// hasSuit checks if the player has any cards matching the led suit in hand.
func hasSuit(hand []Card, ledSuit int) bool {
	for _, c := range hand {
		if c.IsSmallJoker() || c.IsBigJoker() {
			continue
		}
		if c.Suit() == ledSuit {
			return true
		}
	}
	return false
}

// validateFollow checks that played cards follow the led suit+type+category rules.
func validateFollow(cards []Card, hand []Card, ledPlay Play, ledCards []Card,
	trumpSuit int, levelRank int) error {

	if len(cards) == 0 {
		return &game.GameError{Code: game.ErrInvalidCards}
	}

	play := ParsePlayWithContext(cards, trumpSuit, levelRank)
	if play.Type == PlayInvalid {
		return &game.GameError{Code: game.ErrInvalidCards}
	}

	if play.Type != ledPlay.Type {
		return &game.GameError{Code: game.ErrInvalidCards}
	}
	if ledPlay.Type == PlayTractor && play.Length != ledPlay.Length {
		return &game.GameError{Code: game.ErrInvalidCards}
	}

	ledSuit := ledCards[0].Suit()
	ledCat := ClassifyCard(ledCards[0], trumpSuit, levelRank)
	playCat := ClassifyCard(cards[0], trumpSuit, levelRank)
	playSuit := cards[0].Suit()

	// Same suit and same category = valid follow
	if playSuit == ledSuit && playCat == ledCat {
		return nil
	}

	// If no cards of led suit in hand, can trump (枪毙) with main cards
	if !hasSuit(hand, ledSuit) {
		if playCat == CatTrump || playCat == CatNativeMain || playCat == CatJoker {
			return nil
		}
	}

	return &game.GameError{Code: game.ErrInvalidCards}
}

// SerializeForAI returns JSON for AI consumption.
func (e *Engine) SerializeForAI(state game.GameState) string {
	gs, ok := state.(*GameState)
	if !ok {
		return "{}"
	}
	data, err := gs.ToJSON()
	if err != nil {
		return "{}"
	}
	return string(data)
}

// FilterForPlayer creates a copy hiding other players' hands.
func (e *Engine) FilterForPlayer(state game.GameState, seat int) game.GameState {
	gs, ok := state.(*GameState)
	if !ok {
		return nil
	}
	filtered := *gs
	filtered.Players = make([]PlayerHand, len(gs.Players))
	for i, p := range gs.Players {
		filtered.Players[i] = p
		if p.Seat != seat {
			filtered.Players[i].Hand = nil
		}
	}
	return &filtered
}
