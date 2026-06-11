package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type AIRunStore interface {
	ListAIRuns(ctx context.Context, filter model.AIRunFilter) ([]model.AIRunListItem, error)
	GetAIRunDetail(ctx context.Context, id int64) (*model.AIRunDetail, error)
}

type AIRunHandler struct {
	store AIRunStore
}

func NewAIRunHandler(store AIRunStore) *AIRunHandler {
	return &AIRunHandler{store: store}
}

func (h *AIRunHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.AIRunFilter{
		RoomID: q.Get("room_id"),
		UserID: q.Get("user_id"),
		Game:   q.Get("game_type"),
		Phase:  q.Get("phase"),
	}
	if raw := q.Get("success"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			http.Error(w, `{"error":"invalid success filter"}`, http.StatusBadRequest)
			return
		}
		filter.Success = &value
	}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = offset
	}
	runs, err := h.store.ListAIRuns(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list ai runs"}`, http.StatusInternalServerError)
		return
	}
	if runs == nil {
		runs = []model.AIRunListItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func (h *AIRunHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	detail, err := h.store.GetAIRunDetail(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"ai run not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}
