package game

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

// mockEngine implements GameEngine for testing.
type mockEngine struct{}

func (m *mockEngine) Init(players []PlayerInfo) (GameState, error) {
	return map[string]interface{}{
		"initialized": true,
		"players":     players,
	}, nil
}

func (m *mockEngine) ExecuteAction(state GameState, action PlayerAction) (GameState, error) {
	return state, nil
}

func (m *mockEngine) ValidateAction(state GameState, action PlayerAction) error {
	return nil
}

func (m *mockEngine) IsRoundEnd(state GameState) bool {
	return false
}

func (m *mockEngine) CalculateScore(state GameState) ([]PlayerScore, error) {
	return nil, nil
}

func (m *mockEngine) SerializeForAI(state GameState) string {
	return ""
}

func (m *mockEngine) FilterForPlayer(state GameState, seat int) GameState {
	return state
}

func TestNewGameRoom(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	if room.ID != "room-1" {
		t.Errorf("ID = %s, want room-1", room.ID)
	}
	if room.GameType != "doudizhu" {
		t.Errorf("GameType = %s, want doudizhu", room.GameType)
	}
	if room.Status != "waiting" {
		t.Errorf("Status = %s, want waiting", room.Status)
	}
	if room.Engine == nil {
		t.Error("Engine should not be nil")
	}
	if len(room.Players) != 0 {
		t.Errorf("Players = %d, want 0", len(room.Players))
	}
}

func TestRoomManagerGetOrCreateRoom(t *testing.T) {
	rm := NewRoomManager(nil, nil)
	room1 := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})
	if room1 == nil {
		t.Fatal("GetOrCreateRoom returned nil")
	}
	if room1.ID != "room-1" {
		t.Errorf("ID = %s, want room-1", room1.ID)
	}

	room2 := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})
	if room2 != room1 {
		t.Error("GetOrCreateRoom should return the same instance for existing room")
	}
}

func TestRoomManagerGetRoom(t *testing.T) {
	rm := NewRoomManager(nil, nil)
	rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	room := rm.GetRoom("room-1")
	if room == nil {
		t.Fatal("GetRoom returned nil for existing room")
	}
	if room.ID != "room-1" {
		t.Errorf("ID = %s, want room-1", room.ID)
	}

	missing := rm.GetRoom("nonexistent")
	if missing != nil {
		t.Error("GetRoom should return nil for nonexistent room")
	}
}

func TestRoomManagerRemoveRoom(t *testing.T) {
	rm := NewRoomManager(nil, nil)
	rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})
	rm.RemoveRoom("room-1")

	room := rm.GetRoom("room-1")
	if room != nil {
		t.Error("GetRoom should return nil after RemoveRoom")
	}
}

func TestRoomAddPlayer(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	conn := make(chan []byte, 10)

	err := room.AddPlayer("user-1", "", "", conn)
	if err != nil {
		t.Fatalf("AddPlayer failed: %v", err)
	}
	if len(room.Players) != 1 {
		t.Errorf("Players = %d, want 1", len(room.Players))
	}
	if room.Players[0].UserID != "user-1" {
		t.Errorf("UserID = %s, want user-1", room.Players[0].UserID)
	}
	if room.Players[0].Seat < 0 || room.Players[0].Seat >= 3 {
		t.Errorf("Seat = %d, want 0-2", room.Players[0].Seat)
	}
}

func TestRoomAddPlayerReconnect(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	conn1 := make(chan []byte, 10)
	conn2 := make(chan []byte, 10)

	room.AddPlayer("user-1", "", "", conn1)

	// Same user adds again — should reconnect (update Conn), not error
	err := room.AddPlayer("user-1", "", "", conn2)
	if err != nil {
		t.Errorf("AddPlayer reconnection should succeed, got: %v", err)
	}
	if len(room.Players) != 1 {
		t.Errorf("Players = %d, want 1 for reconnected player", len(room.Players))
	}
	if !room.Players[0].Connected {
		t.Error("Player should be connected after reconnection")
	}
}

func TestMarkDisconnectedIgnoresStaleConnection(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	oldConn := make(chan []byte, 10)
	newConn := make(chan []byte, 10)

	if err := room.AddPlayer("user-1", "", "", oldConn); err != nil {
		t.Fatalf("AddPlayer failed: %v", err)
	}
	if err := room.AddPlayer("user-1", "", "", newConn); err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}

	room.MarkDisconnected("user-1", oldConn)
	if !room.Players[0].Connected {
		t.Fatal("stale connection close should not mark reconnected player disconnected")
	}

	room.MarkDisconnected("user-1", newConn)
	if room.Players[0].Connected {
		t.Fatal("current connection close should mark player disconnected")
	}
}

