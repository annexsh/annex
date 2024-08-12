-- name: CreateCaseExecution :one
INSERT INTO case_executions (id, test_execution_id, case_name, schedule_time)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, test_execution_id) DO UPDATE -- safeguard: shouldn't occur in theory
    SET case_name    = excluded.case_name,
        schedule_time = excluded.schedule_time,
        start_time   = null,
        finish_time  = null,
        error        = null
RETURNING *;

-- name: ResetCaseExecution :one
UPDATE case_executions
SET schedule_time = null,
    start_time   = null,
    finish_time  = null,
    error        = null
WHERE id = $1
  AND test_execution_id = $2
RETURNING *;

-- name: UpdateCaseExecutionStarted :one
UPDATE case_executions
SET start_time = $3
WHERE id = $1
  AND test_execution_id = $2
RETURNING *;

-- name: UpdateCaseExecutionFinished :one
UPDATE case_executions
SET finish_time = $3,
    error       = $4
WHERE id = $1
  AND test_execution_id = $2
RETURNING *;

-- name: DeleteCaseExecution :exec
DELETE
FROM case_executions
WHERE id = $1
  AND test_execution_id = $2;

-- name: GetCaseExecution :one
SELECT *
FROM case_executions
WHERE id = $1
  AND test_execution_id = $2;

-- name: ListCaseExecutions :many
SELECT *
FROM case_executions
WHERE test_execution_id = $1
ORDER BY schedule_time;
