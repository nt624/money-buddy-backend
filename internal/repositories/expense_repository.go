package repositories

import "money-buddy-backend/internal/models"

// ExpenseRepository は経費リポジトリの振る舞いを表します。
type ExpenseRepository interface {
	CreateExpense(input models.CreateExpenseInput) (models.Expense, error)
	FindAll() ([]models.Expense, error)
	GetExpenseByID(id int32) (models.Expense, error)
	DeleteExpense(id int32) error
}
