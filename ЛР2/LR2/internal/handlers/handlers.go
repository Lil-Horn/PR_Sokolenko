package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"Laba2/internal/models"
)

func Index(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(
		"webntemp/layout.html",
		"webntemp/index.html",
	)
	if err != nil {
		http.Error(w, "Невозможно открыть шаблон", 500)
		return
	}
	temp.ExecuteTemplate(w, "layout", models.Expenses)

}

func Add(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	desc := r.FormValue("description")
	date := r.FormValue("date")

	models.Add(amount, desc, date)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