func TestRoomAddPlayerFull(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)

	for i := 0; i < 3; i++ {
		err := room.AddPlayer(
			string(rune('a'+i)),
			"",
			"",
			make(chan []byte, 10),
		)
		if err != nil {
			t.Fatalf("AddPlayer %d failed: %v", i, err)
		}
	}

	err := room.AddPlayer("extra", "", "", make(chan []byte, 10))
	if err == nil {
		t.Error("AddPlayer should return error for full room")
	}
}

func TestRoomRemovePlayer(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	conn1 := make(chan []byte, 10)
	conn2 := make(chan []byte, 10)

	room.AddPlayer("user-1", "", "", conn1)
	room.AddPlayer("user-2", "", "", conn2)
	room.RemovePlayer("user-1")

	if len(room.Players) != 1 {
		t.Errorf("Players = %d, want 1", len(room.Players))
	}
	if room.Players[0].UserID != "user-2" {
		t.Errorf("Remaining player = %s, want user-2", room.Players[0].UserID)
	}
	if room.Players[0].Seat != 0 {
		t.Errorf("Remaining player seat = %d, want 0 after removal", room.Players[0].Seat)
	}
}

func TestRoomRemovePlayerNonexistent(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	room.AddPlayer("user-1", "", "", make(chan []byte, 10))

	// This should not panic
	room.RemovePlayer("nonexistent")
	if len(room.Players) != 1 {
		t.Errorf("Players = %d, want 1", len(room.Players))
	}
}

func TestRoomRemoveLastPlayer(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	room.AddPlayer("user-1", "", "", make(chan []byte, 10))
	room.RemovePlayer("user-1")

	if len(room.Players) != 0 {
		t.Errorf("Players = %d, want 0", len(room.Players))
	}
	if room.Status != "waiting" {
		t.Errorf("Status = %s, want waiting", room.Status)
	}
}

func TestRoomSetReady(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	conn := make(chan []byte, 10)

	room.AddPlayer("user-1", "", "", conn)

	// Consume the player_joined message
	<-conn

	room.SetReady("user-1")

	// Should get a player_ready message
	select {
	case msg := <-conn:
		var parsed map[string]interface{}
		if err := json.Unmarshal(msg, &parsed); err != nil {
			t.Fatalf("Failed to parse message: %v", err)
		}
		if parsed["type"] != "player_ready" {
			t.Errorf("message type = %v, want player_ready", parsed["type"])
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for player_ready message")
	}

	if !room.Players[0].Ready {
		t.Error("Player should be ready after SetReady")
	}
}

func TestRoomAllReadyStartsGame(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)
	conn1 := make(chan []byte, 20)
	conn2 := make(chan []byte, 20)
	conn3 := make(chan []byte, 20)

	room.AddPlayer("user-1", "", "", conn1)
	room.AddPlayer("user-2", "", "", conn2)
	room.AddPlayer("user-3", "", "", conn3)

	// Drain AddPlayer broadcasts from all connections.
	// AddPlayer broadcasts to every player in the room.
	// After 3 adds:
	//   conn1: [joined(u1), joined(u2), joined(u3)]
	//   conn2: [joined(u2), joined(u3)]
	//   conn3: [joined(u3)]
	drainN(t, conn1, 3, "player_joined")
	drainN(t, conn2, 2, "player_joined")
	drainN(t, conn3, 1, "player_joined")

	// SetReady broadcasts player_ready to all players in the room.
	room.SetReady("user-1")
	drainN(t, conn1, 1, "player_ready")
	drainN(t, conn2, 1, "player_ready")
	drainN(t, conn3, 1, "player_ready")

	if room.Players[0].Ready != true {
		t.Error("user-1 should be ready")
	}

	room.SetReady("user-2")
	drainN(t, conn1, 1, "player_ready")
	drainN(t, conn2, 1, "player_ready")
	drainN(t, conn3, 1, "player_ready")

	if room.Status != "waiting" {
		t.Errorf("Status = %s, want waiting before all ready", room.Status)
	}

	// SetReady("user-3") toggles user-3 ready
	room.SetReady("user-3")
	// Drain player_ready
	drainN(t, conn1, 1, "player_ready")
	drainN(t, conn2, 1, "player_ready")
	drainN(t, conn3, 1, "player_ready")

	// StartGame triggers game_start
	if err := room.StartGame("user-1"); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Now each conn should have 1 game_start message
	drainN(t, conn1, 1, "game_start")
	drainN(t, conn2, 1, "game_start")
	drainN(t, conn3, 1, "game_start")

	if room.Status != "playing" {
		t.Errorf("Status = %s, want playing after game start", room.Status)
	}
	if room.State == nil {
		t.Error("State should not be nil after game start")
	}
}

