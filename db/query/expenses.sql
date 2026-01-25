-- name: CreateExpense :one
INSERT INTO expenses (
  user_id,
  amount,
  category_id,
  memo,
  spent_at,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id;

-- name: ListExpenses :many
SELECT 
  e.id,
  e.amount,
  e.memo,
  e.spent_at,
  e.status,
  c.id AS category_id,
  c.name AS category_name
FROM expenses e
JOIN categories c ON e.category_id = c.id
WHERE e.user_id = $1
ORDER BY spent_at DESC;

-- name: GetExpenseWithCategoryByID :one
SELECT
  e.id,
  e.amount,
  e.memo,
  e.spent_at,
  e.status,
  c.id AS category_id,
  c.name AS category_name
FROM expenses e
JOIN categories c ON e.category_id = c.id
WHERE e.user_id = $1 AND e.id = $2;

-- name: GetExpenseByID :one
SELECT
  id,
  amount,
  category_id,
  memo,
  spent_at,
  status
FROM expenses
WHERE user_id = $1 AND id = $2;

-- name: UpdateExpense :exec
UPDATE expenses
SET
  amount = $2,
  category_id = $3,
  memo = $4,
  spent_at = $5,
  status = $6,
  update_at = now()
WHERE id = $1 AND user_id = $7;

-- name: UpdateExpenseStatus :exec
UPDATE expenses
SET
  status = $2,
  update_at = now()
WHERE id = $1 AND user_id = $3;

-- name: DeleteExpense :exec
DELETE FROM expenses
WHERE id = $1 AND user_id = $2;