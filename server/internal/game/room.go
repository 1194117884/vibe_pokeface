package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	"github.com/yongkl/vibe-pokeface/internal/ai"
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
	agents   map[string]*ai.AIAgent
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
		agents:   make(map[string]*ai.AIAgent),
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
	return r.nextBotNumberLocked()
}

// nextBotNumberLocked returns the next bot number assuming r.mu is already held.
func (r *GameRoom) nextBotNumberLocked() int {
	n := 1
	for _, p := range r.Players {
		var m int
		if _, err := fmt.Sscanf(p.UserID, "ai:bot:%d", &m); err == nil && m >= n {
			n = m + 1
		}
	}
	return n
}

// ChangeSeat moves a player to a different seat.
func (r *GameRoom) ChangeSeat(userID string, newSeat int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != "waiting" {
		return fmt.Errorf("cannot change seat after game started")
	}
	if newSeat < 0 || newSeat >= 3 {
		return fmt.Errorf("invalid seat number")
	}

	var player *PlayerSession
	for _, p := range r.Players {
		if p.UserID == userID {
			player = p
		}
		if p.Seat == newSeat {
			return fmt.Errorf("seat %d is already occupied", newSeat)
		}
	}
	if player == nil {
		return fmt.Errorf("player not in room")
	}

	oldSeat := player.Seat
	player.Seat = newSeat

	r.broadcastMsg("seat_changed", map[string]interface{}{
		"user_id":  userID,
		"old_seat": oldSeat,
		"new_seat": newSeat,
		"players":  r.playerList(),
	})
	return nil
}

// OwnerID returns the user ID of the room owner (first non-bot player).
func (r *GameRoom) OwnerID() string {
	for _, p := range r.Players {
		if !p.IsBot {
			return p.UserID
		}
	}
	if len(r.Players) > 0 {
		return r.Players[0].UserID
	}
	return ""
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

	// Find all occupied seats
	occupied := map[int]bool{}
	for _, p := range r.Players {
		occupied[p.Seat] = true
	}

	// Pick a random empty seat from 0..2
	available := make([]int, 0)
	for seat := 0; seat < 3; seat++ {
		if !occupied[seat] {
			available = append(available, seat)
		}
	}
	if len(available) == 0 {
		return fmt.Errorf("no available seats")
	}
	seat := available[rand.Intn(len(available))]

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

	// Stop AI agent if present
	if agent, ok := r.agents[userID]; ok {
		agent.Stop()
		delete(r.agents, userID)
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
func (r *GameRoom) FillWithBot(botID string, conn chan []byte, opts ...BotOption) error {
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

	// Apply bot options and create AI agent if a provider is configured
	botCfg := &botConfig{}
	for _, opt := range opts {
		opt(botCfg)
	}
	if botCfg.Provider != nil {
		agent := ai.NewAIAgent(botID, seat, botCfg.Character, botCfg.Provider, r)
		agent.Start()
		r.agents[botID] = agent
	}

	r.broadcastMsg("player_joined", map[string]interface{}{
		"user_id": botID,
		"seat":    seat,
		"is_bot":  true,
		"players": r.playerList(),
	})

	return nil
}

// AddBot adds an AI bot to the room by owner action.
func (r *GameRoom) AddBot(ownerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != "waiting" {
		if ownerID != r.OwnerID() {
			return fmt.Errorf("only the room owner can start the game")
		}
		return fmt.Errorf("game already started")
	}
	if len(r.Players) >= 3 {
		return fmt.Errorf("room is full")
	}

	nextN := r.nextBotNumberLocked()
	botID := fmt.Sprintf("ai:bot:%d", nextN)
	conn := make(chan []byte, 256)

	seat := len(r.Players)
	bot := &PlayerSession{
		UserID: botID,
		Seat:   seat,
		Conn:   conn,
		IsBot:  true,
		Ready:  true,
	}
	r.Players = append(r.Players, bot)

	r.broadcastMsg("player_joined", map[string]interface{}{
		"user_id": botID,
		"seat":    seat,
		"is_bot":  true,
		"players": r.playerList(),
	})
	return nil
}

// PlayerCount returns the current number of players in the room.
func (r *GameRoom) PlayerCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Players)
}

// SetReady toggles the ready status for a player.
func (r *GameRoom) SetReady(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setReady(userID)
}

// setReady is the internal implementation of SetReady, assumes the lock is held.
func (r *GameRoom) setReady(userID string) {
	for _, p := range r.Players {
		if p.UserID == userID {
			p.Ready = !p.Ready // toggle ready
			break
		}
	}
	r.broadcastMsg("player_ready", r.playerList())
}

