package dashengji

import (
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

func makePlayers() []game.PlayerInfo {
	return []game.PlayerInfo{
		{ID: 1, Name: "P0", Seat: 0},
		{ID: 2, Name: "P1", Seat: 1},
		{ID: 3, Name: "P2", Seat: 2},
		{ID: 4, Name: "P3", Seat: 3},
	}
}

func TestInit_RequiresFourPlayers(t *testing.T) {
	eng := &Engine{}
	_, err := eng.Init([]game.PlayerInfo{{ID: 1, Seat: 0}})
	if err == nil {
		t.Error("Init with 1 player should fail")
	}
}

func TestInit_CreatesValidState(t *testing.T) {
	eng := &Engine{}
	state, err := eng.Init(makePlayers())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	gs := state.(*GameState)

	if gs.Phase != PhaseSetTrump {
		t.Errorf("expected PhaseSetTrump, got %v", gs.Phase)
	}
	if len(gs.Players) != 4 {
		t.Errorf("expected 4 players, got %d", len(gs.Players))
	}
	for i, p := range gs.Players {
		if len(p.Hand) != 39 {
			t.Errorf("player %d: expected 39 cards, got %d", i, len(p.Hand))
		}
	}
	if len(gs.BottomCards) != 6 {
		t.Errorf("expected 6 bottom cards, got %d", len(gs.BottomCards))
	}
	if gs.CurrentLevel != 3 {
		t.Errorf("expected starting level 3, got %d", gs.CurrentLevel)
	}
}

func TestInit_DealerTeamValid(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	gs := state.(*GameState)

	if gs.DealerSeats != [2]int{0, 2} && gs.DealerSeats != [2]int{1, 3} {
		t.Errorf("invalid dealer seats: %v", gs.DealerSeats)
	}
	if gs.DealerSeats[1]-gs.DealerSeats[0] != 2 {
		t.Errorf("dealer seats should be opposite, got %v", gs.DealerSeats)
	}
}

func TestIsRoundEnd(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	if eng.IsRoundEnd(state) {
		t.Error("new game should not be round end")
	}
	gs := state.(*GameState)
	gs.Phase = PhaseEnded
	if !eng.IsRoundEnd(state) {
		t.Error("ended game should report round end")
	}
}

func TestFilterForPlayer_HidesOtherHands(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	filtered := eng.FilterForPlayer(state, 0)
	fgs := filtered.(*GameState)

	for i, p := range fgs.Players {
		if i == 0 {
			if len(p.Hand) != 39 {
				t.Errorf("player 0 should see own 39 cards, sees %d", len(p.Hand))
			}
		} else {
			if p.Hand != nil {
				t.Errorf("player %d hand should be hidden, got %d cards", i, len(p.Hand))
			}
		}
	}
}

func TestValidateAction_WrongTurn(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	gs := state.(*GameState)
	// Try action from seat 1 when current seat is dealer[0]
	if gs.CurrentSeat == 1 {
		gs.CurrentSeat = 0
	}
	action := game.PlayerAction{PlayerID: gs.Players[1].UserID, Action: "set_trump"}
	err := eng.ValidateAction(state, action)
	if err == nil {
		t.Error("should reject action from wrong seat")
	}
}

func TestValidateAction_WrongPhase(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	gs := state.(*GameState)
	action := game.PlayerAction{PlayerID: gs.Players[gs.CurrentSeat].UserID, Action: "play"}
	err := eng.ValidateAction(state, action)
	if err == nil {
		t.Error("should reject 'play' action during set_trump phase")
	}
}
