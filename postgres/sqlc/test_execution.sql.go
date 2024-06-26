// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: test_execution.sql

package sqlc

import (
	"context"

	"github.com/annexsh/annex/test"
	"github.com/google/uuid"
)

const createTestExecution = `-- name: CreateTestExecution :one
INSERT INTO test_executions (id, test_id, has_input, schedule_time)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
    SET test_id       = excluded.test_id,
        has_input     = excluded.has_input,
        schedule_time = excluded.schedule_time,
        start_time    = null,
        finish_time   = null,
        error         = null
RETURNING id, test_id, has_input, schedule_time, start_time, finish_time, error
`

type CreateTestExecutionParams struct {
	ID           test.TestExecutionID `json:"id"`
	TestID       uuid.UUID            `json:"test_id"`
	HasInput     bool                 `json:"has_input"`
	ScheduleTime Timestamp            `json:"schedule_time"`
}

func (q *Queries) CreateTestExecution(ctx context.Context, arg CreateTestExecutionParams) (*TestExecution, error) {
	row := q.db.QueryRow(ctx, createTestExecution,
		arg.ID,
		arg.TestID,
		arg.HasInput,
		arg.ScheduleTime,
	)
	var i TestExecution
	err := row.Scan(
		&i.ID,
		&i.TestID,
		&i.HasInput,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}

const createTestExecutionInput = `-- name: CreateTestExecutionInput :exec
INSERT INTO test_execution_inputs (test_execution_id, data)
VALUES ($1, $2)
ON CONFLICT (test_execution_id) DO UPDATE
    SET data = excluded.data
`

type CreateTestExecutionInputParams struct {
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	Data            []byte               `json:"data"`
}

func (q *Queries) CreateTestExecutionInput(ctx context.Context, arg CreateTestExecutionInputParams) error {
	_, err := q.db.Exec(ctx, createTestExecutionInput, arg.TestExecutionID, arg.Data)
	return err
}

const getTestExecution = `-- name: GetTestExecution :one
SELECT id, test_id, has_input, schedule_time, start_time, finish_time, error
FROM test_executions
WHERE id = $1
`

func (q *Queries) GetTestExecution(ctx context.Context, id test.TestExecutionID) (*TestExecution, error) {
	row := q.db.QueryRow(ctx, getTestExecution, id)
	var i TestExecution
	err := row.Scan(
		&i.ID,
		&i.TestID,
		&i.HasInput,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}

const getTestExecutionInput = `-- name: GetTestExecutionInput :one
SELECT test_execution_id, data
FROM test_execution_inputs
WHERE test_execution_id = $1
`

func (q *Queries) GetTestExecutionInput(ctx context.Context, testExecutionID test.TestExecutionID) (*TestExecutionInput, error) {
	row := q.db.QueryRow(ctx, getTestExecutionInput, testExecutionID)
	var i TestExecutionInput
	err := row.Scan(&i.TestExecutionID, &i.Data)
	return &i, err
}

const listTestExecutions = `-- name: ListTestExecutions :many
SELECT id, test_id, has_input, schedule_time, start_time, finish_time, error
FROM test_executions
WHERE ($1 = test_id)
  AND (
    ($2::timestamp IS NULL AND $3::uuid IS NULL)
        OR (schedule_time, id) < ($2::timestamp, $4::uuid)
    )
ORDER BY schedule_time DESC, id DESC
LIMIT ($5::integer)
`

type ListTestExecutionsParams struct {
	TestID              uuid.UUID  `json:"test_id"`
	LastScheduleTime    Timestamp  `json:"last_schedule_time"`
	LastExecID          *uuid.UUID `json:"last_exec_id"`
	LastTestExecutionID uuid.UUID  `json:"last_test_execution_id"`
	PageSize            *int32     `json:"page_size"`
}

func (q *Queries) ListTestExecutions(ctx context.Context, arg ListTestExecutionsParams) ([]*TestExecution, error) {
	rows, err := q.db.Query(ctx, listTestExecutions,
		arg.TestID,
		arg.LastScheduleTime,
		arg.LastExecID,
		arg.LastTestExecutionID,
		arg.PageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*TestExecution
	for rows.Next() {
		var i TestExecution
		if err := rows.Scan(
			&i.ID,
			&i.TestID,
			&i.HasInput,
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

const updateTestExecutionFinished = `-- name: UpdateTestExecutionFinished :one
UPDATE test_executions
SET finish_time = $2,
    error       = $3
WHERE id = $1
RETURNING id, test_id, has_input, schedule_time, start_time, finish_time, error
`

type UpdateTestExecutionFinishedParams struct {
	ID         test.TestExecutionID `json:"id"`
	FinishTime Timestamp            `json:"finish_time"`
	Error      *string              `json:"error"`
}

func (q *Queries) UpdateTestExecutionFinished(ctx context.Context, arg UpdateTestExecutionFinishedParams) (*TestExecution, error) {
	row := q.db.QueryRow(ctx, updateTestExecutionFinished, arg.ID, arg.FinishTime, arg.Error)
	var i TestExecution
	err := row.Scan(
		&i.ID,
		&i.TestID,
		&i.HasInput,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}

const updateTestExecutionStarted = `-- name: UpdateTestExecutionStarted :one
UPDATE test_executions
SET start_time  = $2,
    finish_time = null,
    error       = null
WHERE id = $1
RETURNING id, test_id, has_input, schedule_time, start_time, finish_time, error
`

type UpdateTestExecutionStartedParams struct {
	ID        test.TestExecutionID `json:"id"`
	StartTime Timestamp            `json:"start_time"`
}

func (q *Queries) UpdateTestExecutionStarted(ctx context.Context, arg UpdateTestExecutionStartedParams) (*TestExecution, error) {
	row := q.db.QueryRow(ctx, updateTestExecutionStarted, arg.ID, arg.StartTime)
	var i TestExecution
	err := row.Scan(
		&i.ID,
		&i.TestID,
		&i.HasInput,
		&i.ScheduleTime,
		&i.StartTime,
		&i.FinishTime,
		&i.Error,
	)
	return &i, err
}
