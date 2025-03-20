-- name: GetRecentTransactions :many
SELECT *
FROM transactions
ORDER BY transaction_date DESC
LIMIT $1;
