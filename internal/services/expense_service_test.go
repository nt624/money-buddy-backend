package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"money-buddy-backend/internal/models"
)

type mockRepo struct {
	called bool
	in     models.CreateExpenseInput
}

func (m *mockRepo) CreateExpense(input models.CreateExpenseInput) (models.Expense, error) {
	m.called = true
	m.in = input
	return models.Expense{ID: 1, Amount: *input.Amount, CategoryID: *input.CategoryID, Memo: input.Memo, SpentAt: input.SpentAt}, nil
}

func (m *mockRepo) FindAll() ([]models.Expense, error) {
	return nil, errors.New("not implemented")
}

// mockCategoryRepo satisfies CategoryRepository for testing
type mockCategoryRepo struct {
	exists map[int32]bool
	err    error
}

func (m *mockCategoryRepo) ListCategories(ctx context.Context) ([]models.Category, error) {
	return nil, errors.New("not implemented")
}

func (m *mockCategoryRepo) CategoryExists(ctx context.Context, id int32) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.exists == nil {
		return false, nil
	}
	v, ok := m.exists[id]
	if !ok {
		return false, nil
	}
	return v, nil
}

func TestCreateExpenseValidation(t *testing.T) {
	cases := []struct {
		name       string
		input      models.CreateExpenseInput
		wantErr    bool
		wantCalled bool
	}{
		{name: "金額が0以下の場合はエラーになる", input: models.CreateExpenseInput{Amount: intPtr(0), CategoryID: intPtr(1), SpentAt: "2020-01-01"}, wantErr: true, wantCalled: false},
		{name: "金額が1Bを超える場合はエラーになる", input: models.CreateExpenseInput{Amount: intPtr(1000000001), CategoryID: intPtr(1), SpentAt: "2020-01-02"}, wantErr: true, wantCalled: false},
		{name: "カテゴリIDが0以下の場合はエラーになる", input: models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(0), SpentAt: "2020-01-01"}, wantErr: true, wantCalled: false},
		{name: "カテゴリが存在しない場合はエラーになる（カテゴリテーブル参照）", input: models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(9999), SpentAt: "2020-01-02"}, wantErr: true, wantCalled: false},
		{name: "spentAtがzero time（0001-01-01T00:00:00Z）の場合はエラーになる", input: models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(1), SpentAt: "0001-01-01T00:00:00Z"}, wantErr: true, wantCalled: false},
		{name: "日付のみ（YYYY-MM-DD）で正常に作成される", input: models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(1), SpentAt: "2020-01-02"}, wantErr: false, wantCalled: true},
		{name: "RFC3339形式で正常に作成される", input: models.CreateExpenseInput{Amount: intPtr(200), CategoryID: intPtr(2), SpentAt: "2020-01-02T15:04:05Z"}, wantErr: false, wantCalled: true},
		{name: "spentAtが空文字の場合はエラーになる", input: models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(1), SpentAt: ""}, wantErr: true, wantCalled: false},
		{name: "メモが最大長を超える場合はエラーになる", input: models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(1), SpentAt: "2020-01-02", Memo: strings.Repeat("a", 5001)}, wantErr: true, wantCalled: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := &mockRepo{}
			// Prepare category repo: for cases where repo should be called, mark category as existing
			exists := map[int32]bool{}
			if tc.input.CategoryID != nil && tc.wantCalled {
				exists[int32(*tc.input.CategoryID)] = true
			}
			cr := &mockCategoryRepo{exists: exists}
			s := NewExpenseService(m, cr)

			out, err := s.CreateExpense(tc.input)

			if tc.wantErr {
				if !assert.Error(t, err, "expected error for case %s", tc.name) {
					return
				}
				var ve *ValidationError
				assert.ErrorAs(t, err, &ve, "expected ValidationError for case %s", tc.name)
			} else {
				assert.NoError(t, err, "unexpected error for case %s: %v", tc.name, err)
			}

			assert.Equal(t, tc.wantCalled, m.called, "repo called mismatch for case %s", tc.name)

			if !tc.wantErr {
				assert.Equal(t, *tc.input.Amount, out.Amount, "amount mismatch for case %s", tc.name)
			}
		})
	}
}

func intPtr(v int) *int { return &v }

func TestCreateExpense_DBErrorMapping(t *testing.T) {
	t.Parallel()

	validInput := models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(1), SpentAt: "2020-01-02"}

	cases := []struct {
		name     string
		repoErr  error
		wantType string
		wantMsg  string
	}{
		{name: "sql.ErrNoRows の場合は NotFoundError", repoErr: sqlErrNoRows(), wantType: "NotFoundError", wantMsg: "not found"},
		{name: "外部キー(category_id)違反は ValidationError", repoErr: errors.New("pq: insert or update on table \"expenses\" violates foreign key constraint \"expenses_category_id_fkey\""), wantType: "ValidationError", wantMsg: "category_id is invalid"},
		{name: "その他の DB エラーは InternalError", repoErr: errors.New("some db problem"), wantType: "InternalError", wantMsg: "internal error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := &mockRepoErr{returnErr: tc.repoErr}
			cr := &mockCategoryRepo{exists: map[int32]bool{1: true}}
			s := NewExpenseService(m, cr)

			_, err := s.CreateExpense(validInput)
			if !assert.Error(t, err) {
				return
			}

			// 型の検証とメッセージ検証
			switch tc.wantType {
			case "NotFoundError":
				var e *NotFoundError
				if !assert.ErrorAs(t, err, &e) {
					return
				}
				assert.Contains(t, e.Error(), tc.wantMsg)
			case "ValidationError":
				var e *ValidationError
				if !assert.ErrorAs(t, err, &e) {
					return
				}
				assert.Equal(t, tc.wantMsg, e.Message)
			case "InternalError":
				var e *InternalError
				if !assert.ErrorAs(t, err, &e) {
					return
				}
				assert.Contains(t, e.Error(), tc.wantMsg)
			default:
				t.Fatalf("unknown wantType: %s", tc.wantType)
			}
		})
	}
}

// mockRepoErr は CreateExpense でエラーを返すモック
type mockRepoErr struct {
	returnErr error
}

func (m *mockRepoErr) CreateExpense(input models.CreateExpenseInput) (models.Expense, error) {
	return models.Expense{}, m.returnErr
}

func (m *mockRepoErr) FindAll() ([]models.Expense, error) { return nil, errors.New("not implemented") }

// sqlErrNoRows returns sql.ErrNoRows from database/sql
func sqlErrNoRows() error { return sql.ErrNoRows }

func TestCreateExpense_CategoryExistsError(t *testing.T) {
	t.Parallel()

	input := models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(1), SpentAt: "2020-01-02"}

	m := &mockRepo{}
	cr := &mockCategoryRepo{err: errors.New("db error")}
	s := NewExpenseService(m, cr)

	_, err := s.CreateExpense(input)
	if err == nil {
		t.Fatalf("expected error")
	}
	var ie *InternalError
	if !assert.ErrorAs(t, err, &ie) {
		t.Fatalf("expected InternalError, got %v", err)
	}
	assert.False(t, m.called, "repo should not be called when category existence check fails")
}
