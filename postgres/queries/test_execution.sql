-- name: CreateTestExecutionScheduled :one
INSERT INTO test_executions (id, test_id, has_input, schedule_time)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
    SET test_id       = excluded.test_id,
        has_input     = excluded.has_input,
        schedule_time = excluded.schedule_time,
        start_time    = null,
        finish_time   = null,
        error         = null
RETURNING *;

-- name: CreateTestExecutionInput :exec
INSERT INTO test_execution_inputs (test_execution_id, data)
VALUES ($1, $2)
ON CONFLICT (test_execution_id) DO UPDATE
    SET data = excluded.data;

-- name: GetTestExecutionInput :one
SELECT *
FROM test_execution_inputs
WHERE test_execution_id = $1;

-- name: UpdateTestExecutionStarted :one
UPDATE test_executions
SET start_time  = $2,
    finish_time = null,
    error       = null
WHERE id = $1
RETURNING *;

-- name: UpdateTestExecutionFinished :one
UPDATE test_executions
SET finish_time = $2,
    error       = $3
WHERE id = $1
RETURNING *;

-- name: ResetTestExecution :one
UPDATE test_executions
SET schedule_time = @reset_time,
    start_time    = null,
    finish_time   = null,
    error         = null
WHERE id = $1
RETURNING *;

-- name: GetTestExecution :one
SELECT *
FROM test_executions
WHERE id = $1;

-- name: ListTestExecutions :many
SELECT *
FROM test_executions
WHERE test_id = @test_id
  -- Cast as uuid required below since sqlc.narg doesn't work with overridden column type
  AND (sqlc.narg('offset_id')::uuid IS NULL OR id < sqlc.narg('offset_id')::uuid)
ORDER BY id DESC
LIMIT @page_size;
