package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type AdminUserStore interface {
	ListUsers(ctx context.Context, offset, limit int) ([]model.User, error)
	SearchUsers(ctx context.Context, query string) ([]model.User, error)
	UpdateUserStatus(ctx context.Context, userID int64, status int8) error
	GetUserCount(ctx context.Context) (int, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
}

type AdminUserHandler struct {
	store AdminUserStore
}

func NewAdminUserHandler(store AdminUserStore) *AdminUserHandler {
	return &AdminUserHandler{store: store}
}

type userListResponse struct {
	Users []model.User `json:"users"`
	Total int          `json:"total"`
}

type updateStatusRequest struct {
	Status int8 `json:"status"`
}

func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}

	ctx := r.Context()

	if q != "" {
		users, err := h.store.SearchUsers(ctx, q)
		if err != nil {
			http.Error(w, `{"error":"search failed"}`, http.StatusInternalServerError)
			return
		}
		if users == nil {
			users = []model.User{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userListResponse{Users: users, Total: len(users)})
		return
	}

	offset := (page - 1) * size
	users, err := h.store.ListUsers(ctx, offset, size)
	if err != nil {
		http.Error(w, `{"error":"list failed"}`, http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []model.User{}
	}

	total, _ := h.store.GetUserCount(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userListResponse{Users: users, Total: total})
}

func (h *AdminUserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateUserStatus(r.Context(), userID, req.Status); err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
