package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/services"
	"strconv"
)

type ExpenseHandler struct {
	service services.ExpenseService
}

func NewExpenseHandler(r *gin.Engine, service services.ExpenseService) {
	handler := &ExpenseHandler{service: service}

	r.POST("/expenses", handler.CreateExpense)
	r.GET("/expenses", handler.ListExpenses)
	r.PUT("/expenses/:id", handler.UpdateExpense)
	r.DELETE("/expenses/:id", handler.DeleteExpense)
}

func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	var input models.CreateExpenseInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service.CreateExpense(input)
	if err != nil {
		var ve *services.ValidationError
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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

func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	err = h.service.DeleteExpense(int(id))
	if err != nil {
		var ve *services.ValidationError
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateExpense handles PUT /expenses/:id to update an expense.
func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	// Path param ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	// Bind JSON body without ID (ID comes from path)
	type updateBody struct {
		Amount     *int   `json:"amount" binding:"required"`
		CategoryID *int   `json:"category_id" binding:"required"`
		Memo       string `json:"memo"`
		SpentAt    string `json:"spent_at" binding:"required"`
		Status     string `json:"status"`
	}
	var body updateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build service input
	input := models.UpdateExpenseInput{
		ID:         int(id),
		Amount:     body.Amount,
		CategoryID: body.CategoryID,
		Memo:       body.Memo,
		SpentAt:    body.SpentAt,
		Status:     body.Status,
	}

	exp, err := h.service.UpdateExpense(input)
	if err != nil {
		// Validation errors -> 400
		var ve *services.ValidationError
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
			return
		}
		// Status transition error -> 409
		if errors.Is(err, services.ErrInvalidStatusTransition) {
			c.JSON(http.StatusConflict, gin.H{"error": "invalid status transition"})
			return
		}
		// Others -> 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expense": exp})
}
