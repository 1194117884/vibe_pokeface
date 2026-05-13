package game

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/yongkl/vibe-pokeface/internal/ai"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

// PlayerSession represents a player in a game room.
type PlayerSession struct {
	UserID         string
	PlayerID       int64
	Seat           int
	Conn           chan []byte
	Ready          bool
	IsBot          bool
	Connected      bool
	DisconnectedAt *time.Time
	Nickname       string // display name for seat display
	CharacterID    string // avatar character ID (e.g. "panda", "fox")
}

// GameRoom represents a game room with players and game state.
type GameRoom struct {
	ID       string
	GameType string
	Players  []*PlayerSession
	Engine   GameEngine
	State    GameState
	Status   string
	Theme    string
	Closed   bool
	store    *model.GameStore
	mu       sync.Mutex
	notify   chan []byte
	agents   map[string]*ai.AIAgent
	createdAt    time.Time
	lastActiveAt time.Time
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
	now := time.Now()
	return &GameRoom{
		ID:           id,
		GameType:     gameType,
		Engine:       engine,
		Players:      make([]*PlayerSession, 0),
		Status:       "waiting",
		Theme:        "classic-poker",
		store:        store,
		notify:       make(chan []byte, 256),
		agents:       make(map[string]*ai.AIAgent),
		createdAt:    now,
		lastActiveAt: now,
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
// Returns nil if the room exists but is closed.
func (rm *RoomManager) GetOrCreateRoom(roomID string, gameType string, engine GameEngine) *GameRoom {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if room, ok := rm.rooms[roomID]; ok {
		if room.Closed {
			return nil
		}
		return room
	}
	room := NewGameRoom(roomID, gameType, engine, rm.store)
	rm.rooms[roomID] = room
	return room
}

// GetRoom returns a room by ID, or nil if it does not exist or is closed.
func (rm *RoomManager) GetRoom(roomID string) *GameRoom {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	room := rm.rooms[roomID]
	if room != nil && room.Closed {
		return nil
	}
	return room
}

// RunCleanup starts a background goroutine that periodically removes
// disconnected players past their timeout and closes idle empty rooms.
func (rm *RoomManager) RunCleanup(ctx context.Context, interval time.Duration, disconnectTimeout time.Duration, roomIdleTimeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rm.cleanup(disconnectTimeout, roomIdleTimeout)
			}
		}
	}()
}

// cleanup removes stale disconnected players and closes idle rooms.
func (rm *RoomManager) cleanup(disconnectTimeout time.Duration, roomIdleTimeout time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	now := time.Now()
	for id, room := range rm.rooms {
		if room.Closed {
			delete(rm.rooms, id)
			continue
		}
		// Remove disconnected players past the timeout
		room.RemoveDisconnectedPlayers(disconnectTimeout)
		// Close empty room past idle timeout
		if len(room.Players) == 0 && now.Sub(room.lastActiveAt) > roomIdleTimeout {
			closeRoom(room)
			delete(rm.rooms, id)
		}
	}
}

// closeRoom marks a room as closed (no more joins allowed).
func closeRoom(r *GameRoom) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Closed = true
	r.broadcastMsg("room_closed", map[string]interface{}{})
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

// SetTheme changes the room theme (owner only).
func (r *GameRoom) SetTheme(userID string, themeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.OwnerID() != userID {
		return fmt.Errorf("only the room owner can change the theme")
	}
	r.Theme = themeID
	r.broadcastMsg("theme_changed", map[string]interface{}{
		"theme": themeID,
	})
	return nil
}

// AddPlayer adds a player to the room and broadcasts the join event.
// If the player is reconnecting (already present), their connection is updated.
// Returns an error if the room is closed or full.
func (r *GameRoom) AddPlayer(userID string, nickname string, characterID string, conn chan []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Closed {
		return fmt.Errorf("room is closed")
	}

	// Reconnection: player already in the room
	for _, p := range r.Players {
		if p.UserID == userID {
			p.Conn = conn
			p.Connected = true
			p.DisconnectedAt = nil
			r.lastActiveAt = time.Now()
			// Send current state to reconnected player (fires async)
			r.sendStateTo(p)
			return nil
		}
	}

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
		UserID:      userID,
		Seat:        seat,
		Conn:        conn,
		Connected:   true,
		Nickname:    nickname,
		CharacterID: characterID,
	}
	r.Players = append(r.Players, player)

	r.broadcastMsg("player_joined", map[string]interface{}{
		"user_id": userID,
		"seat":    seat,
		"players": r.playerList(),
		"theme":   r.Theme,
	})
	return nil
}

// RemovePlayer removes a player from the room and broadcasts the leave event.
func (r *GameRoom) RemovePlayer(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removePlayer(userID)
}

