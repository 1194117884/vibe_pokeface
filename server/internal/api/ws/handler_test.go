package ws

import (
	"encoding/json"
	"testing"
)

func TestMessageSerialization(t *testing.T) {
	msg := C2SMessage{
		Type:   "room_action",
		RoomID: "room-1",
		Data:   json.RawMessage(`{"action":"ready"}`),
	}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded C2SMessage
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if decoded.Type != "room_action" {
		t.Errorf("type = %s, want room_action", decoded.Type)
	}
	if decoded.RoomID != "room-1" {
		t.Errorf("room_id = %s, want room-1", decoded.RoomID)
	}
}

func TestS2CMessage(t *testing.T) {
	msg := S2CMessage{
		Type: "game_start",
		Data: map[string]interface{}{"hand": []int{0, 1, 2}},
	}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if decoded["type"] != "game_start" {
		t.Errorf("type = %v, want game_start", decoded["type"])
	}
}

func TestC2SMessageWithAction(t *testing.T) {
	msg := C2SMessage{
		Type:   "room_action",
		RoomID: "room-1",
		Data:   json.RawMessage(`{"action":"play","cards":[0,1,2]}`),
	}
	b, _ := json.Marshal(msg)

	var decoded C2SMessage
	json.Unmarshal(b, &decoded)

	var action RoomAction
	if err := json.Unmarshal(decoded.Data, &action); err != nil {
		t.Fatalf("Failed to unmarshal action: %v", err)
	}
	if action.Action != "play" {
		t.Errorf("action = %s, want play", action.Action)
	}
	if len(action.Cards) != 3 {
		t.Errorf("cards = %v, want [0 1 2]", action.Cards)
	}
}

func TestS2CMessageRoundTrip(t *testing.T) {
	original := S2CMessage{
		Type: "state_update",
		Data: map[string]interface{}{
			"phase":    1,
			"turn":     0,
			"hand":     []int{3, 4, 5},
			"discards": []int{},
		},
	}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded S2CMessage
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if decoded.Type != original.Type {
		t.Errorf("type = %s, want %s", decoded.Type, original.Type)
	}
}
