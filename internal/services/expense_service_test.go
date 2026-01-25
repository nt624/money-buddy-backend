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

func (m *mockRepo) CreateExpense(userID string, input models.CreateExpenseInput) (models.Expense, error) {
	m.called = true
	m.in = input
	return models.Expense{ID: 1, Amount: *input.Amount, Memo: input.Memo, SpentAt: input.SpentAt, Category: models.Category{ID: *input.CategoryID, Name: ""}}, nil
}

func (m *mockRepo) FindAll(userID string) ([]models.Expense, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepo) GetExpenseByID(userID string, id int32) (models.Expense, error) {
	return models.Expense{}, errors.New("not implemented")
}

func (m *mockRepo) DeleteExpense(userID string, id int32) error { return errors.New("not implemented") }

func (m *mockRepo) UpdateExpense(userID string, input models.UpdateExpenseInput) (models.Expense, error) {
	return models.Expense{}, errors.New("not implemented")
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

func (m *mockRepoErr) CreateExpense(userID string, input models.CreateExpenseInput) (models.Expense, error) {
	return models.Expense{}, m.returnErr
}

func (m *mockRepoErr) FindAll(userID string) ([]models.Expense, error) { return nil, errors.New("not implemented") }

func (m *mockRepoErr) GetExpenseByID(userID string, id int32) (models.Expense, error) {
	return models.Expense{}, errors.New("not implemented")
}

func (m *mockRepoErr) DeleteExpense(userID string, id int32) error { return errors.New("not implemented") }

func (m *mockRepoErr) UpdateExpense(userID string, input models.UpdateExpenseInput) (models.Expense, error) {
	return models.Expense{}, errors.New("not implemented")
}

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

func TestCreateExpense_StatusValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		input          models.CreateExpenseInput
		wantErr        bool
		wantCalled     bool
		wantNormalized string
	}{
		{
			name:           "status 'planned' accepted and normalized",
			input:          models.CreateExpenseInput{Amount: intPtr(100), CategoryID: intPtr(1), SpentAt: "2020-01-02", Status: "PlAnNeD"},
			wantErr:        false,
			wantCalled:     true,
			wantNormalized: "planned",
		},
		{
			name:           "status 'confirmed' accepted",
			input:          models.CreateExpenseInput{Amount: intPtr(200), CategoryID: intPtr(2), SpentAt: "2020-01-03", Status: "confirmed"},
			wantErr:        false,
			wantCalled:     true,
			wantNormalized: "confirmed",
		},
		{
			name:       "invalid status causes ValidationError",
			input:      models.CreateExpenseInput{Amount: intPtr(300), CategoryID: intPtr(3), SpentAt: "2020-01-04", Status: "draft"},
			wantErr:    true,
			wantCalled: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &mockRepo{}
			exists := map[int32]bool{}
			if tc.input.CategoryID != nil {
				exists[int32(*tc.input.CategoryID)] = true
			}
			cr := &mockCategoryRepo{exists: exists}
			s := NewExpenseService(m, cr)

			_, err := s.CreateExpense(tc.input)

			if tc.wantErr {
				if !assert.Error(t, err) {
					return
				}
				var ve *ValidationError
				assert.ErrorAs(t, err, &ve)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantNormalized, m.in.Status)
			}

			assert.Equal(t, tc.wantCalled, m.called)
		})
	}
}

// mockDeleteRepo satisfies repositories.ExpenseRepository and adds DeleteExpense for tests
type mockDeleteRepo struct {
	called    bool
	deletedID int32
	returnErr error
}

func (m *mockDeleteRepo) CreateExpense(userID string, input models.CreateExpenseInput) (models.Expense, error) {
	return models.Expense{}, errors.New("not implemented")
}

func (m *mockDeleteRepo) FindAll(userID string) ([]models.Expense, error) {
	return nil, errors.New("not implemented")
}

// DeleteExpense is the method under test expectation
func (m *mockDeleteRepo) DeleteExpense(userID string, id int32) error {
	m.called = true
	m.deletedID = id
	return m.returnErr
}

func (m *mockDeleteRepo) UpdateExpense(userID string, input models.UpdateExpenseInput) (models.Expense, error) {
	return models.Expense{}, errors.New("not implemented")
}

