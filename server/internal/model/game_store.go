package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Room struct {
	ID         string     `db:"id" json:"id"`
	GameType   string     `db:"game_type" json:"game_type"`
	OwnerID    int64      `db:"owner_id" json:"owner_id"`
	Status     string     `db:"status" json:"status"`
	MaxPlayers int8       `db:"max_players" json:"max_players"`
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

type GameSnapshot struct {
	ID         int64     `db:"id" json:"id"`
	RoomID     string    `db:"room_id" json:"room_id"`
	GameID     int64     `db:"game_id" json:"game_id"`
	SnapshotAt time.Time `db:"snapshot_at" json:"snapshot_at"`
	FullState  string    `db:"full_state" json:"full_state"`
	IsCurrent  bool      `db:"is_current" json:"is_current"`
}

type GameStore struct {
	db *sqlx.DB
}

func NewGameStore(db *sqlx.DB) *GameStore {
	return &GameStore{db: db}
}

func (s *GameStore) CreateRoom(ctx context.Context, room *Room) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO rooms (id, game_type, owner_id, max_players, bot_enabled) VALUES (?, ?, ?, ?, ?)",
		room.ID, room.GameType, room.OwnerID, room.MaxPlayers, room.BotEnabled)
	return err
}

func (s *GameStore) UpdateRoomStatus(ctx context.Context, roomID, status string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE rooms SET status = ? WHERE id = ?", status, roomID)
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

func (s *GameStore) SaveSnapshot(ctx context.Context, snap *GameSnapshot) error {
	s.db.ExecContext(ctx, "UPDATE game_snapshots SET is_current = FALSE WHERE room_id = ? AND game_id = ?", snap.RoomID, snap.GameID)
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO game_snapshots (room_id, game_id, full_state, is_current) VALUES (?, ?, ?, TRUE)",
		snap.RoomID, snap.GameID, snap.FullState)
	return err
}

func (s *GameStore) GetLatestSnapshot(ctx context.Context, roomID string) (*GameSnapshot, error) {
	var snap GameSnapshot
	err := s.db.GetContext(ctx, &snap,
		"SELECT * FROM game_snapshots WHERE room_id = ? AND is_current = TRUE ORDER BY snapshot_at DESC LIMIT 1", roomID)
	if err != nil {
		return nil, err
	}
	return &snap, nil
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