// TestRoomAddBot_SeatAssignment verifies that AddBot assigns unique seats.
func TestRoomAddBot_SeatAssignment(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)

	conn1 := make(chan []byte, 20)
	conn2 := make(chan []byte, 20)

	room.AddPlayer("user-1", "", "", conn1) // owner
	room.AddPlayer("user-2", "", "", conn2)

	// Drain broadcast messages
	drainN(t, conn1, 2, "player_joined")
	drainN(t, conn2, 1, "player_joined")

	// Add a bot via AddBot
	err := room.AddBot("user-1")
	if err != nil {
		t.Fatalf("AddBot failed: %v", err)
	}

	if len(room.Players) != 3 {
		t.Fatalf("Players = %d, want 3", len(room.Players))
	}

	// Verify all 3 players have unique seats
	seatSet := map[int]bool{}
	for _, p := range room.Players {
		if seatSet[p.Seat] {
			t.Errorf("Seat collision: seat %d is used by multiple players", p.Seat)
		}
		seatSet[p.Seat] = true
	}

	// Verify owner is still in the room
	ownerFound := false
	for _, p := range room.Players {
		if p.UserID == "user-1" {
			ownerFound = true
			break
		}
	}
	if !ownerFound {
		t.Error("Owner was removed from room after AddBot")
	}

	// Verify room is full
	err = room.AddBot("user-1")
	if err == nil {
		t.Error("AddBot should fail when room is full")
	}
}

// TestRoomAddBot_NonSequentialSeats verifies AddBot doesn't collide with
// non-contiguous seat numbers caused by AddPlayer's random seat selection.
func TestRoomAddBot_NonSequentialSeats(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)

	// Set up non-contiguous seats directly: owner at seat 0, second player at seat 2.
	// This simulates what AddPlayer does when rand picks a non-sequential seat.
	room.mu.Lock()
	room.Players = []*PlayerSession{
		{UserID: "user-1", Seat: 0, Conn: make(chan []byte, 10), Connected: true},
		{UserID: "user-2", Seat: 2, Conn: make(chan []byte, 10), Connected: true},
	}
	room.mu.Unlock()

	// Debug: check state before AddBot
	t.Logf("Players before AddBot:")
	for _, p := range room.Players {
		t.Logf("  %s at seat %d", p.UserID, p.Seat)
	}

	// Players = [owner(seat=0), user-2(seat=2)], len=2, seat=1 is free
	// Old bug: AddBot would assign seat := len(r.Players) = 2 → COLLISION with user-2
	err := room.AddBot("user-1")
	if err != nil {
		t.Fatalf("AddBot failed: %v", err)
	}

	if len(room.Players) != 3 {
		t.Fatalf("Players = %d, want 3", len(room.Players))
	}

	// Debug: check state after AddBot
	t.Logf("Players after AddBot:")
	for _, p := range room.Players {
		t.Logf("  %s at seat %d (bot=%v)", p.UserID, p.Seat, p.IsBot)
	}

	// Verify all seats are unique
	seatSet := map[int]bool{}
	for _, p := range room.Players {
		if seatSet[p.Seat] {
			t.Errorf("Seat collision: seat %d is used by multiple players", p.Seat)
		}
		seatSet[p.Seat] = true
	}

	// Verify seat 1 is now occupied (was the only free seat)
	seat1found := false
	for _, p := range room.Players {
		if p.Seat == 1 {
			seat1found = true
			break
		}
	}
	if !seat1found {
		t.Error("Seat 1 should be occupied after AddBot, but it's still empty")
	}
}

