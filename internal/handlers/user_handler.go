package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"money-buddy-backend/internal/services"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(r *gin.Engine, service services.UserService) {
	h := &UserHandler{service: service}
	r.GET("/user/me", h.GetCurrentUser)
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// TODO: Extract userID from authentication context when auth is implemented
	userID := DummyUserID

	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
