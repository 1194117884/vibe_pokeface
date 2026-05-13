package ai

import (
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