// TestFillEmptySeats_NoOverfill verifies FillEmptySeats correctly fills
// only up to capacity and doesn't exceed it.
func TestFillEmptySeats_NoOverfill(t *testing.T) {
	rm := NewRoomManager(nil, nil)
	room := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	// Add 1 human player (owner)
	room.AddPlayer("user-1", "", "", make(chan []byte, 10))

	// Fill empty seats
	added := rm.FillEmptySeats("room-1")
	if added != 2 {
		t.Errorf("Added %d bots, want 2", added)
	}

	if len(room.Players) != 3 {
		t.Errorf("Players = %d, want 3", len(room.Players))
	}

	// Verify owner is still in the room
	ownerFound := false
	for _, p := range room.Players {
		if p.UserID == "user-1" {
			ownerFound = true
			break
		}
	}
	if !ownerFound {
		t.Error("Owner was removed after FillEmptySeats")
	}

	// All seats should be unique
	seatSet := map[int]bool{}
	for _, p := range room.Players {
		if seatSet[p.Seat] {
			t.Errorf("Seat collision: seat %d is used by multiple players", p.Seat)
		}
		seatSet[p.Seat] = true
	}

	// FillEmptySeats on already full room should add 0
	added = rm.FillEmptySeats("room-1")
	if added != 0 {
		t.Errorf("Added %d bots on full room, want 0", added)
	}
}

func TestRoomHumanCount(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)

	// Empty room
	if n := room.humanCount(); n != 0 {
		t.Errorf("humanCount = %d, want 0 for empty room", n)
	}

	// Add 1 human + 2 bots
	room.AddPlayer("user-1", "", "", make(chan []byte, 10))
	room.AddPlayer("ai:bot:1", "", "", make(chan []byte, 10))
	room.AddPlayer("ai:bot:2", "", "", make(chan []byte, 10))

	// Mark bots
	room.mu.Lock()
	for _, p := range room.Players {
		if p.UserID == "ai:bot:1" || p.UserID == "ai:bot:2" {
			p.IsBot = true
		}
	}
	room.mu.Unlock()

	if n := room.humanCount(); n != 1 {
		t.Errorf("humanCount = %d, want 1", n)
	}
}

func TestRoomAllBots(t *testing.T) {
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, nil)

	// Empty room — should be false (no players at all)
	if room.allBots() {
		t.Error("allBots = true for empty room, want false")
	}

	// Add only bots (2 bots so there is room for a human)
	room.mu.Lock()
	room.Players = []*PlayerSession{
		{UserID: "ai:bot:1", Seat: 0, IsBot: true, Connected: true},
		{UserID: "ai:bot:2", Seat: 1, IsBot: true, Connected: true},
	}
	room.mu.Unlock()

	if !room.allBots() {
		t.Error("allBots = false for all-bot room, want true")
	}

	// Add a human (room has capacity 3, 2 bots + 1 human = 3)
	room.AddPlayer("user-1", "", "", make(chan []byte, 10))
	if room.allBots() {
		t.Error("allBots = true for mixed room, want false")
	}

	// Verify we have 3 players (2 bots + 1 human)
	if len(room.Players) != 3 {
		t.Errorf("Players = %d, want 3 (2 bots + 1 human)", len(room.Players))
	}
}

// mockRoomStore implements RoomStore for testing lifecycle logic.
type mockRoomStore struct {
	mu         sync.Mutex
	closedIDs  []string
	ensuredIDs []string
}

func (m *mockRoomStore) CloseRoom(ctx context.Context, roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closedIDs = append(m.closedIDs, roomID)
	return nil
}

func (m *mockRoomStore) EnsureRoom(ctx context.Context, roomID, gameType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensuredIDs = append(m.ensuredIDs, roomID)
	return nil
}

func (m *mockRoomStore) CloseStaleRooms(ctx context.Context, waitingTimeout, playingTimeout time.Duration) (int64, error) {
	return 0, nil
}

func (m *mockRoomStore) AddRoomPlayer(ctx context.Context, rp *model.RoomPlayer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *mockRoomStore) SaveScore(ctx context.Context, userID int64, gameType string, amount, balance int, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *mockRoomStore) SaveChatMessage(ctx context.Context, msg *model.ChatMessage) error {
	return nil
}

func (m *mockRoomStore) CreateGameRecord(ctx context.Context, record *model.GameRecord) (int64, error) {
	return 0, nil
}

func (m *mockRoomStore) EndGameRecord(ctx context.Context, gameID int64, resultJSON string) error {
	return nil
}

func (m *mockRoomStore) AddGameAction(ctx context.Context, action *model.GameAction) error {
	return nil
}

