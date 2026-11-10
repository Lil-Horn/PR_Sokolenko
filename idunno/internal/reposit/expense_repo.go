package reposit

import (
	"database/sql"
	"strings"

	"idunno/internal/models"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}

}

func (r *ExpenseRepository) Create(e models.Expense) (int, error) {
	res, err := r.db.Exec(
		"INSERT INTO expenses (amount, description, date, category_id, user_id) VALUES (?, ?, ?, ?, ?)",
		e.Amount, e.Description, e.Date, e.CategoryID, e.UserID,
	)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *ExpenseRepository) GetAll(userID int, f models.ExpenseFilter) ([]models.Expense, error) {
	query := `SELECT e.id, e.amount, e.description, e.date,
	                 COALESCE(e.category_id, 0),
	                 COALESCE(c.name, '—')
	          FROM expenses e
	          LEFT JOIN categories c ON c.id = e.category_id`
	// Всегда отбираем только траты текущего пользователя
	conds := []string{"e.user_id = ?"}
	args := []any{userID}
	if f.CategoryID > 0 {
		conds = append(conds, "e.category_id = ?")
		args = append(args, f.CategoryID)
	}
	if f.DateFrom != "" {
		conds = append(conds, "e.date >= ?")
		args = append(args, f.DateFrom)
	}
	if f.DateTo != "" {
		conds = append(conds, "e.date <= ?")
		args = append(args, f.DateTo)
	}
	query += " WHERE " + strings.Join(conds, " AND ")
	query += " ORDER BY e.id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Date,
			&e.CategoryID, &e.CategoryName); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *ExpenseRepository) GetByID(id, userID int) (models.Expense, error) {
	var e models.Expense
	err := r.db.QueryRow(
		`SELECT e.id, e.amount, e.description, e.date,
		        COALESCE(e.category_id, 0), COALESCE(c.name, '—')
		 FROM expenses e
		 LEFT JOIN categories c ON c.id = e.category_id
		 WHERE e.id = ? AND e.user_id = ?`, id, userID,
	).Scan(&e.ID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.CategoryName)
	return e, err
}

// Update меняет трату только если она принадлежит e.UserID, иначе sql.ErrNoRows.
func (r *ExpenseRepository) Update(e models.Expense) error {
	res, err := r.db.Exec(
		"UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ? AND user_id = ?",
		e.Amount, e.Description, e.Date, e.CategoryID, e.ID, e.UserID,
	)
	return affected(res, err)
}

func (r *ExpenseRepository) GetAllCategories() ([]models.Category, error) {
	rows, err := r.db.Query("SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// Delete удаляет трату только если она принадлежит userID, иначе sql.ErrNoRows.
func (r *ExpenseRepository) Delete(id, userID int) error {
	res, err := r.db.Exec("DELETE FROM expenses WHERE id = ? AND user_id = ?", id, userID)
	return affected(res, err)
}

func affected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
