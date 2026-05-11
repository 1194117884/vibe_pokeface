package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type testUserStore struct {
	getUserCountFn func(ctx context.Context) (int, error)
}

func (m *testUserStore) GetUserCount(ctx context.Context) (int, error) {
	if m.getUserCountFn != nil {
		return m.getUserCountFn(ctx)
	}
	return 0, nil
}
func (m *testUserStore) ListUsers(ctx context.Context, offset, limit int) ([]model.User, error) { return nil, nil }
func (m *testUserStore) SearchUsers(ctx context.Context, query string) ([]model.User, error)    { return nil, nil }
func (m *testUserStore) UpdateUserStatus(ctx context.Context, userID int64, status int8) error  { return nil }
func (m *testUserStore) FindByID(ctx context.Context, id int64) (*model.User, error)            { return nil, nil }

type testGameStore struct {
	listActiveRoomsFn func(ctx context.Context) ([]model.Room, error)
}

func (m *testGameStore) ListActiveRooms(ctx context.Context) ([]model.Room, error) {
	if m.listActiveRoomsFn != nil {
		return m.listActiveRoomsFn(ctx)
	}
	return nil, nil
}
func (m *testGameStore) GetRoom(ctx context.Context, roomID string) (*model.Room, error)             { return nil, nil }
func (m *testGameStore) GetRoomPlayers(ctx context.Context, roomID string) ([]model.RoomPlayer, error) { return nil, nil }
func (m *testGameStore) GetUserBalance(ctx context.Context, userID int64) (int, error)               { return 0, nil }
func (m *testGameStore) GetScoreHistory(ctx context.Context, userID int64, limit int) ([]model.ScoreRecord, error) {
	return nil, nil
}
func (m *testGameStore) SaveScore(ctx context.Context, userID int64, gameType string, amount, balance int, reason string) error {
	return nil
}

type testHub struct {
	onlineCountFn func() int
}

func (m *testHub) GetOnlineCount() int {
	if m.onlineCountFn != nil {
		return m.onlineCountFn()
	}
	return 0
}

func TestDashboardHandler_Success(t *testing.T) {
	userStore := &testUserStore{getUserCountFn: func(ctx context.Context) (int, error) { return 42, nil }}
	gameStore := &testGameStore{listActiveRoomsFn: func(ctx context.Context) ([]model.Room, error) {
		return []model.Room{{ID: "r1"}, {ID: "r2"}}, nil
	}}
	hub := &testHub{onlineCountFn: func() int { return 5 }}

	handler := NewDashboardHandler(userStore, gameStore, hub)

	req := httptest.NewRequest("GET", "/api/admin/dashboard", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		OnlinePlayers int `json:"online_players"`
		ActiveRooms   int `json:"active_rooms"`
		TotalUsers    int `json:"total_users"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.OnlinePlayers != 5 {
		t.Errorf("online_players = %d, want 5", resp.OnlinePlayers)
	}
	if resp.ActiveRooms != 2 {
		t.Errorf("active_rooms = %d, want 2", resp.ActiveRooms)
	}
	if resp.TotalUsers != 42 {
		t.Errorf("total_users = %d, want 42", resp.TotalUsers)
	}
}

func TestDashboardHandler_ErrorsGracefully(t *testing.T) {
	// When gameStore returns an error, handler should return 0, not crash
	userStore := &testUserStore{getUserCountFn: func(ctx context.Context) (int, error) { return 0, nil }}
	gameStore := &testGameStore{listActiveRoomsFn: func(ctx context.Context) ([]model.Room, error) {
		return nil, errors.New("db error")
	}}
	hub := &testHub{onlineCountFn: func() int { return 0 }}

	handler := NewDashboardHandler(userStore, gameStore, hub)

	req := httptest.NewRequest("GET", "/api/admin/dashboard", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
