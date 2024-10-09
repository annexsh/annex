// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: case_execution.sql

package sqlc

import (
	"context"
	"time"

	"github.com/annexsh/annex/test"
)

const createCaseExecutionScheduled = `-- name: CreateCaseExecutionScheduled :one
INSERT INTO case_executions (id, test_execution_id, case_name, schedule_time)
VALUES (?, ?, ?, ?)
ON CONFLICT(id, test_execution_id) DO UPDATE
    SET case_name     = excluded.case_name,
        schedule_time = excluded.schedule_time,
        start_time    = NULL,
        finish_time   = NULL,
        error         = NULL
RETURNING id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
`

type CreateCaseExecutionScheduledParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	CaseName        string               `json:"case_name"`
	ScheduleTime    time.Time            `json:"schedule_time"`
}

func (q *Queries) CreateCaseExecutionScheduled(ctx context.Context, arg CreateCaseExecutionScheduledParams) (*CaseExecution, error) {
	row := q.db.QueryRowContext(ctx, createCaseExecutionScheduled,
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
WHERE id = ?
  AND test_execution_id = ?
`

type DeleteCaseExecutionParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
}

func (q *Queries) DeleteCaseExecution(ctx context.Context, arg DeleteCaseExecutionParams) error {
	_, err := q.db.ExecContext(ctx, deleteCaseExecution, arg.ID, arg.TestExecutionID)
	return err
}

const getCaseExecution = `-- name: GetCaseExecution :one
SELECT id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
FROM case_executions
WHERE id = ?
  AND test_execution_id = ?
`

type GetCaseExecutionParams struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
}

func (q *Queries) GetCaseExecution(ctx context.Context, arg GetCaseExecutionParams) (*CaseExecution, error) {
	row := q.db.QueryRowContext(ctx, getCaseExecution, arg.ID, arg.TestExecutionID)
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
WHERE (test_execution_id = ?1)
  -- Cast as integer required below since sqlc.narg doesn't work with overridden column type
  AND (CAST(?2 AS INTEGER) IS NULL OR id > CAST(?2 AS INTEGER))
ORDER BY id
LIMIT ?3
`

type ListCaseExecutionsParams struct {
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	OffsetID        *int64               `json:"offset_id"`
	PageSize        int64                `json:"page_size"`
}

func (q *Queries) ListCaseExecutions(ctx context.Context, arg ListCaseExecutionsParams) ([]*CaseExecution, error) {
	rows, err := q.db.QueryContext(ctx, listCaseExecutions, arg.TestExecutionID, arg.OffsetID, arg.PageSize)
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
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCaseExecutionFinished = `-- name: UpdateCaseExecutionFinished :one
UPDATE case_executions
SET finish_time = ?,
    error       = ?
WHERE id = ?
  AND test_execution_id = ?
RETURNING id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
`

type UpdateCaseExecutionFinishedParams struct {
	FinishTime      *time.Time           `json:"finish_time"`
	Error           *string              `json:"error"`
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
}

func (q *Queries) UpdateCaseExecutionFinished(ctx context.Context, arg UpdateCaseExecutionFinishedParams) (*CaseExecution, error) {
	row := q.db.QueryRowContext(ctx, updateCaseExecutionFinished,
		arg.FinishTime,
		arg.Error,
		arg.ID,
		arg.TestExecutionID,
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
SET start_time = ?
WHERE id = ?
  AND test_execution_id = ?
RETURNING id, test_execution_id, case_name, schedule_time, start_time, finish_time, error
`

type UpdateCaseExecutionStartedParams struct {
	StartTime       *time.Time           `json:"start_time"`
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
}

func (q *Queries) UpdateCaseExecutionStarted(ctx context.Context, arg UpdateCaseExecutionStartedParams) (*CaseExecution, error) {
	row := q.db.QueryRowContext(ctx, updateCaseExecutionStarted, arg.StartTime, arg.ID, arg.TestExecutionID)
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