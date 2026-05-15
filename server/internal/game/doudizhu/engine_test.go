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
	gs := state.(*GameState)

	// Phase 1 — 叫地主: first player passes, next player calls immediately
	firstSeat := gs.CurrentSeat
	secondSeat := (firstSeat + 1) % 3
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[firstSeat].ID, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[secondSeat].ID, Action: "bid_call"})

	gs = state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after call, phase = %d, want %d (PhaseSnatching)", gs.Phase, PhaseSnatching)
	}
	if gs.LandlordSeat != secondSeat {
		t.Errorf("landlord seat = %d, want %d", gs.LandlordSeat, secondSeat)
	}

	// Phase 2 — 抢地主: follow dynamic turn order, pass for all remaining
	for gs.SnatchCount < 3 {
		seat := gs.CurrentSeat
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "bid_pass"})
		gs = state.(*GameState)
	}

	if gs.Phase != PhasePlaying {
		t.Errorf("after snatching, phase = %d, want %d (PhasePlaying)", gs.Phase, PhasePlaying)
	}
	if !gs.Players[secondSeat].IsLandlord {
		t.Error("expected second player to be landlord")
	}
	if len(gs.Players[secondSeat].Hand) != 20 {
		t.Errorf("landlord hand = %d cards, want 20", len(gs.Players[secondSeat].Hand))
	}

	// Playing phase — landlord leads
	landlordHand := gs.Players[secondSeat].Hand
	if len(landlordHand) > 0 {
		firstCard := []int{landlordHand[0].ID}
		var err error
		state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[secondSeat].ID, Action: "play", Cards: firstCard})
		if err != nil {
			t.Fatalf("landlord play error: %v", err)
		}
		gs = state.(*GameState)
		if len(gs.Players[secondSeat].Hand) != 19 {
			t.Errorf("landlord hand after play = %d cards, want 19", len(gs.Players[secondSeat].Hand))
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

	gs := state.(*GameState)
	oldHand0 := make([]Card, len(gs.Players[0].Hand))
	copy(oldHand0, gs.Players[0].Hand)

	// Send pass for each seat in current turn order
	for i := 0; i < 3; i++ {
		seat := gs.CurrentSeat
		pid := players[seat].ID
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: pid, Action: "bid_pass"})
		gs = state.(*GameState)
	}

	// After all pass, should re-deal and restart PhaseCalling
	if gs.Phase != PhaseCalling {
		t.Errorf("after all pass re-deal, phase = %d, want %d (PhaseCalling)", gs.Phase, PhaseCalling)
	}
	if gs.RoundNum != 2 {
		t.Errorf("after re-deal, round = %d, want 2", gs.RoundNum)
	}
	if gs.LandlordSeat != -1 {
		t.Errorf("after re-deal, landlord seat = %d, want -1", gs.LandlordSeat)
	}
	if gs.Multiplier != 1 {
		t.Errorf("after re-deal, multiplier = %d, want 1", gs.Multiplier)
	}
	// Verify hands were re-dealt (different cards)
	sameCount := 0
	for i := 0; i < len(oldHand0); i++ {
		for j := 0; j < len(gs.Players[0].Hand); j++ {
			if oldHand0[i].ID == gs.Players[0].Hand[j].ID {
				sameCount++
				break
			}
		}
	}
	if sameCount == 17 {
		t.Error("hands should be different after re-deal")
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
	gs := state.(*GameState)

	// Send from a seat that is NOT the current seat
	wrongSeat := (gs.CurrentSeat + 1) % 3
	valid := e.ValidateAction(state, game.PlayerAction{PlayerID: players[wrongSeat].ID, Action: "bid_call"})
	if valid {
		t.Error("expected invalid: not their turn")
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
	gs := state.(*GameState)

	cur := gs.CurrentSeat
	if !e.ValidateAction(state, game.PlayerAction{PlayerID: players[cur].ID, Action: "bid_call"}) {
		t.Error("expected valid: current player can bid call")
	}
	if !e.ValidateAction(state, game.PlayerAction{PlayerID: players[cur].ID, Action: "bid_pass"}) {
		t.Error("expected valid: current player can bid pass")
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
	gs := state.(*GameState)

	wrongSeat := (gs.CurrentSeat + 1) % 3
	_, err := e.ExecuteAction(state, game.PlayerAction{PlayerID: players[wrongSeat].ID, Action: "bid_call"})
	if err == nil {
		t.Error("expected error: not their turn")
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
	gs := state.(*GameState)

	_, err := e.ExecuteAction(state, game.PlayerAction{PlayerID: players[gs.CurrentSeat].ID, Action: "bid_invalid"})
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
	gs := state.(*GameState)

	// First player (random) calls 叫地主
	first := gs.CurrentSeat
	second := (first + 1) % 3
	third := (first + 2) % 3
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_call"})
	gs = state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after call, phase = %d, want PhaseSnatching", gs.Phase)
	}

	// Second player snatches 抢地主 → multiplier ×2
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[second].ID, Action: "bid_call"})
	gs = state.(*GameState)
	if gs.LandlordSeat != second {
		t.Errorf("after snatch, landlord seat = %d, want %d", gs.LandlordSeat, second)
	}
	if gs.Multiplier != 2 {
		t.Errorf("after 1 snatch, multiplier = %d, want 2", gs.Multiplier)
	}

	// Third player 不抢
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[third].ID, Action: "bid_pass"})

	// First player 不抢 (final snatch turn)
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_pass"})

	gs = state.(*GameState)
	if gs.Phase != PhasePlaying {
		t.Errorf("after snatching, phase = %d, want PhasePlaying", gs.Phase)
	}
	if gs.LandlordSeat != second {
		t.Errorf("final landlord seat = %d, want %d", gs.LandlordSeat, second)
	}
	if gs.Multiplier != 2 {
		t.Errorf("final multiplier = %d, want 2", gs.Multiplier)
	}
	if !gs.Players[second].IsLandlord {
		t.Error("expected second player to be landlord")
	}
	if len(gs.Players[second].Hand) != 20 {
		t.Errorf("landlord hand = %d cards, want 20", len(gs.Players[second].Hand))
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
	gs := state.(*GameState)

	first := gs.CurrentSeat
	second := (first + 1) % 3
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[second].ID, Action: "bid_call"})

	gs = state.(*GameState)
	if !gs.HasPassed[first] {
		t.Error("first player should be marked as HasPassed")
	}

	for gs.SnatchCount < 3 {
		seat := gs.CurrentSeat
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "bid_pass"})
		gs = state.(*GameState)
	}

	if gs.Phase != PhasePlaying {
		t.Errorf("phase = %d, want PhasePlaying", gs.Phase)
	}
	if gs.LandlordSeat != second {
		t.Errorf("final landlord = %d, want %d", gs.LandlordSeat, second)
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
	gs := state.(*GameState)

	first := gs.CurrentSeat
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_call"})

	gs = state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after immediate call, phase = %d, want PhaseSnatching", gs.Phase)
	}
	if gs.LandlordSeat != first {
		t.Errorf("landlord seat = %d, want %d", gs.LandlordSeat, first)
	}
	for i := 0; i < 3; i++ {
		if i != first && gs.HasPassed[i] {
			t.Errorf("seat %d should not be in HasPassed (never bid)", i)
		}
	}
}
