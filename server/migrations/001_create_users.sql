CREATE TABLE IF NOT EXISTS users (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    nickname    VARCHAR(32) NOT NULL,
    avatar_url  VARCHAR(256),
    role        ENUM('user','admin') DEFAULT 'user',
    status      TINYINT DEFAULT 1,
    created_at  DATETIME DEFAULT NOW(),
    updated_at  DATETIME DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_auths (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id      BIGINT NOT NULL,
    provider     ENUM('google','wechat','phone','guest'),
    provider_uid VARCHAR(128) NOT NULL,
    credential   VARCHAR(256),
    UNIQUE KEY uk_provider (provider, provider_uid)
);
