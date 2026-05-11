package game

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

// PlayerSession represents a connected player in a game room.
type PlayerSession struct {
	UserID   string
	PlayerID int64
	Seat     int
	Conn     chan []byte
	Ready    bool
	IsBot    bool
}

// GameRoom represents a game room with players and game state.
type GameRoom struct {
	ID       string
	GameType string
	Players  []*PlayerSession
	Engine   GameEngine
	State    GameState
	Status   string
	store    *model.GameStore
	mu       sync.Mutex
	notify   chan []byte
}

// RoomManager manages all active game rooms.
type RoomManager struct {
	rooms map[string]*GameRoom
	mu    sync.RWMutex
	store *model.GameStore
}

// NewGameRoom creates a new game room with the given engine and store.
// The engine must be provided by the caller based on the game type.
func NewGameRoom(id string, gameType string, engine GameEngine, store *model.GameStore) *GameRoom {
	return &GameRoom{
		ID:       id,
		GameType: gameType,
		Engine:   engine,
		Players:  make([]*PlayerSession, 0),
		Status:   "waiting",
		store:    store,
		notify:   make(chan []byte, 256),
	}
}

// NewRoomManager creates a new RoomManager.
func NewRoomManager(store *model.GameStore) *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*GameRoom),
		store: store,
	}
}

// GetOrCreateRoom returns an existing room or creates a new one with the given engine.
func (rm *RoomManager) GetOrCreateRoom(roomID string, gameType string, engine GameEngine) *GameRoom {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if room, ok := rm.rooms[roomID]; ok {
		return room
	}
	room := NewGameRoom(roomID, gameType, engine, rm.store)
	rm.rooms[roomID] = room
	return room
}

// GetRoom returns a room by ID, or nil if it does not exist.
func (rm *RoomManager) GetRoom(roomID string) *GameRoom {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

// RemoveRoom removes a room from the manager.
func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.rooms, roomID)
}

// FillEmptySeats fills all empty seats in a room with bots.
// Returns the number of bots added.
func (rm *RoomManager) FillEmptySeats(roomID string) int {
	room := rm.GetRoom(roomID)
	if room == nil {
		return 0
	}
	if room.Status != "waiting" {
		return 0
	}

	nextN := room.nextBotNumber()
	added := 0
	for room.PlayerCount() < 3 {
		botID := fmt.Sprintf("ai:bot:%d", nextN)
		if err := room.FillWithBot(botID, make(chan []byte, 256)); err != nil {
			break
		}
		added++
		nextN++
	}
	return added
}

// nextBotNumber returns the next available bot sequence number for the room.
// Acquires r.mu to safely read Players.
func (r *GameRoom) nextBotNumber() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 1
	for _, p := range r.Players {
		var m int
		if _, err := fmt.Sscanf(p.UserID, "ai:bot:%d", &m); err == nil && m >= n {
			n = m + 1
		}
	}
	return n
}

// AddPlayer adds a player to the room and broadcasts the join event.
// Returns an error if the room is full or the player is already in the room.
func (r *GameRoom) AddPlayer(userID string, conn chan []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= 3 {
		return fmt.Errorf("room is full")
	}

	for _, p := range r.Players {
		if p.UserID == userID {
			return fmt.Errorf("player already in room")
		}
	}

	seat := len(r.Players)
	player := &PlayerSession{
		UserID: userID,
		Seat:   seat,
		Conn:   conn,
	}
	r.Players = append(r.Players, player)

	r.broadcastMsg("player_joined", map[string]interface{}{
		"user_id": userID,
		"seat":    seat,
		"players": r.playerList(),
	})
	return nil
}

// RemovePlayer removes a player from the room and broadcasts the leave event.
func (r *GameRoom) RemovePlayer(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := -1
	for i, p := range r.Players {
		if p.UserID == userID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}

	r.Players = append(r.Players[:idx], r.Players[idx+1:]...)
	for i, p := range r.Players {
		p.Seat = i
	}

	r.broadcastMsg("player_left", map[string]interface{}{
		"user_id": userID,
		"players": r.playerList(),
	})

	if len(r.Players) == 0 {
		r.Status = "waiting"
		r.State = nil
	}
}

