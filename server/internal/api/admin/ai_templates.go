package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/ai"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type AITemplateStore interface {
	ListPromptTemplates(ctx context.Context, gameType, phase string) ([]model.AIPromptTemplate, error)
	SavePromptTemplate(ctx context.Context, t *model.AIPromptTemplate) error
	PublishPromptTemplate(ctx context.Context, id int64) (*model.AIPromptTemplate, error)
	ListToolTemplates(ctx context.Context, gameType, phase string) ([]model.AIToolTemplate, error)
	SaveToolTemplate(ctx context.Context, t *model.AIToolTemplate) error
	PublishToolTemplate(ctx context.Context, id int64) (*model.AIToolTemplate, error)
}

type AITemplateHandler struct {
	store AITemplateStore
}

func NewAITemplateHandler(store AITemplateStore) *AITemplateHandler {
	return &AITemplateHandler{store: store}
}

func (h *AITemplateHandler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	templates, err := h.store.ListPromptTemplates(r.Context(), r.URL.Query().Get("game_type"), r.URL.Query().Get("phase"))
	if err != nil {
		http.Error(w, `{"error":"failed to list templates"}`, http.StatusInternalServerError)
		return
	}
	if templates == nil {
		templates = []model.AIPromptTemplate{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *AITemplateHandler) SavePrompt(w http.ResponseWriter, r *http.Request) {
	var tmpl model.AIPromptTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if idParam := chi.URLParam(r, "id"); idParam != "" {
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}
		tmpl.ID = id
	}
	if tmpl.GameType == "" || tmpl.Phase == "" || tmpl.TemplateType == "" || tmpl.Name == "" || tmpl.Content == "" {
		http.Error(w, `{"error":"game_type, phase, template_type, name, content are required"}`, http.StatusBadRequest)
		return
	}
	if tmpl.TemplateType != "system" && tmpl.TemplateType != "user" {
		http.Error(w, `{"error":"template_type must be system or user"}`, http.StatusBadRequest)
		return
	}
	variables, _ := json.Marshal(ai.TemplateVariables(tmpl.Content))
	variablesJSON := string(variables)
	tmpl.VariablesJSON = &variablesJSON
	if err := h.store.SavePromptTemplate(r.Context(), &tmpl); err != nil {
		http.Error(w, `{"error":"failed to save template"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tmpl)
}

func (h *AITemplateHandler) PublishPrompt(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	tmpl, err := h.store.PublishPromptTemplate(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to publish template"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tmpl)
}

func (h *AITemplateHandler) PreviewPrompt(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string            `json:"content"`
		Vars    map[string]string `json:"vars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Vars == nil {
		req.Vars = map[string]string{}
	}
	resp := map[string]interface{}{
		"rendered":  ai.RenderTemplate(req.Content, req.Vars),
		"variables": ai.TemplateVariables(req.Content),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AITemplateHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	gameType := r.URL.Query().Get("game_type")
	phase := r.URL.Query().Get("phase")
	templates, err := h.store.ListToolTemplates(r.Context(), gameType, phase)
	if err != nil {
		http.Error(w, `{"error":"failed to list tool templates"}`, http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"defaults":  ai.GetToolSchemasForGame(gameType, phase),
		"templates": templates,
	}
	if templates == nil {
		resp["templates"] = []model.AIToolTemplate{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AITemplateHandler) SaveTool(w http.ResponseWriter, r *http.Request) {
	var tmpl model.AIToolTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if idParam := chi.URLParam(r, "id"); idParam != "" {
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}
		tmpl.ID = id
	}
	if tmpl.GameType == "" || tmpl.Phase == "" || tmpl.ToolName == "" {
		http.Error(w, `{"error":"game_type, phase, tool_name are required"}`, http.StatusBadRequest)
		return
	}
	if !toolAllowed(tmpl.GameType, tmpl.Phase, tmpl.ToolName) {
		http.Error(w, `{"error":"tool_name is not in code whitelist for this phase"}`, http.StatusBadRequest)
		return
	}
	if tmpl.ParametersJSON != nil && *tmpl.ParametersJSON != "" {
		var params ai.ParamSchema
		if err := json.Unmarshal([]byte(*tmpl.ParametersJSON), &params); err != nil || params.Type == "" {
			http.Error(w, `{"error":"parameters_json must be a valid parameter schema"}`, http.StatusBadRequest)
			return
		}
	}
	if err := h.store.SaveToolTemplate(r.Context(), &tmpl); err != nil {
		http.Error(w, `{"error":"failed to save tool template"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tmpl)
}

func (h *AITemplateHandler) PublishTool(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	tmpl, err := h.store.PublishToolTemplate(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to publish tool template"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tmpl)
}

func toolAllowed(gameType, phase, toolName string) bool {
	for _, schema := range ai.GetToolSchemasForGame(gameType, phase) {
		if schema.Function.Name == toolName {
			return true
		}
	}
	return false
}