func (m *mockRoomStore) closedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.closedIDs)
}

func TestCloseRoom_FullCleanup(t *testing.T) {
	store := &mockRoomStore{}
	room := NewGameRoom("room-1", "doudizhu", &mockEngine{}, store)

	// Add 2 bots + 1 human
	conn1 := make(chan []byte, 20)
	conn2 := make(chan []byte, 20)
	conn3 := make(chan []byte, 20)
	room.AddPlayer("ai:bot:1", "", "", conn1)
	room.AddPlayer("ai:bot:2", "", "", conn2)
	room.AddPlayer("user-1", "", "", conn3)

	// Mark bots
	room.mu.Lock()
	for _, p := range room.Players {
		if p.UserID == "ai:bot:1" || p.UserID == "ai:bot:2" {
			p.IsBot = true
		}
	}
	room.mu.Unlock()

	// Drain join broadcasts (3 adds = 3 broadcasts per conn for the first player, 2 for second, etc.)
	drainN(t, conn1, 3, "player_joined")
	drainN(t, conn2, 2, "player_joined")
	drainN(t, conn3, 1, "player_joined")

	// Call closeRoom directly
	room.CloseRoom()

	// Assert Closed flag
	if !room.Closed {
		t.Error("room.Closed = false, want true")
	}

	// Assert DB was called
	if store.closedCount() != 1 {
		t.Errorf("CloseRoom called %d times, want 1", store.closedCount())
	}
	if store.closedIDs[0] != "room-1" {
		t.Errorf("closed room ID = %s, want room-1", store.closedIDs[0])
	}

	// Assert EnsureRoom was called (3 AddPlayers + 1 CloseRoom = 4)
	if len(store.ensuredIDs) != 4 {
		t.Errorf("EnsureRoom called %d times, want 4", len(store.ensuredIDs))
	}

	// Assert room_closed broadcast was sent to connected players
	drainN(t, conn1, 1, "room_closed")
	drainN(t, conn2, 1, "room_closed")
	drainN(t, conn3, 1, "room_closed")
}

func TestCleanup_TriggerB_AllHumansLeave(t *testing.T) {
	store := &mockRoomStore{}
	rm := NewRoomManager(store, nil)
	room := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	// 1 human + 2 bots
	room.AddPlayer("user-1", "", "", make(chan []byte, 10))
	room.mu.Lock()
	room.Players = append(room.Players,
		&PlayerSession{UserID: "ai:bot:1", Seat: 1, IsBot: true, Connected: true},
		&PlayerSession{UserID: "ai:bot:2", Seat: 2, IsBot: true, Connected: true},
	)
	room.mu.Unlock()

	// Remove the human — now 0 humans
	room.RemovePlayer("user-1")

	// Run cleanup — should close the room
	rm.cleanup(0, 0)

	// Room should be removed from manager
	if got := rm.GetRoom("room-1"); got != nil {
		t.Error("room-1 should be removed from manager after all humans leave")
	}
	if store.closedCount() != 1 {
		t.Errorf("CloseRoom called %d times, want 1", store.closedCount())
	}
}

func TestCleanup_TriggerC_IdleEmpty(t *testing.T) {
	store := &mockRoomStore{}
	rm := NewRoomManager(store, nil)
	room := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	// Room has 0 players, set lastActiveAt to 6 minutes ago
	room.mu.Lock()
	room.lastActiveAt = time.Now().Add(-6 * time.Minute)
	room.mu.Unlock()

	// Run cleanup with 5min idle timeout
	rm.cleanup(0, 5*time.Minute)

	if got := rm.GetRoom("room-1"); got != nil {
		t.Error("room-1 should be removed from manager when idle-empty")
	}
	if store.closedCount() != 1 {
		t.Errorf("CloseRoom called %d times, want 1", store.closedCount())
	}
}

func TestCleanup_TriggerE_AllDisconnected(t *testing.T) {
	store := &mockRoomStore{}
	rm := NewRoomManager(store, nil)
	room := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	// All players disconnected for > 2 minutes
	room.mu.Lock()
	discAt := time.Now().Add(-3 * time.Minute)
	room.Players = []*PlayerSession{
		{UserID: "user-1", Seat: 0, IsBot: false, Connected: false, DisconnectedAt: &discAt},
		{UserID: "user-2", Seat: 1, IsBot: false, Connected: false, DisconnectedAt: &discAt},
		{UserID: "ai:bot:1", Seat: 2, IsBot: true, Connected: false, DisconnectedAt: &discAt},
	}
	room.mu.Unlock()

	rm.cleanup(2*time.Minute, 5*time.Minute)

	if got := rm.GetRoom("room-1"); got != nil {
		t.Error("room-1 should be removed when all players disconnected > timeout")
	}
}

