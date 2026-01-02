package services

import (
	"context"

	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/repositories"
)

type CategoryService interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) ListCategories(ctx context.Context) ([]models.Category, error) {
	return s.repo.ListCategories(ctx)
}
