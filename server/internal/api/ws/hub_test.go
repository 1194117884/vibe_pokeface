package ws

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	h := NewHub(nil, nil)
	if h == nil {
		t.Fatal("NewHub(nil, nil) returned nil")
	}
}

func TestHub_RegisterClient(t *testing.T) {
	h := NewHub(nil, nil)
	go h.Run()
	client := &Client{ID: "test-1", RoomID: "room-1"}
	h.Register <- client
	// Give hub time to process
	time.Sleep(50 * time.Millisecond)
	h.mu.RLock()
	_, exists := h.Rooms["room-1"]
	h.mu.RUnlock()
	if !exists {
		t.Error("expected room-1 to exist after registering client")
	}
}

func TestHub_UnregisterClient(t *testing.T) {
	h := NewHub(nil, nil)
	go h.Run()
	client := &Client{ID: "test-1", RoomID: "room-1"}
	h.Register <- client
	h.Unregister <- client
}

func TestRoomHub_AddAndCount(t *testing.T) {
	rh := NewRoomHub()
	rh.Add(&Client{ID: "c1", RoomID: "r1"})
	rh.Add(&Client{ID: "c2", RoomID: "r1"})
	if rh.Count() != 2 {
		t.Errorf("Count() = %d, want %d", rh.Count(), 2)
	}
}

func TestRoomHub_Remove(t *testing.T) {
	rh := NewRoomHub()
	rh.Add(&Client{ID: "c1", RoomID: "r1"})
	rh.Remove("c1")
	if rh.Count() != 0 {
		t.Errorf("Count() = %d, want %d", rh.Count(), 0)
	}
}

func TestRoomHub_Broadcast(t *testing.T) {
	rh := NewRoomHub()
	rh.Add(&Client{ID: "c1", RoomID: "r1", Send: make(chan []byte, 10)})
	rh.Add(&Client{ID: "c2", RoomID: "r1", Send: make(chan []byte, 10)})

	rh.Broadcast([]byte("hello"))

	// Each client should receive the message
	msg1 := <-rh.Clients["c1"].Send
	msg2 := <-rh.Clients["c2"].Send
	if string(msg1) != "hello" {
		t.Errorf("client 1 got %q, want %q", string(msg1), "hello")
	}
	if string(msg2) != "hello" {
		t.Errorf("client 2 got %q, want %q", string(msg2), "hello")
	}
}

func TestGetRoom(t *testing.T) {
	h := NewHub(nil, nil)
	go h.Run()
	client := &Client{ID: "test-1", RoomID: "room-1"}
	h.Register <- client

	time.Sleep(50 * time.Millisecond)

	room := h.GetRoom("room-1")
	if room == nil {
		t.Fatal("GetRoom() returned nil for existing room")
	}
	if room.Count() != 1 {
		t.Errorf("room count = %d, want %d", room.Count(), 1)
	}

	missing := h.GetRoom("nonexistent")
	if missing != nil {
		t.Error("GetRoom() should return nil for nonexistent room")
	}
}
