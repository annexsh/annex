// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: case_execution.sql

package sqlc

import (
	"context"
	"time"

	"github.com/annexsh/annex/test"
)

const createCaseExecution = `-- name: CreateCaseExecution :one
INSERT INTO case_executions (id, test_execution_id, case_name, schedule_time)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, test_execution_id) DO UPDATE -- safeguard: shouldn't occur in theory
    SET case_name    = excluded.case_name,
        schedule_time = excluded.schedule_time,
        start_time   = null,
        finish_time  = null,
        error        = null
RETURNING id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
`

type CreateCaseExecutionParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	CaseName        string               `json:"case_name"`
	ScheduleTime    time.Time            `json:"schedule_time"`
}

func (q *Queries) CreateCaseExecution(ctx context.Context, arg CreateCaseExecutionParams) (*CaseExecution, error) {
	row := q.db.QueryRow(ctx, createCaseExecution,
		arg.ID,
		arg.TestExecutionID,
		arg.CaseName,
		arg.ScheduleTime,
	)
	var i CaseExecution
	err := row.Scan(
		&i.ID,
		&i.TestExecutionID,
		&i.CaseName,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}

const deleteCaseExecution = `-- name: DeleteCaseExecution :exec
DELETE
FROM case_executions
WHERE id = $1
  AND test_execution_id = $2
`

type DeleteCaseExecutionParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
}

func (q *Queries) DeleteCaseExecution(ctx context.Context, arg DeleteCaseExecutionParams) error {
	_, err := q.db.Exec(ctx, deleteCaseExecution, arg.ID, arg.TestExecutionID)
	return err
}

const getCaseExecution = `-- name: GetCaseExecution :one
SELECT id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
FROM case_executions
WHERE id = $1
  AND test_execution_id = $2
`

type GetCaseExecutionParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
}

func (q *Queries) GetCaseExecution(ctx context.Context, arg GetCaseExecutionParams) (*CaseExecution, error) {
	row := q.db.QueryRow(ctx, getCaseExecution, arg.ID, arg.TestExecutionID)
	var i CaseExecution
	err := row.Scan(
		&i.ID,
		&i.TestExecutionID,
		&i.CaseName,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}

const listCaseExecutions = `-- name: ListCaseExecutions :many
SELECT id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
FROM case_executions
WHERE test_execution_id = $1
`

func (q *Queries) ListCaseExecutions(ctx context.Context, testExecutionID test.TestExecutionID) ([]*CaseExecution, error) {
	rows, err := q.db.Query(ctx, listCaseExecutions, testExecutionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*CaseExecution
	for rows.Next() {
		var i CaseExecution
		if err := rows.Scan(
			&i.ID,
			&i.TestExecutionID,
			&i.CaseName,
			&i.ScheduleTime,
			&i.StartTime,
			&i.FinishTime,
			&i.Error,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const resetCaseExecution = `-- name: ResetCaseExecution :one
UPDATE case_executions
SET schedule_time = null,
    start_time   = null,
    finish_time  = null,
    error        = null
WHERE id = $1
  AND test_execution_id = $2
RETURNING id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
`

type ResetCaseExecutionParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
}

func (q *Queries) ResetCaseExecution(ctx context.Context, arg ResetCaseExecutionParams) (*CaseExecution, error) {
	row := q.db.QueryRow(ctx, resetCaseExecution, arg.ID, arg.TestExecutionID)
	var i CaseExecution
	err := row.Scan(
		&i.ID,
		&i.TestExecutionID,
		&i.CaseName,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}

const updateCaseExecutionFinished = `-- name: UpdateCaseExecutionFinished :one
UPDATE case_executions
SET finish_time = $3,
    error       = $4
WHERE id = $1
  AND test_execution_id = $2
RETURNING id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
`

type UpdateCaseExecutionFinishedParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	FinishTime      *time.Time           `json:"finish_time"`
	Error           *string              `json:"error"`
}

func (q *Queries) UpdateCaseExecutionFinished(ctx context.Context, arg UpdateCaseExecutionFinishedParams) (*CaseExecution, error) {
	row := q.db.QueryRow(ctx, updateCaseExecutionFinished,
		arg.ID,
		arg.TestExecutionID,
		arg.FinishTime,
		arg.Error,
	)
	var i CaseExecution
	err := row.Scan(
		&i.ID,
		&i.TestExecutionID,
		&i.CaseName,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}

const updateCaseExecutionStarted = `-- name: UpdateCaseExecutionStarted :one
UPDATE case_executions
SET start_time = $3
WHERE id = $1
  AND test_execution_id = $2
RETURNING id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
`

type UpdateCaseExecutionStartedParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	StartTime       *time.Time           `json:"start_time"`
}

func (q *Queries) UpdateCaseExecutionStarted(ctx context.Context, arg UpdateCaseExecutionStartedParams) (*CaseExecution, error) {
	row := q.db.QueryRow(ctx, updateCaseExecutionStarted, arg.ID, arg.TestExecutionID, arg.StartTime)
	var i CaseExecution
	err := row.Scan(
		&i.ID,
		&i.TestExecutionID,
		&i.CaseName,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}
