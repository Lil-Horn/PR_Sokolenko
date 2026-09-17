package models

type Expense struct {
	Amount      float64
	Description string
	Date        string
}

var Expenses []Expense

func Add(amount float64, desc string, date string) {
	Expenses = append(Expenses, Expense{
		Amount:      amount,
		Description: desc,
		Date:        date,
	})
}