// removePlayer removes a player from the room. The caller must hold r.mu.
func (r *GameRoom) removePlayer(userID string) {
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
	r.lastActiveAt = time.Now()
}

// MarkDisconnected marks a player as disconnected without removing them,
// allowing reconnection within the grace period.
func (r *GameRoom) MarkDisconnected(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.Players {
		if p.UserID == userID {
			p.Connected = false
			now := time.Now()
			p.DisconnectedAt = &now
			return
		}
	}
}

// RemoveDisconnectedPlayers removes players disconnected longer than timeout.
// Returns the number of players remaining.
func (r *GameRoom) RemoveDisconnectedPlayers(timeout time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	changed := false
	for i := len(r.Players) - 1; i >= 0; i-- {
		p := r.Players[i]
		if !p.Connected && p.DisconnectedAt != nil && now.Sub(*p.DisconnectedAt) > timeout {
			if agent, ok := r.agents[p.UserID]; ok {
				agent.Stop()
				delete(r.agents, p.UserID)
			}
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			changed = true
		}
	}
	if changed {
		// Reassign seats after removals
		for i, p := range r.Players {
			p.Seat = i
		}
		r.broadcastMsg("player_left", map[string]interface{}{
			"players": r.playerList(),
		})
		if len(r.Players) == 0 {
			r.Status = "waiting"
			r.State = nil
		}
	}
	r.lastActiveAt = now
	return len(r.Players)
}

// sendStateTo sends the current game state to a single player (for reconnection).
func (r *GameRoom) sendStateTo(p *PlayerSession) {
	// Send player_joined with current room state
	msg, err := json.Marshal(map[string]interface{}{
		"type": "player_joined",
		"data": map[string]interface{}{
			"user_id": p.UserID,
			"seat":    p.Seat,
			"players": r.playerList(),
			"theme":   r.Theme,
		},
	})
	if err != nil {
		return
	}
	select {
	case p.Conn <- msg:
	default:
	}

	// If a game is in progress, send the current state too
	if r.State != nil {
		stateMsg, err := json.Marshal(map[string]interface{}{
			"type": "state_update",
			"data": r.State,
		})
		if err != nil {
			return
		}
		select {
		case p.Conn <- stateMsg:
		default:
		}
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
		UserID:   botID,
		Seat:     seat,
		Conn:     conn,
		IsBot:    true,
		Connected: true,
		Ready:    true, // bots are always ready
	}
	// Set nickname from bot sequence number
	var botN int
	if _, err := fmt.Sscanf(botID, "ai:bot:%d", &botN); err == nil {
		bot.Nickname = fmt.Sprintf("AI Player %d", botN)
	} else {
		bot.Nickname = "AI Player"
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
		UserID:   botID,
		Seat:     seat,
		Conn:     conn,
		IsBot:    true,
		Connected: true,
		Ready:    true,
		Nickname: fmt.Sprintf("AI Player %d", nextN),
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

	// Check for 报单/报双 (cards left announcement)
	cardsLeftMsg := checkCardsLeft(newState)
	if cardsLeftMsg != "" {
		r.broadcastMsg("cards_left", map[string]interface{}{
			"message": cardsLeftMsg,
		})
	}

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
		entry := map[string]interface{}{
			"user_id":  p.UserID,
			"seat":     p.Seat,
			"ready":    p.Ready,
			"is_bot":   p.IsBot,
			"is_owner": p.UserID == ownerID,
			"nickname": p.Nickname,
		}
		if p.CharacterID != "" {
			entry["character_id"] = p.CharacterID
		}
		list[i] = entry
	}
	return list
}

// SetPlayerInfo updates a player display info after they have joined.
func (r *GameRoom) SetPlayerInfo(userID string, nickname string, characterID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.Players {
		if p.UserID == userID {
			p.Nickname = nickname
			p.CharacterID = characterID
			return
		}
	}
}

// findSeat returns the seat number for a given user ID, or -1 if not found.
func (r *GameRoom) findSeat(userID string) int {
	for _, p := range r.Players {
		if p.UserID == userID {
			return p.Seat
		}
	}
	return -1
}

// checkCardsLeft checks if any player has 1 or 2 cards remaining
// and returns an announcement message, or empty string.
func checkCardsLeft(state GameState) string {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return ""
	}
	var data struct {
		Players []struct {
			Seat int  `json:"seat"`
			Hand []struct {
				ID int `json:"id"`
			} `json:"hand"`
		} `json:"players"`
	}
	if err := json.Unmarshal(stateJSON, &data); err != nil {
		return ""
	}
	for _, p := range data.Players {
		if len(p.Hand) == 1 {
			return fmt.Sprintf("seat_%d_baodan", p.Seat)
		}
		if len(p.Hand) == 2 {
			return fmt.Sprintf("seat_%d_baoshuang", p.Seat)
		}
	}
	return ""
}
