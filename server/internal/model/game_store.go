package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Room struct {
	ID         string     `db:"id" json:"id"`
	Name       string     `db:"name" json:"name"`
	Theme      string     `db:"theme" json:"theme"`
	GameType   string     `db:"game_type" json:"game_type"`
	OwnerID    int64      `db:"owner_id" json:"owner_id"`
	Status     string     `db:"status" json:"status"`
	MaxPlayers int8       `db:"max_players" json:"max_players"`
	IsOpen     bool       `db:"is_open" json:"is_open"`
	Password   *string    `db:"password" json:"password,omitempty"`
	BotEnabled bool       `db:"bot_enabled" json:"bot_enabled"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	EndedAt    *time.Time `db:"ended_at" json:"ended_at,omitempty"`
}

type RoomPlayer struct {
	ID          int64  `db:"id" json:"id"`
	RoomID      string `db:"room_id" json:"room_id"`
	UserID      *int64 `db:"user_id" json:"user_id,omitempty"`
	IsBot       bool   `db:"is_bot" json:"is_bot"`
	CharacterID *int   `db:"character_id" json:"character_id,omitempty"`
	SeatIndex   int8   `db:"seat_index" json:"seat_index"`
	Score       int    `db:"score" json:"score"`
	Status      string `db:"status" json:"status"`
}

type GameRecord struct {
	ID        int64     `db:"id" json:"id"`
	RoomID    string    `db:"room_id" json:"room_id"`
	GameType  string    `db:"game_type" json:"game_type"`
	RoundNum  int       `db:"round_num" json:"round_num"`
	StateData *string   `db:"state_data" json:"state_data,omitempty"`
	Result    *string   `db:"result" json:"result,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type GameAction struct {
	ID         int64     `db:"id" json:"id"`
	GameID     int64     `db:"game_id" json:"game_id"`
	RoundNum   int       `db:"round_num" json:"round_num"`
	ActionSeq  int       `db:"action_seq" json:"action_seq"`
	PlayerID   *int64    `db:"player_id" json:"player_id,omitempty"`
	SeatIndex  int8      `db:"seat_index" json:"seat_index"`
	IsBot      bool      `db:"is_bot" json:"is_bot"`
	ActionType string    `db:"action_type" json:"action_type"`
	Cards      *string   `db:"cards" json:"cards,omitempty"`
	FullState  *string   `db:"full_state" json:"full_state,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type ScoreRecord struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	GameType  string    `db:"game_type" json:"game_type"`
	Amount    int       `db:"amount" json:"amount"`
	Balance   int       `db:"balance" json:"balance"`
	Reason    string    `db:"reason" json:"reason"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type GameStore struct {
	db *sqlx.DB
}

func NewGameStore(db *sqlx.DB) *GameStore {
	return &GameStore{db: db}
}

func (s *GameStore) CreateRoom(ctx context.Context, room *Room) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO rooms (id, name, game_type, owner_id, status, max_players, is_open, password, bot_enabled) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		room.ID, room.Name, room.GameType, room.OwnerID, room.Status, room.MaxPlayers, room.IsOpen, room.Password, room.BotEnabled)
	return err
}

func (s *GameStore) UpdateRoomStatus(ctx context.Context, roomID, status string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE rooms SET status = ? WHERE id = ?", status, roomID)
	return err
}

// CloseRoom marks a room as ended and sets the ended_at timestamp.
func (s *GameStore) CloseRoom(ctx context.Context, roomID string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE rooms SET status = 'ended', ended_at = NOW() WHERE id = ?", roomID)
	return err
}

