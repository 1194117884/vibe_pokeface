CREATE TABLE IF NOT EXISTS ai_characters (
    id              INT PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(32) NOT NULL,
    avatar_url      VARCHAR(256),
    personality     TEXT,          -- character description
    play_style      VARCHAR(16) DEFAULT 'balanced',  -- aggressive/balanced/conservative
    catchphrase     VARCHAR(128),  -- 口头禅
    occupation      VARCHAR(64),
    voice           VARCHAR(64),   -- voice type for TTS
    greeting        TEXT,          -- initial chat message on game start
    enabled         BOOLEAN DEFAULT TRUE,
    created_at      DATETIME DEFAULT NOW(),
    updated_at      DATETIME DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS llm_configs (
    id              INT PRIMARY KEY AUTO_INCREMENT,
    provider        VARCHAR(32) NOT NULL,  -- 'openai' | 'anthropic' | 'custom'
    api_key         VARCHAR(512) NOT NULL,
    api_url         VARCHAR(256),           -- custom endpoint
    model           VARCHAR(64) NOT NULL,   -- 'gpt-4o', 'claude-sonnet-4', etc.
    temperature     DECIMAL(3,2) DEFAULT 0.7,
    max_tokens      INT DEFAULT 1024,
    is_active       BOOLEAN DEFAULT FALSE,
    created_at      DATETIME DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id         VARCHAR(32) NOT NULL,
    user_id         VARCHAR(64) NOT NULL,  -- could be "ai:char_id" for bots
    content         TEXT NOT NULL,
    msg_type        VARCHAR(16) DEFAULT 'text',  -- text/emoji/system
    created_at      DATETIME(3) DEFAULT NOW(3),
    INDEX idx_room (room_id, created_at)
);

CREATE TABLE IF NOT EXISTS llm_call_logs (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    provider        VARCHAR(32) NOT NULL,
    model           VARCHAR(64) NOT NULL,
    prompt_tokens   INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    duration_ms     INT DEFAULT 0,
    success         BOOLEAN DEFAULT TRUE,
    error_message   TEXT,
    call_type       VARCHAR(32) NOT NULL,  -- 'play_decision' | 'chat' | 'greeting'
    created_at      DATETIME DEFAULT NOW()
);
