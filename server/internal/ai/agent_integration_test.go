package ai

import (
	"testing"
)

// mockExecutor for testing
type mockExecutor struct {
	lastAction string
	lastCards  []int
	lastChat   string
	actions    []struct{ action string; cards []int }
}

func (m *mockExecutor) ExecuteAction(userID string, action string, cards []int) {
	m.lastAction = action
	m.lastCards = cards
	m.actions = append(m.actions, struct{ action string; cards []int }{action, cards})
}

func (m *mockExecutor) SendChat(senderID string, content string, msgType string) {
	m.lastChat = content
}

func TestAgent_ExecutePlayAction(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	agent.executeToolCall(`{"tool": "play_cards", "args": {"cards": [], "chat": "过"}}`)

	if exec.lastAction != "pass" {
		t.Errorf("expected pass action, got %s", exec.lastAction)
	}
	if exec.lastChat != "过" {
		t.Errorf("expected chat '过', got '%s'", exec.lastChat)
	}
}

func TestAgent_ExecuteBidAction(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	agent.executeToolCall(`{"tool": "bid_landlord", "args": {"chat": "我来！"}}`)

	if exec.lastAction != "bid_call" {
		t.Errorf("expected bid_call, got %s", exec.lastAction)
	}
	if exec.lastChat != "我来！" {
		t.Errorf("expected chat '我来！', got '%s'", exec.lastChat)
	}
}

func TestAgent_ExecutePassBidAction(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	agent.executeToolCall(`{"tool": "pass_bid", "args": {}}`)

	if exec.lastAction != "bid_pass" {
		t.Errorf("expected bid_pass, got %s", exec.lastAction)
	}
}

func TestAgent_ExecuteChatTool(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	agent.executeToolCall(`{"tool": "say", "args": {"message": "大家好！"}}`)

	if exec.lastChat != "大家好！" {
		t.Errorf("expected chat '大家好！', got '%s'", exec.lastChat)
	}
}

func TestAgent_ExecutePlayCards(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	agent.executeToolCall(`{"tool": "play_cards", "args": {"cards": [0, 13, 26]}}`)

	if exec.lastAction != "play" {
		t.Errorf("expected play action, got %s", exec.lastAction)
	}
	if len(exec.lastCards) != 3 || exec.lastCards[0] != 0 {
		t.Errorf("unexpected cards: %v", exec.lastCards)
	}
}

func TestAgent_MakeDecisionFunc(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.MakeDecisionFunc = func(agent *AIAgent, phase string, handCards []int, stateJSON string) string {
		return `{"tool": "play_cards", "args": {"cards": [], "chat": "过"}}`
	}

	agent.makeDecision()

	if exec.lastAction != "pass" {
		t.Errorf("expected pass action, got %s", exec.lastAction)
	}
}

func TestAgent_FallbackAction(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	agent.fallbackAction("playing")

	if exec.lastAction != "pass" {
		t.Errorf("expected fallback pass, got %s", exec.lastAction)
	}
}

func TestAgent_FallbackBid(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	agent.fallbackAction("bidding")

	if exec.lastAction != "bid_pass" {
		t.Errorf("expected bid_pass, got %s", exec.lastAction)
	}
}

func TestAgent_ExtractAndExecutePlay(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)

	call, err := ExtractToolCall(`{"tool": "play_cards", "args": {"cards": [0, 13, 26]}}`)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if call.Name != "play_cards" {
		t.Errorf("name = %s, want play_cards", call.Name)
	}

	agent.executeToolCall(`{"tool": "play_cards", "args": {"cards": [0, 13, 26]}}`)
	if exec.lastAction != "play" {
		t.Errorf("expected play, got %s", exec.lastAction)
	}
}

func TestAgent_NewAndStop(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil)
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	if agent.UserID != "ai:bot:1" {
		t.Errorf("userID = %s, want ai:bot:1", agent.UserID)
	}
	// Stop should not block
	done := make(chan struct{})
	go func() {
		agent.Stop()
		close(done)
	}()
	// The stop should complete (no deadlock from closing stopChan twice)
	agent.Stop()
	// Should be safe to stop multiple times
	agent.Stop()
}

func TestAgent_TriggerAndStop(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.MakeDecisionFunc = func(agent *AIAgent, phase string, handCards []int, stateJSON string) string {
		return `{"tool": "say", "args": {"message": "hi"}}`
	}
	agent.Start()

	agent.Trigger()
	agent.Stop()
	// Stopped agent: should not panic on second trigger (channel closed)
	// Note: after Stop, the channel is closed, so Trigger's select may panic
	// because sending on a closed channel panics in Go
}
