package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"money-buddy-backend/internal/services"
)

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(r *gin.Engine, service services.CategoryService) {
	h := &CategoryHandler{service: service}
	r.GET("/categories", h.ListCategories)
}

func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
