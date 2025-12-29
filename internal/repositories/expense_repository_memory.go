package repositories

import "money-buddy-backend/internal/models"

type expenseRepositoryMemory struct {
	expenses []models.Expense
	nextID   int
}

func NewExpenseRepositoryMemory() ExpenseRepository {
	return &expenseRepositoryMemory{
		expenses: []models.Expense{},
		nextID:   1,
	}
}

func (r *expenseRepositoryMemory) CreateExpense(input models.CreateExpenseInput) (models.Expense, error) {
	expense := models.Expense{
		ID:         r.nextID,
		Amount:     input.Amount,
		CategoryID: input.CategoryID,
		Memo:       input.Memo,
		SpentAt:    input.SpentAt,
	}
	r.nextID++

	r.expenses = append(r.expenses, expense)

	return expense, nil
}

func (r *expenseRepositoryMemory) FindAll() ([]models.Expense, error) {
	return r.expenses, nil
}
