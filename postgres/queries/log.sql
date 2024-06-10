-- name: CreateLog :exec
INSERT INTO logs (id, test_execution_id, case_execution_id, level, message, create_time)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetLog :one
SELECT *
FROM logs
WHERE id = $1;

-- name: ListLogs :many
SELECT *
FROM logs
WHERE test_execution_id = $1
ORDER BY create_time;

-- name: DeleteLog :exec
DELETE
FROM logs
WHERE id = $1;