func (m *mockDeleteRepo) GetExpenseByID(userID string, id int32) (models.Expense, error) {
	// simulate existence: 9999 -> not found, others exist
	if id == 9999 {
		return models.Expense{}, sqlErrNoRows()
	}
	status := "confirmed"
	if id == 10 {
		status = "planned"
	}
	return models.Expense{ID: int(id), Status: status}, nil
}

func TestDeleteExpense_Success(t *testing.T) {
	t.Parallel()

	repo := &mockDeleteRepo{returnErr: nil}
	// category repo is unused for delete
	cr := &mockCategoryRepo{}
	// Construct concrete service to allow calling DeleteExpense (to be implemented)
	s := &expenseService{repo: repo, categoryRepo: cr}

	err := s.DeleteExpense(1)
	assert.NoError(t, err)
	assert.True(t, repo.called, "repo should be called")
	assert.Equal(t, int32(1), repo.deletedID)
}

func TestDeleteExpense_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockDeleteRepo{returnErr: sqlErrNoRows()}
	cr := &mockCategoryRepo{}
	s := &expenseService{repo: repo, categoryRepo: cr}

	err := s.DeleteExpense(9999)
	var nfe *NotFoundError
	if !assert.ErrorAs(t, err, &nfe) {
		return
	}
}

func TestDeleteExpense_StatusAgnostic(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		id   int
	}{
		{name: "planned item can delete", id: 10},
		{name: "confirmed item can delete", id: 11},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockDeleteRepo{returnErr: nil}
			cr := &mockCategoryRepo{}
			s := &expenseService{repo: repo, categoryRepo: cr}

			err := s.DeleteExpense(tc.id)
			assert.NoError(t, err)
			assert.True(t, repo.called)
			assert.Equal(t, int32(tc.id), repo.deletedID)
		})
	}
}

// mockUpdateRepo satisfies repositories.ExpenseRepository and simulates update behavior
type mockUpdateRepo struct {
	current   models.Expense
	called    bool
	in        models.UpdateExpenseInput
	returnErr error
	getErr    error
}

func (m *mockUpdateRepo) CreateExpense(userID string, input models.CreateExpenseInput) (models.Expense, error) {
	return models.Expense{}, errors.New("not implemented")
}

func (m *mockUpdateRepo) FindAll(userID string) ([]models.Expense, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUpdateRepo) GetExpenseByID(userID string, id int32) (models.Expense, error) {
	if m.getErr != nil {
		return models.Expense{}, m.getErr
	}
	return m.current, nil
}

func (m *mockUpdateRepo) DeleteExpense(userID string, id int32) error { return errors.New("not implemented") }

// UpdateExpense updates fields; if Status is empty, keep current status
func (m *mockUpdateRepo) UpdateExpense(userID string, input models.UpdateExpenseInput) (models.Expense, error) {
	m.called = true
	m.in = input
	if m.returnErr != nil {
		return models.Expense{}, m.returnErr
	}
	e := m.current
	if input.Amount != nil {
		e.Amount = *input.Amount
	}
	if input.CategoryID != nil {
		e.Category = models.Category{ID: *input.CategoryID}
	}
	e.Memo = input.Memo
	e.SpentAt = input.SpentAt
	if input.Status != "" {
		e.Status = strings.ToLower(input.Status)
	}
	m.current = e
	return e, nil
}

