-- Make users.nickname unique to prevent duplicate accounts from race conditions
-- Handles existing duplicates by appending the user id as suffix

START TRANSACTION;

-- Deduplicate existing nicknames by appending _<id> to all but the oldest account
UPDATE users u
INNER JOIN (
    SELECT nickname, MIN(id) as keep_id
    FROM users
    GROUP BY nickname
    HAVING COUNT(*) > 1
) dups ON u.nickname = dups.nickname AND u.id != dups.keep_id
SET u.nickname = CONCAT(u.nickname, '_', u.id);

ALTER TABLE users ADD UNIQUE INDEX uk_nickname (nickname);

COMMIT;
