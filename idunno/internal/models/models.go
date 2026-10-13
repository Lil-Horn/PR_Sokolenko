package models

type Expense struct {
	ID           int
	Amount       float64
	Description  string
	Date         string
	CategoryID   int
	CategoryName string
}

type Category struct {
	ID   int
	Name string
}

type ExpenseFilter struct {
	CategoryID int
	DateFrom   string
	DateTo     string
}
