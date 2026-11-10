package middleware

import (
	"context"
	"net/http"
)

const CookieName = "session_id"

type ctxKey struct{}

// SessionStore — то, что нужно middleware от хранилища сессий.
type SessionStore interface {
	GetUserID(token string) (int, error)
}

// UserFromRequest достаёт id пользователя по cookie (0, false — не авторизован / сессия просрочена).
func UserFromRequest(s SessionStore, r *http.Request) (int, bool) {
	c, err := r.Cookie(CookieName)
	if err != nil || c.Value == "" {
		return 0, false
	}
	id, err := s.GetUserID(c.Value)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}

// RequireAuth — middleware: пропускает только авторизованных,
// остальных перенаправляет на /login. ID пользователя кладёт в context запроса.
func RequireAuth(s SessionStore) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, ok := UserFromRequest(s, r)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, id)
			next(w, r.WithContext(ctx))
		}
	}
}

// UserID достаёт id авторизованного пользователя из контекста (в обработчиках за RequireAuth).
func UserID(ctx context.Context) int {
	id, _ := ctx.Value(ctxKey{}).(int)
	return id
}
