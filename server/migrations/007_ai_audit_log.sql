-- 007_ai_audit_log.sql
-- Store full LLM request/response payloads and tool execution logs.

ALTER TABLE llm_call_logs
  ADD COLUMN room_id       VARCHAR(32)   AFTER call_type,
  ADD COLUMN user_id       VARCHAR(64)   AFTER room_id,
  ADD COLUMN seat          INT DEFAULT 0 AFTER user_id,
  ADD COLUMN phase         VARCHAR(16)   AFTER seat,
  ADD COLUMN turn_number   INT DEFAULT 0 AFTER phase,
  ADD COLUMN request_json  MEDIUMTEXT    AFTER turn_number,
  ADD COLUMN response_json MEDIUMTEXT    AFTER request_json;

CREATE TABLE IF NOT EXISTS ai_tool_executions (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    call_log_id     BIGINT NOT NULL,
    tool_name       VARCHAR(32) NOT NULL,
    tool_type       VARCHAR(8) NOT NULL DEFAULT 'action',
    args_json       TEXT,
    result_json     TEXT,
    created_at      DATETIME DEFAULT NOW(),
    INDEX idx_call_log (call_log_id)
);
