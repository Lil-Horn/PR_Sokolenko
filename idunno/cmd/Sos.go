package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"idunno/internal/handlers"
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

func main() {
	db, err := sql.Open("sqlite", "expense.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	migration, err := os.ReadFile("migrations/create_expense.sql")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(string(migration))
	if err != nil {
		panic(err)
	}

	repo := reposit.NewExpenseRepository(db)
	h := handlers.New(repo)

	fs := http.FileServer(http.Dir("webntemp/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/add", h.Add)
	http.HandleFunc("/edit", h.EditPage)
	http.HandleFunc("/update", h.Update)
	http.HandleFunc("/delete", h.Delete)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/ping", pingHandler)

	fmt.Println("Сервер запущен. http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
