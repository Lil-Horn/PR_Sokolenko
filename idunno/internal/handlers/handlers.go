package handlers

import (
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"idunno/internal/middleware"
	"idunno/internal/models"
	"idunno/internal/reposit"
)

type Handler struct {
	repo     *reposit.ExpenseRepository
	users    *reposit.UserRepository
	sessions *reposit.SessionRepository
}

func New(repo *reposit.ExpenseRepository, users *reposit.UserRepository, sessions *reposit.SessionRepository) *Handler {
	return &Handler{repo: repo, users: users, sessions: sessions}
}

// Page — общие поля всех страниц (встраивается в данные шаблонов).
type Page struct {
	Authed bool   // показывать ли навигацию авторизованного пользователя
	Error  string // сообщение об ошибке
	Info   string // информационное сообщение
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

// render собирает layout + конкретную страницу и выводит результат.
func render(w http.ResponseWriter, page string, data any) {
	temp, err := template.New("").Funcs(template.FuncMap{
		"money": Money,
	}).ParseFiles("webntemp/layout.html", "webntemp/"+page)
	if err != nil {
		http.Error(w, "Невозможно открыть шаблон", http.StatusInternalServerError)
		return
	}
	if err := temp.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Ошибка рендера шаблона", http.StatusInternalServerError)
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())
	categoryID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	f := models.ExpenseFilter{
		CategoryID: categoryID,
		DateFrom:   r.URL.Query().Get("from"),
		DateTo:     r.URL.Query().Get("to"),
	}

	expenses, err := h.repo.GetAll(userID, f)
	if err != nil {
		http.Error(w, "Ошибка чтения из базы", http.StatusInternalServerError)
		return
	}

	categories, err := h.repo.GetAllCategories()
	if err != nil {
		http.Error(w, "Ошибка чтения категорий", http.StatusInternalServerError)
		return
	}

	data := struct {
		Page
		Expenses   []models.Expense
		Categories []models.Category
		Filter     models.ExpenseFilter
	}{
		Page:       Page{Authed: true},
		Expenses:   expenses,
		Categories: categories,
		Filter:     f,
	}
	render(w, "index.html", data)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
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
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	if desc == "" || date == "" {
		http.Error(w, "Заполните все поля", http.StatusBadRequest)
		return
	}

	_, err = h.repo.Create(models.Expense{
		UserID:      middleware.UserID(r.Context()),
		Amount:      amount,
		Description: desc,
		Date:        date,
		CategoryID:  categoryID,
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

	expense, err := h.repo.GetByID(id, middleware.UserID(r.Context()))
	if err != nil {
		http.Error(w, "Трата не найдена", http.StatusNotFound)
		return
	}

	categories, err := h.repo.GetAllCategories()
	if err != nil {
		http.Error(w, "Ошибка чтения категорий", http.StatusInternalServerError)
		return
	}

	data := struct {
		Page
		Expense    models.Expense
		Categories []models.Category
	}{Page{Authed: true}, expense, categories}

	render(w, "edit.html", data)
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
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	if desc == "" || date == "" {
		http.Error(w, "Заполните все поля", http.StatusBadRequest)
		return
	}

	err = h.repo.Update(models.Expense{
		ID:          id,
		UserID:      middleware.UserID(r.Context()),
		Amount:      amount,
		Description: desc,
		Date:        date,
		CategoryID:  categoryID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Трата не найдена", http.StatusNotFound)
		return
	}
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

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id, middleware.UserID(r.Context()))
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Трата не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
