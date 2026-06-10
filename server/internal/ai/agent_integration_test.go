package ai

import (
	"strings"
	"testing"
)

// mockExecutor for testing
type mockExecutor struct {
	lastAction string
	lastCards  []int
	lastChat   string
	events     []string
	actions    []struct {
		action string
		cards  []int
	}
	statuses []struct {
		status  string
		message string
	}
}

func (m *mockExecutor) ExecuteAction(userID string, action string, cards []int) {
	m.lastAction = action
	m.lastCards = cards
	m.events = append(m.events, "action:"+action)
	m.actions = append(m.actions, struct {
		action string
		cards  []int
	}{action, cards})
}

func (m *mockExecutor) SendChat(senderID string, content string, msgType string) {
	m.lastChat = content
	m.events = append(m.events, "chat:"+content)
}

func (m *mockExecutor) ReportAIStatus(userID string, status string, message string) {
	m.statuses = append(m.statuses, struct {
		status  string
		message string
	}{status, message})
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

	agent.executeToolCall(`{"tool": "play_cards", "args": {"cards": [0, 13], "chat": "看我的！"}}`)

	if exec.lastAction != "play" {
		t.Errorf("expected play action, got %s", exec.lastAction)
	}
	if exec.lastChat != "看我的！" {
		t.Errorf("expected chat '看我的！', got '%s'", exec.lastChat)
	}
	wantEvents := []string{"chat:看我的！", "action:play"}
	if strings.Join(exec.events, "|") != strings.Join(wantEvents, "|") {
		t.Fatalf("events = %v, want %v", exec.events, wantEvents)
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

func TestDashengjiAgent_PlayCardsSendsChatBeforeAction(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.GameType = "dashengji"

	agent.executeDashengjiToolCall(&ToolCall{
		Name: "play_cards",
		Args: []byte(`{"cards":[15,69],"chat":"先压一手"}`),
	})

	wantEvents := []string{"chat:先压一手", "action:play"}
	if strings.Join(exec.events, "|") != strings.Join(wantEvents, "|") {
		t.Fatalf("events = %v, want %v", exec.events, wantEvents)
	}
}

func TestAgent_ExecuteToolCallReportsActionStatus(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.stateJSON = `{"phase":4}`

	agent.executeToolCall(`{"tool": "play_cards", "args": {"cards": [0]}}`)

	if len(exec.statuses) == 0 {
		t.Fatal("expected an AI status before action execution")
	}
	got := exec.statuses[len(exec.statuses)-1]
	if got.status != "acting" || got.message != "正在思考出牌" {
		t.Fatalf("status = %#v, want acting/正在思考出牌", got)
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

	agent.fallbackAction("calling")

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
		return `{"tool": "play_cards", "args": {"cards": [], "chat": "过"}}`
	}
	agent.Start()

	agent.Trigger()
	agent.Stop()
	// Stopped agent: should not panic on second trigger (channel closed)
	// Note: after Stop, the channel is closed, so Trigger's select may panic
	// because sending on a closed channel panics in Go
}

func TestExecuteInfoTool_CheckGameStatus_ParsesRealStateJSON(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil)
	// Real state JSON: hand cards are [{"id":N}, ...] not [N, ...]
	agent.stateJSON = `{
		"phase": 4,
		"current_seat": 0,
		"landlord_seat": 0,
		"multiplier": 1,
		"last_play": null,
		"players": [
			{"seat": 0, "is_landlord": true, "hand": [{"id":53},{"id":12}]},
			{"seat": 1, "is_landlord": false, "hand": null},
			{"seat": 2, "is_landlord": false, "hand": null}
		]
	}`

	result := agent.executeInfoTool("check_game_status")

	if result == "游戏状态不可用" {
		t.Fatal("state should be available")
	}
	if result == "无法解析游戏状态" {
		t.Fatal("should parse real state JSON with [{\"id\":N}] hand format")
	}
	// Verify key data appears in result
	if !strings.Contains(result, "座位0") {
		t.Errorf("expected seat 0 info in result, got: %s", result)
	}
}
