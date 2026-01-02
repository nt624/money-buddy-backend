package repositories

import (
	"context"

	db "money-buddy-backend/db/generated"
	"money-buddy-backend/internal/models"
)

type CategoryRepositorySQLC struct {
	q *db.Queries
}

func NewCategoryRepositorySQLC(q *db.Queries) *CategoryRepositorySQLC {
	return &CategoryRepositorySQLC{q: q}
}

func (r *CategoryRepositorySQLC) ListCategories(ctx context.Context) ([]models.Category, error) {
	items, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	var out []models.Category
	for _, it := range items {
		out = append(out, dbCategoryToModel(it))
	}

	return out, nil
}

func dbCategoryToModel(c db.ListCategoriesRow) models.Category {
	return models.Category{
		ID:   int(c.ID),
		Name: c.Name,
	}
}
