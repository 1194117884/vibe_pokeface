package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestExtractToolCall_Play(t *testing.T) {
	llmOutput := `{"tool": "play_cards", "args": {"cards": [0, 13, 26], "chat": "接招！"}}`
	call, err := ExtractToolCall(llmOutput)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if call.Name != "play_cards" {
		t.Errorf("name = %s, want play_cards", call.Name)
	}
}

func TestExtractToolCall_Bid(t *testing.T) {
	llmOutput := `{"tool": "bid_landlord", "args": {"chat": "我来当地主！"}}`
	call, err := ExtractToolCall(llmOutput)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if call.Name != "bid_landlord" {
		t.Errorf("name = %s, want bid_landlord", call.Name)
	}
}

func TestExtractToolCall_Info(t *testing.T) {
	llmOutput := `{"tool": "check_my_hand", "args": {}}`
	call, err := ExtractToolCall(llmOutput)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if call.Name != "check_my_hand" {
		t.Errorf("name = %s, want check_my_hand", call.Name)
	}
}

func TestExtractToolCall_WithCodeFences(t *testing.T) {
	llmOutput := "```json\n{\"tool\": \"play_cards\", \"args\": {\"cards\": []}}\n```"
	call, err := ExtractToolCall(llmOutput)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if call.Name != "play_cards" {
		t.Errorf("name = %s, want play_cards", call.Name)
	}
}

func TestExtractToolCall_Invalid(t *testing.T) {
	_, err := ExtractToolCall("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAIAgent_NewAndStop(t *testing.T) {
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
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop() timed out")
	}
}

func TestAIAgent_Trigger(t *testing.T) {
	triggered := make(chan struct{}, 1)
	decisionFn := func(agent *AIAgent, phase string, handCards []int, stateJSON string) string {
		triggered <- struct{}{}
		return `{"tool": "play_cards", "args": {"cards": []}}`
	}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil)
	agent.MakeDecisionFunc = decisionFn
	agent.Start()

	agent.Trigger()

	select {
	case <-triggered:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("trigger was not processed")
	}
	agent.Stop()
}

type mockToolProvider struct {
	responses []*LLMResultWithTools
	callCount int
	requests  [][]ChatMessage
}

