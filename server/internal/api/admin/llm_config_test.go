package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockLLMConfigStore struct {
	listConfigsFn  func(ctx context.Context) ([]model.LLMConfig, error)
	saveConfigFn   func(ctx context.Context, cfg *model.LLMConfig) error
	deleteConfigFn func(ctx context.Context, id int) error
}

func (m *mockLLMConfigStore) ListConfigs(ctx context.Context) ([]model.LLMConfig, error) {
	return m.listConfigsFn(ctx)
}
func (m *mockLLMConfigStore) SaveConfig(ctx context.Context, cfg *model.LLMConfig) error {
	return m.saveConfigFn(ctx, cfg)
}
func (m *mockLLMConfigStore) DeleteConfig(ctx context.Context, id int) error {
	return m.deleteConfigFn(ctx, id)
}

type mockLLMStatsStore struct {
	getLLMStatsFn func(ctx context.Context) (*model.LLMStatsSummary, error)
}

func (m *mockLLMStatsStore) GetLLMStats(ctx context.Context) (*model.LLMStatsSummary, error) {
	return m.getLLMStatsFn(ctx)
}

func TestLLMConfigList(t *testing.T) {
	store := &mockLLMConfigStore{
		listConfigsFn: func(ctx context.Context) ([]model.LLMConfig, error) {
			return []model.LLMConfig{
				{ID: 1, Provider: "openai", Model: "gpt-4o", APIKey: "sk-abcdefghijklmnop", IsActive: true},
			}, nil
		},
	}
	handler := NewLLMConfigHandler(store, &mockLLMStatsStore{})

	req := httptest.NewRequest("GET", "/api/admin/llm-configs", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var configs []model.LLMConfig
	json.NewDecoder(w.Body).Decode(&configs)
	if len(configs) != 1 || configs[0].Provider != "openai" {
		t.Errorf("unexpected: %+v", configs)
	}
	// Verify API key is masked
	if len(configs[0].APIKey) > 0 && configs[0].APIKey == "sk-abcdefghijklmnop" {
		t.Errorf("API key was not masked: %s", configs[0].APIKey)
	}
}

func TestLLMConfigList_Empty(t *testing.T) {
	store := &mockLLMConfigStore{
		listConfigsFn: func(ctx context.Context) ([]model.LLMConfig, error) { return []model.LLMConfig{}, nil },
	}
	handler := NewLLMConfigHandler(store, &mockLLMStatsStore{})

	req := httptest.NewRequest("GET", "/api/admin/llm-configs", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestLLMConfigCreate_Success(t *testing.T) {
	store := &mockLLMConfigStore{
		saveConfigFn: func(ctx context.Context, cfg *model.LLMConfig) error {
			cfg.ID = 1
			return nil
		},
	}
	handler := NewLLMConfigHandler(store, &mockLLMStatsStore{})

	body := map[string]interface{}{
		"provider": "anthropic", "model": "claude-sonnet-4-20250514",
		"api_key": "sk-test", "temperature": 0.7, "max_tokens": 2048, "is_active": true,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/llm-configs", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Create(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp model.LLMConfig
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Provider != "anthropic" {
		t.Errorf("provider = %s, want anthropic", resp.Provider)
	}
	if resp.APIKey != "" {
		t.Errorf("API key should be empty in response, got %s", resp.APIKey)
	}
}

func TestLLMConfigCreate_MissingFields(t *testing.T) {
	handler := NewLLMConfigHandler(&mockLLMConfigStore{}, &mockLLMStatsStore{})

	body := map[string]interface{}{"provider": "openai"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/llm-configs", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLLMConfigGetStats(t *testing.T) {
	statsStore := &mockLLMStatsStore{
		getLLMStatsFn: func(ctx context.Context) (*model.LLMStatsSummary, error) {
			return &model.LLMStatsSummary{
				TotalCalls: 100, TotalTokens: 50000, AvgLatencyMs: 1200, SuccessRate: 95,
			}, nil
		},
	}
	handler := NewLLMConfigHandler(&mockLLMConfigStore{}, statsStore)

	req := httptest.NewRequest("GET", "/api/admin/llm-stats", nil)
	w := httptest.NewRecorder()
	handler.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var stats model.LLMStatsSummary
	json.NewDecoder(w.Body).Decode(&stats)
	if stats.TotalCalls != 100 {
		t.Errorf("total_calls = %d, want 100", stats.TotalCalls)
	}
}
