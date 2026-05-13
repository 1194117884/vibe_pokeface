package ws

import (
	"context"
	"sync"
	"time"

	"github.com/yongkl/vibe-pokeface/internal/game"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type Client struct {
	ID     string
	RoomID string
	Send   chan []byte
}

type RoomHub struct {
	Clients map[string]*Client
	mu      sync.RWMutex
}

func NewRoomHub() *RoomHub {
	return &RoomHub{
		Clients: make(map[string]*Client),
	}
}

func (rh *RoomHub) Add(client *Client) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.Clients[client.ID] = client
}

func (rh *RoomHub) Remove(clientID string) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	delete(rh.Clients, clientID)
}

func (rh *RoomHub) Broadcast(msg []byte) {
	rh.mu.RLock()
	defer rh.mu.RUnlock()
	for _, client := range rh.Clients {
		select {
		case client.Send <- msg:
		default:
		}
	}
}

func (rh *RoomHub) Count() int {
	rh.mu.RLock()
	defer rh.mu.RUnlock()
	return len(rh.Clients)
}

type Hub struct {
	Rooms       map[string]*RoomHub
	mu          sync.RWMutex
	Register    chan *Client
	Unregister  chan *Client
	RoomManager *game.RoomManager
	AIStore     *model.AIStore
	UserStore   model.UserStore
}

func NewHub(store *model.GameStore, aiStore *model.AIStore, userStore model.UserStore) *Hub {
	return &Hub{
		Rooms:       make(map[string]*RoomHub),
		Register:    make(chan *Client, 256),
		Unregister:  make(chan *Client, 256),
		RoomManager: game.NewRoomManager(store),
		AIStore:     aiStore,
		UserStore:   userStore,
	}
}

func (h *Hub) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run room cleanup every 30s: remove players disconnected >30s, close rooms idle >5min
	h.RoomManager.RunCleanup(ctx, 30*time.Second, 30*time.Second, 5*time.Minute)

	for {
		select {
		case client := <-h.Register:
			h.mu.RLock()
			room, ok := h.Rooms[client.RoomID]
			h.mu.RUnlock()
			if !ok {
				room = NewRoomHub()
				h.mu.Lock()
				h.Rooms[client.RoomID] = room
				h.mu.Unlock()
			}
			room.Add(client)

		case client := <-h.Unregister:
			h.mu.RLock()
			room, ok := h.Rooms[client.RoomID]
			h.mu.RUnlock()
			if ok {
				room.Remove(client.ID)
				if room.Count() == 0 {
					h.mu.Lock()
					delete(h.Rooms, client.RoomID)
					h.mu.Unlock()
				}
			}
		}
	}
}

func (h *Hub) GetRoom(roomID string) *RoomHub {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Rooms[roomID]
}

func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, room := range h.Rooms {
		total += room.Count()
	}
	return total
}
