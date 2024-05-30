-- name: CreateCaseExecution :one
INSERT INTO case_executions (id, test_exec_id, case_name, scheduled_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, test_exec_id) DO UPDATE -- safeguard: shouldn't occur in theory
    SET case_name    = excluded.case_name,
        scheduled_at = excluded.scheduled_at,
        started_at   = null,
        finished_at  = null,
        error        = null
RETURNING *;

-- name: ResetCaseExecution :one
UPDATE case_executions
SET scheduled_at = null,
    started_at   = null,
    finished_at  = null,
    error        = null
WHERE id = $1
  AND test_exec_id = $2
RETURNING *;

-- name: UpdateCaseExecutionStarted :one
UPDATE case_executions
SET started_at = $3
WHERE id = $1
  AND test_exec_id = $2
RETURNING *;

-- name: UpdateCaseExecutionFinished :one
UPDATE case_executions
SET finished_at = $3,
    error       = $4
WHERE id = $1
  AND test_exec_id = $2
RETURNING *;

-- name: DeleteCaseExecution :exec
DELETE
FROM case_executions
WHERE id = $1
  AND test_exec_id = $2;

-- name: GetCaseExecution :one
SELECT *
FROM case_executions
WHERE id = $1
  AND test_exec_id = $2;

-- name: ListTestCaseExecutions :many
SELECT *
FROM case_executions
WHERE test_exec_id = $1;

