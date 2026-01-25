-- name: GetMonthlySummary :one
SELECT
  u.income,
  u.saving_goal,
  COALESCE(SUM(fc.amount), 0) AS fixed_costs
FROM users u
LEFT JOIN fixed_costs fc ON fc.user_id = u.id
WHERE u.id = $1
GROUP BY u.id;
