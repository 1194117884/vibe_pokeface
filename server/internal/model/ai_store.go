package model

import (
	"context"
	"strings"
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
	RoomID           string    `db:"room_id" json:"room_id,omitempty"`
	UserID           string    `db:"user_id" json:"user_id,omitempty"`
	Seat             int       `db:"seat" json:"seat,omitempty"`
	Phase            string    `db:"phase" json:"phase,omitempty"`
	TurnNumber       int       `db:"turn_number" json:"turn_number,omitempty"`
	RequestJSON      *string   `db:"request_json" json:"request_json,omitempty"`
	ResponseJSON     *string   `db:"response_json" json:"response_json,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type AiToolExecution struct {
	ID         int64     `db:"id" json:"id"`
	CallLogID  int64     `db:"call_log_id" json:"call_log_id"`
	ToolName   string    `db:"tool_name" json:"tool_name"`
	ToolType   string    `db:"tool_type" json:"tool_type"`
	ArgsJSON   *string   `db:"args_json" json:"args_json,omitempty"`
	ResultJSON *string   `db:"result_json" json:"result_json,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type AIRunListItem struct {
	ID               int64     `db:"id" json:"id"`
	Provider         string    `db:"provider" json:"provider"`
	Model            string    `db:"model" json:"model"`
	PromptTokens     int       `db:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int       `db:"completion_tokens" json:"completion_tokens"`
	DurationMs       int       `db:"duration_ms" json:"duration_ms"`
	Success          bool      `db:"success" json:"success"`
	ErrorMessage     *string   `db:"error_message" json:"error_message,omitempty"`
	CallType         string    `db:"call_type" json:"call_type"`
	RoomID           string    `db:"room_id" json:"room_id,omitempty"`
	UserID           string    `db:"user_id" json:"user_id,omitempty"`
	Seat             int       `db:"seat" json:"seat"`
	Phase            string    `db:"phase" json:"phase,omitempty"`
	TurnNumber       int       `db:"turn_number" json:"turn_number"`
	ToolCount        int       `db:"tool_count" json:"tool_count"`
	ActionCount      int       `db:"action_count" json:"action_count"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type AIRunDetail struct {
	LLMCallLog
	Tools       []AiToolExecution `json:"tools"`
	GameActions []GameAction      `json:"game_actions"`
}

type AIRunFilter struct {
	RoomID  string
	UserID  string
	Game    string
	Phase   string
	Success *bool
	Limit   int
	Offset  int
}

type AIPromptTemplate struct {
	ID            int64      `db:"id" json:"id"`
	GameType      string     `db:"game_type" json:"game_type"`
	Phase         string     `db:"phase" json:"phase"`
	TemplateType  string     `db:"template_type" json:"template_type"`
	Name          string     `db:"name" json:"name"`
	Content       string     `db:"content" json:"content"`
	VariablesJSON *string    `db:"variables_json" json:"variables_json,omitempty"`
	Status        string     `db:"status" json:"status"`
	Version       int        `db:"version" json:"version"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	PublishedAt   *time.Time `db:"published_at" json:"published_at,omitempty"`
}

type AIToolTemplate struct {
	ID             int64      `db:"id" json:"id"`
	GameType       string     `db:"game_type" json:"game_type"`
	Phase          string     `db:"phase" json:"phase"`
	ToolName       string     `db:"tool_name" json:"tool_name"`
	Enabled        bool       `db:"enabled" json:"enabled"`
	Description    *string    `db:"description" json:"description,omitempty"`
	ParametersJSON *string    `db:"parameters_json" json:"parameters_json,omitempty"`
	Status         string     `db:"status" json:"status"`
	Version        int        `db:"version" json:"version"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
	PublishedAt    *time.Time `db:"published_at" json:"published_at,omitempty"`
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
func (s *AIStore) LogLLMCall(ctx context.Context, log *LLMCallLog) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO llm_call_logs
		 (provider, model, prompt_tokens, completion_tokens, duration_ms, success,
		  error_message, call_type, room_id, user_id, seat, phase, turn_number,
		  request_json, response_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.Provider, log.Model, log.PromptTokens, log.CompletionTokens,
		log.DurationMs, log.Success, log.ErrorMessage, log.CallType,
		log.RoomID, log.UserID, log.Seat, log.Phase, log.TurnNumber,
		log.RequestJSON, log.ResponseJSON)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	log.ID = id
	return id, nil
}

func (s *AIStore) LogToolExecution(ctx context.Context, exec *AiToolExecution) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO ai_tool_executions (call_log_id, tool_name, tool_type, args_json, result_json)
		 VALUES (?, ?, ?, ?, ?)`,
		exec.CallLogID, exec.ToolName, exec.ToolType, exec.ArgsJSON, exec.ResultJSON)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	exec.ID = id
	return id, nil
}

func (s *AIStore) ListAIRuns(ctx context.Context, filter AIRunFilter) ([]AIRunListItem, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	where := []string{"1=1"}
	args := []interface{}{}
	if filter.RoomID != "" {
		where = append(where, "l.room_id = ?")
		args = append(args, filter.RoomID)
	}
	if filter.UserID != "" {
		where = append(where, "l.user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Phase != "" {
		where = append(where, "l.phase = ?")
		args = append(args, filter.Phase)
	}
	if filter.Success != nil {
		where = append(where, "l.success = ?")
		args = append(args, *filter.Success)
	}
	if filter.Game != "" {
		where = append(where, "gr.game_type = ?")
		args = append(args, filter.Game)
	}
	args = append(args, filter.Limit, filter.Offset)
	var runs []AIRunListItem
	query := `
		SELECT l.id, l.provider, l.model, l.prompt_tokens, l.completion_tokens,
		       l.duration_ms, l.success, l.error_message, l.call_type, l.room_id,
		       l.user_id, l.seat, l.phase, l.turn_number, l.created_at,
		       COUNT(DISTINCT t.id) AS tool_count,
		       COUNT(DISTINCT ga.id) AS action_count
		  FROM llm_call_logs l
		  LEFT JOIN ai_tool_executions t ON t.call_log_id = l.id
		  LEFT JOIN game_actions ga ON ga.tool_execution_id = t.id
		  LEFT JOIN game_records gr ON gr.id = ga.game_id
		 WHERE ` + strings.Join(where, " AND ") + `
		 GROUP BY l.id
		 ORDER BY l.created_at DESC
		 LIMIT ? OFFSET ?`
	err := s.db.SelectContext(ctx, &runs, query, args...)
	return runs, err
}

func (s *AIStore) GetAIRunDetail(ctx context.Context, id int64) (*AIRunDetail, error) {
	var log LLMCallLog
	if err := s.db.GetContext(ctx, &log, "SELECT * FROM llm_call_logs WHERE id = ?", id); err != nil {
		return nil, err
	}
	var tools []AiToolExecution
	if err := s.db.SelectContext(ctx, &tools, "SELECT * FROM ai_tool_executions WHERE call_log_id = ? ORDER BY id", id); err != nil {
		return nil, err
	}
	var actions []GameAction
	if len(tools) > 0 {
		ids := make([]int64, len(tools))
		for i, tool := range tools {
			ids[i] = tool.ID
		}
		query, args, err := sqlx.In("SELECT * FROM game_actions WHERE tool_execution_id IN (?) ORDER BY action_seq", ids)
		if err != nil {
			return nil, err
		}
		query = s.db.Rebind(query)
		if err := s.db.SelectContext(ctx, &actions, query, args...); err != nil {
			return nil, err
		}
	}
	return &AIRunDetail{LLMCallLog: log, Tools: tools, GameActions: actions}, nil
}

func (s *AIStore) ListPromptTemplates(ctx context.Context, gameType, phase string) ([]AIPromptTemplate, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if gameType != "" {
		where = append(where, "game_type = ?")
		args = append(args, gameType)
	}
	if phase != "" {
		where = append(where, "phase = ?")
		args = append(args, phase)
	}
	var templates []AIPromptTemplate
	err := s.db.SelectContext(ctx, &templates,
		"SELECT * FROM ai_prompt_templates WHERE "+strings.Join(where, " AND ")+" ORDER BY game_type, phase, template_type, version DESC",
		args...)
	return templates, err
}

func (s *AIStore) SavePromptTemplate(ctx context.Context, t *AIPromptTemplate) error {
	if t.Status == "" {
		t.Status = "draft"
	}
	if t.ID > 0 {
		_, err := s.db.ExecContext(ctx,
			`UPDATE ai_prompt_templates SET game_type=?, phase=?, template_type=?, name=?, content=?, variables_json=?, status=?, updated_at=NOW() WHERE id=?`,
			t.GameType, t.Phase, t.TemplateType, t.Name, t.Content, t.VariablesJSON, t.Status, t.ID)
		return err
	}
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO ai_prompt_templates (game_type, phase, template_type, name, content, variables_json, status, version)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
		t.GameType, t.Phase, t.TemplateType, t.Name, t.Content, t.VariablesJSON, t.Status)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	t.ID = id
	t.Version = 1
	return nil
}

func (s *AIStore) PublishPromptTemplate(ctx context.Context, id int64) (*AIPromptTemplate, error) {
	var src AIPromptTemplate
	if err := s.db.GetContext(ctx, &src, "SELECT * FROM ai_prompt_templates WHERE id = ?", id); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`UPDATE ai_prompt_templates SET status='archived', updated_at=NOW()
		  WHERE game_type=? AND phase=? AND template_type=? AND status='published'`,
		src.GameType, src.Phase, src.TemplateType); err != nil {
		return nil, err
	}
	var maxVersion int
	if err := tx.GetContext(ctx, &maxVersion,
		`SELECT COALESCE(MAX(version), 0) FROM ai_prompt_templates WHERE game_type=? AND phase=? AND template_type=?`,
		src.GameType, src.Phase, src.TemplateType); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx,
		`INSERT INTO ai_prompt_templates (game_type, phase, template_type, name, content, variables_json, status, version, published_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'published', ?, NOW())`,
		src.GameType, src.Phase, src.TemplateType, src.Name, src.Content, src.VariablesJSON, maxVersion+1)
	if err != nil {
		return nil, err
	}
	newID, _ := result.LastInsertId()
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	var published AIPromptTemplate
	if err := s.db.GetContext(ctx, &published, "SELECT * FROM ai_prompt_templates WHERE id = ?", newID); err != nil {
		return nil, err
	}
	return &published, nil
}

func (s *AIStore) GetPublishedPromptTemplate(ctx context.Context, gameType, phase, templateType string) (*AIPromptTemplate, error) {
	var t AIPromptTemplate
	err := s.db.GetContext(ctx, &t,
		`SELECT * FROM ai_prompt_templates
		  WHERE game_type=? AND phase=? AND template_type=? AND status='published'
		  ORDER BY version DESC LIMIT 1`,
		gameType, phase, templateType)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *AIStore) ListToolTemplates(ctx context.Context, gameType, phase string) ([]AIToolTemplate, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if gameType != "" {
		where = append(where, "game_type = ?")
		args = append(args, gameType)
	}
	if phase != "" {
		where = append(where, "phase = ?")
		args = append(args, phase)
	}
	var templates []AIToolTemplate
	err := s.db.SelectContext(ctx, &templates,
		"SELECT * FROM ai_tool_templates WHERE "+strings.Join(where, " AND ")+" ORDER BY game_type, phase, tool_name, version DESC",
		args...)
	return templates, err
}

func (s *AIStore) SaveToolTemplate(ctx context.Context, t *AIToolTemplate) error {
	if t.Status == "" {
		t.Status = "draft"
	}
	if t.ID > 0 {
		_, err := s.db.ExecContext(ctx,
			`UPDATE ai_tool_templates SET game_type=?, phase=?, tool_name=?, enabled=?, description=?, parameters_json=?, status=?, updated_at=NOW() WHERE id=?`,
			t.GameType, t.Phase, t.ToolName, t.Enabled, t.Description, t.ParametersJSON, t.Status, t.ID)
		return err
	}
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO ai_tool_templates (game_type, phase, tool_name, enabled, description, parameters_json, status, version)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
		t.GameType, t.Phase, t.ToolName, t.Enabled, t.Description, t.ParametersJSON, t.Status)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	t.ID = id
	t.Version = 1
	return nil
}

func (s *AIStore) PublishToolTemplate(ctx context.Context, id int64) (*AIToolTemplate, error) {
	var src AIToolTemplate
	if err := s.db.GetContext(ctx, &src, "SELECT * FROM ai_tool_templates WHERE id = ?", id); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`UPDATE ai_tool_templates SET status='archived', updated_at=NOW()
		  WHERE game_type=? AND phase=? AND tool_name=? AND status='published'`,
		src.GameType, src.Phase, src.ToolName); err != nil {
		return nil, err
	}
	var maxVersion int
	if err := tx.GetContext(ctx, &maxVersion,
		`SELECT COALESCE(MAX(version), 0) FROM ai_tool_templates WHERE game_type=? AND phase=? AND tool_name=?`,
		src.GameType, src.Phase, src.ToolName); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx,
		`INSERT INTO ai_tool_templates (game_type, phase, tool_name, enabled, description, parameters_json, status, version, published_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'published', ?, NOW())`,
		src.GameType, src.Phase, src.ToolName, src.Enabled, src.Description, src.ParametersJSON, maxVersion+1)
	if err != nil {
		return nil, err
	}
	newID, _ := result.LastInsertId()
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	var published AIToolTemplate
	if err := s.db.GetContext(ctx, &published, "SELECT * FROM ai_tool_templates WHERE id = ?", newID); err != nil {
		return nil, err
	}
	return &published, nil
}

func (s *AIStore) ListPublishedToolTemplates(ctx context.Context, gameType, phase string) ([]AIToolTemplate, error) {
	var tools []AIToolTemplate
	err := s.db.SelectContext(ctx, &tools,
		`SELECT t.* FROM ai_tool_templates t
		  JOIN (
		    SELECT game_type, phase, tool_name, MAX(version) AS version
		      FROM ai_tool_templates
		     WHERE game_type=? AND phase=? AND status='published'
		     GROUP BY game_type, phase, tool_name
		  ) latest ON latest.game_type=t.game_type AND latest.phase=t.phase AND latest.tool_name=t.tool_name AND latest.version=t.version
		 WHERE t.status='published'
		 ORDER BY t.tool_name`,
		gameType, phase)
	return tools, err
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
