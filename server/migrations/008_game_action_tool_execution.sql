-- 008_game_action_tool_execution.sql
-- Add nullable tool_execution_id to game_actions so AI tool executions can be traced.
-- No FK constraint — purely a loose reference for audit lookups.

ALTER TABLE game_actions
  ADD COLUMN tool_execution_id BIGINT NULL AFTER full_state,
  ADD INDEX idx_tool_exec (tool_execution_id);
