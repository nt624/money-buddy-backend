-- name: CreateFixedCost :one
INSERT INTO fixed_costs (
  user_id,
  name,
  amount
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: ListFixedCostsByUser :many
SELECT 
  id,
  user_id,
  name,
  amount,
  created_at,
  updated_at
FROM fixed_costs
WHERE user_id = $1
ORDER BY id ASC;

-- name: UpdateFixedCost :exec
UPDATE fixed_costs
SET
  name = $2,
  amount = $3,
  updated_at = now()
WHERE id = $1 AND user_id = $4;

-- name: DeleteFixedCost :exec
DELETE FROM fixed_costs
WHERE id = $1 AND user_id = $2;