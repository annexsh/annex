-- name: CreateTest :one
INSERT INTO tests (id, project, name, has_payload, runner_id, runner_heartbeat_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (project, name) DO UPDATE
    SET project             = excluded.project,
        name                = excluded.name,
        has_payload         = excluded.has_payload,
        runner_id           = excluded.runner_id,
        runner_heartbeat_at = excluded.runner_heartbeat_at,
        created_at          = now()
RETURNING *;

-- name: GetTest :one
SELECT *
FROM tests
WHERE id = $1;

-- name: GetTestByName :one
SELECT *
FROM tests
WHERE name = $1
  AND project = $2;

-- name: ListTests :many
SELECT *
FROM tests;

-- name: SetTestRunnerHeartbeat :exec
UPDATE tests
SET runner_id           = $2,
    runner_heartbeat_at = NOW()
WHERE id = $1;

-- name: CreateTestDefaultPayload :exec
INSERT INTO test_default_payloads (test_id, payload, is_zero)
VALUES ($1, $2, $3)
ON CONFLICT (test_id) DO UPDATE
    SET payload = excluded.payload,
        is_zero = excluded.is_zero;

-- name: GetTestDefaultPayload :one
SELECT *
FROM test_default_payloads
WHERE test_id = $1;

-- name: CreateTestExecution :one
INSERT INTO test_executions (id, test_id, has_payload, scheduled_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
    SET test_id      = excluded.test_id,
        has_payload  = excluded.has_payload,
        scheduled_at = excluded.scheduled_at,
        started_at   = null,
        finished_at  = null,
        error        = null
RETURNING *;

-- name: CreateTestExecutionPayload :exec
INSERT INTO test_execution_payloads (test_exec_id, payload)
VALUES ($1, $2)
ON CONFLICT (test_exec_id) DO UPDATE
    SET payload = excluded.payload;

-- name: GetTestExecutionPayload :one
SELECT *
FROM test_execution_payloads
WHERE test_exec_id = $1;

-- name: UpdateTestExecutionStarted :one
UPDATE test_executions
SET started_at  = $2,
    finished_at = null,
    error       = null
WHERE id = $1
RETURNING *;

-- name: UpdateTestExecutionFinished :one
UPDATE test_executions
SET finished_at = $2,
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
    (sqlc.narg('last_scheduled_at')::timestamp IS NULL AND sqlc.narg('last_exec_id')::uuid IS NULL)
        OR (scheduled_at, id) < (@last_scheduled_at::timestamp, @last_exec_id::uuid)
    )
ORDER BY scheduled_at DESC, id DESC
LIMIT (sqlc.narg('page_size')::integer);
