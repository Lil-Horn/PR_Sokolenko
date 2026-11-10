package reposit

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

type SessionRepository struct {
	db  *sql.DB
	ttl time.Duration
}

func NewSessionRepository(db *sql.DB, ttl time.Duration) *SessionRepository {
	return &SessionRepository{db: db, ttl: ttl}
}

func (r *SessionRepository) TTL() time.Duration { return r.ttl }

// Create генерирует криптографически случайный идентификатор сессии и сохраняет его в БД.
func (r *SessionRepository) Create(userID int) (string, time.Time, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(buf)
	expires := time.Now().Add(r.ttl)

	_, err := r.db.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expires.Unix(),
	)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

// GetUserID возвращает id пользователя, если сессия существует И не просрочена.
func (r *SessionRepository) GetUserID(token string) (int, error) {
	var userID int
	err := r.db.QueryRow(
		"SELECT user_id FROM sessions WHERE session_id = ? AND expires_at > ?",
		token, time.Now().Unix(),
	).Scan(&userID)
	return userID, err
}

func (r *SessionRepository) Delete(token string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE session_id = ?", token)
	return err
}

// DeleteExpired чистит просроченные сессии.
func (r *SessionRepository) DeleteExpired() error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE expires_at <= ?", time.Now().Unix())
	return err
}
