package dashengji

import "encoding/json"

// GamePhase represents the current phase of a Dashengji game.
type GamePhase int

const (
	PhaseSetTrump      GamePhase = iota // 定主
	PhaseCounterTrump                   // 反主
	PhaseTakeBottom                     // 起底
	PhaseDiscardBottom                  // 扣底
	PhasePlaying                        // 出牌
	PhaseEnded                          // 结束
)

var phaseNames = map[GamePhase]string{
	PhaseSetTrump:      "set_trump",
	PhaseCounterTrump:  "counter_trump",
	PhaseTakeBottom:    "take_bottom",
	PhaseDiscardBottom: "discard_bottom",
	PhasePlaying:       "playing",
	PhaseEnded:         "ended",
}

func (p GamePhase) String() string {
	if name, ok := phaseNames[p]; ok {
		return name
	}
	return "unknown"
}

// phaseActions maps each game phase to allowed action strings.
var phaseActions = map[GamePhase][]string{
	PhaseSetTrump:      {"set_trump", "pass_trump"},
	PhaseCounterTrump:  {"counter_trump", "pass_counter"},
	PhaseTakeBottom:    {"take_bottom", "pass_take_bottom"},
	PhaseDiscardBottom: {"discard_bottom"},
	PhasePlaying:       {"play"},
}

// AllowedActions returns the list of valid action strings for a phase.
func AllowedActions(phase GamePhase) []string {
	return phaseActions[phase]
}

// SeatTeam returns 0 for team A (seats 0,2) or 1 for team B (seats 1,3).
func SeatTeam(seat int) int {
	return seat % 2
}

// IsDealerTeam returns true if the seat belongs to the dealer team.
func IsDealerTeam(seat int) bool {
	return SeatTeam(seat) == 0
}

// PartnerSeat returns the partner's seat (0<->2, 1<->3).
func PartnerSeat(seat int) int {
	return (seat + 2) % 4
}

// PlayerHand represents a player's hand and metadata.
type PlayerHand struct {
	UserID int64  `json:"user_id"`
	Seat   int    `json:"seat"`
	Hand   []Card `json:"hand"`
}

// PlayRecord records a play made by a player.
type PlayRecord struct {
	Seat  int    `json:"seat"`
	Play  Play   `json:"play"`
	Cards []Card `json:"cards"`
}

// Notice is a state-carried event that clients display once as a toast.
type Notice struct {
	Seq    int    `json:"seq"`
	Kind   string `json:"kind"`
	Seat   int    `json:"seat"`
	Action string `json:"action,omitempty"`
}

// GameState represents the full state of a Dashengji game.
type GameState struct {
	Phase       GamePhase    `json:"phase"`
	Players     []PlayerHand `json:"players"`
	CurrentSeat int          `json:"current_seat"`

	// Dealer team info
	DealerSeats         [2]int `json:"dealer_seats"`
	TeamLevels          [2]int `json:"team_levels"`
	CurrentLevel        int    `json:"current_level"` // 3-14 (A=14)
	LevelRank           int    `json:"level_rank"`    // base rank of level card (3->3, ..., A->14)
	SwappedDealer       bool   `json:"swapped_dealer"`
	OriginalDealerSeats [2]int `json:"original_dealer_seats"`

	// Trump
	TrumpSuit     int    `json:"trump_suit"` // -1 if not set, 0-3 otherwise
	IsDeadTrump   bool   `json:"is_dead_trump"`
	TrumpCards    []Card `json:"trump_cards"`
	TrumpRevealed bool   `json:"trump_revealed"`

	// Bottom cards
	BottomCards    []Card `json:"bottom_cards"`
	BottomTaken    bool   `json:"bottom_taken"`
	TakeBottomSeat int    `json:"take_bottom_seat"`
	DiscardedCards []Card `json:"discarded_cards"`
	BottomRevealed bool   `json:"bottom_revealed"`

	// Play state
	LastPlay          *PlayRecord  `json:"last_play"`
	PlayHistory       []PlayRecord `json:"play_history"`
	ConsecutivePasses int          `json:"consecutive_passes"`
	WinnerSeat        *int         `json:"winner_seat,omitempty"`
	RoundPlays        []PlayRecord `json:"round_plays"`  // plays in current round (0-4)
	RoundLeader       int          `json:"round_leader"` // seat that started the current round
	PendingEnd        bool         `json:"pending_end"`

	// Pass tracking
	HasPassedTrump   map[int]bool `json:"has_passed_trump"`
	HasPassedCounter map[int]bool `json:"has_passed_counter"`

	// Score tracking
	RoundPoints   int      `json:"round_points"`
	RoundNum      int      `json:"round_num"`
	DealerHistory []int    `json:"dealer_history"` // which player took bottom each round (alternating)
	Notices       []Notice `json:"notices"`
	NoticeSeq     int      `json:"notice_seq"`
}

// ToJSON serializes the GameState to JSON.
func (s *GameState) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON deserializes JSON data into the GameState.
func (s *GameState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
