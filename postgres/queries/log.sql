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
WHERE (test_execution_id = @test_execution_id)
  AND (sqlc.narg('offset_id')::uuid IS NULL OR id < sqlc.narg('offset_id')::uuid)
ORDER BY id DESC
LIMIT @page_size;

-- name: DeleteLog :exec
DELETE
FROM logs
WHERE id = $1;
