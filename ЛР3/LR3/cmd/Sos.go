package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"Laba2/internal/handlers"
	"Laba2/internal/reposit"

	_ "modernc.org/sqlite"
)

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

	fmt.Println("Сервер запущен. http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
