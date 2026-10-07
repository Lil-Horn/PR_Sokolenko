package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"idunno/internal/models"
	"idunno/internal/reposit"
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
		http.Error(w, "Ошибка чтения из базы", http.StatusInternalServerError)
		return
	}

	temp, err := parseTemplates()
	if err != nil {
		http.Error(w, "Невозможно открыть шаблон", http.StatusInternalServerError)
		return
	}

	if err := temp.ExecuteTemplate(w, "layout", expenses); err != nil {
		http.Error(w, "Ошибка рендера шаблона", http.StatusInternalServerError)
	}
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	// Проверка метода: добавление — только POST
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
		return
	}

	// Проверка корректности суммы
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Некорректная сумма!", http.StatusBadRequest)
		return
	}

	if amount <= 0 {
		http.Error(w, "Сумма должна быть положительной и больше нуля!", http.StatusBadRequest)
		return
	}

	desc := strings.TrimSpace(r.FormValue("description"))
	date := strings.TrimSpace(r.FormValue("date"))
	if desc == "" || date == "" {
		http.Error(w, "Заполните все поля", http.StatusBadRequest)
		return
	}

	_, err = h.repo.Create(models.Expense{
		Amount:      amount,
		Description: desc,
		Date:        date,
	})
	if err != nil {
		http.Error(w, "Ошибка записи в базу", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) EditPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	expense, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Трата не найдена", http.StatusNotFound)
		return
	}

	temp, err := template.ParseFiles("webntemp/edit.html")
	if err != nil {
		http.Error(w, "Невозможно открыть шаблон", http.StatusInternalServerError)
		return
	}
	if err := temp.Execute(w, expense); err != nil {
		http.Error(w, "Ошибка рендера шаблона", http.StatusInternalServerError)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Некорректная сумма!", http.StatusBadRequest)
		return
	}

	if amount <= 0 {
		http.Error(w, "Сумма должна быть положительной и больше нуля!", http.StatusBadRequest)
		return
	}

	desc := strings.TrimSpace(r.FormValue("description"))
	date := strings.TrimSpace(r.FormValue("date"))
	if desc == "" || date == "" {
		http.Error(w, "Заполните все поля", http.StatusBadRequest)
		return
	}

	err = h.repo.Update(models.Expense{
		ID:          id,
		Amount:      amount,
		Description: desc,
		Date:        date,
	})
	if err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
