-- 010_ai_templates.sql
-- Admin-managed AI prompt/tool templates. Code defaults remain the fallback.

CREATE TABLE IF NOT EXISTS ai_prompt_templates (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    game_type       VARCHAR(32) NOT NULL,
    phase           VARCHAR(32) NOT NULL,
    template_type   VARCHAR(32) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    content         MEDIUMTEXT NOT NULL,
    variables_json  TEXT,
    status          VARCHAR(16) NOT NULL DEFAULT 'draft',
    version         INT NOT NULL DEFAULT 1,
    created_at      DATETIME DEFAULT NOW(),
    updated_at      DATETIME DEFAULT NOW(),
    published_at    DATETIME NULL,
    INDEX idx_prompt_lookup (game_type, phase, template_type, status, version)
);

CREATE TABLE IF NOT EXISTS ai_tool_templates (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    game_type       VARCHAR(32) NOT NULL,
    phase           VARCHAR(32) NOT NULL,
    tool_name       VARCHAR(64) NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    description     TEXT,
    parameters_json MEDIUMTEXT,
    status          VARCHAR(16) NOT NULL DEFAULT 'draft',
    version         INT NOT NULL DEFAULT 1,
    created_at      DATETIME DEFAULT NOW(),
    updated_at      DATETIME DEFAULT NOW(),
    published_at    DATETIME NULL,
    UNIQUE KEY uniq_tool_template_version (game_type, phase, tool_name, status, version),
    INDEX idx_tool_lookup (game_type, phase, status)
);
