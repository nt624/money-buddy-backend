-- name: CreateUser :exec
INSERT INTO users (
    id,
    income,
    saving_goal
) VALUES (
    $1,$2,$3
);

-- name: GetUserByID :one
SELECT
    id,
    income,
    saving_goal,
    created_at,
    updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserSettings :exec
UPDATE users
SET
    income = $2,
    saving_goal = $3,
    updated_at = now()
WHERE id = $1;