func (m *mockToolProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (*LLMResult, error) {
	return &LLMResult{Content: "{}"}, nil
}

func (m *mockToolProvider) CompleteWithTools(ctx context.Context, messages []ChatMessage, tools []ToolSchema) (*LLMResultWithTools, error) {
	copied := make([]ChatMessage, len(messages))
	copy(copied, messages)
	m.requests = append(m.requests, copied)
	if m.callCount >= len(m.responses) {
		return &LLMResultWithTools{}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *mockToolProvider) ProviderName() string { return "mock" }
func (m *mockToolProvider) ModelName() string    { return "mock" }

func TestAgent_MultiTurnInfoThenAction(t *testing.T) {
	exec := &mockExecutor{}
	provider := &mockToolProvider{
		responses: []*LLMResultWithTools{
			{
				ToolCalls: []AssistantToolCall{
					{ID: "call_1", Type: "function"},
				},
			},
			{
				ToolCalls: []AssistantToolCall{
					{ID: "call_2", Type: "function"},
				},
			},
		},
	}
	provider.responses[0].ToolCalls[0].Function.Name = "check_my_hand"
	provider.responses[0].ToolCalls[0].Function.Arguments = "{}"
	provider.responses[1].ToolCalls[0].Function.Name = "play_cards"
	provider.responses[1].ToolCalls[0].Function.Arguments = `{"cards":[0,13]}`

	agent := NewAIAgent("ai:bot:1", 0, nil, provider, exec)
	agent.HandCards = []int{0, 13, 26}
	agent.stateJSON = `{"phase":4,"current_seat":0,"landlord_seat":0,"multiplier":1,"players":[{"seat":0,"is_landlord":true,"hand":[{"id":0},{"id":13},{"id":26}]},{"seat":1,"is_landlord":false,"hand":[{"id":1}]},{"seat":2,"is_landlord":false,"hand":[{"id":2}]}]}`
	agent.makeDecisionWithTools()

	if exec.lastAction != "play" {
		t.Errorf("expected play action, got %s", exec.lastAction)
	}
	if len(exec.lastCards) != 2 {
		t.Errorf("expected 2 cards, got %v", exec.lastCards)
	}
	if provider.callCount != 2 {
		t.Errorf("expected 2 LLM calls, got %d", provider.callCount)
	}
}

func TestAgent_MultiTurnMaxLoop(t *testing.T) {
	exec := &mockExecutor{}
	responses := make([]*LLMResultWithTools, 11)
	for i := range responses {
		responses[i] = &LLMResultWithTools{
			ToolCalls: []AssistantToolCall{
				{ID: fmt.Sprintf("call_%d", i), Type: "function"},
			},
		}
		responses[i].ToolCalls[0].Function.Name = "check_my_hand"
		responses[i].ToolCalls[0].Function.Arguments = "{}"
	}
	provider := &mockToolProvider{responses: responses}

	agent := NewAIAgent("ai:bot:1", 0, nil, provider, exec)
	agent.HandCards = []int{}
	agent.stateJSON = `{"phase":4,"current_seat":0,"landlord_seat":0,"players":[]}`
	agent.makeDecisionWithTools()

	// Should stop at 10 iterations then fallback
	if provider.callCount != 10 {
		t.Errorf("expected 10 calls (max loop), got %d", provider.callCount)
	}
	if exec.lastAction != "pass" {
		t.Errorf("expected fallback pass, got %s", exec.lastAction)
	}
}

func TestAgent_MultiTurnActionAndChat(t *testing.T) {
	exec := &mockExecutor{}
	provider := &mockToolProvider{
		responses: []*LLMResultWithTools{
			{
				ToolCalls: []AssistantToolCall{
					{ID: "call_1", Type: "function"},
				},
			},
		},
	}
	provider.responses[0].ToolCalls[0].Function.Name = "play_cards"
	provider.responses[0].ToolCalls[0].Function.Arguments = `{"cards":[],"chat":"过"}`

	agent := NewAIAgent("ai:bot:1", 0, nil, provider, exec)
	agent.HandCards = []int{0}
	agent.stateJSON = `{"phase":4,"current_seat":0,"landlord_seat":0,"players":[]}`
	agent.makeDecisionWithTools()

	if exec.lastAction != "pass" {
		t.Errorf("expected pass, got %s", exec.lastAction)
	}
	if exec.lastChat != "过" {
		t.Errorf("expected chat '过', got '%s'", exec.lastChat)
	}
}

func TestAgent_PersistentMemoryCarriesAcrossDecisions(t *testing.T) {
	exec := &mockExecutor{}
	provider := &mockToolProvider{
		responses: []*LLMResultWithTools{
			{ToolCalls: []AssistantToolCall{{ID: "call_1", Type: "function"}}},
			{ToolCalls: []AssistantToolCall{{ID: "call_2", Type: "function"}}},
		},
	}
	provider.responses[0].ToolCalls[0].Function.Name = "play_cards"
	provider.responses[0].ToolCalls[0].Function.Arguments = `{"cards":[0],"chat":"先出小牌探路"}`
	provider.responses[1].ToolCalls[0].Function.Name = "play_cards"
	provider.responses[1].ToolCalls[0].Function.Arguments = `{"cards":[13],"chat":"延续刚才的节奏"}`

	agent := NewAIAgent("ai:bot:1", 0, nil, provider, exec)
	agent.HandCards = []int{0, 13, 26}
	agent.stateJSON = `{"phase":4,"round_num":1,"current_seat":0,"landlord_seat":0,"multiplier":1,"players":[{"seat":0,"is_landlord":true,"hand":[{"id":0},{"id":13},{"id":26}]},{"seat":1,"is_landlord":false,"hand":[{"id":1}]},{"seat":2,"is_landlord":false,"hand":[{"id":2}]}],"play_history":[]}`
	agent.makeDecisionWithTools()

	agent.HandCards = []int{13, 26}
	agent.stateJSON = `{"phase":4,"round_num":1,"current_seat":0,"landlord_seat":0,"multiplier":1,"players":[{"seat":0,"is_landlord":true,"hand":[{"id":13},{"id":26}]},{"seat":1,"is_landlord":false,"hand":[]},{"seat":2,"is_landlord":false,"hand":[{"id":2}]}],"play_history":[{"seat":0,"play":{"type":1,"main_rank":3,"length":1},"cards":[{"id":0}]},{"seat":1,"play":{"type":0},"cards":[]}]}`
	agent.makeDecisionWithTools()

	if provider.callCount != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", provider.callCount)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("expected 2 captured requests, got %d", len(provider.requests))
	}

	var second strings.Builder
	for _, msg := range provider.requests[1] {
		second.WriteString(msg.Content)
		second.WriteString("\n")
	}
	got := second.String()
	if !strings.Contains(got, "先出小牌探路") {
		t.Fatalf("second request should include previous decision memory, got:\n%s", got)
	}
	if !strings.Contains(got, "座位1：过牌") {
		t.Fatalf("second request should include new play history since last decision, got:\n%s", got)
	}
}
