ALTER TABLE users
  ADD COLUMN `character_id` VARCHAR(32) NOT NULL DEFAULT 'panda' AFTER `nickname`;

ALTER TABLE rooms
  ADD COLUMN `theme` VARCHAR(32) NOT NULL DEFAULT 'classic-poker' AFTER `name`;
