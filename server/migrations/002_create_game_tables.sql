CREATE TABLE IF NOT EXISTS rooms (
    id           VARCHAR(32) PRIMARY KEY,
    game_type    VARCHAR(32) NOT NULL,
    owner_id     BIGINT NOT NULL,
    status       ENUM('waiting','playing','ended') DEFAULT 'waiting',
    max_players  TINYINT DEFAULT 3,
    bot_enabled  BOOLEAN DEFAULT TRUE,
    created_at   DATETIME DEFAULT NOW(),
    ended_at     DATETIME
);

CREATE TABLE IF NOT EXISTS room_players (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id       VARCHAR(32) NOT NULL,
    user_id       BIGINT,
    is_bot        BOOLEAN DEFAULT FALSE,
    character_id  INT,
    seat_index    TINYINT NOT NULL,
    score         INT DEFAULT 0,
    status        ENUM('ready','playing','left') DEFAULT 'ready',
    UNIQUE KEY uk_room_seat (room_id, seat_index)
);

CREATE TABLE IF NOT EXISTS game_records (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id     VARCHAR(32) NOT NULL,
    game_type   VARCHAR(32) NOT NULL,
    round_num   INT DEFAULT 1,
    state_data  JSON,
    result      JSON,
    created_at  DATETIME DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS game_snapshots (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id      VARCHAR(32) NOT NULL,
    game_id      BIGINT NOT NULL,
    snapshot_at  DATETIME(3) DEFAULT NOW(),
    full_state   JSON,
    is_current   BOOLEAN DEFAULT TRUE,
    INDEX idx_game (room_id, game_id)
);

CREATE TABLE IF NOT EXISTS game_actions (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    game_id      BIGINT NOT NULL,
    round_num    INT NOT NULL,
    action_seq   INT NOT NULL,
    player_id    BIGINT,
    seat_index   TINYINT NOT NULL,
    is_bot       BOOLEAN DEFAULT FALSE,
    action_type  VARCHAR(16) NOT NULL,
    cards        JSON,
    full_state   JSON,
    created_at   DATETIME(3) DEFAULT NOW(),
    INDEX idx_game (game_id, round_num, action_seq)
);

CREATE TABLE IF NOT EXISTS scores (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    game_type   VARCHAR(32) NOT NULL,
    amount      INT NOT NULL,
    balance     INT NOT NULL,
    reason      VARCHAR(64),
    created_at  DATETIME DEFAULT NOW(),
    INDEX idx_user (user_id, created_at)
);
