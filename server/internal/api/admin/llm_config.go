package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type LLMConfigStore interface {
	ListConfigs(ctx context.Context) ([]model.LLMConfig, error)
	SaveConfig(ctx context.Context, cfg *model.LLMConfig) error
	DeleteConfig(ctx context.Context, id int) error
}

type LLMStatsStore interface {
	GetLLMStats(ctx context.Context) (*model.LLMStatsSummary, error)
}

type LLMConfigHandler struct {
	configStore LLMConfigStore
	statsStore  LLMStatsStore
}

func NewLLMConfigHandler(configStore LLMConfigStore, statsStore LLMStatsStore) *LLMConfigHandler {
	return &LLMConfigHandler{configStore: configStore, statsStore: statsStore}
}

func (h *LLMConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	configs, err := h.configStore.ListConfigs(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list configs"}`, http.StatusInternalServerError)
		return
	}
	if configs == nil {
		configs = []model.LLMConfig{}
	}

	// Mask API keys in response
	for i := range configs {
		if len(configs[i].APIKey) > 8 {
			configs[i].APIKey = configs[i].APIKey[:4] + "..." + configs[i].APIKey[len(configs[i].APIKey)-4:]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

func (h *LLMConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	var cfg model.LLMConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if cfg.Provider == "" || cfg.Model == "" || cfg.APIKey == "" {
		http.Error(w, `{"error":"provider, model, and api_key are required"}`, http.StatusBadRequest)
		return
	}
	if err := h.configStore.SaveConfig(r.Context(), &cfg); err != nil {
		http.Error(w, `{"error":"failed to save config"}`, http.StatusInternalServerError)
		return
	}
	cfg.APIKey = "" // Don't return the key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func (h *LLMConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.configStore.DeleteConfig(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete config"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *LLMConfigHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsStore.GetLLMStats(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get stats"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
