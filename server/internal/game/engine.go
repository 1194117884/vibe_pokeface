package game

import (
	"context"
	"time"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

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

// GameErrorCode identifies the category of a game rule violation.
type GameErrorCode string

const (
	ErrPhaseMismatch GameErrorCode = "PHASE_MISMATCH"
	ErrNotYourTurn   GameErrorCode = "NOT_YOUR_TURN"
	ErrInvalidAction GameErrorCode = "INVALID_ACTION"
	ErrInvalidCards  GameErrorCode = "INVALID_CARDS"
	ErrCannotPass    GameErrorCode = "CANNOT_PASS"
	ErrCannotBeat    GameErrorCode = "CANNOT_BEAT"
	ErrWaitTeammate  GameErrorCode = "WAIT_TEAMMATE"
	ErrAlreadyActed  GameErrorCode = "ALREADY_ACTED"
)

// GameError is a structured error returned when a game rule is violated.
// It carries enough context for consumers (WS, AI, frontend) to produce
// localized messages.
type GameError struct {
	Code   GameErrorCode `json:"code"`
	Phase  string        `json:"phase,omitempty"`
	Action string        `json:"action,omitempty"`
}

func (e *GameError) Error() string {
	if e.Phase != "" && e.Action != "" {
		return string(e.Code) + ":" + e.Phase + ":" + e.Action
	}
	return string(e.Code)
}

// GameEngine is the interface that all card game engines must implement.
type GameEngine interface {
	Init(players []PlayerInfo) (GameState, error)
	ExecuteAction(state GameState, action PlayerAction) (GameState, error)
	ValidateAction(state GameState, action PlayerAction) error
	IsRoundEnd(state GameState) bool
	CalculateScore(state GameState) ([]PlayerScore, error)
	SerializeForAI(state GameState) string
	FilterForPlayer(state GameState, seat int) GameState
}

// RoomStore abstracts DB operations needed by GameRoom for lifecycle management.
type RoomStore interface {
	CloseRoom(ctx context.Context, roomID string) error
	EnsureRoom(ctx context.Context, roomID, gameType string) error
	CloseStaleRooms(ctx context.Context, waitingTimeout, playingTimeout time.Duration) (int64, error)
	AddRoomPlayer(ctx context.Context, rp *model.RoomPlayer) error
	SaveScore(ctx context.Context, userID int64, gameType string, amount, balance int, reason string) error
	SaveChatMessage(ctx context.Context, msg *model.ChatMessage) error
	CreateGameRecord(ctx context.Context, record *model.GameRecord) (int64, error)
	EndGameRecord(ctx context.Context, gameID int64, resultJSON string) error
	AddGameAction(ctx context.Context, action *model.GameAction) error
}