func TestCleanup_TriggerF_OnlyBotsInPlaying(t *testing.T) {
	store := &mockRoomStore{}
	rm := NewRoomManager(store, nil)
	room := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	// Room in "playing" with only bots
	room.mu.Lock()
	room.Status = "playing"
	room.Players = []*PlayerSession{
		{UserID: "ai:bot:1", Seat: 0, IsBot: true, Connected: true},
		{UserID: "ai:bot:2", Seat: 1, IsBot: true, Connected: true},
		{UserID: "ai:bot:3", Seat: 2, IsBot: true, Connected: true},
	}
	room.lastActiveAt = time.Now().Add(-6 * time.Minute)
	room.mu.Unlock()

	rm.cleanup(0, 5*time.Minute)

	if got := rm.GetRoom("room-1"); got != nil {
		t.Error("room-1 should be removed when only bots in playing state")
	}
}

func TestRemoveDisconnectedPlayers_DoesNotReassignSeatsWhilePlaying(t *testing.T) {
	room := NewGameRoom("room-1", "dashengji", &mockEngine{}, nil)
	discAt := time.Now().Add(-2 * time.Minute)
	room.Players = []*PlayerSession{
		{UserID: "user-1", Seat: 0, PlayerID: 0, Connected: true},
		{UserID: "user-2", Seat: 1, PlayerID: 1, Connected: false, DisconnectedAt: &discAt},
		{UserID: "user-3", Seat: 2, PlayerID: 2, Connected: true},
		{UserID: "user-4", Seat: 3, PlayerID: 3, Connected: true},
	}
	room.Status = "playing"

	remaining := room.RemoveDisconnectedPlayers(time.Minute)

	if remaining != 4 {
		t.Fatalf("remaining players = %d, want 4", remaining)
	}
	for i, p := range room.Players {
		if p.Seat != i {
			t.Fatalf("player %s seat = %d, want %d", p.UserID, p.Seat, i)
		}
		if p.PlayerID != int64(i) {
			t.Fatalf("player %s playerID = %d, want %d", p.UserID, p.PlayerID, i)
		}
	}
}

func TestCleanup_NotClosed_HumanPresent(t *testing.T) {
	store := &mockRoomStore{}
	rm := NewRoomManager(store, nil)
	room := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	// 1 human + 2 bots
	room.mu.Lock()
	room.Players = []*PlayerSession{
		{UserID: "user-1", Seat: 0, IsBot: false, Connected: true},
		{UserID: "ai:bot:1", Seat: 1, IsBot: true, Connected: true},
		{UserID: "ai:bot:2", Seat: 2, IsBot: true, Connected: true},
	}
	room.mu.Unlock()

	rm.cleanup(0, 5*time.Minute)

	// Room should still exist
	if got := rm.GetRoom("room-1"); got == nil {
		t.Error("room-1 should NOT be closed — human is still present")
	}
}

func TestCleanup_NotClosed_RecentlyActive(t *testing.T) {
	store := &mockRoomStore{}
	rm := NewRoomManager(store, nil)
	room := rm.GetOrCreateRoom("room-1", "doudizhu", &mockEngine{})

	// Empty room but just recently active
	room.mu.Lock()
	room.lastActiveAt = time.Now() // just now
	room.mu.Unlock()

	rm.cleanup(0, 5*time.Minute)

	// Room should still exist — not idle long enough
	if got := rm.GetRoom("room-1"); got == nil {
		t.Error("room-1 should NOT be closed — recently active")
	}
}

// drainN reads n messages from conn and verifies each has the expected type.
func drainN(t *testing.T, conn chan []byte, n int, expectedType string) {
	t.Helper()
	for i := 0; i < n; i++ {
		select {
		case msg := <-conn:
			var parsed map[string]interface{}
			if err := json.Unmarshal(msg, &parsed); err != nil {
				t.Fatalf("Failed to parse message: %v", err)
			}
			if parsed["type"] != expectedType {
				t.Errorf("expected %q, got %q", expectedType, parsed["type"])
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %q (got %d/%d)", expectedType, i, n)
		}
	}
}