// StartGame validates conditions and starts the game.
func (r *GameRoom) StartGame(ownerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != "waiting" {
		return fmt.Errorf("game already started")
	}
	if ownerID != r.OwnerID() {
		return fmt.Errorf("only the room owner can start the game")
	}
	if len(r.Players) < 3 {
		return fmt.Errorf("need 3 players to start")
	}
	for _, p := range r.Players {
		if !p.Ready {
			return fmt.Errorf("player %s is not ready", p.UserID)
		}
	}
	r.startGame()
	return nil
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

	// Trigger AI agent if the first player to act is a bot
	r.triggerAIAgent()
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

		// Clean up all AI agents on round end
		for id, agent := range r.agents {
			agent.Stop()
			delete(r.agents, id)
		}
	} else {
		r.broadcastMsg("state_update", newState)

		// Trigger AI agent if next player is a bot
		r.triggerAIAgent()
	}
}

// BotOption configures optional AI agent behavior for a bot player.
type BotOption func(*botConfig)

type botConfig struct {
	Character *model.AICharacter
	Provider  ai.LLMProvider
}

// WithAICharacter sets the AI character for the bot's personality.
func WithAICharacter(char *model.AICharacter) BotOption {
	return func(c *botConfig) {
		c.Character = char
	}
}

// WithLLMProvider sets the LLM provider for the bot's AI decision-making.
func WithLLMProvider(provider ai.LLMProvider) BotOption {
	return func(c *botConfig) {
		c.Provider = provider
	}
}

// createAIAgent creates and starts an AI agent for a bot player.
// The caller must hold r.mu.
func (r *GameRoom) createAIAgent(userID string, seat int, character *model.AICharacter, provider ai.LLMProvider) *ai.AIAgent {
	agent := ai.NewAIAgent(userID, seat, character, provider, r)
	agent.Start()
	r.agents[userID] = agent
	return agent
}

// stopAIAgent stops and removes an AI agent for the given user ID.
// The caller must hold r.mu.
func (r *GameRoom) stopAIAgent(userID string) {
	if agent, ok := r.agents[userID]; ok {
		agent.Stop()
		delete(r.agents, userID)
	}
}

// ExecuteAction implements ai.ActionExecutor by delegating to HandleAction.
func (r *GameRoom) ExecuteAction(userID string, action string, cards []int) {
	r.HandleAction(userID, action, cards)
}

// SendChat implements ai.ActionExecutor by delegating to BroadcastChat.
func (r *GameRoom) SendChat(senderID string, content string, msgType string) {
	r.BroadcastChat(senderID, content, msgType)
}

// triggerAIAgent checks if the current player is a bot and triggers its AI agent.
// Uses JSON marshaling to avoid circular import of the concrete game state type.
// The caller must hold r.mu.
func (r *GameRoom) triggerAIAgent() {
	if r.State == nil {
		return
	}

	stateJSON, err := json.Marshal(r.State)
	if err != nil {
		return
	}

	var stateData struct {
		CurrentSeat int `json:"current_seat"`
		Players     []struct {
			Seat int `json:"seat"`
			Hand []struct {
				ID int `json:"id"`
			} `json:"hand"`
		} `json:"players"`
	}
	if err := json.Unmarshal(stateJSON, &stateData); err != nil {
		return
	}

	currentSeat := stateData.CurrentSeat
	for _, p := range r.Players {
		if p.Seat == currentSeat && p.IsBot {
			if agent, ok := r.agents[p.UserID]; ok {
				// Update the agent's hand cards
				for _, ph := range stateData.Players {
					if ph.Seat == currentSeat {
						cards := make([]int, len(ph.Hand))
						for i, c := range ph.Hand {
							cards[i] = c.ID
						}
						agent.UpdateHand(cards)
						break
					}
				}
				agent.UpdateState(string(stateJSON))
				agent.Trigger()
			}
			return
		}
	}
}

// BroadcastChat sends a chat message to all players in the room.
func (r *GameRoom) BroadcastChat(senderID string, content string, msgType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.broadcastMsg("chat", map[string]interface{}{
		"user_id": senderID,
		"content": content,
		"type":    msgType,
	})
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
	ownerID := r.OwnerID()
	list := make([]map[string]interface{}, len(r.Players))
	for i, p := range r.Players {
		list[i] = map[string]interface{}{
			"user_id":  p.UserID,
			"seat":     p.Seat,
			"ready":    p.Ready,
			"is_bot":   p.IsBot,
			"is_owner": p.UserID == ownerID,
		}
	}
	return list
}
