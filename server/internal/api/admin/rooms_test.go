package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockRoomAdminStore struct {
	listActiveRoomsFn func(ctx context.Context) ([]model.Room, error)
	getRoomFn         func(ctx context.Context, roomID string) (*model.Room, error)
	getRoomPlayersFn  func(ctx context.Context, roomID string) ([]model.RoomPlayer, error)
}

func (m *mockRoomAdminStore) ListActiveRooms(ctx context.Context) ([]model.Room, error) {
	return m.listActiveRoomsFn(ctx)
}
func (m *mockRoomAdminStore) GetRoom(ctx context.Context, roomID string) (*model.Room, error) {
	return m.getRoomFn(ctx, roomID)
}
func (m *mockRoomAdminStore) GetRoomPlayers(ctx context.Context, roomID string) ([]model.RoomPlayer, error) {
	return m.getRoomPlayersFn(ctx, roomID)
}

func TestRoomList(t *testing.T) {
	now := time.Now()
	store := &mockRoomAdminStore{
		listActiveRoomsFn: func(ctx context.Context) ([]model.Room, error) {
			return []model.Room{
				{ID: "room1", GameType: "doudizhu", OwnerID: 1, Status: "playing", MaxPlayers: 3, CreatedAt: now},
				{ID: "room2", GameType: "doudizhu", OwnerID: 2, Status: "waiting", MaxPlayers: 3, CreatedAt: now},
			}, nil
		},
	}
	handler := NewAdminRoomHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/rooms", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp []model.Room
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("room count = %d, want 2", len(resp))
	}
}

func TestRoomList_Empty(t *testing.T) {
	store := &mockRoomAdminStore{
		listActiveRoomsFn: func(ctx context.Context) ([]model.Room, error) { return []model.Room{}, nil },
	}
	handler := NewAdminRoomHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/rooms", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var resp []model.Room
	json.NewDecoder(w.Body).Decode(&resp)
	if resp == nil || len(resp) != 0 {
		t.Errorf("expected empty array, got %+v", resp)
	}
}

func TestRoomGet_Success(t *testing.T) {
	store := &mockRoomAdminStore{
		getRoomFn: func(ctx context.Context, roomID string) (*model.Room, error) {
			return &model.Room{ID: "room1", GameType: "doudizhu", OwnerID: 1, Status: "playing"}, nil
		},
		getRoomPlayersFn: func(ctx context.Context, roomID string) ([]model.RoomPlayer, error) {
			return []model.RoomPlayer{
				{RoomID: "room1", SeatIndex: 0, Status: "playing"},
				{RoomID: "room1", SeatIndex: 1, Status: "playing"},
			}, nil
		},
	}
	handler := NewAdminRoomHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/rooms/room1", nil)
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "room1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	w := httptest.NewRecorder()
	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp roomDetailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Room.ID != "room1" {
		t.Errorf("room id = %s, want room1", resp.Room.ID)
	}
	if len(resp.Players) != 2 {
		t.Errorf("players = %d, want 2", len(resp.Players))
	}
}

func TestRoomGet_NotFound(t *testing.T) {
	store := &mockRoomAdminStore{
		getRoomFn: func(ctx context.Context, roomID string) (*model.Room, error) {
			return nil, errors.New("not found")
		},
	}
	handler := NewAdminRoomHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/rooms/nonexistent", nil)
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	w := httptest.NewRecorder()
	handler.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
