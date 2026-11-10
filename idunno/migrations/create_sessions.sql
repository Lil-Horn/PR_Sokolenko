CREATE TABLE IF NOT EXISTS sessions (
    session_id TEXT    PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL  -- unix-время (секунды), после него сессия недействительна
);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
