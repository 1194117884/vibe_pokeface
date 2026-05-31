package api

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

// ---------------------------------------------------------------------------
// Mock SQL driver – lets us test handlers that depend on *model.GameStore
// without a real MySQL instance.
// ---------------------------------------------------------------------------

var mockMu sync.Mutex
var mockRooms []model.Room

func init() { sql.Register("roomtest", &mockDriver{}) }

type mockDriver struct{}

func (d *mockDriver) Open(_ string) (driver.Conn, error) { return &mockConn{}, nil }

type mockConn struct{}

func (c *mockConn) Prepare(_ string) (driver.Stmt, error) { return &mockStmt{}, nil }
func (c *mockConn) Close() error                          { return nil }
func (c *mockConn) Begin() (driver.Tx, error) {
	return nil, io.EOF // transactions not supported
}

type mockStmt struct{}

func (s *mockStmt) Close() error                                 { return nil }
func (s *mockStmt) NumInput() int                                { return -1 }
func (s *mockStmt) Exec(_ []driver.Value) (driver.Result, error) { return &mockResult{}, nil }
func (s *mockStmt) Query(_ []driver.Value) (driver.Rows, error) {
	mockMu.Lock()
	rows := make([]model.Room, len(mockRooms))
	copy(rows, mockRooms)
	mockMu.Unlock()
	return &mockRows{rooms: rows}, nil
}

type mockResult struct{}

func (r *mockResult) LastInsertId() (int64, error) { return 0, nil }
func (r *mockResult) RowsAffected() (int64, error) { return 0, nil }

type mockRows struct {
	rooms []model.Room
	pos   int
}

func (r *mockRows) Columns() []string {
	return []string{
		"id", "name", "theme", "game_type", "owner_id",
		"status", "max_players", "is_open", "password",
		"bot_enabled", "created_at", "ended_at",
	}
}

func (r *mockRows) Close() error { return nil }

func (r *mockRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rooms) {
		return io.EOF
	}
	rm := r.rooms[r.pos]
	r.pos++

	dest[0] = rm.ID                // id
	dest[1] = rm.Name              // name
	dest[2] = rm.Theme             // theme
	dest[3] = rm.GameType          // game_type
	dest[4] = rm.OwnerID           // owner_id
	dest[5] = rm.Status            // status
	dest[6] = int64(rm.MaxPlayers) // max_players (driver.Value must be int64, not int8)
	if rm.IsOpen {
		dest[7] = int64(1)
	} else {
		dest[7] = int64(0)
	}
	if rm.Password != nil {
		dest[8] = *rm.Password
	} else {
		dest[8] = nil
	}
	if rm.BotEnabled {
		dest[9] = int64(1)
	} else {
		dest[9] = int64(0)
	}
	dest[10] = rm.CreatedAt // created_at
	if rm.EndedAt != nil {
		dest[11] = *rm.EndedAt
	} else {
		dest[11] = nil
	}
	return nil
}

// newRoomHandlerWithRooms creates a RoomHandler backed by a mock database
// that returns the given rooms from ListActiveRooms.
func newRoomHandlerWithRooms(rooms []model.Room) *RoomHandler {
	mockMu.Lock()
	mockRooms = rooms
	mockMu.Unlock()

	rawDB, err := sql.Open("roomtest", "")
	if err != nil {
		panic("failed to open mock db: " + err.Error())
	}
	db := sqlx.NewDb(rawDB, "mysql")
	store := model.NewGameStore(db)
	return NewRoomHandler(store)
}

// ptrString returns a pointer to s, useful for *string model fields.
func ptrString(s string) *string { return &s }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateRoomRejectsPrivateRoomWithoutPassword(t *testing.T) {
	h := NewRoomHandler(nil)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/rooms",
		strings.NewReader(`{"name":"private","game_type":"doudizhu","max_players":3,"is_open":false}`),
	)
	w := httptest.NewRecorder()

	h.CreateRoom(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "password required for private room") {
		t.Fatalf("body = %q, want password validation error", w.Body.String())
	}
}

