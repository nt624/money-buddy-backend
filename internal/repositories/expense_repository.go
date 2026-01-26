package repositories

import "money-buddy-backend/internal/models"

// ExpenseRepository は経費リポジトリの振る舞いを表します。
type ExpenseRepository interface {
	CreateExpense(userID string, input models.CreateExpenseInput) (models.Expense, error)
	FindAll(userID string) ([]models.Expense, error)
	GetExpenseByID(userID string, id int32) (models.Expense, error)
	DeleteExpense(userID string, id int32) error
	UpdateExpense(userID string, input models.UpdateExpenseInput) (models.Expense, error)
}
