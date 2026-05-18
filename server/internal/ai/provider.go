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

// ChatMessage represents a message in a tool-use conversation
type ChatMessage struct {
	Role       string              `json:"role"`
	Content    string              `json:"content,omitempty"`
	ToolCalls  []AssistantToolCall `json:"tool_calls,omitempty"`
	ToolCallID string              `json:"tool_call_id,omitempty"`
	Name       string              `json:"name,omitempty"`
}

// AssistantToolCall represents a tool call from the LLM
type AssistantToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// LLMResultWithTools extends LLMResult with tool calls
type LLMResultWithTools struct {
	LLMResult
	ToolCalls []AssistantToolCall
}

type LLMProvider interface {
	Complete(ctx context.Context, systemPrompt, userPrompt string) (*LLMResult, error)
	CompleteWithTools(ctx context.Context, messages []ChatMessage, tools []ToolSchema) (*LLMResultWithTools, error)
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
		"stream":      false,
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

func (p *OpenAIProvider) CompleteWithTools(ctx context.Context, messages []ChatMessage, tools []ToolSchema) (*LLMResultWithTools, error) {
	body := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"tools":       tools,
		"temperature": p.temperature,
		"max_tokens":  p.maxTokens,
		"stream":      false,
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
				Content   string              `json:"content"`
				ToolCalls []AssistantToolCall `json:"tool_calls"`
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

	return &LLMResultWithTools{
		LLMResult: LLMResult{
			Content:          result.Choices[0].Message.Content,
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			DurationMs:       duration,
		},
		ToolCalls: result.Choices[0].Message.ToolCalls,
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

func (p *AnthropicProvider) CompleteWithTools(ctx context.Context, messages []ChatMessage, tools []ToolSchema) (*LLMResultWithTools, error) {
	// Convert tools to Anthropic format
	anthropicTools := make([]map[string]interface{}, len(tools))
	for i, t := range tools {
		anthropicTools[i] = map[string]interface{}{
			"name":         t.Function.Name,
			"description":  t.Function.Description,
			"input_schema": t.Function.Parameters,
		}
	}

	// Convert ChatMessage to Anthropic messages format
	var systemPrompt string
	anthropicMsgs := make([]map[string]interface{}, 0)
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			if msg.Content != "" {
				anthropicMsgs = append(anthropicMsgs, map[string]interface{}{
					"role":    "user",
					"content": msg.Content,
				})
			}
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				content := make([]map[string]interface{}, 0)
				if msg.Content != "" {
					content = append(content, map[string]interface{}{
						"type": "text",
						"text": msg.Content,
					})
				}
				for _, tc := range msg.ToolCalls {
					var input map[string]interface{}
					json.Unmarshal([]byte(tc.Function.Arguments), &input)
					content = append(content, map[string]interface{}{
						"type":  "tool_use",
						"id":    tc.ID,
						"name":  tc.Function.Name,
						"input": input,
					})
				}
				anthropicMsgs = append(anthropicMsgs, map[string]interface{}{
					"role":    "assistant",
					"content": content,
				})
			} else if msg.Content != "" {
				anthropicMsgs = append(anthropicMsgs, map[string]interface{}{
					"role":    "assistant",
					"content": msg.Content,
				})
			}
		case "tool":
			anthropicMsgs = append(anthropicMsgs, map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type":        "tool_result",
						"tool_use_id": msg.ToolCallID,
						"content":     msg.Content,
					},
				},
			})
		}
	}

	body := map[string]interface{}{
		"model":       p.model,
		"max_tokens":  p.maxTokens,
		"messages":    anthropicMsgs,
		"tools":       anthropicTools,
		"temperature": p.temperature,
	}
	if systemPrompt != "" {
		body["system"] = systemPrompt
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
			Type  string          `json:"type"`
			Text  string          `json:"text"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("Anthropic response parse error: %w", err)
	}

	textContent := ""
	toolCalls := make([]AssistantToolCall, 0)
	for _, c := range result.Content {
		if c.Type == "text" {
			textContent += c.Text
		}
		if c.Type == "tool_use" {
			args, _ := json.Marshal(c.Input)
			tc := AssistantToolCall{}
			tc.ID = c.ID
			tc.Type = "function"
			tc.Function.Name = c.Name
			tc.Function.Arguments = string(args)
			toolCalls = append(toolCalls, tc)
		}
	}

	return &LLMResultWithTools{
		LLMResult: LLMResult{
			Content:          textContent,
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			DurationMs:       duration,
		},
		ToolCalls: toolCalls,
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
