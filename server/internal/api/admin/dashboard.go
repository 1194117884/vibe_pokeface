package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type DashboardUserStore interface {
	GetUserCount(ctx context.Context) (int, error)
}

type DashboardGameStore interface {
	ListActiveRooms(ctx context.Context) ([]model.Room, error)
}

type OnlineCounter interface {
	GetOnlineCount() int
}

type DashboardResponse struct {
	OnlinePlayers int `json:"online_players"`
	ActiveRooms   int `json:"active_rooms"`
	TotalUsers    int `json:"total_users"`
}

type DashboardHandler struct {
	userStore DashboardUserStore
	gameStore DashboardGameStore
	hub       OnlineCounter
}

func NewDashboardHandler(us DashboardUserStore, gs DashboardGameStore, hub OnlineCounter) *DashboardHandler {
	return &DashboardHandler{userStore: us, gameStore: gs, hub: hub}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	onlinePlayers := h.hub.GetOnlineCount()

	rooms, err := h.gameStore.ListActiveRooms(r.Context())
	activeRooms := 0
	if err == nil {
		activeRooms = len(rooms)
	}

	totalUsers, err := h.userStore.GetUserCount(r.Context())
	if err != nil {
		totalUsers = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DashboardResponse{
		OnlinePlayers: onlinePlayers,
		ActiveRooms:   activeRooms,
		TotalUsers:    totalUsers,
	})
}
