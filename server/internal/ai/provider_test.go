package ai

import (
    "fmt"
    "strings"
    "testing"

    "github.com/yongkl/vibe-pokeface/internal/model"
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

func TestNewProvider_SupportedProviders(t *testing.T) {
    tests := []struct {
        name        string
        provider    string
        wantType    string
        wantDefault string
    }{
        {"openai", "openai", "*ai.OpenAIProvider", "https://api.openai.com"},
        {"anthropic", "anthropic", "*ai.AnthropicProvider", "https://api.anthropic.com"},
        {"deepseek", "deepseek", "*ai.OpenAIProvider", "https://api.deepseek.com"},
        {"minimax", "minimax", "*ai.OpenAIProvider", "https://api.minimax.chat"},
        {"glm", "glm", "*ai.OpenAIProvider", "https://open.bigmodel.cn"},
        {"qwen", "qwen", "*ai.OpenAIProvider", "https://dashscope.aliyuncs.com"},
        {"custom", "custom", "*ai.OpenAIProvider", "https://api.openai.com"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := &model.LLMConfig{
                Provider:    tt.provider,
                APIKey:      "sk-test",
                Model:       "test-model",
                Temperature: 0.7,
                MaxTokens:   1024,
            }
            p, err := NewProvider(cfg)
            if err != nil {
                t.Fatalf("NewProvider(%q) unexpected error: %v", tt.provider, err)
            }
            typeName := fmt.Sprintf("%T", p)
            if typeName != tt.wantType {
                t.Errorf("NewProvider(%q) type = %s, want %s", tt.provider, typeName, tt.wantType)
            }
            var apiURL string
            switch v := p.(type) {
            case *OpenAIProvider:
                apiURL = v.apiURL
            case *AnthropicProvider:
                apiURL = v.apiURL
            }
            if !strings.HasPrefix(apiURL, tt.wantDefault) {
                t.Errorf("NewProvider(%q) apiURL = %s, want prefix %s", tt.provider, apiURL, tt.wantDefault)
            }
        })
    }
}

func TestNewProvider_UnsupportedProvider(t *testing.T) {
    cfg := &model.LLMConfig{Provider: "unknown", APIKey: "sk-test", Model: "test"}
    _, err := NewProvider(cfg)
    if err == nil {
        t.Fatal("expected error for unsupported provider")
    }
}

func TestNewProvider_CustomURLOverride(t *testing.T) {
    customURL := "https://my-proxy.example.com/v1/chat/completions"
    cfg := &model.LLMConfig{
        Provider: "deepseek",
        APIKey:   "sk-test",
        Model:    "test-model",
        APIURL:   &customURL,
    }
    p, err := NewProvider(cfg)
    if err != nil {
        t.Fatalf("NewProvider error: %v", err)
    }
    provider := p.(*OpenAIProvider)
    if provider.apiURL != customURL {
        t.Errorf("apiURL = %s, want %s", provider.apiURL, customURL)
    }
}
