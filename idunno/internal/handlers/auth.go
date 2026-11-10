package handlers

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"idunno/internal/middleware"
	"idunno/internal/reposit"
)

const (
	minPasswordLen = 6
	maxPasswordLen = 72 // ограничение алгоритма bcrypt (в байтах)
)

type authData struct {
	Page
	Email string
}

func (h *Handler) alreadyLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := middleware.UserFromRequest(h.sessions, r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return true
	}
	return false
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if h.alreadyLoggedIn(w, r) {
		return
	}
	render(w, "register.html", authData{})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	fail := func(msg string) {
		w.WriteHeader(http.StatusBadRequest)
		render(w, "register.html", authData{Page: Page{Error: msg}, Email: email})
	}

	if _, err := mail.ParseAddress(email); err != nil {
		fail("Введите корректный email")
		return
	}
	if len(password) < minPasswordLen {
		fail("Пароль должен быть не короче 6 символов")
		return
	}
	if len(password) > maxPasswordLen {
		fail("Пароль слишком длинный (максимум 72 байта)")
		return
	}
	if password != password2 {
		fail("Пароли не совпадают")
		return
	}

	// В БД попадает только bcrypt-хэш (с солью внутри), сам пароль не хранится
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка обработки пароля", http.StatusInternalServerError)
		return
	}

	if _, err := h.users.Create(email, string(hash)); err != nil {
		if err == reposit.ErrEmailTaken {
			fail("Пользователь с таким email уже существует")
			return
		}
		http.Error(w, "Ошибка записи в базу", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if h.alreadyLoggedIn(w, r) {
		return
	}
	d := authData{}
	if r.URL.Query().Get("registered") == "1" {
		d.Info = "Регистрация прошла успешно. Теперь войдите."
	}
	render(w, "login.html", d)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	fail := func() {
		// Одно и то же сообщение для "нет такого email" и "неверный пароль" —
		// чтобы нельзя было перебором узнать, какие email зарегистрированы
		w.WriteHeader(http.StatusUnauthorized)
		render(w, "login.html", authData{Page: Page{Error: "Неверный email или пароль"}, Email: email})
	}

	user, err := h.users.GetByEmail(email)
	if err != nil {
		fail()
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		fail()
		return
	}

	_ = h.sessions.DeleteExpired()
	token, expires, err := h.sessions.Create(user.ID)
	if err != nil {
		http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,                 // недоступна из JavaScript (защита от кражи через XSS)
		SameSite: http.SameSiteLaxMode, // защита от CSRF
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(middleware.CookieName); err == nil {
		_ = h.sessions.Delete(c.Value) // инвалидируем сессию на сервере
	}
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