func TestUpdateExpense_NormalCases(t *testing.T) {
	t.Parallel()

	// Pattern 1: current planned -> update to confirmed
	t.Run("planned to confirmed", func(t *testing.T) {
		t.Parallel()

		repo := &mockUpdateRepo{current: models.Expense{ID: 1, Amount: 100, Memo: "old", SpentAt: "2025-01-01", Status: "planned", Category: models.Category{ID: 1}}}
		cr := &mockCategoryRepo{}
		s := &expenseService{repo: repo, categoryRepo: cr}

		input := models.UpdateExpenseInput{
			ID:         1,
			Amount:     intPtr(200),
			CategoryID: intPtr(2),
			Memo:       "updated memo",
			SpentAt:    "2025-02-01",
			Status:     "confirmed",
		}
		out, err := s.UpdateExpense(input)

		assert.NoError(t, err)
		assert.True(t, repo.called)
		assert.Equal(t, "confirmed", out.Status)
		assert.Equal(t, 200, out.Amount)
		assert.Equal(t, 2, out.Category.ID)
		assert.Equal(t, "updated memo", out.Memo)
		assert.Equal(t, "2025-02-01", out.SpentAt)
	})

	// Pattern 2: current confirmed -> update content without changing status
	t.Run("confirmed no status change", func(t *testing.T) {
		t.Parallel()

		repo := &mockUpdateRepo{current: models.Expense{ID: 2, Amount: 300, Memo: "c-old", SpentAt: "2025-03-01", Status: "confirmed", Category: models.Category{ID: 3}}}
		cr := &mockCategoryRepo{}
		s := &expenseService{repo: repo, categoryRepo: cr}

		input := models.UpdateExpenseInput{
			ID:         2,
			Amount:     intPtr(350),
			CategoryID: intPtr(4),
			Memo:       "c-updated",
			SpentAt:    "2025-03-15",
			Status:     "", // no change
		}
		out, err := s.UpdateExpense(input)

		assert.NoError(t, err)
		assert.True(t, repo.called)
		assert.Equal(t, "confirmed", out.Status)
		assert.Equal(t, 350, out.Amount)
		assert.Equal(t, 4, out.Category.ID)
		assert.Equal(t, "c-updated", out.Memo)
		assert.Equal(t, "2025-03-15", out.SpentAt)
	})

	// Pattern 3: current planned -> update content without changing status
	t.Run("planned no status change", func(t *testing.T) {
		t.Parallel()

		repo := &mockUpdateRepo{current: models.Expense{ID: 3, Amount: 500, Memo: "p-old", SpentAt: "2025-04-01", Status: "planned", Category: models.Category{ID: 5}}}
		cr := &mockCategoryRepo{}
		s := &expenseService{repo: repo, categoryRepo: cr}

		input := models.UpdateExpenseInput{
			ID:         3,
			Amount:     intPtr(550),
			CategoryID: intPtr(6),
			Memo:       "p-updated",
			SpentAt:    "2025-04-10",
			Status:     "", // no change
		}
		out, err := s.UpdateExpense(input)

		assert.NoError(t, err)
		assert.True(t, repo.called)
		assert.Equal(t, "planned", out.Status)
		assert.Equal(t, 550, out.Amount)
		assert.Equal(t, 6, out.Category.ID)
		assert.Equal(t, "p-updated", out.Memo)
		assert.Equal(t, "2025-04-10", out.SpentAt)
	})
}

func TestUpdateExpense_InvalidTransition_ConfirmedToPlanned(t *testing.T) {
	t.Parallel()

	repo := &mockUpdateRepo{current: models.Expense{ID: 100, Amount: 1000, Memo: "confirmed item", SpentAt: "2025-05-01", Status: "confirmed", Category: models.Category{ID: 10}}}
	cr := &mockCategoryRepo{}
	s := &expenseService{repo: repo, categoryRepo: cr}

	input := models.UpdateExpenseInput{
		ID:         100,
		Amount:     intPtr(1200),
		CategoryID: intPtr(11),
		Memo:       "try revert to planned",
		SpentAt:    "2025-05-02",
		Status:     "planned",
	}

	_, err := s.UpdateExpense(input)
	if err == nil {
		t.Fatalf("expected error")
	}
	// ErrInvalidStatusTransition を返すこと
	if !assert.ErrorIs(t, err, ErrInvalidStatusTransition) {
		return
	}
	// Update が呼ばれないこと
	assert.False(t, repo.called)
}

func TestUpdateExpense_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockUpdateRepo{getErr: sqlErrNoRows()}
	cr := &mockCategoryRepo{}
	s := &expenseService{repo: repo, categoryRepo: cr}

	input := models.UpdateExpenseInput{
		ID:         9999,
		Amount:     intPtr(500),
		CategoryID: intPtr(1),
		Memo:       "not found case",
		SpentAt:    "2025-06-01",
		Status:     "planned",
	}

	_, err := s.UpdateExpense(input)
	if err == nil {
		t.Fatalf("expected error")
	}
	// 指定のIDが存在しない場合も ErrInvalidStatusTransition を返す
	if !assert.ErrorIs(t, err, ErrInvalidStatusTransition) {
		return
	}
	// Update が呼ばれないこと
	assert.False(t, repo.called)
}
