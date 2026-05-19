package doudizhu

import (
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

func TestEngine_TwoPassesResetLastPlay_FreeLead(t *testing.T) {
	e := &Engine{}
	players := []game.PlayerInfo{
		{ID: 1, Name: "P0", Seat: 0},
		{ID: 2, Name: "P1", Seat: 1},
		{ID: 3, Name: "P2", Seat: 2},
	}
	state, _ := e.Init(players)
	gs := state.(*GameState)

	// 快速过到出牌阶段
	first := gs.CurrentSeat
	state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[first].ID, Action: "bid_call"})
	gs = state.(*GameState)
	for gs.SnatchCount < 3 {
		state, _ = e.ExecuteAction(state, game.PlayerAction{PlayerID: players[gs.CurrentSeat].ID, Action: "bid_pass"})
		gs = state.(*GameState)
	}
	state = passRevealAndDouble(t, e, players, state)
	gs = state.(*GameState)

	// 强制指定角色：seat0=地主, seat1=对手1, seat2=对手2
	// 直接写入手牌确保测试可重复
	gs.Players[0].IsLandlord = true
	gs.Players[1].IsLandlord = false
	gs.Players[2].IsLandlord = false

	// 座0(地主)手牌放一张小单(♠3=0)和底牌
	gs.Players[0].Hand = []Card{{ID: 0}, {ID: 4}, {ID: 5}, {ID: 6}, {ID: 7}, {ID: 8},
		{ID: 9}, {ID: 10}, {ID: 11}, {ID: 12}, {ID: 13}, {ID: 14},
		{ID: 15}, {ID: 16}, {ID: 17}, {ID: 18}, {ID: 19}, {ID: 20}, {ID: 21}, {ID: 22}}
	// 座1(对手1)手牌：单张2=♠4=1,  一对♠5=4+♦5=43
	gs.Players[1].Hand = []Card{{ID: 1}, {ID: 4}, {ID: 43}, {ID: 23}, {ID: 24},
		{ID: 25}, {ID: 26}, {ID: 27}, {ID: 28}, {ID: 29},
		{ID: 30}, {ID: 31}, {ID: 32}, {ID: 33}, {ID: 34}, {ID: 35}, {ID: 36}}
	// 座2(对手2)手牌
	gs.Players[2].Hand = []Card{{ID: 2}, {ID: 3}, {ID: 37}, {ID: 38}, {ID: 39},
		{ID: 40}, {ID: 41}, {ID: 42}, {ID: 44}, {ID: 45},
		{ID: 46}, {ID: 47}, {ID: 48}, {ID: 49}, {ID: 50}, {ID: 51}, {ID: 52}}

	gs.CurrentSeat = 0
	gs.Phase = PhasePlaying
	gs.LastPlay = nil
	gs.ConsecutivePasses = 0
	gs.LandlordSeat = 0
	gs.Multiplier = 1

	t.Logf("地主=seat0, 对手1=seat1, 对手2=seat2")

	// 地主出最小单张♠3(ID=0, rank=3)
	state, err := e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "play", Cards: []int{0}})
	if err != nil {
		t.Fatalf("地主出牌失败: %v", err)
	}
	gs = state.(*GameState)
	t.Logf("1. 地主出♠3, LastPlay.seat=%d, CurrentSeat=%d", gs.LastPlay.Seat, gs.CurrentSeat)

	// 对手1出♠4(ID=1, rank=4) 压牌
	state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "play", Cards: []int{1}})
	if err != nil {
		t.Fatalf("对手1压牌失败: %v", err)
	}
	gs = state.(*GameState)
	t.Logf("2. 对手1出♠4压牌, LastPlay.seat=%d, CurrentSeat=%d", gs.LastPlay.Seat, gs.CurrentSeat)

	// 对手2过牌
	state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: 3, Action: "pass"})
	if err != nil {
		t.Fatalf("对手2过牌失败: %v", err)
	}
	gs = state.(*GameState)
	t.Logf("3. 对手2过牌, LastPlay.seat=%d, ConsecutivePasses=%d, CurrentSeat=%d",
		gs.LastPlay.Seat, gs.ConsecutivePasses, gs.CurrentSeat)

	// 地主也过牌
	state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: 1, Action: "pass"})
	if err != nil {
		t.Fatalf("地主过牌失败: %v", err)
	}
	gs = state.(*GameState)
	t.Logf("4. 地主过牌, LastPlay=%v, ConsecutivePasses=%d, CurrentSeat=%d",
		gs.LastPlay, gs.ConsecutivePasses, gs.CurrentSeat)

	// === 验证 ===
	if gs.CurrentSeat != 1 {
		t.Errorf("❌ 应该轮到对手1(seat1)，但 CurrentSeat=%d", gs.CurrentSeat)
	} else {
		t.Logf("✅ CurrentSeat=1 正确")
	}
	if gs.LastPlay != nil {
		t.Errorf("❌ 两轮过牌后 LastPlay 应为 nil，但 seat=%d cards=%v", gs.LastPlay.Seat, gs.LastPlay.Cards)
	} else {
		t.Logf("✅ LastPlay=nil 正确（回合已重置）")
	}
	if gs.ConsecutivePasses != 0 {
		t.Errorf("❌ ConsecutivePasses 应为 0，实际=%d", gs.ConsecutivePasses)
	} else {
		t.Logf("✅ ConsecutivePasses=0 正确")
	}

	// 对手1 自由出牌：打一对 ♠5(ID=4)+♦5(ID=43)
	state, err = e.ExecuteAction(state, game.PlayerAction{PlayerID: 2, Action: "play", Cards: []int{4, 43}})
	if err != nil {
		t.Errorf("❌ 自由出牌失败（应该允许出对子）：%v", err)
	} else {
		t.Logf("✅ 自由出牌成功！出了对子 [4,43] 不受限于之前单张牌型")
	}
}
