package game

// PlayerInfo represents a player in a game session.
type PlayerInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Seat int    `json:"seat"`
}

// PlayerAction represents an action taken by a player.
type PlayerAction struct {
	PlayerID int64  `json:"player_id"`
	Action   string `json:"action"`
	Cards    []int  `json:"cards,omitempty"`
}

// PlayerScore represents the score for a player at the end of a round.
type PlayerScore struct {
	PlayerID int64 `json:"player_id"`
	Score    int   `json:"score"`
}

// GameState is a marker interface for game-specific state types.
type GameState interface{}

// GameEngine is the interface that all card game engines must implement.
type GameEngine interface {
	Init(players []PlayerInfo) (GameState, error)
	ExecuteAction(state GameState, action PlayerAction) (GameState, error)
	ValidateAction(state GameState, action PlayerAction) bool
	IsRoundEnd(state GameState) bool
	CalculateScore(state GameState) ([]PlayerScore, error)
	SerializeForAI(state GameState) string
}
