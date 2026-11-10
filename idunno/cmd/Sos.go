package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"idunno/internal/handlers"
	"idunno/internal/middleware"
	"idunno/internal/reposit"

	_ "modernc.org/sqlite"
)

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "О проекте: Это простой сайтик, написанный с использованием пакета net/http.")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
}

// ensureUserIDColumn: для БД, созданных на этапах 3-4 (без user_id) добавляет колонку.
func ensureUserIDColumn(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(expenses)")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "user_id" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	migration, err := os.ReadFile("migrations/add_expenses_user_id.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(migration))
	return err
}

func main() {
	db, err := sql.Open("sqlite", "expense.db?_pragma=foreign_keys(1)")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// порядок важен: users и categories должны существовать до expenses/sessions
	for _, file := range []string{
		"migrations/create_categories.sql",
		"migrations/create_users.sql",
		"migrations/create_sessions.sql",
		"migrations/create_expense.sql",
	} {
		migration, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		if _, err = db.Exec(string(migration)); err != nil {
			panic(err)
		}
	}
	if err := ensureUserIDColumn(db); err != nil {
		panic(err)
	}

	repo := reposit.NewExpenseRepository(db)
	users := reposit.NewUserRepository(db)
	sessions := reposit.NewSessionRepository(db, 24*time.Hour)
	_ = sessions.DeleteExpired()
	h := handlers.New(repo, users, sessions)

	auth := middleware.RequireAuth(sessions)

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("webntemp/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// публичные маршруты
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("GET /register", h.RegisterPage)
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)

	// защищённые маршруты (только для авторизованных)
	mux.HandleFunc("GET /{$}", auth(h.Index))
	mux.HandleFunc("/add", auth(h.Add))
	mux.HandleFunc("/edit", auth(h.EditPage))
	mux.HandleFunc("/update", auth(h.Update))
	mux.HandleFunc("/delete", auth(h.Delete))

	fmt.Println("Сервер запущен. http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
