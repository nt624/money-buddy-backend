package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/services"
)

type InitialSetupHandler struct {
	service services.InitialSetupService
}

type initialSetupRequest struct {
	Income     int                     `json:"income"`
	SavingGoal int                     `json:"savingGoal"`
	FixedCosts []models.FixedCostInput `json:"fixedCosts"`
}

func NewInitialSetupHandler(r *gin.Engine, service services.InitialSetupService) {
	h := &InitialSetupHandler{service: service}
	r.POST("/setup", h.CompleteInitialSetup)
}

func (h *InitialSetupHandler) CompleteInitialSetup(c *gin.Context) {
	var req initialSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Extract userID from authentication context when auth is implemented
	userID := DummyUserID

	err := h.service.CompleteInitialSetup(c.Request.Context(), userID, req.Income, req.SavingGoal, req.FixedCosts)
	if err != nil {
		var ve *services.ValidationError
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
			return
		}
		var be *services.NotFoundError
		if errors.As(err, &be) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": be.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
