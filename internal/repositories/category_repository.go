package repositories

import (
	"context"

	"money-buddy-backend/internal/models"
)

type CategoryRepository interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
	CategoryExists(ctx context.Context, id int32) (bool, error)
}
