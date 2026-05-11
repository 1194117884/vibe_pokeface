package ai

import (
    "testing"
)

func TestPromptBuilder_IncludesCards(t *testing.T) {
    pb := NewPromptBuilder()
    prompt := pb.BuildPlayDecision(3, []int{3, 4, 5, 6, 7, 8, 9}, []int{5})
    if len(prompt) < 50 {
        t.Error("prompt too short, expected detailed play decision prompt")
    }
}

func TestPromptBuilder_IncludesBidding(t *testing.T) {
    pb := NewPromptBuilder()
    prompt := pb.BuildBidDecision([]int{3, 4, 5, 6, 7}, 3)
    if len(prompt) < 30 {
        t.Error("bid decision prompt too short")
    }
}

func TestParsePlayResponse_Valid(t *testing.T) {
    action, cards, err := ParseAIPlayResponse(`{"action": "play", "cards": [0, 13, 26]}`)
    if err != nil {
        t.Fatalf("parse error: %v", err)
    }
    if action != "play" {
        t.Errorf("action = %s, want play", action)
    }
    if len(cards) != 3 {
        t.Errorf("cards = %d, want 3", len(cards))
    }
}

func TestParseAIPlayResponse_Pass(t *testing.T) {
    action, cards, err := ParseAIPlayResponse(`{"action": "pass", "cards": []}`)
    if err != nil {
        t.Fatalf("parse error: %v", err)
    }
    if action != "pass" {
        t.Errorf("action = %s, want pass", action)
    }
    if len(cards) != 0 {
        t.Errorf("cards = %d, want 0", len(cards))
    }
}

func TestParseAIPlayResponse_Invalid(t *testing.T) {
    _, _, err := ParseAIPlayResponse(`not json`)
    if err == nil {
        t.Error("expected error for invalid JSON")
    }
}

func TestParseBidResponse(t *testing.T) {
    bid, err := ParseAIBidResponse(`{"action": "bid_call"}`)
    if err != nil {
        t.Fatalf("parse error: %v", err)
    }
    if bid != "bid_call" {
        t.Errorf("bid = %s, want bid_call", bid)
    }
}

func TestParseChatResponse(t *testing.T) {
    msg, err := ParseAIChatResponse(`{"message": "你好，我叫张三！"}`)
    if err != nil {
        t.Fatalf("parse error: %v", err)
    }
    if msg != "你好，我叫张三！" {
        t.Errorf("message = %s, want 你好，我叫张三！", msg)
    }
}
