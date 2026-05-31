package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/api/middleware"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

// RegistrationCodeAdminStore is the subset of RegistrationCodeStore used by admin handlers.
type RegistrationCodeAdminStore interface {
	Create(ctx context.Context, codes []*model.RegistrationCode) error
	List(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error)
	Disable(ctx context.Context, id int64) error
	CountByStatus(ctx context.Context, used bool) (int, error)
}

type RegistrationCodesHandler struct {
	store RegistrationCodeAdminStore
}

func NewRegistrationCodesHandler(store RegistrationCodeAdminStore) *RegistrationCodesHandler {
	return &RegistrationCodesHandler{store: store}
}

type generateRequest struct {
	Count int    `json:"count"`
	Note  string `json:"note"`
}

type codesListResponse struct {
	Codes []model.RegistrationCodeWithUser `json:"codes"`
	Total int                              `json:"total"`
}

// List returns paginated registration codes.
func (h *RegistrationCodesHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}

	var usedFilter *bool
	if f := r.URL.Query().Get("used"); f != "" {
		v := f == "true"
		usedFilter = &v
	}

	offset := (page - 1) * size

	codes, total, err := h.store.List(r.Context(), offset, size, usedFilter)
	if err != nil {
		http.Error(w, `{"error":"list failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(codesListResponse{Codes: codes, Total: total})
}

// Generate creates new registration codes.
func (h *RegistrationCodesHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Count < 1 || req.Count > 100 {
		http.Error(w, `{"error":"count must be between 1 and 100"}`, http.StatusBadRequest)
		return
	}

	claims := middleware.GetClaims(r.Context())

	codes := make([]*model.RegistrationCode, req.Count)
	for i := range codes {
		codes[i] = &model.RegistrationCode{
			CreatedBy: claims.UserID,
			Note:      req.Note,
		}
	}

	if err := h.store.Create(r.Context(), codes); err != nil {
		http.Error(w, `{"error":"failed to generate codes"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"codes": codes,
		"count": len(codes),
	})
}

// Disable disables an unused registration code.
func (h *RegistrationCodesHandler) Disable(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid code id"}`, http.StatusBadRequest)
		return
	}

	if err := h.store.Disable(r.Context(), id); err != nil {
		http.Error(w, `{"error":"disable failed — code may already be used or disabled"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
