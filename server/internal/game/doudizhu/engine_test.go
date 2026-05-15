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
	if gs.Phase != PhaseCalling {
		t.Errorf("phase = %d, want %d (PhaseCalling)", gs.Phase, PhaseCalling)
	}
	if len(gs.Players) != 3 {
		t.Errorf("players = %d, want 3", len(gs.Players))
	}
	if gs.Multiplier != 1 {
		t.Errorf("multiplier = %d, want 1", gs.Multiplier)
	}
	if gs.HasPassed == nil {
		t.Error("HasPassed map not initialized")
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

	// Phase 1 — 叫地主: Alice passes, Bob calls immediately
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_call"})

	gs := state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after call, phase = %d, want %d (PhaseSnatching)", gs.Phase, PhaseSnatching)
	}
	if gs.LandlordSeat != 1 {
		t.Errorf("after Bob call, landlord seat = %d, want 1", gs.LandlordSeat)
	}

	// Phase 2 — 抢地主: Charlie 不抢, Alice skipped (HasPassed), Bob 不抢
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 3, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_pass"})

	gs = state.(*GameState)
	if gs.Phase != PhasePlaying {
		t.Errorf("after snatching, phase = %d, want %d (PhasePlaying)", gs.Phase, PhasePlaying)
	}
	if !gs.Players[1].IsLandlord {
		t.Error("expected player 1 to be landlord")
	}
	if len(gs.Players[1].Hand) != 20 {
		t.Errorf("landlord hand = %d cards, want 20", len(gs.Players[1].Hand))
	}

	// Playing phase — landlord leads
	landlordHand := gs.Players[1].Hand
	if len(landlordHand) > 0 {
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
		t.Errorf("after all pass, phase = %d, want %d (PhaseEnded)", gs.Phase, PhaseEnded)
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
		Multiplier: 1,
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
		Multiplier: 1,
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

func TestEngine_CalculateScore_WithMultiplier(t *testing.T) {
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
		Multiplier: 4, // 2 snatches
	}
	scores, _ := e.CalculateScore(state)
	for _, s := range scores {
		if s.PlayerID == 1 && s.Score != 8 {
			t.Errorf("landlord score with ×4 = %d, want 8", s.Score)
		}
		if s.PlayerID != 1 && s.Score != -4 {
			t.Errorf("farmer %d score with ×4 = %d, want -4", s.PlayerID, s.Score)
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
	state := &GameState{Phase: PhaseCalling}
	if e.IsRoundEnd(state) {
		t.Error("calling phase should not be round end")
	}

	state.Phase = PhaseSnatching
	if e.IsRoundEnd(state) {
		t.Error("snatching phase should not be round end")
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

func TestEngine_Bidding_SnatchLandlord(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	// Alice calls 叫地主 → nominee = 0
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_call"})
	gs := state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after call, phase = %d, want PhaseSnatching", gs.Phase)
	}
	if gs.LandlordSeat != 0 {
		t.Errorf("after Alice call, landlord seat = %d, want 0", gs.LandlordSeat)
	}

	// Bob snatches 抢地主 → nominee = 1, multiplier ×2
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_call"})
	gs = state.(*GameState)
	if gs.LandlordSeat != 1 {
		t.Errorf("after Bob snatch, landlord seat = %d, want 1", gs.LandlordSeat)
	}
	if gs.Multiplier != 2 {
		t.Errorf("after 1 snatch, multiplier = %d, want 2", gs.Multiplier)
	}

	// Charlie 不抢
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 3, Action: "bid_pass"})

	// Alice 不抢 (final snatch turn)
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_pass"})

	gs = state.(*GameState)
	if gs.Phase != PhasePlaying {
		t.Errorf("after snatching, phase = %d, want PhasePlaying", gs.Phase)
	}
	if gs.LandlordSeat != 1 {
		t.Errorf("final landlord seat = %d, want 1", gs.LandlordSeat)
	}
	if gs.Multiplier != 2 {
		t.Errorf("final multiplier = %d, want 2", gs.Multiplier)
	}
	if !gs.Players[1].IsLandlord {
		t.Error("expected Bob to be landlord")
	}
	if len(gs.Players[1].Hand) != 20 {
		t.Errorf("Bob hand = %d cards, want 20", len(gs.Players[1].Hand))
	}
}

func TestEngine_Bidding_PassRestriction(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	// Alice 不叫 → HasPassed[0] = true
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_pass"})
	// Bob 叫地主 → nominee = 1, enter snatching
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_call"})

	gs := state.(*GameState)
	if gs.LandlordSeat != 1 {
		t.Errorf("landlord seat = %d, want 1", gs.LandlordSeat)
	}
	if !gs.HasPassed[0] {
		t.Error("Alice should be marked as HasPassed")
	}

	// Charlie 不抢
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 3, Action: "bid_pass"})
	// Alice auto-skipped → seat moves to Bob
	// Bob 不抢 → finalize
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "bid_pass"})

	gs = state.(*GameState)
	if gs.Phase != PhasePlaying {
		t.Errorf("phase = %d, want PhasePlaying", gs.Phase)
	}
	if gs.LandlordSeat != 1 {
		t.Errorf("final landlord = %d, want 1", gs.LandlordSeat)
	}
	if gs.Multiplier != 1 {
		t.Errorf("multiplier = %d, want 1 (no snatches)", gs.Multiplier)
	}
}

func TestEngine_Bidding_FirstCallWinsImmediately(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)

	// Alice calls immediately → bidding ends, enter snatching
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "bid_call"})

	gs := state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after immediate call, phase = %d, want PhaseSnatching", gs.Phase)
	}
	if gs.LandlordSeat != 0 {
		t.Errorf("landlord seat = %d, want 0", gs.LandlordSeat)
	}
	// Bob and Charlie never got a turn in 叫地主 — they should NOT be in HasPassed
	if gs.HasPassed[1] || gs.HasPassed[2] {
		t.Error("players who didn't bid should not be in HasPassed")
	}
}
