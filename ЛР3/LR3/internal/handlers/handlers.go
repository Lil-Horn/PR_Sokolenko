package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"Laba2/internal/models"
	"Laba2/internal/reposit"
)

type Handler struct {
	repo *reposit.ExpenseRepository
}

func New(repo *reposit.ExpenseRepository) *Handler {
	return &Handler{repo: repo}
}

func Money(amount float64) string {
	s := strconv.FormatFloat(amount, 'f', 2, 64)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]

	var result []string
	for len(intPart) > 3 {
		result = append([]string{intPart[len(intPart)-3:]}, result...)
		intPart = intPart[:len(intPart)-3]
	}
	result = append([]string{intPart}, result...)

	return strings.Join(result, " ") + "," + decPart
}

func parseTemplates() (*template.Template, error) {
	return template.New("").Funcs(template.FuncMap{
		"money": Money,
	}).ParseFiles(
		"webntemp/layout.html",
		"webntemp/index.html",
		"webntemp/edit.html",
	)
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка чтения из базы", 500)
		return
	}

	temp, err := parseTemplates()
	if err != nil {
		http.Error(w, "Невозможно открыть шаблон", 500)
		return
	}
	temp.ExecuteTemplate(w, "layout", expenses)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	desc := r.FormValue("description")
	date := r.FormValue("date")

	_, err := h.repo.Create(models.Expense{
		Amount:      amount,
		Description: desc,
		Date:        date,
	})
	if err != nil {
		http.Error(w, "Ошибка записи в базу", 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) EditPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	expense, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Трата не найдена", 404)
		return
	}

	temp, err := template.ParseFiles("webntemp/edit.html")
	if err != nil {
		http.Error(w, "Невозможно открыть шаблон", 500)
		return
	}
	temp.Execute(w, expense)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id, _ := strconv.Atoi(r.FormValue("id"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	desc := r.FormValue("description")
	date := r.FormValue("date")

	err := h.repo.Update(models.Expense{
		ID:          id,
		Amount:      amount,
		Description: desc,
		Date:        date,
	})
	if err != nil {
		http.Error(w, "Ошибка обновления", 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	err := h.repo.Delete(id)
	if err != nil {
		http.Error(w, "Ошибка удаления", 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
