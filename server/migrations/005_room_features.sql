ALTER TABLE rooms
  ADD COLUMN `name` VARCHAR(64) NOT NULL DEFAULT '' AFTER `id`,
  ADD COLUMN `is_open` BOOLEAN NOT NULL DEFAULT TRUE AFTER `max_players`,
  ADD COLUMN `password` VARCHAR(64) DEFAULT NULL AFTER `is_open`,
  ADD INDEX `idx_rooms_status` (`status`),
  ADD INDEX `idx_rooms_is_open` (`is_open`);
