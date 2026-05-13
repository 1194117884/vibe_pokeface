package game

import (
	"encoding/json"
	"testing"
	"time"
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

func (m *mockEngine) ValidateAction(state GameState, action PlayerAction) bool {
	return true
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
	rm := NewRoomManager(nil)
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
	rm := NewRoomManager(nil)
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
	rm := NewRoomManager(nil)
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
