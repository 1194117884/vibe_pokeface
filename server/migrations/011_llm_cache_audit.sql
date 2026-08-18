-- Store LLM cache audit metadata for provider-side prompt cache analysis.

ALTER TABLE llm_call_logs
  ADD COLUMN request_hash VARCHAR(64) AFTER turn_number,
  ADD COLUMN prompt_cache_hit_tokens INT DEFAULT 0 AFTER completion_tokens,
  ADD COLUMN prompt_cache_miss_tokens INT DEFAULT 0 AFTER prompt_cache_hit_tokens,
  ADD COLUMN raw_response_json MEDIUMTEXT AFTER response_json,
  ADD INDEX idx_llm_request_hash (request_hash);

UPDATE llm_call_logs
   SET request_hash = SHA2(request_json, 256)
 WHERE request_hash IS NULL
   AND request_json IS NOT NULL
   AND request_json <> '';
