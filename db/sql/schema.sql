CREATE TABLE IF NOT EXISTS users (
    user_id      TEXT PRIMARY KEY,
    lastfm_username TEXT NOT NULL,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_lastfm_username
ON users(lastfm_username);
