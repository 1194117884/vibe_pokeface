# LLM Multi-Platform Provider Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add DeepSeek, MiniMax, GLM (Zhipu), and Qwen (Alibaba Cloud) as named LLM providers for the AI player system, with correct default API URLs and model compatibility guidance.

**Architecture:** All four new providers use OpenAI-compatible API formats (messages array, Bearer auth, same response structure). They reuse the existing `OpenAIProvider` implementation with provider-specific default API URLs. The `NewProvider` factory gains a URL resolution helper that picks defaults per provider while still allowing custom `api_url` overrides.

**Tech Stack:** Go server (`server/internal/ai/`), Next.js frontend (`frontend/app/admin/llm-config/`)

---

### Task 1: Add provider constants and default URL resolution

**Files:**
- Modify: `server/internal/ai/provider.go:200-209`

- [ ] **Step 1: Add provider string constants at top of file**

Add after the import block:

```go
const (
    ProviderOpenAI    = "openai"
    ProviderAnthropic = "anthropic"
    ProviderDeepSeek  = "deepseek"
    ProviderMinimax   = "minimax"
    ProviderGLM       = "glm"
    ProviderQwen      = "qwen"
    ProviderCustom    = "custom"
)
```

- [ ] **Step 2: Add default URL resolution function**

Add before `NewOpenAIProvider`:

```go
// defaultAPIURL returns the official API endpoint for a given provider.
// The user-set api_url in LLMConfig overrides this default.
func defaultAPIURL(provider string) string {
    switch provider {
    case ProviderDeepSeek:
        return "https://api.deepseek.com/v1/chat/completions"
    case ProviderMinimax:
        return "https://api.minimax.chat/v1/chat/completions"
    case ProviderGLM:
        return "https://open.bigmodel.cn/api/paas/v4/chat/completions"
    case ProviderQwen:
        return "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
    default:
        return "https://api.openai.com/v1/chat/completions"
    }
}
```

- [ ] **Step 3: Update `NewOpenAIProvider` to use the helper**

Replace the URL logic in `NewOpenAIProvider` (lines 37-39):

```go
func NewOpenAIProvider(cfg *model.LLMConfig) *OpenAIProvider {
    url := defaultAPIURL(cfg.Provider)
    if cfg.APIURL != nil && *cfg.APIURL != "" {
        url = *cfg.APIURL
    }
    // rest unchanged ...
```

- [ ] **Step 4: Update `NewProvider` factory to handle all new providers**

Replace the switch statement (lines 200-209):

```go
func NewProvider(cfg *model.LLMConfig) (LLMProvider, error) {
    switch cfg.Provider {
    case ProviderOpenAI, ProviderDeepSeek, ProviderMinimax, ProviderGLM, ProviderQwen, ProviderCustom:
        return NewOpenAIProvider(cfg), nil
    case ProviderAnthropic:
        return NewAnthropicProvider(cfg), nil
    default:
        return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
    }
}
```

- [ ] **Step 5: Verify the file compiles**

Run: `cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server && go vet ./internal/ai/...`
Expected: no errors

- [ ] **Step 6: Commit**

```bash
cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface
git add server/internal/ai/provider.go
git commit -m "feat(ai): add DeepSeek, MiniMax, GLM, Qwen provider constants with default API URLs"
```

---

### Task 2: Update frontend LLM config page with new providers

**Files:**
- Modify: `frontend/app/admin/llm-config/page.tsx:76-86`

- [ ] **Step 1: Add model placeholder map and set default when provider changes**

Add this function before the component (after the `LLMConfig` interface):

```typescript
const MODEL_PLACEHOLDERS: Record<string, string> = {
  openai: "gpt-4o",
  deepseek: "deepseek-chat",
  minimax: "minimax-text-01",
  glm: "glm-4-plus",
  qwen: "qwen-max",
  custom: "your-model-name",
};

```

- [ ] **Step 2: Update dropdown options and add placeholder logic**

Replace the `<select>` block (lines 76-86):

```typescript
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Provider</label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent bg-white"
                value={form.provider}
                onChange={(e) => {
                  const p = e.target.value;
                  setForm({ ...form, provider: p, model: "" });
                }}
              >
                <option value="openai">OpenAI</option>
                <option value="anthropic">Anthropic</option>
                <option value="deepseek">DeepSeek</option>
                <option value="minimax">MiniMax</option>
                <option value="glm">GLM (智谱)</option>
                <option value="qwen">Qwen (通义千问)</option>
                <option value="custom">Custom</option>
              </select>
```

Replace the model `<input>` placeholder (line 94):

```typescript
                placeholder={MODEL_PLACEHOLDERS[form.provider] || "model-name"}
```

- [ ] **Step 3: Run frontend type check**

Run: `cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface
git add frontend/app/admin/llm-config/page.tsx
git commit -m "feat(admin): add DeepSeek, MiniMax, GLM, Qwen to LLM config dropdown"
```

---

### Task 3: Add provider factory tests

**Files:**
- Modify: `server/internal/ai/provider_test.go`

- [ ] **Step 1: Add test for all provider factory routing**

Append to `provider_test.go`:

```go
func TestNewProvider_SupportedProviders(t *testing.T) {
    tests := []struct {
        name        string
        provider    string
        wantType    string // "*ai.OpenAIProvider" or "*ai.AnthropicProvider"
        wantDefault string // expected default URL prefix
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
            // Check default URL is set
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
```

- [ ] **Step 2: Add the `strings` import at the top of the file**

```go
import (
    "fmt"
    "strings"
    "testing"
)
```

- [ ] **Step 3: Run tests to verify they pass**

Run: `cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server && go test ./internal/ai/... -v`
Expected: All tests PASS including the new ones

- [ ] **Step 4: Commit**

```bash
cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface
git add server/internal/ai/provider_test.go
git commit -m "test(ai): add factory tests for DeepSeek, MiniMax, GLM, Qwen providers"
```
