package doudizhu

import (
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

func TestEngine_Init_CreatesState(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, err := e.Init(players)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	gs := state.(*GameState)
	if gs == nil {
		t.Fatal("Init() returned nil state")
	}
	if gs.Phase != PhaseBidding {
		t.Errorf("phase = %d, want %d", gs.Phase, PhaseBidding)
	}
	if len(gs.Players) != 3 {
		t.Errorf("players = %d, want 3", len(gs.Players))
	}
	totalCards := 0
	for _, p := range gs.Players {
		totalCards += len(p.Hand)
	}
	if totalCards+len(gs.LandlordCards) != 54 {
		t.Errorf("total cards = %d, want 54", totalCards+len(gs.LandlordCards))
	}
}

func TestEngine_Init_RequiresThreePlayers(t *testing.T) {
	e := &Engine{}
	_, err := e.Init([]game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
	})
	if err == nil {
		t.Error("expected error for <3 players")
	}

	_, err = e.Init([]game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
		{ID: 4, Name: "Dave", Seat: 3},
	})
	if err == nil {
		t.Error("expected error for >3 players")
	}
}

func TestEngine_FullGameFlow(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	// Alice passes bid, Bob calls -> bidding ends immediately
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_call"})

	gs := state.(*GameState)

	if gs.Phase != PhasePlaying {
		t.Errorf("after bidding, phase = %d, want %d", gs.Phase, PhasePlaying)
	}
	if gs.LandlordSeat != 1 {
		t.Errorf("landlord seat = %d, want 1", gs.LandlordSeat)
	}
	if len(gs.Players[1].Hand) != 20 {
		t.Errorf("landlord hand = %d cards, want 20", len(gs.Players[1].Hand))
	}

	// Playing phase — make a play with some cards from landlord's hand
	landlordHand := gs.Players[1].Hand
	if len(landlordHand) > 0 {
		// Play a single card (smallest)
		firstCard := []int{landlordHand[0].ID}
		var err error
		state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "play", Cards: firstCard})
		if err != nil {
			t.Fatalf("landlord play error: %v", err)
		}
		gs = state.(*GameState)
		if len(gs.Players[1].Hand) != 19 {
			t.Errorf("landlord hand after play = %d cards, want 19", len(gs.Players[1].Hand))
		}
	}
}

func TestEngine_FullGameFlow_AllPass(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 3, Action: "bid_pass"})

	gs := state.(*GameState)
	if gs.Phase != PhaseEnded {
		t.Errorf("after all pass, phase = %d, want %d", gs.Phase, PhaseEnded)
	}
}

func TestEngine_ValidateAction_WrongPlayer(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	// Bob (seat 1) tries to bid, but it's Alice's turn (seat 0)
	valid := e.ValidateAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_call"})
	if valid {
		t.Error("expected invalid: not Bob's turn")
	}
}

func TestEngine_ValidateAction_ValidBid(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	if !e.ValidateAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_call"}) {
		t.Error("expected valid: Alice can bid call")
	}
	if !e.ValidateAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_pass"}) {
		t.Error("expected valid: Alice can bid pass")
	}
}

func TestEngine_CalculateScore_LandlordWins(t *testing.T) {
	e := &Engine{}
	winnerSeat := 0
	state := &GameState{
		Phase:        PhaseEnded,
		LandlordSeat: 0,
		Players: []PlayerHand{
			{UserID: 1, Hand: []Card{}, IsLandlord: true},
			{UserID: 2, Hand: []Card{}},
			{UserID: 3, Hand: []Card{}},
		},
		WinnerSeat: &winnerSeat,
	}
	scores, err := e.CalculateScore(state)
	if err != nil {
		t.Fatalf("CalculateScore() error = %v", err)
	}
	if len(scores) != 3 {
		t.Errorf("scores = %d, want 3", len(scores))
	}
	for _, s := range scores {
		if s.PlayerID == 1 && s.Score != 2 {
			t.Errorf("landlord score = %d, want 2", s.Score)
		}
		if s.PlayerID != 1 && s.Score != -1 {
			t.Errorf("farmer %d score = %d, want -1", s.PlayerID, s.Score)
		}
	}
}

func TestEngine_CalculateScore_FarmersWin(t *testing.T) {
	e := &Engine{}
	winnerSeat := 1 // a farmer wins
	state := &GameState{
		Phase:        PhaseEnded,
		LandlordSeat: 0,
		Players: []PlayerHand{
			{UserID: 1, Hand: []Card{}, IsLandlord: true},
			{UserID: 2, Hand: []Card{}},
			{UserID: 3, Hand: []Card{}},
		},
		WinnerSeat: &winnerSeat,
	}
	scores, _ := e.CalculateScore(state)
	for _, s := range scores {
		if s.PlayerID == 1 && s.Score != -2 {
			t.Errorf("landlord score = %d, want -2", s.Score)
		}
		if s.PlayerID != 1 && s.Score != 1 {
			t.Errorf("farmer %d score = %d, want 1", s.PlayerID, s.Score)
		}
	}
}

func TestEngine_CalculateScore_NoWinner(t *testing.T) {
	e := &Engine{}
	state := &GameState{
		Phase:        PhasePlaying,
		LandlordSeat: 0,
		Players: []PlayerHand{
			{UserID: 1, Hand: []Card{}},
			{UserID: 2, Hand: []Card{}},
			{UserID: 3, Hand: []Card{}},
		},
	}
	_, err := e.CalculateScore(state)
	if err == nil {
		t.Error("expected error when no winner")
	}
}

func TestEngine_IsRoundEnd(t *testing.T) {
	e := &Engine{}
	state := &GameState{Phase: PhaseBidding}
	if e.IsRoundEnd(state) {
		t.Error("bidding phase should not be round end")
	}

	state.Phase = PhasePlaying
	if e.IsRoundEnd(state) {
		t.Error("playing phase should not be round end")
	}

	state.Phase = PhaseEnded
	if !e.IsRoundEnd(state) {
		t.Error("ended phase should be round end")
	}
}

func TestEngine_SerializeForAI(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)
	s := e.SerializeForAI(state)
	if s == "" {
		t.Error("SerializeForAI() returned empty string")
	}
}

func TestEngine_ExecuteAction_WrongPlayer(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	_, err := e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_call"})
	if err == nil {
		t.Error("expected error: not Bob's turn")
	}
}

func TestEngine_ExecuteAction_InvalidBid(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	_, err := e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_invalid"})
	if err == nil {
		t.Error("expected error: invalid bid action")
	}
}

func TestEngine_GameEnded(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	gs := state.(*GameState)
	gs.Phase = PhaseEnded
	_, err := e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_pass"})
	if err == nil {
		t.Error("expected error: game already ended")
	}
}
