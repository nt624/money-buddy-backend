package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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
	DeleteExpense(id int) error
	UpdateExpense(input models.UpdateExpenseInput) (models.Expense, error)
}

type expenseService struct {
	repo         repositories.ExpenseRepository
	categoryRepo repositories.CategoryRepository
}

func NewExpenseService(repo repositories.ExpenseRepository, categoryRepo repositories.CategoryRepository) ExpenseService {
	return &expenseService{repo: repo, categoryRepo: categoryRepo}
}

func (s *expenseService) CreateExpense(input models.CreateExpenseInput) (models.Expense, error) {
	// 金額チェック: 入力が存在するかをまず確認し、その後業務上の制約を確認する
	if input.Amount == nil {
		return models.Expense{}, &ValidationError{Message: "amount must be provided"}
	}
	if *input.Amount <= 0 {
		return models.Expense{}, &ValidationError{Message: "amount must be greater than 0"}
	}
	if *input.Amount > BusinessMaxAmount {
		return models.Expense{}, &ValidationError{Message: "amount exceeds maximum allowed"}
	}

	// カテゴリID チェック
	if input.CategoryID == nil {
		return models.Expense{}, &ValidationError{Message: "category_id must be provided"}
	}
	if *input.CategoryID <= 0 {
		return models.Expense{}, &ValidationError{Message: "category_id must be greater than 0"}
	}

	// SpentAt の非空チェック
	if input.SpentAt == "" {
		return models.Expense{}, &ValidationError{Message: "spent_at must be provided"}
	}

	// 日付フォーマットの検証（RFC3339 をまず試し、失敗したら日付のみフォーマットを試す）
	var spentAt time.Time
	var err error
	spentAt, err = time.Parse(time.RFC3339, input.SpentAt)
	if err != nil {
		spentAt, err = time.Parse("2006-01-02", input.SpentAt)
		if err != nil {
			return models.Expense{}, &ValidationError{Message: "spent_at is invalid"}
		}
		// 日付のみの場合は UTC の 00:00 として扱う
		spentAt = time.Date(spentAt.Year(), spentAt.Month(), spentAt.Day(), 0, 0, 0, 0, time.UTC)
	}
	if spentAt.IsZero() {
		return models.Expense{}, &ValidationError{Message: "spent_at must be a non-zero time"}
	}

	// Memo 長チェック
	if len(input.Memo) > MemoMaxLen {
		return models.Expense{}, &ValidationError{Message: "memo exceeds maximum length"}
	}

	// Status の検証（任意入力、指定されている場合のみチェック）
	if input.Status != "" {
		if normalized, ok := models.NormalizeStatus(input.Status); ok {
			// 正規化: DB は小文字で扱う前提
			input.Status = normalized
		} else {
			return models.Expense{}, &ValidationError{Message: "status must be 'planned' or 'confirmed'"}
		}
	}

	// カテゴリ存在チェック（CategoryExists を用いる）
	exists, err := s.categoryRepo.CategoryExists(context.Background(), int32(*input.CategoryID))
	if err != nil {
		// リポジトリ/DB からのエラーは内部エラーとして扱う
		return models.Expense{}, &InternalError{Message: "internal error"}
	}
	if !exists {
		return models.Expense{}, &ValidationError{Message: "category_id is invalid"}
	}

	exp, err := s.repo.CreateExpense(input)
	if err != nil {
		// sql.ErrNoRows -> NotFoundError
		if errors.Is(err, sql.ErrNoRows) {
			return models.Expense{}, &NotFoundError{Message: "expense not found"}
		}

		// 外部キー制約（category_id）の検出。
		// ドライバ固有の型へアサートするよりも、エラーメッセージに含まれる
		// 文言を確認して判定する（安全策）。
		lerr := strings.ToLower(err.Error())
		if strings.Contains(lerr, "foreign key") && (strings.Contains(lerr, "category") || strings.Contains(lerr, "category_id")) {
			return models.Expense{}, &ValidationError{Message: "category_id is invalid"}
		}

		// その他は内部エラーとしてラップして返す
		return models.Expense{}, &InternalError{Message: "internal error"}
	}

	return exp, nil
}

func (s *expenseService) ListExpenses() ([]models.Expense, error) {
	return s.repo.FindAll()
}

func (s *expenseService) DeleteExpense(id int) error {
	expense, err := s.repo.GetExpenseByID(int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{Message: "expense not found"}
		}
		return &InternalError{Message: "internal error"}
	}
	if expense == (models.Expense{}) {
		return &NotFoundError{Message: "expense not found"}
	}

	return s.repo.DeleteExpense(int32(id))
}

func (s *expenseService) UpdateExpense(input models.UpdateExpenseInput) (models.Expense, error) {
	// 現在の状態を取得し、ステータス遷移のバリデーションを行う
	current, err := s.repo.GetExpenseByID(int32(input.ID))
	if err != nil {
		// テスト仕様に合わせ、見つからない場合も遷移エラーとして扱う
		if errors.Is(err, sql.ErrNoRows) {
			return models.Expense{}, ErrInvalidStatusTransition
		}
		return models.Expense{}, &InternalError{Message: "internal error"}
	}

	// 変更後ステータスの決定（未指定なら現状維持）
	desiredStatus := input.Status
	if desiredStatus == "" {
		desiredStatus = current.Status
	} else {
		if normalized, ok := models.NormalizeStatus(desiredStatus); ok {
			desiredStatus = normalized
		} else {
			return models.Expense{}, &ValidationError{Message: "status must be 'planned' or 'confirmed'"}
		}
	}

	// 遷移ルール: confirmed → planned は禁止
	if strings.ToLower(current.Status) == "confirmed" && desiredStatus == "planned" {
		return models.Expense{}, ErrInvalidStatusTransition
	}

	// リポジトリに渡す前に正規化済みステータスをセット
	input.Status = desiredStatus
	return s.repo.UpdateExpense(input)
}
