package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type AICharacterStore interface {
	ListCharacters(ctx context.Context) ([]model.AICharacter, error)
	GetAllCharacters(ctx context.Context) ([]model.AICharacter, error)
	GetCharacter(ctx context.Context, id int) (*model.AICharacter, error)
	CreateCharacter(ctx context.Context, c *model.AICharacter) error
	UpdateCharacter(ctx context.Context, c *model.AICharacter) error
	DeleteCharacter(ctx context.Context, id int) error
}

type AICharacterHandler struct {
	store AICharacterStore
}

func NewAICharacterHandler(store AICharacterStore) *AICharacterHandler {
	return &AICharacterHandler{store: store}
}

func (h *AICharacterHandler) List(w http.ResponseWriter, r *http.Request) {
	chars, err := h.store.GetAllCharacters(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list characters"}`, http.StatusInternalServerError)
		return
	}
	if chars == nil {
		chars = []model.AICharacter{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chars)
}

func (h *AICharacterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c model.AICharacter
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if c.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	if err := h.store.CreateCharacter(r.Context(), &c); err != nil {
		http.Error(w, `{"error":"failed to create character"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *AICharacterHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var c model.AICharacter
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	c.ID = id

	if err := h.store.UpdateCharacter(r.Context(), &c); err != nil {
		http.Error(w, `{"error":"failed to update character"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *AICharacterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteCharacter(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete character"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
