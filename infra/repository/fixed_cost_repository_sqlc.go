package repository

import (
	"context"
	"time"

	db "money-buddy-backend/db/generated"
	"money-buddy-backend/infra/transaction"
	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/repositories"
)

type fixedCostRepositorySQLC struct {
	q *db.Queries
}

func NewFixedCostRepositorySQLC(q *db.Queries) repositories.FixedCostRepository {
	return &fixedCostRepositorySQLC{q: q}
}

func (r *fixedCostRepositorySQLC) queries(ctx context.Context) *db.Queries {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *fixedCostRepositorySQLC) CreateFixedCost(ctx context.Context, userID string, name string, amount int) (models.FixedCost, error) {
	params := db.CreateFixedCostParams{
		UserID: userID,
		Name:   name,
		Amount: int32(amount),
	}
	row, err := r.queries(ctx).CreateFixedCost(ctx, params)
	if err != nil {
		return models.FixedCost{}, err
	}

	return dbFixedCostToModel(row), nil
}

func (r *fixedCostRepositorySQLC) ListFixedCostsByUser(ctx context.Context, userID string) ([]models.FixedCost, error) {
	items, err := r.queries(ctx).ListFixedCostsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]models.FixedCost, 0, len(items))
	for _, it := range items {
		out = append(out, dbFixedCostToModel(it))
	}

	return out, nil
}

func (r *fixedCostRepositorySQLC) DeleteFixedCostsByUser(ctx context.Context, userID string) error {
	return r.queries(ctx).DeleteFixedCostsByUser(ctx, userID)
}

func (r *fixedCostRepositorySQLC) BulkCreateFixedCosts(ctx context.Context, userID string, fixedCosts []models.FixedCostInput) error {
	if len(fixedCosts) == 0 {
		return nil
	}

	userIDs := make([]string, 0, len(fixedCosts))
	names := make([]string, 0, len(fixedCosts))
	amounts := make([]int32, 0, len(fixedCosts))
	for _, fc := range fixedCosts {
		userIDs = append(userIDs, userID)
		names = append(names, fc.Name)
		amounts = append(amounts, int32(fc.Amount))
	}

	params := db.BulkCreateFixedCostsParams{
		Column1: userIDs,
		Column2: names,
		Column3: amounts,
	}
	return r.queries(ctx).BulkCreateFixedCosts(ctx, params)
}

func (r *fixedCostRepositorySQLC) UpdateFixedCost(ctx context.Context, id int32, userID string, name string, amount int) error {
	params := db.UpdateFixedCostParams{
		ID:     id,
		Name:   name,
		Amount: int32(amount),
		UserID: userID,
	}
	return r.queries(ctx).UpdateFixedCost(ctx, params)
}

func (r *fixedCostRepositorySQLC) DeleteFixedCost(ctx context.Context, id int32, userID string) error {
	return r.queries(ctx).DeleteFixedCost(ctx, db.DeleteFixedCostParams{
		ID:     id,
		UserID: userID,
	})
}

func dbFixedCostToModel(fc db.FixedCost) models.FixedCost {
	createdAt := ""
	if fc.CreatedAt.Valid {
		createdAt = fc.CreatedAt.Time.Format(time.RFC3339)
	}
	updatedAt := ""
	if fc.UpdatedAt.Valid {
		updatedAt = fc.UpdatedAt.Time.Format(time.RFC3339)
	}

	return models.FixedCost{
		ID:        int(fc.ID),
		UserID:    fc.UserID,
		Name:      fc.Name,
		Amount:    int(fc.Amount),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