// FillWithBot adds an AI bot player to fill an empty seat.
func (r *GameRoom) FillWithBot(botID string, conn chan []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= 3 {
		return fmt.Errorf("room is full")
	}

	for _, p := range r.Players {
		if p.UserID == botID {
			return fmt.Errorf("bot already in room")
		}
	}

	seat := len(r.Players)
	bot := &PlayerSession{
		UserID: botID,
		Seat:   seat,
		Conn:   conn,
		IsBot:  true,
		Ready:  true, // bots are always ready
	}
	r.Players = append(r.Players, bot)

	r.broadcastMsg("player_joined", map[string]interface{}{
		"user_id": botID,
		"seat":    seat,
		"is_bot":  true,
		"players": r.playerList(),
	})

	// Auto-start if room is now full and all ready
	if len(r.Players) == 3 {
		allReady := true
		for _, p := range r.Players {
			if !p.Ready {
				allReady = false
				break
			}
		}
		if allReady {
			r.startGame()
		}
	}

	return nil
}

// PlayerCount returns the current number of players in the room.
func (r *GameRoom) PlayerCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Players)
}

// SetReady marks a player as ready. If all 3 players are ready, the game starts.
func (r *GameRoom) SetReady(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setReady(userID)
}

// setReady is the internal implementation of SetReady, assumes the lock is held.
func (r *GameRoom) setReady(userID string) {
	for _, p := range r.Players {
		if p.UserID == userID {
			p.Ready = true
			break
		}
	}

	r.broadcastMsg("player_ready", r.playerList())

	if r.Status == "waiting" && len(r.Players) == 3 {
		allReady := true
		for _, p := range r.Players {
			if !p.Ready {
				allReady = false
				break
			}
		}
		if allReady {
			r.startGame()
		}
	}
}

// startGame initializes the game engine and broadcasts the initial game state.
func (r *GameRoom) startGame() {
	players := make([]PlayerInfo, len(r.Players))
	for i, p := range r.Players {
		p.PlayerID = int64(i)
		players[i] = PlayerInfo{
			ID:   p.PlayerID,
			Name: p.UserID,
			Seat: i,
		}
	}

	state, err := r.Engine.Init(players)
	if err != nil {
		return
	}
	r.State = state
	r.Status = "playing"

	for _, p := range r.Players {
		p.Ready = false
	}

	r.broadcastMsg("game_start", state)
}

// HandleAction processes a player action (game action or "ready").
func (r *GameRoom) HandleAction(userID string, action string, cards []int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if action == "ready" {
		r.setReady(userID)
		return
	}

	if r.Status != "playing" || r.State == nil {
		return
	}

	var player *PlayerSession
	for _, p := range r.Players {
		if p.UserID == userID {
			player = p
			break
		}
	}
	if player == nil {
		return
	}

	gameAction := PlayerAction{
		PlayerID: player.PlayerID,
		Action:   action,
		Cards:    cards,
	}

	newState, err := r.Engine.ExecuteAction(r.State, gameAction)
	if err != nil {
		errMsg, _ := json.Marshal(map[string]interface{}{
			"type": "error",
			"data": err.Error(),
		})
		select {
		case player.Conn <- errMsg:
		default:
		}
		return
	}

	r.State = newState

	if r.Engine.IsRoundEnd(newState) {
		scores, err := r.Engine.CalculateScore(newState)
		if err != nil {
			return
		}

		r.broadcastMsg("round_end", map[string]interface{}{
			"scores": scores,
		})

		r.Status = "waiting"
		r.State = nil
		for _, p := range r.Players {
			p.Ready = false
		}
	} else {
		r.broadcastMsg("state_update", newState)
	}
}

// broadcastMsg sends a JSON message with the given type and data to all players.
func (r *GameRoom) broadcastMsg(msgType string, data interface{}) {
	msg, err := json.Marshal(map[string]interface{}{
		"type": msgType,
		"data": data,
	})
	if err != nil {
		return
	}
	r.broadcast(msg)
}

// broadcast sends raw bytes to all player connections.
// Caller must hold r.mu (read or write).
func (r *GameRoom) broadcast(msg []byte) {
	for _, p := range r.Players {
		select {
		case p.Conn <- msg:
		default:
		}
	}
}

// playerList returns a summary of all players for broadcast.
// Caller must hold r.mu (read or write).
func (r *GameRoom) playerList() []map[string]interface{} {
	list := make([]map[string]interface{}, len(r.Players))
	for i, p := range r.Players {
		list[i] = map[string]interface{}{
			"user_id": p.UserID,
			"seat":    p.Seat,
			"ready":   p.Ready,
			"is_bot":  p.IsBot,
		}
	}
	return list
}