// CloseStaleRooms closes rooms that have been stuck in a state too long.
// waiting rooms older than waitingTimeout are closed (nobody ever joined).
// playing rooms older than playingTimeout are closed (game stuck, probably only bots).
// Returns the number of rooms closed.
func (s *GameStore) CloseStaleRooms(ctx context.Context, waitingTimeout, playingTimeout time.Duration) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE rooms SET status = 'ended', ended_at = NOW()
		 WHERE status IN ('waiting','playing')
		 AND ((status = 'waiting' AND created_at < DATE_SUB(NOW(), INTERVAL ? SECOND))
		   OR (status = 'playing' AND created_at < DATE_SUB(NOW(), INTERVAL ? SECOND)))`,
		int(waitingTimeout.Seconds()), int(playingTimeout.Seconds()))
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return n, nil
}

// EnsureRoom inserts a minimal room row if one does not already exist.
// This is used when a room is created in-memory via WebSocket without
// a prior REST CreateRoom call.
func (s *GameStore) EnsureRoom(ctx context.Context, roomID, gameType string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT IGNORE INTO rooms (id, name, game_type, owner_id, status, max_players, is_open, bot_enabled) VALUES (?, ?, ?, 0, 'waiting', 3, true, true)",
		roomID, roomID, gameType)
	return err
}

func (s *GameStore) GetRoom(ctx context.Context, roomID string) (*Room, error) {
	var room Room
	err := s.db.GetContext(ctx, &room, "SELECT * FROM rooms WHERE id = ?", roomID)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (s *GameStore) AddRoomPlayer(ctx context.Context, rp *RoomPlayer) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO room_players (room_id, user_id, is_bot, seat_index, status) VALUES (?, ?, ?, ?, ?)",
		rp.RoomID, rp.UserID, rp.IsBot, rp.SeatIndex, rp.Status)
	return err
}

func (s *GameStore) GetRoomPlayers(ctx context.Context, roomID string) ([]RoomPlayer, error) {
	var players []RoomPlayer
	err := s.db.SelectContext(ctx, &players, "SELECT * FROM room_players WHERE room_id = ? ORDER BY seat_index", roomID)
	if err != nil {
		return nil, err
	}
	return players, nil
}

func (s *GameStore) SaveScore(ctx context.Context, userID int64, gameType string, amount, balance int, reason string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO scores (user_id, game_type, amount, balance, reason) VALUES (?, ?, ?, ?, ?)",
		userID, gameType, amount, balance, reason)
	return err
}

func (s *GameStore) GetUserBalance(ctx context.Context, userID int64) (int, error) {
	var balance sql.NullInt64
	err := s.db.GetContext(ctx, &balance, "SELECT MAX(balance) FROM scores WHERE user_id = ?", userID)
	if err != nil || !balance.Valid {
		return 0, err
	}
	return int(balance.Int64), nil
}

func (s *GameStore) ListActiveRooms(ctx context.Context) ([]Room, error) {
	var rooms []Room
	err := s.db.SelectContext(ctx, &rooms, "SELECT * FROM rooms WHERE status IN ('waiting','playing') ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (s *GameStore) GetScoreHistory(ctx context.Context, userID int64, limit int) ([]ScoreRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	var records []ScoreRecord
	err := s.db.SelectContext(ctx, &records, "SELECT * FROM scores WHERE user_id = ? ORDER BY created_at DESC LIMIT ?", userID, limit)
	return records, err
}

// SaveChatMessage persists a chat message.
func (s *GameStore) SaveChatMessage(ctx context.Context, msg *ChatMessage) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO chat_messages (room_id, user_id, content, msg_type) VALUES (?, ?, ?, ?)",
		msg.RoomID, msg.UserID, msg.Content, msg.MsgType)
	return err
}

// CreateGameRecord creates a new game record and returns its ID.
func (s *GameStore) CreateGameRecord(ctx context.Context, record *GameRecord) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO game_records (room_id, game_type, round_num) VALUES (?, ?, ?)",
		record.RoomID, record.GameType, record.RoundNum)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// EndGameRecord updates a game record with the final result.
func (s *GameStore) EndGameRecord(ctx context.Context, gameID int64, resultJSON string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE game_records SET result = ? WHERE id = ?", resultJSON, gameID)
	return err
}

// AddGameAction records a single game action.
func (s *GameStore) AddGameAction(ctx context.Context, action *GameAction) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO game_actions (game_id, round_num, action_seq, player_id, seat_index, is_bot, action_type, cards, full_state)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		action.GameID, action.RoundNum, action.ActionSeq, action.PlayerID,
		action.SeatIndex, action.IsBot, action.ActionType, action.Cards, action.FullState)
	return err
}
