-- name: CreateTestExecutionScheduled :one
INSERT INTO test_executions (id, test_id, has_input, schedule_time)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE
    SET test_id       = excluded.test_id,
        has_input     = excluded.has_input,
        schedule_time = excluded.schedule_time,
        start_time    = NULL,
        finish_time   = NULL,
        error         = NULL
RETURNING *;

-- name: CreateTestExecutionInput :exec
INSERT INTO test_execution_inputs (test_execution_id, data)
VALUES (?, ?)
ON CONFLICT(test_execution_id) DO UPDATE
    SET data = excluded.data;

-- name: GetTestExecutionInput :one
SELECT *
FROM test_execution_inputs
WHERE test_execution_id = ?;

-- name: UpdateTestExecutionStarted :one
UPDATE test_executions
SET start_time  = ?,
    finish_time = NULL,
    error       = NULL
WHERE id = ?
RETURNING *;

-- name: UpdateTestExecutionFinished :one
UPDATE test_executions
SET finish_time = ?,
    error       = ?
WHERE id = ?
RETURNING *;

-- name: ResetTestExecution :one
UPDATE test_executions
SET schedule_time = @reset_time,
    start_time    = NULL,
    finish_time   = NULL,
    error         = NULL
WHERE id = ?
RETURNING *;

-- name: GetTestExecution :one
SELECT *
FROM test_executions
WHERE id = ?;

-- name: ListTestExecutions :many
SELECT *
FROM test_executions
WHERE (test_id = @test_id)
  -- Cast as text required below since sqlc.narg doesn't work with overridden column type
  AND (CAST(sqlc.narg('offset_id') AS TEXT) IS NULL OR id < CAST(sqlc.narg('offset_id') AS TEXT))
ORDER BY id DESC
LIMIT @page_size;
