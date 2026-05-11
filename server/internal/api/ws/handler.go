package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yongkl/vibe-pokeface/internal/game/doudizhu"
)

// C2SMessage is a client-to-server WebSocket message.
type C2SMessage struct {
	Type   string          `json:"type"`
	RoomID string          `json:"room_id"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// S2CMessage is a server-to-client WebSocket message.
type S2CMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// RoomAction represents an action taken in a room.
type RoomAction struct {
	Action string `json:"action"`
	Cards  []int  `json:"cards,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// HandleWS upgrades an HTTP connection to WebSocket and starts the read loop.
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		log.Println("WebSocket connection missing user_id")
		conn.Close()
		return
	}

	client := &Client{
		ID:   userID,
		Send: make(chan []byte, 256),
	}

	go client.writePump(conn)

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	defer func() {
		conn.Close()
		h.Unregister <- client
		if client.RoomID != "" {
			if room := h.RoomManager.GetRoom(client.RoomID); room != nil {
				room.RemovePlayer(userID)
				// Auto-fill empty seats with bots
				h.RoomManager.FillEmptySeats(client.RoomID)
			}
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg C2SMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		switch msg.Type {
		case "join_room":
			h.handleJoinRoom(client, msg)
		case "leave_room":
			h.handleLeaveRoom(client, msg)
		case "room_action":
			h.handleRoomAction(client, msg)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
			errMsg, _ := json.Marshal(S2CMessage{Type: "error", Data: "unknown message type: " + msg.Type})
			select {
			case client.Send <- errMsg:
			default:
			}
		}
	}
}

// handleJoinRoom processes a join_room message from a client.
func (h *Hub) handleJoinRoom(client *Client, msg C2SMessage) {
	if client.RoomID != "" {
		if oldRoom := h.RoomManager.GetRoom(client.RoomID); oldRoom != nil {
			oldRoom.RemovePlayer(client.ID)
		}
		h.Unregister <- client
	}

	roomID := msg.RoomID
	if roomID == "" {
		errMsg, _ := json.Marshal(S2CMessage{Type: "error", Data: "room_id is required"})
		select {
		case client.Send <- errMsg:
		default:
		}
		return
	}

	gameType := "doudizhu"
	if len(msg.Data) > 0 {
		var joinData struct {
			GameType string `json:"game_type"`
		}
		if err := json.Unmarshal(msg.Data, &joinData); err == nil && joinData.GameType != "" {
			gameType = joinData.GameType
		}
	}

	room := h.RoomManager.GetOrCreateRoom(roomID, gameType, &doudizhu.Engine{})

	client.RoomID = roomID
	h.Register <- client

	if err := room.AddPlayer(client.ID, client.Send); err != nil {
		errMsg, _ := json.Marshal(S2CMessage{Type: "error", Data: err.Error()})
		select {
		case client.Send <- errMsg:
		default:
		}
	}
}

// handleLeaveRoom processes a leave_room message from a client.
func (h *Hub) handleLeaveRoom(client *Client, msg C2SMessage) {
	if client.RoomID == "" {
		return
	}

	if room := h.RoomManager.GetRoom(client.RoomID); room != nil {
		room.RemovePlayer(client.ID)
	}

	h.RoomManager.FillEmptySeats(client.RoomID)

	h.Unregister <- client
	client.RoomID = ""
}

// handleRoomAction processes a room_action message from a client.
func (h *Hub) handleRoomAction(client *Client, msg C2SMessage) {
	if client.RoomID == "" || client.RoomID != msg.RoomID {
		errMsg, _ := json.Marshal(S2CMessage{Type: "error", Data: "not in room"})
		select {
		case client.Send <- errMsg:
		default:
		}
		return
	}

	var action RoomAction
	if err := json.Unmarshal(msg.Data, &action); err != nil {
		errMsg, _ := json.Marshal(S2CMessage{Type: "error", Data: "invalid action"})
		select {
		case client.Send <- errMsg:
		default:
		}
		return
	}

	room := h.RoomManager.GetRoom(msg.RoomID)
	if room == nil {
		return
	}

	room.HandleAction(client.ID, action.Action, action.Cards)
}

// writePump reads messages from the client's Send channel and writes them to the WebSocket connection.
func (c *Client) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
