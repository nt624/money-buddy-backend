-- name: CreateExpense :one
INSERT INTO expenses (
  amount,
  category_id,
  memo,
  spent_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING id;

-- name: ListExpenses :many
SELECT 
  e.id,
  e.amount,
  e.memo,
  e.spent_at,
  c.id AS category_id,
  c.name AS category_name
FROM expenses e
JOIN categories c ON e.category_id = c.id
ORDER BY spent_at DESC;

-- name: GetExpenseByID :one
SELECT
  e.id,
  e.amount,
  e.memo,
  e.spent_at,
  c.id AS category_id,
  c.name AS category_name
FROM expenses e
JOIN categories c ON e.category_id = c.id
WHERE e.id = $1;