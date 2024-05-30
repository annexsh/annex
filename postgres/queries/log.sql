-- name: CreateLog :exec
INSERT INTO logs (id, test_exec_id, case_exec_id, level, message, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetExecutionLog :one
SELECT *
FROM logs
WHERE id = $1;

-- name: ListTestExecutionLogs :many
SELECT *
FROM logs
WHERE test_exec_id = $1
ORDER BY created_at;

-- name: DeleteLog :exec
DELETE
FROM logs
WHERE id = $1;