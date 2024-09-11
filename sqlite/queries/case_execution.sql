-- name: CreateCaseExecutionScheduled :one
INSERT INTO case_executions (id, test_execution_id, case_name, schedule_time)
VALUES (?, ?, ?, ?)
ON CONFLICT(id, test_execution_id) DO UPDATE
    SET case_name     = excluded.case_name,
        schedule_time = excluded.schedule_time,
        start_time    = NULL,
        finish_time   = NULL,
        error         = NULL
RETURNING *;

-- name: UpdateCaseExecutionStarted :one
UPDATE case_executions
SET start_time = ?
WHERE id = ?
  AND test_execution_id = ?
RETURNING *;

-- name: UpdateCaseExecutionFinished :one
UPDATE case_executions
SET finish_time = ?,
    error       = ?
WHERE id = ?
  AND test_execution_id = ?
RETURNING *;

-- name: DeleteCaseExecution :exec
DELETE
FROM case_executions
WHERE id = ?
  AND test_execution_id = ?;

-- name: GetCaseExecution :one
SELECT *
FROM case_executions
WHERE id = ?
  AND test_execution_id = ?;

-- name: ListCaseExecutions :many
SELECT *
FROM case_executions
WHERE (test_execution_id = @test_execution_id)
  -- Cast as integer required below since sqlc.narg doesn't work with overridden column type
  AND (CAST(sqlc.narg('offset_id') AS INTEGER) IS NULL OR id > CAST(sqlc.narg('offset_id') AS INTEGER))
ORDER BY id
LIMIT @page_size;
