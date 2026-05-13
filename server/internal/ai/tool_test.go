package ai

import (
	"encoding/json"
	"testing"
)

func TestToolCallJSON_Parse(t *testing.T) {
	raw := `{"tool": "check_my_hand", "args": {}}`
	var call ToolCall
	if err := json.Unmarshal([]byte(raw), &call); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if call.Name != "check_my_hand" {
		t.Errorf("name = %s, want check_my_hand", call.Name)
	}
}

func TestToolCallJSON_ParseWithArgs(t *testing.T) {
	raw := `{"tool": "play_cards", "args": {"cards": [0, 13, 26]}}`
	var call ToolCall
	json.Unmarshal([]byte(raw), &call)
	var args PlayCardsArgs
	json.Unmarshal(call.Args, &args)
	if len(args.Cards) != 3 || args.Cards[0] != 0 {
		t.Errorf("unexpected cards: %v", args.Cards)
	}
}

func TestToolSchemas_AreValidJSON(t *testing.T) {
	schemas := GetToolSchemas()
	for _, s := range schemas {
		data, err := json.Marshal(s)
		if err != nil {
			t.Errorf("schema %s failed marshal: %v", s.Function.Name, err)
		}
		if len(data) < 10 {
			t.Errorf("schema %s too short", s.Function.Name)
		}
	}
}
