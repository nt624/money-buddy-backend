package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"money-buddy-backend/internal/models"
)

// mockExpenseService implements the services.ExpenseService interface for tests.
type mockExpenseService struct{}

func (m *mockExpenseService) CreateExpense(input models.CreateExpenseInput) (models.Expense, error) {
	return models.Expense{
		ID:         1,
		Amount:     input.Amount,
		CategoryID: input.CategoryID,
		Memo:       input.Memo,
		SpentAt:    input.SpentAt,
	}, nil
}

func (m *mockExpenseService) ListExpenses() ([]models.Expense, error) {
	return nil, nil
}

func TestCreateExpenseHandler_Created(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	svc := &mockExpenseService{}
	NewExpenseHandler(router, svc)

	body := `{"amount":1000,"category_id":2,"memo":"lunch","spent_at":"2025-12-30"}`
	req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]models.Expense
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	expense, ok := resp["expense"]
	require.True(t, ok)
	require.Equal(t, 1, expense.ID)
	require.Equal(t, 1000, expense.Amount)
	require.Equal(t, 2, expense.CategoryID)
	require.Equal(t, "lunch", expense.Memo)
	require.Equal(t, "2025-12-30", expense.SpentAt)
}

func TestCreateExpenseHandler_InvalidJSON_Skip(t *testing.T) {
	t.Skip("SKIP: JSON異常系テストは後ほど追加します")
}
