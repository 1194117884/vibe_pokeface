package ai

import "testing"

func TestDashengjiAgent_DetectPhaseAndTools(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.stateJSON = `{"phase":3}`

	if got := agent.detectPhase(); got != "discard_bottom" {
		t.Fatalf("phase = %s, want discard_bottom", got)
	}

	tools := GetToolSchemasForGame("dashengji", "discard_bottom")
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Function.Name] = true
	}
	if !names["discard_bottom"] {
		t.Fatal("discard_bottom tool is missing")
	}
	if names["bid_landlord"] {
		t.Fatal("doudizhu bid tool should not be exposed for dashengji")
	}
}

func TestDashengjiAgent_FallbackSetTrump(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{53, 25, 16, 1}
	agent.stateJSON = `{"phase":0,"level_rank":6}`

	agent.fallbackDashengjiAction("set_trump")

	if exec.lastAction != "set_trump" {
		t.Fatalf("action = %s, want set_trump", exec.lastAction)
	}
	if len(exec.lastCards) != 3 {
		t.Fatalf("cards = %v, want 3 trump cards", exec.lastCards)
	}
}

func TestDashengjiAgent_FallbackDiscardBottomChoosesSixCards(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{0, 1, 2, 3, 4, 5, 9, 10, 13, 16}
	agent.stateJSON = `{"phase":3,"trump_suit":1,"level_rank":6}`

	agent.fallbackDashengjiAction("discard_bottom")

	if exec.lastAction != "discard_bottom" {
		t.Fatalf("action = %s, want discard_bottom", exec.lastAction)
	}
	if len(exec.lastCards) != 6 {
		t.Fatalf("cards = %v, want 6 cards", exec.lastCards)
	}
	handSet := make(map[int]bool)
	for _, id := range agent.HandCards {
		handSet[id] = true
	}
	for _, id := range exec.lastCards {
		if !handSet[id] {
			t.Fatalf("discarded card %d is not in hand %v", id, agent.HandCards)
		}
	}
}

func TestDashengjiAgent_FallbackPlayFollowsLedSuitSingle(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{0, 14, 28}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"trump_suit":1,
		"level_rank":6,
		"round_plays":[{"seat":0,"play":{"type":1,"main_rank":13,"length":1},"cards":[{"id":10}]}]
	}`

	agent.fallbackDashengjiAction("playing")

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if len(exec.lastCards) != 1 || exec.lastCards[0] != 0 {
		t.Fatalf("cards = %v, want spade follow card [0]", exec.lastCards)
	}
}

func TestDashengjiAgent_PlayCardsSendsChat(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, exec)
	agent.SetGameType("dashengji")

	agent.executeToolCall(`{"tool":"play_cards","args":{"cards":[0],"chat":"这手先探路"}}`)

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if exec.lastChat != "这手先探路" {
		t.Fatalf("chat = %q, want 这手先探路", exec.lastChat)
	}
}
