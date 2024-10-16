-- name: CreateLog :exec
INSERT INTO logs (id, test_execution_id, case_execution_id, level, message, create_time)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetLog :one
SELECT *
FROM logs
WHERE id = ?;

-- name: ListLogs :many
SELECT *
FROM logs
WHERE (test_execution_id = @test_execution_id)
  -- Cast as text required below since sqlc.narg doesn't work with overridden column type
  AND (CAST(sqlc.narg('offset_id') AS TEXT) IS NULL OR id < CAST(sqlc.narg('offset_id') AS TEXT))
LIMIT @page_size;

-- name: DeleteLog :exec
DELETE
FROM logs
WHERE id = ?;
