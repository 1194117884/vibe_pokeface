-- 009_registration_codes: invite-only registration codes
CREATE TABLE IF NOT EXISTS registration_codes (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    code            VARCHAR(6) NOT NULL,
    is_used         BOOLEAN DEFAULT FALSE,
    is_disabled     BOOLEAN DEFAULT FALSE,
    used_by_user_id BIGINT NULL,
    used_at         DATETIME NULL,
    created_by      BIGINT NOT NULL,
    note            VARCHAR(255) DEFAULT '',
    created_at      DATETIME DEFAULT NOW(),
    updated_at      DATETIME DEFAULT NOW(),
    UNIQUE KEY uk_code (code),
    INDEX idx_used (is_used),
    INDEX idx_disabled (is_disabled)
);
