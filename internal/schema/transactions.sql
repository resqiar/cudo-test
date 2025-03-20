-- name: GetUserTransactions :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY transaction_date DESC
LIMIT 100;

-- name: GetUserTransactionsWithinTimeframe :many
SELECT 
    DATE_TRUNC('hour', transaction_date)::timestamp AS transac_hour,
    user_id,
    COUNT(*) AS transac_count
FROM transactions
WHERE user_id = $1
GROUP BY transac_hour, user_id
ORDER BY transac_hour DESC;
