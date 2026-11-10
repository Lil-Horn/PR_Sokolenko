package reposit

import (
	"database/sql"
	"errors"

	"idunno/internal/models"
)

var ErrEmailTaken = errors.New("email уже зарегистрирован")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create сохраняет пользователя. В passwordHash передаётся ТОЛЬКО bcrypt-хэш.
func (r *UserRepository) Create(email, passwordHash string) (int, error) {
	res, err := r.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, passwordHash,
	)
	if err != nil {
		// нарушение UNIQUE(email) (в т.ч. гонка двух одновременных регистраций)
		if existing, e := r.GetByEmail(email); e == nil && existing.ID > 0 {
			return 0, ErrEmailTaken
		}
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *UserRepository) GetByEmail(email string) (models.User, error) {
	var u models.User
	err := r.db.QueryRow(
		"SELECT id, email, password_hash FROM users WHERE email = ?", email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	return u, err
}
