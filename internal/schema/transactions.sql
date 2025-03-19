-- name: GetTransactions :many
SELECT * FROM transactions;

-- name: GetUserTransactionsWithinTimeframe :many
SELECT 
    DATE_TRUNC('hour', transaction_date) AS transac_hour,
    user_id,
    COUNT(*) AS transac_count
FROM transactions
GROUP BY transac_hour, user_id
ORDER BY transac_hour DESC, transac_count DESC
LIMIT 10000;
