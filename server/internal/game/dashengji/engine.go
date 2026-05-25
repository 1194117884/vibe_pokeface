package dashengji

import (
	"fmt"
	"log"
	"math/rand"
	"sort"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

func init() {
	game.RegisterEngine("dashengji", func() game.GameEngine { return &Engine{} })
}

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
			TakeBottomSeat:   -1,
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
	log.Printf("[DEBUG] ExecuteAction: seat=%d currentSeat=%d phase=%s action=%s", seat, gs.CurrentSeat, gs.Phase.String(), action.Action)
	if !isValidSeatForPhase(gs, seat) {
		log.Printf("[DEBUG] ExecuteAction: seat mismatch! seat=%d != currentSeat=%d", seat, gs.CurrentSeat)
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
	// Delegate bottom-taking to teammate
	if action.Action == "pass_take_bottom" {
		other := -1
		for _, s := range gs.DealerSeats {
			if s != seat {
				other = s
				break
			}
		}
		if other == -1 {
			return nil, fmt.Errorf("no teammate to delegate to")
		}
		gs.CurrentSeat = other
		return gs, nil
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
	gs.RoundPlays = nil
	gs.RoundLeader = seat

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

	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	isLeading := len(gs.RoundPlays) == 0
	log.Printf("[HANDLEPLAY] seat=%d isLeading=%t roundPlays=%d action=%s cardCount=%d",
		seat, isLeading, len(gs.RoundPlays), action.Action, len(action.Cards))

	if isLeading {
		play := ParsePlayWithContext(cards, gs.TrumpSuit, gs.LevelRank)
		if play.Type == PlayInvalid {
			return nil, &game.GameError{Code: game.ErrInvalidCards}
		}
		gs.RoundLeader = seat
	} else {
		_, err := validateFollow(cards, gs.Players[playerIdx].Hand, gs.RoundPlays[0].Play,
			gs.RoundPlays[0].Cards, gs.TrumpSuit, gs.LevelRank)
		if err != nil {
			return nil, err
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

	play := ParsePlayWithContext(cards, gs.TrumpSuit, gs.LevelRank)
	record := PlayRecord{Seat: seat, Play: play, Cards: cards}
	gs.LastPlay = &record
	gs.PlayHistory = append(gs.PlayHistory, record)
	gs.RoundPlays = append(gs.RoundPlays, record)

	gs.RoundPoints += CountRoundPoints(cards)

	if len(newHand) == 0 {
		gs.Phase = PhaseEnded
		gs.WinnerSeat = &seat
		return gs, nil
	}

	if len(gs.RoundPlays) == 4 {
		e.resolveRound(gs)
	} else {
		e.advanceSeat(gs)
	}
	return gs, nil
}

// resolveRound determines the winner of the current round and sets up the next round.
func (e *Engine) resolveRound(gs *GameState) {
	winner := gs.RoundPlays[0].Seat
	bestPlay := gs.RoundPlays[0].Play
	bestIsTrump := false
	bestIsPadding := true

	ledPlay := gs.RoundPlays[0]
	ledSuit := ledPlay.Cards[0].Suit()
	ledCat := ClassifyCard(ledPlay.Cards[0], gs.TrumpSuit, gs.LevelRank)

	for _, r := range gs.RoundPlays {
		playCat := ClassifyCard(r.Cards[0], gs.TrumpSuit, gs.LevelRank)
		playSuit := r.Cards[0].Suit()

		isFollowing := playSuit == ledSuit && playCat == ledCat
		isTrump := !isFollowing && (playCat == CatTrump || playCat == CatNativeMain || playCat == CatJoker)
		isPadding := !isFollowing && !isTrump

		if isPadding {
			continue
		}

		if bestIsPadding {
			// First non-padding play wins by default
			winner = r.Seat
			bestPlay = r.Play
			bestIsTrump = isTrump
			bestIsPadding = false
			continue
		}

		// Trump beats following
		if isTrump && !bestIsTrump {
			winner = r.Seat
			bestPlay = r.Play
			bestIsTrump = true
			continue
		}

		// Following loses to existing trump
		if !isTrump && bestIsTrump {
			continue
		}

		// Same category: compare main rank
		if r.Play.MainRank > bestPlay.MainRank {
			winner = r.Seat
			bestPlay = r.Play
			bestIsTrump = isTrump
		}
	}

	log.Printf("[RESOLVEROUND] winner=seat%d roundPlays=%d", winner, len(gs.RoundPlays))
	gs.RoundPlays = nil
	gs.RoundLeader = winner
	gs.CurrentSeat = winner
}

func (e *Engine) advanceSeat(gs *GameState) {
	gs.CurrentSeat = (gs.CurrentSeat + 1) % 4
}

// isValidSeatForPhase checks whether the given seat can act in the current phase.
func isValidSeatForPhase(gs *GameState, seat int) bool {
	return seat == gs.CurrentSeat
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
		gs.TakeBottomSeat = -1
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
	if !isValidSeatForPhase(gs, seat) {
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

// canFollow checks whether the player can match the led play type with same suit+category.
func canFollow(hand []Card, ledPlay Play, ledCards []Card, trumpSuit int, levelRank int) bool {
	ledSuit := ledCards[0].Suit()
	ledCat := ClassifyCard(ledCards[0], trumpSuit, levelRank)

	matching := make([]Card, 0)
	for _, c := range hand {
		if c.Suit() == ledSuit && ClassifyCard(c, trumpSuit, levelRank) == ledCat {
			matching = append(matching, c)
		}
	}

	var result bool
	switch ledPlay.Type {
	case PlaySingle:
		result = len(matching) >= 1
	case PlayPair:
		faceCount := make(map[int]int)
		for _, c := range matching {
			faceCount[c.Face()]++
		}
		for _, n := range faceCount {
			if n >= 2 {
				result = true
				break
			}
		}
	case PlayTriple:
		faceCount := make(map[int]int)
		for _, c := range matching {
			faceCount[c.Face()]++
		}
		for _, n := range faceCount {
			if n >= 3 {
				result = true
				break
			}
		}
	case PlayTractor:
		faceCount := make(map[int]int)
		for _, c := range matching {
			faceCount[c.Face()]++
		}
		paired := make([]int, 0)
		for face, n := range faceCount {
			if n >= 2 {
				paired = append(paired, face%13)
			}
		}
		sort.Ints(paired)
		if len(paired) >= ledPlay.Length {
			run := 1
			for i := 1; i < len(paired); i++ {
				if paired[i]-paired[i-1] == 1 {
					run++
					if run >= ledPlay.Length {
						result = true
						break
					}
				} else {
					run = 1
				}
			}
		}
	}

	log.Printf("[CANFOLLOW] ledSuit=%d ledCat=%d trumpSuit=%d levelRank=%d matching=%d ledType=%d result=%t",
		ledSuit, ledCat, trumpSuit, levelRank, len(matching), ledPlay.Type, result)
	return result
}

// validateFollow checks that played cards follow the led suit+type+category rules.
// Returns isPadding=true when the player cannot follow and is padding (垫牌),
// meaning the CanBeat check should be skipped.
func validateFollow(cards []Card, hand []Card, ledPlay Play, ledCards []Card,
	trumpSuit int, levelRank int) (bool, error) {

	if len(cards) != len(ledCards) {
		return false, &game.GameError{Code: game.ErrInvalidCards}
	}

	canFollowResult := canFollow(hand, ledPlay, ledCards, trumpSuit, levelRank)
	log.Printf("[VALIDATEFOLLOW] canFollow=%t ledType=%d cardCount=%d", canFollowResult, ledPlay.Type, len(cards))

	if canFollowResult {
		// Player CAN follow => MUST follow exactly (same type, suit, category)
		play := ParsePlayWithContext(cards, trumpSuit, levelRank)
		if play.Type == PlayInvalid || play.Type != ledPlay.Type {
			log.Printf("[VALIDATEFOLLOW] FOLLOW-REJECT: playType=%d playInvalid=%t typeMismatch=%t",
				play.Type, play.Type == PlayInvalid, play.Type != ledPlay.Type)
			return false, &game.GameError{Code: game.ErrInvalidCards}
		}
		if ledPlay.Type == PlayTractor && play.Length != ledPlay.Length {
			return false, &game.GameError{Code: game.ErrInvalidCards}
		}

		ledSuit := ledCards[0].Suit()
		ledCat := ClassifyCard(ledCards[0], trumpSuit, levelRank)
		if cards[0].Suit() != ledSuit || ClassifyCard(cards[0], trumpSuit, levelRank) != ledCat {
			log.Printf("[VALIDATEFOLLOW] FOLLOW-REJECT: suitMismatch cardSuit=%d ledSuit=%d cardCat=%d ledCat=%d",
				cards[0].Suit(), ledSuit, ClassifyCard(cards[0], trumpSuit, levelRank), ledCat)
			return false, &game.GameError{Code: game.ErrInvalidCards}
		}
		log.Printf("[VALIDATEFOLLOW] FOLLOW-ACCEPT: valid follow, will check CanBeat")
		return false, nil // valid follow, must also beat
	}

	// Player CANNOT follow => 垫牌 (pad) or 枪毙 (trump)
	// 枪毙: trump cards forming the same play type, must beat
	play := ParsePlayWithContext(cards, trumpSuit, levelRank)
	if play.Type != PlayInvalid && play.Type == ledPlay.Type {
		if ledPlay.Type == PlayTractor && play.Length != ledPlay.Length {
			log.Printf("[VALIDATEFOLLOW] PADDING: mismatched tractor length")
			return true, nil // mismatched tractor length => padding
		}
		cat := ClassifyCard(cards[0], trumpSuit, levelRank)
		if cat == CatTrump || cat == CatNativeMain || cat == CatJoker {
			log.Printf("[VALIDATEFOLLOW] TRUMPING: cat=%d playType=%d mainRank=%d", cat, play.Type, play.MainRank)
			return false, nil // trumping play, must beat
		}
	}

	log.Printf("[VALIDATEFOLLOW] PADDING-ACCEPT: playType=%d (ledType=%d) - accepted as padding",
		play.Type, ledPlay.Type)
	return true, nil // padding, no need to beat
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
