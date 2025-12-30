package services

import (
	"errors"
	"time"

	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/repositories"
)

const (
	// BusinessMaxAmount は業務上の上限（個人向け家計簿の想定）
	BusinessMaxAmount = 1000000000
	// MemoMaxLen はメモの最大長
	MemoMaxLen = 5000
)

type ExpenseService interface {
	CreateExpense(input models.CreateExpenseInput) (models.Expense, error)
	ListExpenses() ([]models.Expense, error)
}

type expenseService struct {
	repo repositories.ExpenseRepository
}

func NewExpenseService(repo repositories.ExpenseRepository) ExpenseService {
	return &expenseService{repo: repo}
}

func (s *expenseService) CreateExpense(input models.CreateExpenseInput) (models.Expense, error) {
	// 金額チェック
	if input.Amount <= 0 {
		return models.Expense{}, errors.New("amount must be greater than 0")
	}
	if input.Amount > BusinessMaxAmount {
		return models.Expense{}, errors.New("amount exceeds maximum allowed")
	}

	// カテゴリID チェック
	if input.CategoryID <= 0 {
		return models.Expense{}, errors.New("category_id must be greater than 0")
	}

	// SpentAt の非空チェック
	if input.SpentAt == "" {
		return models.Expense{}, errors.New("spent_at must be provided")
	}

	// 日付フォーマットの検証（RFC3339 をまず試し、失敗したら日付のみフォーマットを試す）
	var spentAt time.Time
	var err error
	spentAt, err = time.Parse(time.RFC3339, input.SpentAt)
	if err != nil {
		spentAt, err = time.Parse("2006-01-02", input.SpentAt)
		if err != nil {
			return models.Expense{}, errors.New("spent_at is invalid")
		}
		// 日付のみの場合は UTC の 00:00 として扱う
		spentAt = time.Date(spentAt.Year(), spentAt.Month(), spentAt.Day(), 0, 0, 0, 0, time.UTC)
	}
	if spentAt.IsZero() {
		return models.Expense{}, errors.New("spent_at must be a non-zero time")
	}

	// Memo 長チェック
	if len(input.Memo) > MemoMaxLen {
		return models.Expense{}, errors.New("memo exceeds maximum length")
	}

	// 現在はカテゴリ存在チェックは行わない（将来追加予定）

	return s.repo.CreateExpense(input)
}

func (s *expenseService) ListExpenses() ([]models.Expense, error) {
	return s.repo.FindAll()
}
