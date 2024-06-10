-- name: CreateTest :one
INSERT INTO tests (id, "group", name, has_input, runner_id, runner_heartbeat_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT ("group", name) DO UPDATE
    SET "group"             = excluded."group",
        name                = excluded.name,
        has_input         = excluded.has_input,
        runner_id           = excluded.runner_id,
        runner_heartbeat_at = excluded.runner_heartbeat_at,
        create_time          = now()
RETURNING *;

-- name: GetTest :one
SELECT *
FROM tests
WHERE id = $1;

-- name: GetTestByName :one
SELECT *
FROM tests
WHERE name = $1
  AND "group" = $2;

-- name: ListTests :many
SELECT *
FROM tests;

-- name: CreateTestDefaultInput :exec
INSERT INTO test_default_inputs (test_id, data)
VALUES ($1, $2)
ON CONFLICT (test_id) DO UPDATE
    SET data = excluded.data;

-- name: GetTestDefaultInput :one
SELECT *
FROM test_default_inputs
WHERE test_id = $1;

-- name: CreateTestExecution :one
INSERT INTO test_executions (id, test_id, has_input, schedule_time)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
    SET test_id      = excluded.test_id,
        has_input  = excluded.has_input,
        schedule_time = excluded.schedule_time,
        start_time   = null,
        finish_time  = null,
        error        = null
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

-- name: GetTestExecution :one
SELECT *
FROM test_executions
WHERE id = $1;

-- name: ListTestExecutions :many
SELECT *
FROM test_executions
WHERE (@test_id = test_id)
  AND (
    (sqlc.narg('last_schedule_time')::timestamp IS NULL AND sqlc.narg('last_exec_id')::uuid IS NULL)
        OR (schedule_time, id) < (@last_schedule_time::timestamp, @last_test_execution_id::uuid)
    )
ORDER BY schedule_time DESC, id DESC
LIMIT (sqlc.narg('page_size')::integer);
