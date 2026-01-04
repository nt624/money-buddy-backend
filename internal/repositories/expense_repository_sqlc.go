package repositories

import (
	"context"
	"database/sql"
	"time"

	db "money-buddy-backend/db/generated"
	"money-buddy-backend/internal/models"
)

// sqlc-backed repository
type expenseRepositorySQLC struct {
	q *db.Queries
}

func NewExpenseRepositorySQLC(q *db.Queries) ExpenseRepository {
	return &expenseRepositorySQLC{q: q}
}

func (r *expenseRepositorySQLC) CreateExpense(input models.CreateExpenseInput) (models.Expense, error) {
	// 複数フォーマットに対応するため、RFC3339 をまず試し、失敗したら日付のみ (2006-01-02) を試す
	var spentAt time.Time
	var err error

	spentAt, err = time.Parse(time.RFC3339, input.SpentAt)
	if err != nil {
		// try date-only format
		spentAt, err = time.Parse("2006-01-02", input.SpentAt)
		if err != nil {
			return models.Expense{}, err
		}
		// 日付のみの場合は UTC の 00:00 として扱う
		spentAt = time.Date(spentAt.Year(), spentAt.Month(), spentAt.Day(), 0, 0, 0, 0, time.UTC)
	}

	params := db.CreateExpenseParams{
		Amount:     int32(*input.Amount),
		CategoryID: int32(*input.CategoryID),
		Memo:       sql.NullString{String: input.Memo, Valid: input.Memo != ""},
		SpentAt:    spentAt,
	}

	e, err := r.q.CreateExpense(context.Background(), params)
	if err != nil {
		return models.Expense{}, err
	}

	return dbExpenseToModel(e), nil
}

func (r *expenseRepositorySQLC) FindAll() ([]models.Expense, error) {
	items, err := r.q.ListExpenses(context.Background())
	if err != nil {
		return nil, err
	}

	var out []models.Expense
	for _, it := range items {
		out = append(out, dbListExpenseRowToModel(it))
	}

	return out, nil
}

func dbExpenseToModel(e db.Expense) models.Expense {
	memo := ""
	if e.Memo.Valid {
		memo = e.Memo.String
	}

	return models.Expense{
		ID:           int(e.ID),
		Amount:       int(e.Amount),
		CategoryID:   int(e.CategoryID),
		CategoryName: "",
		Memo:         memo,
		SpentAt:      e.SpentAt.Format(time.RFC3339),
	}
}

func dbListExpenseRowToModel(e db.ListExpensesRow) models.Expense {
	memo := ""
	if e.Memo.Valid {
		memo = e.Memo.String
	}

	return models.Expense{
		ID:           int(e.ID),
		Amount:       int(e.Amount),
		CategoryID:   int(e.CategoryID),
		CategoryName: e.CategoryName,
		Memo:         memo,
		SpentAt:      e.SpentAt.Format(time.RFC3339),
	}
}
