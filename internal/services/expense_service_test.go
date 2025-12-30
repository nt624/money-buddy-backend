package services

import (
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
	return models.Expense{ID: 1, Amount: input.Amount, CategoryID: input.CategoryID, Memo: input.Memo, SpentAt: input.SpentAt}, nil
}

func (m *mockRepo) FindAll() ([]models.Expense, error) {
	return nil, errors.New("not implemented")
}

func TestCreateExpenseValidation(t *testing.T) {
	cases := []struct {
		name       string
		input      models.CreateExpenseInput
		wantErr    bool
		wantCalled bool
		skip       bool
		skipReason string
	}{
		{name: "金額が0以下の場合はエラーになる", input: models.CreateExpenseInput{Amount: 0, CategoryID: 1, SpentAt: "2020-01-01"}, wantErr: true, wantCalled: false},
		{name: "金額が1Bを超える場合はエラーになる", input: models.CreateExpenseInput{Amount: 1000000001, CategoryID: 1, SpentAt: "2020-01-02"}, wantErr: true, wantCalled: false},
		{name: "カテゴリIDが0以下の場合はエラーになる", input: models.CreateExpenseInput{Amount: 100, CategoryID: 0, SpentAt: "2020-01-01"}, wantErr: true, wantCalled: false},
		{name: "カテゴリが存在しない場合はエラーになる（将来的にカテゴリテーブルを参照）", input: models.CreateExpenseInput{Amount: 100, CategoryID: 9999, SpentAt: "2020-01-02"}, wantErr: true, wantCalled: false, skip: true, skipReason: "カテゴリ存在チェックは未実装"},
		{name: "spentAtがzero time（0001-01-01T00:00:00Z）の場合はエラーになる", input: models.CreateExpenseInput{Amount: 100, CategoryID: 1, SpentAt: "0001-01-01T00:00:00Z"}, wantErr: true, wantCalled: false},
		{name: "日付のみ（YYYY-MM-DD）で正常に作成される", input: models.CreateExpenseInput{Amount: 100, CategoryID: 1, SpentAt: "2020-01-02"}, wantErr: false, wantCalled: true},
		{name: "RFC3339形式で正常に作成される", input: models.CreateExpenseInput{Amount: 200, CategoryID: 2, SpentAt: "2020-01-02T15:04:05Z"}, wantErr: false, wantCalled: true},
		{name: "spentAtが空文字の場合はエラーになる", input: models.CreateExpenseInput{Amount: 100, CategoryID: 1, SpentAt: ""}, wantErr: true, wantCalled: false},
		{name: "メモが最大長を超える場合はエラーになる", input: models.CreateExpenseInput{Amount: 100, CategoryID: 1, SpentAt: "2020-01-02", Memo: strings.Repeat("a", 5001)}, wantErr: true, wantCalled: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip(tc.skipReason)
			}
			t.Parallel()
			m := &mockRepo{}
			s := NewExpenseService(m)

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
				assert.Equal(t, tc.input.Amount, out.Amount, "amount mismatch for case %s", tc.name)
			}
		})
	}
}
