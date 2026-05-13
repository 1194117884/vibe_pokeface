package api

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/yongkl/vibe-pokeface/internal/api/middleware"
	"github.com/yongkl/vibe-pokeface/internal/auth"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type RoomHandler struct {
	store *model.GameStore
}

type CreateRoomRequest struct {
	Name       string `json:"name"`
	GameType   string `json:"game_type"`
	MaxPlayers int8   `json:"max_players"`
	IsOpen     bool   `json:"is_open"`
	Password   string `json:"password,omitempty"`
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

type RoomListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	GameType    string `json:"game_type"`
	Status      string `json:"status"`
	MaxPlayers  int8   `json:"max_players"`
	PlayerCount int    `json:"player_count"`
	IsOpen      bool   `json:"is_open"`
	HasPassword bool   `json:"has_password"`
	OwnerID     int64  `json:"owner_id"`
}

func NewRoomHandler(store *model.GameStore) *RoomHandler {
	return &RoomHandler{store: store}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.GameType == "" {
		req.GameType = "doudizhu"
	}
	if req.MaxPlayers < 2 || req.MaxPlayers > 4 {
		req.MaxPlayers = 3
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Generate a short room ID
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	roomID := string(b)

	room := &model.Room{
		ID:         roomID,
		Name:       req.Name,
		GameType:   req.GameType,
		OwnerID:    claims.UserID,
		Status:     "waiting",
		MaxPlayers: req.MaxPlayers,
		IsOpen:     req.IsOpen,
		BotEnabled: true,
	}
	if req.Password != "" {
		room.Password = &req.Password
	}

	if err := h.store.CreateRoom(r.Context(), room); err != nil {
		http.Error(w, `{"error":"failed to create room"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateRoomResponse{RoomID: roomID})
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.store.ListActiveRooms(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list rooms"}`, http.StatusInternalServerError)
		return
	}
	list := make([]RoomListItem, 0, len(rooms))
	for _, rm := range rooms {
		hasPw := rm.Password != nil && *rm.Password != ""
		list = append(list, RoomListItem{
			ID:          rm.ID,
			Name:        rm.Name,
			GameType:    rm.GameType,
			Status:      rm.Status,
			MaxPlayers:  rm.MaxPlayers,
			PlayerCount: 0,
			IsOpen:      rm.IsOpen,
			HasPassword: hasPw,
			OwnerID:     rm.OwnerID,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
