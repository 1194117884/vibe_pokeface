package ai

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/yongkl/vibe-pokeface/internal/model"
)

const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderDeepSeek  = "deepseek"
	ProviderMinimax   = "minimax"
	ProviderGLM       = "glm"
	ProviderQwen      = "qwen"
	ProviderCustom    = "custom"
)

type LLMResult struct {
    Content          string
    PromptTokens     int
    CompletionTokens int
    DurationMs       int
}

type LLMProvider interface {
    Complete(ctx context.Context, systemPrompt, userPrompt string) (*LLMResult, error)
}

type OpenAIProvider struct {
    apiKey      string
    model       string
    apiURL      string
    temperature float64
    maxTokens   int
    client      *http.Client
}

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

func NewOpenAIProvider(cfg *model.LLMConfig) *OpenAIProvider {
    url := defaultAPIURL(cfg.Provider)
    if cfg.APIURL != nil && *cfg.APIURL != "" {
        url = *cfg.APIURL
    }
    return &OpenAIProvider{
        apiKey:      cfg.APIKey,
        model:       cfg.Model,
        apiURL:      url,
        temperature: cfg.Temperature,
        maxTokens:   cfg.MaxTokens,
        client:      &http.Client{Timeout: 30 * time.Second},
    }
}

func (p *OpenAIProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (*LLMResult, error) {
    body := map[string]interface{}{
        "model": p.model,
        "messages": []map[string]string{
            {"role": "system", "content": systemPrompt},
            {"role": "user", "content": userPrompt},
        },
        "temperature": p.temperature,
        "max_tokens":  p.maxTokens,
    }

    jsonBody, _ := json.Marshal(body)
    start := time.Now()

    req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewReader(jsonBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+p.apiKey)

    resp, err := p.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("LLM request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
    }

    respBytes, _ := io.ReadAll(resp.Body)
    duration := int(time.Since(start).Milliseconds())

    var result struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
        Usage struct {
            PromptTokens     int `json:"prompt_tokens"`
            CompletionTokens int `json:"completion_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal(respBytes, &result); err != nil {
        return nil, fmt.Errorf("LLM response parse error: %w", err)
    }

    if len(result.Choices) == 0 {
        return nil, fmt.Errorf("LLM returned no choices")
    }

    return &LLMResult{
        Content:          result.Choices[0].Message.Content,
        PromptTokens:     result.Usage.PromptTokens,
        CompletionTokens: result.Usage.CompletionTokens,
        DurationMs:       duration,
    }, nil
}

type AnthropicProvider struct {
    apiKey      string
    model       string
    apiURL      string
    temperature float64
    maxTokens   int
    client      *http.Client
}

func NewAnthropicProvider(cfg *model.LLMConfig) *AnthropicProvider {
    url := "https://api.anthropic.com/v1/messages"
    if cfg.APIURL != nil && *cfg.APIURL != "" {
        url = *cfg.APIURL
    }
    return &AnthropicProvider{
        apiKey:      cfg.APIKey,
        model:       cfg.Model,
        apiURL:      url,
        temperature: cfg.Temperature,
        maxTokens:   cfg.MaxTokens,
        client:      &http.Client{Timeout: 30 * time.Second},
    }
}

func (p *AnthropicProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (*LLMResult, error) {
    body := map[string]interface{}{
        "model":      p.model,
        "max_tokens": p.maxTokens,
        "system":     systemPrompt,
        "messages": []map[string]string{
            {"role": "user", "content": userPrompt},
        },
        "temperature": p.temperature,
    }

    jsonBody, _ := json.Marshal(body)
    start := time.Now()

    req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewReader(jsonBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", p.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")

    resp, err := p.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Anthropic request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
    }

    respBytes, _ := io.ReadAll(resp.Body)
    duration := int(time.Since(start).Milliseconds())

    var result struct {
        Content []struct {
            Text string `json:"text"`
        } `json:"content"`
        Usage struct {
            InputTokens  int `json:"input_tokens"`
            OutputTokens int `json:"output_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal(respBytes, &result); err != nil {
        return nil, fmt.Errorf("Anthropic response parse error: %w", err)
    }

    content := ""
    if len(result.Content) > 0 {
        content = result.Content[0].Text
    }

    return &LLMResult{
        Content:          content,
        PromptTokens:     result.Usage.InputTokens,
        CompletionTokens: result.Usage.OutputTokens,
        DurationMs:       duration,
    }, nil
}

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
