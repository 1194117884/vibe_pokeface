package doudizhu

import "encoding/json"

// GamePhase represents the current phase of a Dou Di Zhu game.
type GamePhase int

const (
	PhaseCalling   GamePhase = iota // 叫地主
	PhaseSnatching                   // 抢地主
	PhaseRevealing                   // 明牌
	PhaseDoubling                    // 加倍
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
	ConsecutivePasses int          `json:"consecutive_passes"`
	WinnerSeat        *int         `json:"winner_seat,omitempty"`
	BidHistory        []BidRecord  `json:"bid_history"`
	RoundNum          int          `json:"round_num"`
	Multiplier        int          `json:"multiplier"`
	HasPassed         map[int]bool `json:"has_passed"`
	SnatchCount       int          `json:"snatch_count"`
	Revealed          map[int]bool `json:"revealed"`   // players who chose 明牌
	Doubled           map[int]bool `json:"doubled"`    // players who chose 加倍
	RevealCount       int          `json:"reveal_count"` // reveal decisions made
	DoubleCount       int          `json:"double_count"` // double decisions made
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
