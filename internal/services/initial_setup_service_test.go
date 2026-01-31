package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"money-buddy-backend/internal/models"
)

type txMock struct{ mock.Mock }

type txManagerMock struct{ mock.Mock }

type userRepoMock struct{ mock.Mock }

type fixedCostRepoMock struct{ mock.Mock }

func (m *txMock) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *txMock) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *txMock) Context(ctx context.Context) context.Context {
	return ctx
}

func (m *txManagerMock) Begin(ctx context.Context) (Tx, error) {
	args := m.Called(ctx)
	if tx, ok := args.Get(0).(Tx); ok {
		return tx, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *userRepoMock) CreateUser(ctx context.Context, id string, income int, savingGoal int) error {
	args := m.Called(ctx, id, income, savingGoal)
	return args.Error(0)
}

func (m *userRepoMock) GetUserByID(ctx context.Context, id string) (models.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(models.User); ok {
		return u, args.Error(1)
	}
	return models.User{}, args.Error(1)
}

func (m *userRepoMock) UpdateUserSettings(ctx context.Context, id string, income int, savingGoal int) error {
	args := m.Called(ctx, id, income, savingGoal)
	return args.Error(0)
}

func (m *fixedCostRepoMock) CreateFixedCost(ctx context.Context, userID string, name string, amount int) (models.FixedCost, error) {
	args := m.Called(ctx, userID, name, amount)
	if fc, ok := args.Get(0).(models.FixedCost); ok {
		return fc, args.Error(1)
	}
	return models.FixedCost{}, args.Error(1)
}

func (m *fixedCostRepoMock) ListFixedCostsByUser(ctx context.Context, userID string) ([]models.FixedCost, error) {
	args := m.Called(ctx, userID)
	if list, ok := args.Get(0).([]models.FixedCost); ok {
		return list, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *fixedCostRepoMock) DeleteFixedCostsByUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *fixedCostRepoMock) BulkCreateFixedCosts(ctx context.Context, userID string, fixedCosts []models.FixedCostInput) error {
	args := m.Called(ctx, userID, fixedCosts)
	return args.Error(0)
}

func (m *fixedCostRepoMock) UpdateFixedCost(ctx context.Context, id int32, userID string, name string, amount int) error {
	args := m.Called(ctx, id, userID, name, amount)
	return args.Error(0)
}

func (m *fixedCostRepoMock) DeleteFixedCost(ctx context.Context, id int32, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func TestCompleteInitialSetup(t *testing.T) {
	userID := "user-1"
	validFixedCosts := []models.FixedCostInput{
		{Name: "rent", Amount: 50000},
		{Name: "phone", Amount: 6000},
	}

	cases := []struct {
		name         string
		income       int
		savingGoal   int
		fixedCosts   []models.FixedCostInput
		setupMocks   func(tx *txMock, tm *txManagerMock, ur *userRepoMock, fr *fixedCostRepoMock, calls *[]string)
		wantErr      bool
		wantValidate bool
		wantCommit   bool
		wantRollback bool
		wantCalls    []string
	}{
		{
			name:       "正常系",
			income:     300000,
			savingGoal: 50000,
			fixedCosts: validFixedCosts,
			setupMocks: func(tx *txMock, tm *txManagerMock, ur *userRepoMock, fr *fixedCostRepoMock, calls *[]string) {
				tm.On("Begin", mock.Anything).Run(func(args mock.Arguments) { *calls = append(*calls, "begin") }).Return(tx, nil)
				ur.On("GetUserByID", mock.Anything, userID).Run(func(args mock.Arguments) { *calls = append(*calls, "get_user") }).Return(models.User{}, sql.ErrNoRows)
				ur.On("CreateUser", mock.Anything, userID, 300000, 50000).Run(func(args mock.Arguments) { *calls = append(*calls, "create_user") }).Return(nil)
				fr.On("DeleteFixedCostsByUser", mock.Anything, userID).Run(func(args mock.Arguments) { *calls = append(*calls, "delete_fixed") }).Return(nil)
				fr.On("BulkCreateFixedCosts", mock.Anything, userID, validFixedCosts).Run(func(args mock.Arguments) { *calls = append(*calls, "bulk_create") }).Return(nil)
				tx.On("Commit").Run(func(args mock.Arguments) { *calls = append(*calls, "commit") }).Return(nil)
			},
			wantCommit:   true,
			wantRollback: false,
			wantCalls:    []string{"begin", "get_user", "create_user", "delete_fixed", "bulk_create", "commit"},
		},
		{
			name:         "income が 0 以下でエラー",
			income:       0,
			savingGoal:   0,
			fixedCosts:   validFixedCosts,
			setupMocks:   func(tx *txMock, tm *txManagerMock, ur *userRepoMock, fr *fixedCostRepoMock, calls *[]string) {},
			wantErr:      true,
			wantValidate: true,
		},
		{
			name:         "固定費 amount が 0 以下でエラー",
			income:       100,
			savingGoal:   0,
			fixedCosts:   []models.FixedCostInput{{Name: "rent", Amount: 0}},
			setupMocks:   func(tx *txMock, tm *txManagerMock, ur *userRepoMock, fr *fixedCostRepoMock, calls *[]string) {},
			wantErr:      true,
			wantValidate: true,
		},
		{
			name:       "fixed_costs 削除失敗で rollback",
			income:     100,
			savingGoal: 0,
			fixedCosts: validFixedCosts,
			setupMocks: func(tx *txMock, tm *txManagerMock, ur *userRepoMock, fr *fixedCostRepoMock, calls *[]string) {
				tm.On("Begin", mock.Anything).Run(func(args mock.Arguments) { *calls = append(*calls, "begin") }).Return(tx, nil)
				ur.On("GetUserByID", mock.Anything, userID).Run(func(args mock.Arguments) { *calls = append(*calls, "get_user") }).Return(models.User{ID: userID}, nil)
				ur.On("UpdateUserSettings", mock.Anything, userID, 100, 0).Run(func(args mock.Arguments) { *calls = append(*calls, "update_user") }).Return(nil)
				fr.On("DeleteFixedCostsByUser", mock.Anything, userID).Run(func(args mock.Arguments) { *calls = append(*calls, "delete_fixed") }).Return(errors.New("delete failed"))
				tx.On("Rollback").Run(func(args mock.Arguments) { *calls = append(*calls, "rollback") }).Return(nil)
			},
			wantErr:      true,
			wantCommit:   false,
			wantRollback: true,
			wantCalls:    []string{"begin", "get_user", "update_user", "delete_fixed", "rollback"},
		},
		{
			name:       "fixed_costs 作成失敗で rollback",
			income:     100,
			savingGoal: 0,
			fixedCosts: validFixedCosts,
			setupMocks: func(tx *txMock, tm *txManagerMock, ur *userRepoMock, fr *fixedCostRepoMock, calls *[]string) {
				tm.On("Begin", mock.Anything).Run(func(args mock.Arguments) { *calls = append(*calls, "begin") }).Return(tx, nil)
				ur.On("GetUserByID", mock.Anything, userID).Run(func(args mock.Arguments) { *calls = append(*calls, "get_user") }).Return(models.User{ID: userID}, nil)
				ur.On("UpdateUserSettings", mock.Anything, userID, 100, 0).Run(func(args mock.Arguments) { *calls = append(*calls, "update_user") }).Return(nil)
				fr.On("DeleteFixedCostsByUser", mock.Anything, userID).Run(func(args mock.Arguments) { *calls = append(*calls, "delete_fixed") }).Return(nil)
				fr.On("BulkCreateFixedCosts", mock.Anything, userID, validFixedCosts).Run(func(args mock.Arguments) { *calls = append(*calls, "bulk_create") }).Return(errors.New("bulk failed"))
				tx.On("Rollback").Run(func(args mock.Arguments) { *calls = append(*calls, "rollback") }).Return(nil)
			},
			wantErr:      true,
			wantCommit:   false,
			wantRollback: true,
			wantCalls:    []string{"begin", "get_user", "update_user", "delete_fixed", "bulk_create", "rollback"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tx := &txMock{}
			tm := &txManagerMock{}
			ur := &userRepoMock{}
			fr := &fixedCostRepoMock{}
			calls := []string{}

			if tc.setupMocks != nil {
				tc.setupMocks(tx, tm, ur, fr, &calls)
			}

			s := NewInitialSetupService(ur, fr, tm)
			err := s.CompleteInitialSetup(context.Background(), userID, tc.income, tc.savingGoal, tc.fixedCosts)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantValidate {
					var ve *ValidationError
					assert.ErrorAs(t, err, &ve)
				}
			} else {
				assert.NoError(t, err)
			}

			if tc.wantCalls != nil {
				assert.Equal(t, tc.wantCalls, calls)
			}

			if tc.wantCommit {
				tx.AssertCalled(t, "Commit")
			} else {
				tx.AssertNotCalled(t, "Commit")
			}
			if tc.wantRollback {
				tx.AssertCalled(t, "Rollback")
			} else {
				tx.AssertNotCalled(t, "Rollback")
			}

			tm.AssertExpectations(t)
			ur.AssertExpectations(t)
			fr.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
