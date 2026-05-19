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
	for _, phase := range []string{"calling", "snatching", "revealing", "doubling", "playing"} {
		schemas := GetToolSchemas(phase)
		for _, s := range schemas {
			data, err := json.Marshal(s)
			if err != nil {
				t.Errorf("[%s] schema %s failed marshal: %v", phase, s.Function.Name, err)
			}
			if len(data) < 10 {
				t.Errorf("[%s] schema %s too short", phase, s.Function.Name)
			}
		}
	}
}

func TestToolSchemas_PhaseGating(t *testing.T) {
	tests := []struct {
		phase   string
		mustHave  []string
		mustNotHave []string
	}{
		{
			phase:     "calling",
			mustHave:    []string{"check_my_hand", "check_game_status", "say", "bid_landlord", "pass_bid"},
			mustNotHave: []string{"play_cards", "reveal_cards", "pass_reveal", "choose_double", "choose_no_double"},
		},
		{
			phase:     "snatching",
			mustHave:    []string{"check_my_hand", "check_game_status", "say", "bid_landlord", "pass_bid"},
			mustNotHave: []string{"play_cards", "reveal_cards", "pass_reveal", "choose_double", "choose_no_double"},
		},
		{
			phase:     "revealing",
			mustHave:    []string{"check_my_hand", "check_game_status", "say", "reveal_cards", "pass_reveal"},
			mustNotHave: []string{"play_cards", "bid_landlord", "pass_bid", "choose_double", "choose_no_double"},
		},
		{
			phase:     "doubling",
			mustHave:    []string{"check_my_hand", "check_game_status", "say", "choose_double", "choose_no_double"},
			mustNotHave: []string{"play_cards", "bid_landlord", "pass_bid", "reveal_cards", "pass_reveal"},
		},
		{
			phase:     "playing",
			mustHave:    []string{"check_my_hand", "check_game_status", "say", "play_cards"},
			mustNotHave: []string{"bid_landlord", "pass_bid", "reveal_cards", "pass_reveal", "choose_double", "choose_no_double"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			schemas := GetToolSchemas(tt.phase)
			names := make(map[string]bool)
			for _, s := range schemas {
				names[s.Function.Name] = true
			}
			for _, name := range tt.mustHave {
				if !names[name] {
					t.Errorf("[%s] must have %s, but it is missing", tt.phase, name)
				}
			}
			for _, name := range tt.mustNotHave {
				if names[name] {
					t.Errorf("[%s] must NOT have %s, but it is present", tt.phase, name)
				}
			}
		})
	}
}
