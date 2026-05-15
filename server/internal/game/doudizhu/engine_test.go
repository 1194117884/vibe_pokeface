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

// passRevealAndDouble runs through 明牌 (all pass) and 加倍 (all no_double).
func passRevealAndDouble(t *testing.T, e *Engine, players []game.PlayerInfo, state game.GameState) game.GameState {
	t.Helper()
	gs := state.(*GameState)

	for gs.RevealCount < 3 {
		seat := gs.CurrentSeat
		var err error
		state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "pass"})
		if err != nil {
			t.Fatalf("reveal pass error: %v", err)
		}
		gs = state.(*GameState)
	}

	for gs.DoubleCount < 3 {
		seat := gs.CurrentSeat
		var err error
		state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "no_double"})
		if err != nil {
			t.Fatalf("double error: %v", err)
		}
		gs = state.(*GameState)
	}

	return state
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

	firstSeat := gs.CurrentSeat
	secondSeat := (firstSeat + 1) % 3
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[firstSeat].ID, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[secondSeat].ID, Action: "bid_call"})

	gs = state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after call, phase = %d, want PhaseSnatching", gs.Phase)
	}

	for gs.SnatchCount < 3 {
		seat := gs.CurrentSeat
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "bid_pass"})
		gs = state.(*GameState)
	}

	state = passRevealAndDouble(t, e, players, state)
	gs = state.(*GameState)

	if gs.Phase != PhasePlaying {
		t.Errorf("after doubling, phase = %d, want PhasePlaying", gs.Phase)
	}
	if !gs.Players[secondSeat].IsLandlord {
		t.Error("expected second player to be landlord")
	}
	if len(gs.Players[secondSeat].Hand) != 20 {
		t.Errorf("landlord hand = %d cards, want 20", len(gs.Players[secondSeat].Hand))
	}

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

	for i := 0; i < 3; i++ {
		seat := gs.CurrentSeat
		pid := players[seat].ID
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: pid, Action: "bid_pass"})
		gs = state.(*GameState)
	}

	if gs.Phase != PhaseCalling {
		t.Errorf("after all pass re-deal, phase = %d, want PhaseCalling", gs.Phase)
	}
	if gs.RoundNum != 2 {
		t.Errorf("after re-deal, round = %d, want 2", gs.RoundNum)
	}
	if gs.LandlordSeat != -1 {
		t.Errorf("after re-deal, landlord seat = %d, want -1", gs.LandlordSeat)
	}
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
			{UserID: 1, Seat: 0, Hand: []Card{}, IsLandlord: true},
			{UserID: 2, Seat: 1, Hand: []Card{}},
			{UserID: 3, Seat: 2, Hand: []Card{}},
		},
		WinnerSeat: &winnerSeat,
		Multiplier: 1,
		Doubled:    map[int]bool{},
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
	winnerSeat := 1
	state := &GameState{
		Phase:        PhaseEnded,
		LandlordSeat: 0,
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{}, IsLandlord: true},
			{UserID: 2, Seat: 1, Hand: []Card{}},
			{UserID: 3, Seat: 2, Hand: []Card{}},
		},
		WinnerSeat: &winnerSeat,
		Multiplier: 1,
		Doubled:    map[int]bool{},
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
			{UserID: 1, Seat: 0, Hand: []Card{}, IsLandlord: true},
			{UserID: 2, Seat: 1, Hand: []Card{}},
			{UserID: 3, Seat: 2, Hand: []Card{}},
		},
		WinnerSeat: &winnerSeat,
		Multiplier: 4,
		Doubled:    map[int]bool{},
	}
	scores, _ := e.CalculateScore(state)
	for _, s := range scores {
		if s.PlayerID == 1 && s.Score != 8 {
			t.Errorf("landlord score with x4 = %d, want 8", s.Score)
		}
		if s.PlayerID != 1 && s.Score != -4 {
			t.Errorf("farmer %d score with x4 = %d, want -4", s.PlayerID, s.Score)
		}
	}
}

