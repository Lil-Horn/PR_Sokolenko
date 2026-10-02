package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"Laba2/internal/models"
)

func Money(amount float64) string {

	s := strconv.FormatFloat(amount, 'f', 2, 64)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]

	var res []string
	for len(intPart) > 3 {
		res = append([]string{intPart[len(intPart)-3:]}, res...)
		intPart = intPart[:len(intPart)-3]
	}
	res = append([]string{intPart}, res...)

	return strings.Join(res, " ") + "," + decPart
}

func Index(w http.ResponseWriter, r *http.Request) {
	temp, err := template.New("").Funcs(template.FuncMap{
		"money": Money,
	}).ParseFiles(

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

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Некорректная сумма!", http.StatusBadRequest)
		return
	}

	if amount <= 0 {
		http.Error(w, "Сумма должна быть положительной и больше нуля!", http.StatusBadRequest)
		return
	}

	desc := r.FormValue("description")
	date := r.FormValue("date")

	models.Add(amount, desc, date)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
