package ai

import (
	"context"
	"fmt"
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

func TestExtractToolCall_Say(t *testing.T) {
	llmOutput := `{"tool": "say", "args": {"message": "大家好！"}}`
	call, err := ExtractToolCall(llmOutput)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if call.Name != "say" {
		t.Errorf("name = %s, want say", call.Name)
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
}

func (m *mockToolProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (*LLMResult, error) {
	return &LLMResult{Content: "{}"}, nil
}

func (m *mockToolProvider) CompleteWithTools(ctx context.Context, messages []ChatMessage, tools []ToolSchema) (*LLMResultWithTools, error) {
	if m.callCount >= len(m.responses) {
		return &LLMResultWithTools{}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

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
					{ID: "call_2", Type: "function"},
				},
			},
		},
	}
	provider.responses[0].ToolCalls[0].Function.Name = "say"
	provider.responses[0].ToolCalls[0].Function.Arguments = `{"message":"大家好"}`
	provider.responses[0].ToolCalls[1].Function.Name = "play_cards"
	provider.responses[0].ToolCalls[1].Function.Arguments = `{"cards":[]}`

	agent := NewAIAgent("ai:bot:1", 0, nil, provider, exec)
	agent.HandCards = []int{0}
	agent.stateJSON = `{"phase":4,"current_seat":0,"landlord_seat":0,"players":[]}`
	agent.makeDecisionWithTools()

	if exec.lastAction != "pass" {
		t.Errorf("expected pass, got %s", exec.lastAction)
	}
	if exec.lastChat != "大家好" {
		t.Errorf("expected chat '大家好', got '%s'", exec.lastChat)
	}
}
