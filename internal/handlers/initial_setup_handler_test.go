package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/services"
)

type initialSetupServiceMock struct {
	CompleteInitialSetupFunc func(userID string, income, savingGoal int, fixedCosts []models.FixedCostInput) error
}

func (m *initialSetupServiceMock) CompleteInitialSetup(ctx context.Context, userID string, income, savingGoal int, fixedCosts []models.FixedCostInput) error {
	if m.CompleteInitialSetupFunc != nil {
		return m.CompleteInitialSetupFunc(userID, income, savingGoal, fixedCosts)
	}
	return nil
}

func TestInitialSetupHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	called := false
	svc := &initialSetupServiceMock{
		CompleteInitialSetupFunc: func(userID string, income, savingGoal int, fixedCosts []models.FixedCostInput) error {
			called = true
			require.Equal(t, DummyUserID, userID)
			require.Equal(t, 300000, income)
			require.Equal(t, 50000, savingGoal)
			require.Equal(t, []models.FixedCostInput{{Name: "家賃", Amount: 80000}, {Name: "通信費", Amount: 5000}}, fixedCosts)
			return nil
		},
	}
	NewInitialSetupHandler(router, svc)

	body := `{"income":300000,"savingGoal":50000,"fixedCosts":[{"name":"家賃","amount":80000},{"name":"通信費","amount":5000}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "ok", resp["status"])
}

func TestInitialSetupHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	svc := &initialSetupServiceMock{}
	NewInitialSetupHandler(router, svc)

	body := `{"income":"bad"}`
	req := httptest.NewRequest(http.MethodPost, "/api/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInitialSetupHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	svc := &initialSetupServiceMock{
		CompleteInitialSetupFunc: func(userID string, income, savingGoal int, fixedCosts []models.FixedCostInput) error {
			return &services.ValidationError{Message: "income must be greater than 0"}
		},
	}
	NewInitialSetupHandler(router, svc)

	body := `{"income":0,"savingGoal":0,"fixedCosts":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "income must be greater than 0", resp["error"])
}

func TestInitialSetupHandler_BusinessError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	svc := &initialSetupServiceMock{
		CompleteInitialSetupFunc: func(userID string, income, savingGoal int, fixedCosts []models.FixedCostInput) error {
			return &services.NotFoundError{Message: "user not found"}
		},
	}
	NewInitialSetupHandler(router, svc)

	body := `{"income":100,"savingGoal":0,"fixedCosts":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "user not found", resp["error"])
}

func TestInitialSetupHandler_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	svc := &initialSetupServiceMock{
		CompleteInitialSetupFunc: func(userID string, income, savingGoal int, fixedCosts []models.FixedCostInput) error {
			return errors.New("boom")
		},
	}
	NewInitialSetupHandler(router, svc)

	body := `{"income":100,"savingGoal":0,"fixedCosts":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
