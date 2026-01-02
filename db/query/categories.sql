-- name: ListCategories :many
SELECT
  id,
  name
FROM categories
ORDER BY id;