package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/services"
)

type ExpenseHandler struct {
	service services.ExpenseService
}

func NewExpenseHandler(r *gin.Engine, service services.ExpenseService) {
	handler := &ExpenseHandler{service: service}

	r.POST("/expenses", handler.CreateExpense)
	r.GET("/expenses", handler.ListExpenses)
}

func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	var input models.CreateExpenseInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service.CreateExpense(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"expense": expense})
}

func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	expenses, err := h.service.ListExpenses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list expenses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expenses": expenses})
}
