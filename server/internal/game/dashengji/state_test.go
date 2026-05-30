package dashengji

import (
	"encoding/json"
	"testing"
)

func TestNoticeJSON_SeatZeroNotOmitted(t *testing.T) {
	n := Notice{Seq: 1, Kind: "test", Seat: 0}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, ok := raw["seat"]; !ok {
		t.Errorf("seat should not be omitted when value is 0, got JSON: %s", string(b))
	}
}
