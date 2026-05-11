package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type AdminRoomStore interface {
	ListActiveRooms(ctx context.Context) ([]model.Room, error)
	GetRoom(ctx context.Context, roomID string) (*model.Room, error)
	GetRoomPlayers(ctx context.Context, roomID string) ([]model.RoomPlayer, error)
}

type AdminRoomHandler struct {
	store AdminRoomStore
}

func NewAdminRoomHandler(store AdminRoomStore) *AdminRoomHandler {
	return &AdminRoomHandler{store: store}
}

func (h *AdminRoomHandler) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.store.ListActiveRooms(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list rooms"}`, http.StatusInternalServerError)
		return
	}
	if rooms == nil {
		rooms = []model.Room{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

type roomDetailResponse struct {
	Room    *model.Room        `json:"room"`
	Players []model.RoomPlayer `json:"players"`
}

func (h *AdminRoomHandler) Get(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	if roomID == "" {
		http.Error(w, `{"error":"room id required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	room, err := h.store.GetRoom(ctx, roomID)
	if err != nil {
		http.Error(w, `{"error":"room not found"}`, http.StatusNotFound)
		return
	}

	players, err := h.store.GetRoomPlayers(ctx, roomID)
	if err != nil {
		players = []model.RoomPlayer{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomDetailResponse{Room: room, Players: players})
}
