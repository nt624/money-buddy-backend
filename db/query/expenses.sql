-- name: CreateExpense :one
INSERT INTO expenses (
  amount,
  category_id,
  memo,
  spent_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: ListExpenses :many
SELECT * FROM expenses
ORDER BY spent_at DESC;
