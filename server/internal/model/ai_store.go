package model

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type AICharacter struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	AvatarURL   *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	Personality *string   `db:"personality" json:"personality,omitempty"`
	PlayStyle   string    `db:"play_style" json:"play_style"`
	Catchphrase *string   `db:"catchphrase" json:"catchphrase,omitempty"`
	Occupation  *string   `db:"occupation" json:"occupation,omitempty"`
	Voice       *string   `db:"voice" json:"voice,omitempty"`
	Greeting    *string   `db:"greeting" json:"greeting,omitempty"`
	Enabled     bool      `db:"enabled" json:"enabled"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type LLMConfig struct {
	ID          int       `db:"id" json:"id"`
	Provider    string    `db:"provider" json:"provider"`
	APIKey      string    `db:"api_key" json:"api_key,omitempty"`
	APIURL      *string   `db:"api_url" json:"api_url,omitempty"`
	Model       string    `db:"model" json:"model"`
	Temperature float64   `db:"temperature" json:"temperature"`
	MaxTokens   int       `db:"max_tokens" json:"max_tokens"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type ChatMessage struct {
	ID        int64     `db:"id" json:"id"`
	RoomID    string    `db:"room_id" json:"room_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Content   string    `db:"content" json:"content"`
	MsgType   string    `db:"msg_type" json:"msg_type"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type LLMCallLog struct {
	ID               int64     `db:"id" json:"id"`
	Provider         string    `db:"provider" json:"provider"`
	Model            string    `db:"model" json:"model"`
	PromptTokens     int       `db:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int       `db:"completion_tokens" json:"completion_tokens"`
	DurationMs       int       `db:"duration_ms" json:"duration_ms"`
	Success          bool      `db:"success" json:"success"`
	ErrorMessage     *string   `db:"error_message" json:"error_message,omitempty"`
	CallType         string    `db:"call_type" json:"call_type"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type AIStore struct {
	db *sqlx.DB
}

func NewAIStore(db *sqlx.DB) *AIStore {
	return &AIStore{db: db}
}

// AI Character CRUD
func (s *AIStore) ListCharacters(ctx context.Context) ([]AICharacter, error) {
	var chars []AICharacter
	err := s.db.SelectContext(ctx, &chars, "SELECT * FROM ai_characters WHERE enabled = TRUE ORDER BY name")
	return chars, err
}

func (s *AIStore) GetAllCharacters(ctx context.Context) ([]AICharacter, error) {
	var chars []AICharacter
	err := s.db.SelectContext(ctx, &chars, "SELECT * FROM ai_characters ORDER BY name")
	return chars, err
}

func (s *AIStore) GetCharacter(ctx context.Context, id int) (*AICharacter, error) {
	var c AICharacter
	err := s.db.GetContext(ctx, &c, "SELECT * FROM ai_characters WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *AIStore) CreateCharacter(ctx context.Context, c *AICharacter) error {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO ai_characters (name, avatar_url, personality, play_style, catchphrase, occupation, voice, greeting, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Name, c.AvatarURL, c.Personality, c.PlayStyle, c.Catchphrase, c.Occupation, c.Voice, c.Greeting, c.Enabled)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	c.ID = int(id)
	return nil
}

func (s *AIStore) UpdateCharacter(ctx context.Context, c *AICharacter) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE ai_characters SET name=?, avatar_url=?, personality=?, play_style=?, catchphrase=?, occupation=?, voice=?, greeting=?, enabled=? WHERE id=?`,
		c.Name, c.AvatarURL, c.Personality, c.PlayStyle, c.Catchphrase, c.Occupation, c.Voice, c.Greeting, c.Enabled, c.ID)
	return err
}

func (s *AIStore) DeleteCharacter(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM ai_characters WHERE id = ?", id)
	return err
}

// LLM Config CRUD
func (s *AIStore) GetActiveConfig(ctx context.Context) (*LLMConfig, error) {
	var cfg LLMConfig
	err := s.db.GetContext(ctx, &cfg, "SELECT * FROM llm_configs WHERE is_active = TRUE LIMIT 1")
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *AIStore) ListConfigs(ctx context.Context) ([]LLMConfig, error) {
	var configs []LLMConfig
	err := s.db.SelectContext(ctx, &configs, "SELECT * FROM llm_configs ORDER BY created_at DESC")
	return configs, err
}

func (s *AIStore) SaveConfig(ctx context.Context, cfg *LLMConfig) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if cfg.IsActive {
		if _, err = tx.ExecContext(ctx, "UPDATE llm_configs SET is_active = FALSE WHERE is_active = TRUE"); err != nil {
			return err
		}
	}
	result, err := tx.ExecContext(ctx,
		`INSERT INTO llm_configs (provider, api_key, api_url, model, temperature, max_tokens, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		cfg.Provider, cfg.APIKey, cfg.APIURL, cfg.Model, cfg.Temperature, cfg.MaxTokens, cfg.IsActive)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	cfg.ID = int(id)
	return tx.Commit()
}

func (s *AIStore) DeleteConfig(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM llm_configs WHERE id = ?", id)
	return err
}

// Chat Messages
func (s *AIStore) SaveChatMessage(ctx context.Context, msg *ChatMessage) error {
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO chat_messages (room_id, user_id, content, msg_type) VALUES (?, ?, ?, ?)",
		msg.RoomID, msg.UserID, msg.Content, msg.MsgType)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	msg.ID = id
	return nil
}

func (s *AIStore) GetChatHistory(ctx context.Context, roomID string, limit int) ([]ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	var msgs []ChatMessage
	err := s.db.SelectContext(ctx, &msgs,
		"SELECT * FROM chat_messages WHERE room_id = ? ORDER BY created_at DESC LIMIT ?", roomID, limit)
	return msgs, err
}

// LLM Call Log
func (s *AIStore) LogLLMCall(ctx context.Context, log *LLMCallLog) error {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO llm_call_logs (provider, model, prompt_tokens, completion_tokens, duration_ms, success, error_message, call_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		log.Provider, log.Model, log.PromptTokens, log.CompletionTokens, log.DurationMs, log.Success, log.ErrorMessage, log.CallType)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	log.ID = id
	return nil
}

type LLMStatsSummary struct {
	TotalCalls   int64   `db:"total_calls" json:"total_calls"`
	TotalTokens  int64   `db:"total_tokens" json:"total_tokens"`
	AvgLatencyMs float64 `db:"avg_latency_ms" json:"avg_latency_ms"`
	SuccessRate  float64 `db:"success_rate" json:"success_rate"`
}

func (s *AIStore) GetLLMStats(ctx context.Context) (*LLMStatsSummary, error) {
	var stats LLMStatsSummary
	err := s.db.GetContext(ctx, &stats, `
		SELECT
			COUNT(*) as total_calls,
			COALESCE(SUM(prompt_tokens + completion_tokens), 0) as total_tokens,
			COALESCE(ROUND(AVG(duration_ms)), 0) as avg_latency_ms,
			COALESCE(ROUND(SUM(success) / COUNT(*) * 100), 0) as success_rate
		FROM llm_call_logs
	`)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
