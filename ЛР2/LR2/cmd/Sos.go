package main

import (
	"fmt"
	"net/http"

	"Laba2/internal/handlers"
)

func main() {
	fileserv := http.FileServer(http.Dir("webntemp/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileserv))

	http.HandleFunc("/", handlers.Index)
	http.HandleFunc("/add", handlers.Add)

	fmt.Println("Сервер запущен. http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
