-- name: ListCategories :many
SELECT
  id,
  name
FROM categories
ORDER BY id;

-- name: CategoryExists :one
SELECT EXISTS (
  SELECT 1 
  FROM categories 
  WHERE id = $1
);