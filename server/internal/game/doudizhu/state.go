package doudizhu

import "encoding/json"

// GamePhase represents the current phase of a Dou Di Zhu game.
type GamePhase int

const (
	PhaseCalling   GamePhase = iota // 叫地主 — first call wins immediately
	PhaseSnatching                   // 抢地主 — each eligible player gets one snatch chance
	PhasePlaying                     // 出牌
	PhaseEnded                       // 结束
)

// PlayerHand represents a player's hand and metadata during a game.
type PlayerHand struct {
	UserID     int64  `json:"user_id"`
	Seat       int    `json:"seat"`
	Hand       []Card `json:"hand"`
	IsLandlord bool   `json:"is_landlord"`
}

// PlayRecord records a play made by a player.
type PlayRecord struct {
	Seat  int    `json:"seat"`
	Play  Play   `json:"play"`
	Cards []Card `json:"cards"`
}

// GameState represents the full state of a Dou Di Zhu game at a point in time.
type GameState struct {
	Phase             GamePhase    `json:"phase"`
	Players           []PlayerHand `json:"players"`
	CurrentSeat       int          `json:"current_seat"`
	LandlordSeat      int          `json:"landlord_seat"`
	LandlordCards     []Card       `json:"landlord_cards"`
	LastPlay          *PlayRecord  `json:"last_play"`
	LastPlaySeat      int          `json:"last_play_seat"`
	ConsecutivePasses int          `json:"consecutive_passes"`
	WinnerSeat        *int         `json:"winner_seat,omitempty"`
	BidHistory        []BidRecord  `json:"bid_history"`
	RoundNum          int          `json:"round_num"`
	Multiplier        int          `json:"multiplier"`  // game multiplier (×2 per snatch)
	HasPassed         map[int]bool `json:"has_passed"`  // seats that passed during 叫地主
	SnatchCount       int          `json:"snatch_count"` // number of snatch decisions made
}

// BidRecord records a single bid action during the bidding phase.
type BidRecord struct {
	Seat   int  `json:"seat"`
	Called bool `json:"called"`
}

// ToJSON serializes the GameState to JSON.
func (s *GameState) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON deserializes JSON data into the GameState.
func (s *GameState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