func TestEngine_CalculateScore_WithDouble(t *testing.T) {
	e := &Engine{}
	winnerSeat := 0
	state := &GameState{
		Phase:        PhaseEnded,
		LandlordSeat: 0,
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{}, IsLandlord: true},
			{UserID: 2, Seat: 1, Hand: []Card{}},
			{UserID: 3, Seat: 2, Hand: []Card{}},
		},
		WinnerSeat: &winnerSeat,
		Multiplier: 1,
		Doubled:    map[int]bool{0: true},
	}
	scores, _ := e.CalculateScore(state)
	for _, s := range scores {
		if s.PlayerID == 1 && s.Score != 4 {
			t.Errorf("landlord doubled score = %d, want 4", s.Score)
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
	state.Phase = PhaseRevealing
	if e.IsRoundEnd(state) {
		t.Error("revealing phase should not be round end")
	}
	state.Phase = PhaseDoubling
	if e.IsRoundEnd(state) {
		t.Error("doubling phase should not be round end")
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

	first := gs.CurrentSeat
	second := (first + 1) % 3
	third := (first + 2) % 3

	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_call"})
	gs = state.(*GameState)
	if gs.Phase != PhaseSnatching {
		t.Errorf("after call, phase = %d, want PhaseSnatching", gs.Phase)
	}

	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[second].ID, Action: "bid_call"})
	gs = state.(*GameState)
	if gs.LandlordSeat != second {
		t.Errorf("after snatch, landlord seat = %d, want %d", gs.LandlordSeat, second)
	}
	if gs.Multiplier != 2 {
		t.Errorf("after 1 snatch, multiplier = %d, want 2", gs.Multiplier)
	}

	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[third].ID, Action: "bid_pass"})
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_pass"})
	gs = state.(*GameState)

	if gs.Phase != PhaseRevealing {
		t.Errorf("after snatching, phase = %d, want PhaseRevealing", gs.Phase)
	}

	state = passRevealAndDouble(t, e, players, state)
	gs = state.(*GameState)

	if gs.Phase != PhasePlaying {
		t.Errorf("after doubling, phase = %d, want PhasePlaying", gs.Phase)
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

	state = passRevealAndDouble(t, e, players, state)
	gs = state.(*GameState)

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

func TestEngine_Reveal_AllPass(t *testing.T) {
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
	for gs.SnatchCount < 3 {
		var err error
		state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[gs.CurrentSeat].ID, Action: "bid_pass"})
		if err != nil {
			t.Fatalf("snatch error: %v", err)
		}
		gs = state.(*GameState)
	}

	gs = state.(*GameState)
	if gs.Phase != PhaseRevealing {
		t.Fatalf("expected PhaseRevealing, got %d", gs.Phase)
	}

	multBefore := gs.Multiplier
	for gs.RevealCount < 3 {
		seat := gs.CurrentSeat
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "pass"})
		gs = state.(*GameState)
	}

	if gs.Phase != PhaseDoubling {
		t.Errorf("after reveal all pass, phase = %d, want PhaseDoubling", gs.Phase)
	}
	if gs.Multiplier != multBefore {
		t.Errorf("multiplier changed from %d to %d, want unchanged", multBefore, gs.Multiplier)
	}
}

func TestEngine_Reveal_WithReveal(t *testing.T) {
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
	for gs.SnatchCount < 3 {
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[gs.CurrentSeat].ID, Action: "bid_pass"})
		gs = state.(*GameState)
	}
	gs = state.(*GameState)

	multBefore := gs.Multiplier
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "reveal_all"})
	gs = state.(*GameState)
	if !gs.Revealed[first] {
		t.Error("first player should be marked as revealed")
	}
	if gs.Multiplier != multBefore*2 {
		t.Errorf("multiplier = %d, want %d", gs.Multiplier, multBefore*2)
	}

	for gs.RevealCount < 3 {
		seat := gs.CurrentSeat
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "pass"})
		gs = state.(*GameState)
	}

	if gs.Phase != PhaseDoubling {
		t.Errorf("after reveal, phase = %d, want PhaseDoubling", gs.Phase)
	}
}

func TestEngine_Double_WithDouble(t *testing.T) {
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
	for gs.SnatchCount < 3 {
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[gs.CurrentSeat].ID, Action: "bid_pass"})
		gs = state.(*GameState)
	}
	for gs.RevealCount < 3 {
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[gs.CurrentSeat].ID, Action: "pass"})
		gs = state.(*GameState)
	}
	gs = state.(*GameState)

	if gs.Phase != PhaseDoubling {
		t.Fatalf("expected PhaseDoubling, got %d", gs.Phase)
	}

	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "double"})
	gs = state.(*GameState)
	if !gs.Doubled[first] {
		t.Error("first player should be marked as doubled")
	}

	for gs.DoubleCount < 3 {
		seat := gs.CurrentSeat
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "no_double"})
		gs = state.(*GameState)
	}

	if gs.Phase != PhasePlaying {
		t.Errorf("after double, phase = %d, want PhasePlaying", gs.Phase)
	}
}

func TestEngine_PassNotAllowedWhenLeading(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "Alice", Seat: 0},
		{ID: 2, Name: "Bob", Seat: 1},
		{ID: 3, Name: "Charlie", Seat: 2},
	}
	state, _ := e.Init(players)
	gs := state.(*GameState)

	// Go through bidding → snatch → reveal → double to reach playing
	first := gs.CurrentSeat
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_call"})
	for gs.SnatchCount < 3 {
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[gs.CurrentSeat].ID, Action: "bid_pass"})
		gs = state.(*GameState)
	}
	state = passRevealAndDouble(t, e, players, state)
	gs = state.(*GameState)

	if gs.Phase != PhasePlaying {
		t.Fatalf("expected PhasePlaying, got %d", gs.Phase)
	}

	landlordSeat := gs.CurrentSeat
	// Landlord plays a card
	landlordHand := gs.Players[landlordSeat].Hand
	state, err := e.ExecuteAction(state, game.PlayerAction{PlayerID: players[landlordSeat].ID, Action: "play", Cards: []int{landlordHand[0].ID}})
	if err != nil {
		t.Fatalf("landlord play error: %v", err)
	}
	gs = state.(*GameState)

	// Two opponents pass
	for i := 0; i < 2; i++ {
		seat := gs.CurrentSeat
		state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[seat].ID, Action: "pass"})
		if err != nil {
			t.Fatalf("opponent pass error: %v", err)
		}
		gs = state.(*GameState)
	}

	// Now landlord leads again — passing should be rejected
	if gs.CurrentSeat != landlordSeat {
		t.Fatalf("expected landlord's turn after 2 passes")
	}
	_, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[landlordSeat].ID, Action: "pass"})
	if err == nil {
		t.Error("expected error: cannot pass when leading")
	}
}