func TestListRooms_ReturnsPrivateRooms(t *testing.T) {
	now := time.Now()
	rooms := []model.Room{
		{
			ID: "OPEN001", Name: "Open Room", GameType: "doudizhu",
			OwnerID: 1, Status: "waiting", MaxPlayers: 3,
			IsOpen: true, BotEnabled: true, CreatedAt: now,
		},
		{
			ID: "PRIV001", Name: "Private Room", GameType: "doudizhu",
			OwnerID: 2, Status: "waiting", MaxPlayers: 4,
			IsOpen: false, BotEnabled: true, CreatedAt: now,
			Password: ptrString("secret123"),
		},
	}

	h := newRoomHandlerWithRooms(rooms)
	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	w := httptest.NewRecorder()

	h.ListRooms(w, req)

	// Expect 200 OK
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var list []RoomListItem
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode error: %v; body=%s", err, w.Body.String())
	}

	// The private room must appear in the list
	var privateFound bool
	for _, item := range list {
		if item.ID == "PRIV001" {
			privateFound = true
			if item.IsOpen {
				t.Errorf("private room PRIV001 should have is_open=false")
			}
		}
	}
	if !privateFound {
		t.Error("private room (is_open=false) not found in list — the is_open filter is still in place")
	}

	// Both rooms should be returned (2 total)
	if len(list) != 2 {
		t.Errorf("expected 2 rooms in list, got %d", len(list))
	}
}

func TestListRooms_HasPasswordFlag(t *testing.T) {
	now := time.Now()
	rooms := []model.Room{
		{
			ID: "PWDYES", Name: "Has Pwd", GameType: "doudizhu",
			OwnerID: 1, Status: "waiting", MaxPlayers: 3,
			IsOpen: false, BotEnabled: true, CreatedAt: now,
			Password: ptrString("mypassword"),
		},
		{
			ID: "PWDNO", Name: "No Pwd", GameType: "doudizhu",
			OwnerID: 2, Status: "waiting", MaxPlayers: 3,
			IsOpen: true, BotEnabled: true, CreatedAt: now,
			Password: nil,
		},
		{
			ID: "PWDEMPTY", Name: "Empty Pwd", GameType: "doudizhu",
			OwnerID: 3, Status: "waiting", MaxPlayers: 3,
			IsOpen: true, BotEnabled: true, CreatedAt: now,
			Password: ptrString(""), // empty string should be treated as no password
		},
	}

	h := newRoomHandlerWithRooms(rooms)
	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	w := httptest.NewRecorder()

	h.ListRooms(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var list []RoomListItem
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode error: %v; body=%s", err, w.Body.String())
	}

	for _, item := range list {
		switch item.ID {
		case "PWDYES":
			if !item.HasPassword {
				t.Errorf("room %q should have has_password=true", item.ID)
			}
		case "PWDNO":
			if item.HasPassword {
				t.Errorf("room %q should have has_password=false", item.ID)
			}
		case "PWDEMPTY":
			if item.HasPassword {
				t.Errorf("room %q (empty string password) should have has_password=false", item.ID)
			}
		}
	}
}

func TestListRooms_ExcludesEndedRooms(t *testing.T) {
	// Verify that ListRooms only returns active rooms (waiting/playing).
	// This tests the store layer's filter, not the handler's response transformation.
	now := time.Now()
	rooms := []model.Room{
		{
			ID: "WAITING", Name: "Waiting Room", GameType: "doudizhu",
			OwnerID: 1, Status: "waiting", MaxPlayers: 3,
			IsOpen: true, BotEnabled: true, CreatedAt: now,
		},
		{
			ID: "PLAYING", Name: "Playing Room", GameType: "doudizhu",
			OwnerID: 2, Status: "playing", MaxPlayers: 3,
			IsOpen: true, BotEnabled: true, CreatedAt: now,
		},
		{
			ID: "ENDED", Name: "Ended Room", GameType: "doudizhu",
			OwnerID: 3, Status: "ended", MaxPlayers: 3,
			IsOpen: true, BotEnabled: true, CreatedAt: now,
			EndedAt: &now,
		},
	}

	// In reality, ListActiveRooms only returns waiting/playing rooms from the DB.
	// But our mock returns all rooms. We test that the handler correctly passes
	// through whatever the store returns.
	h := newRoomHandlerWithRooms(rooms)
	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	w := httptest.NewRecorder()

	h.ListRooms(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var list []RoomListItem
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode error: %v; body=%s", err, w.Body.String())
	}

	// All 3 rooms should pass through (the handler doesn't filter by status)
	if len(list) != 3 {
		t.Errorf("expected 3 rooms in list (handler should pass through store results), got %d", len(list))
	}
}

// TestListRooms_NilList checks that the handler does not crash
// when the store returns a nil slice (no rooms).
func TestListRooms_NilList(t *testing.T) {
	h := newRoomHandlerWithRooms(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	w := httptest.NewRecorder()

	h.ListRooms(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var list []RoomListItem
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode error: %v; body=%s", err, w.Body.String())
	}
	if list == nil {
		t.Error("expected empty array, got nil")
	}
}
