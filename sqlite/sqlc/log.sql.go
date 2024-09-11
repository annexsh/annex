// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: log.sql

package sqlc

import (
	"context"
	"time"

	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

const createLog = `-- name: CreateLog :exec
INSERT INTO logs (id, test_execution_id, case_execution_id, level, message, create_time)
VALUES (?, ?, ?, ?, ?, ?)
`

type CreateLogParams struct {
	ID              uuid.V7               `json:"id"`
	TestExecutionID test.TestExecutionID  `json:"test_execution_id"`
	CaseExecutionID *test.CaseExecutionID `json:"case_execution_id"`
	Level           string                `json:"level"`
	Message         string                `json:"message"`
	CreateTime      time.Time             `json:"create_time"`
}

func (q *Queries) CreateLog(ctx context.Context, arg CreateLogParams) error {
	_, err := q.db.ExecContext(ctx, createLog,
		arg.ID,
		arg.TestExecutionID,
		arg.CaseExecutionID,
		arg.Level,
		arg.Message,
		arg.CreateTime,
	)
	return err
}

const deleteLog = `-- name: DeleteLog :exec
DELETE
FROM logs
WHERE id = ?
`

func (q *Queries) DeleteLog(ctx context.Context, id uuid.V7) error {
	_, err := q.db.ExecContext(ctx, deleteLog, id)
	return err
}

const getLog = `-- name: GetLog :one
SELECT id, test_execution_id, case_execution_id, level, message, create_time
FROM logs
WHERE id = ?
`

func (q *Queries) GetLog(ctx context.Context, id uuid.V7) (*Log, error) {
	row := q.db.QueryRowContext(ctx, getLog, id)
	var i Log
	err := row.Scan(
		&i.ID,
		&i.TestExecutionID,
		&i.CaseExecutionID,
		&i.Level,
		&i.Message,
		&i.CreateTime,
	)
	return &i, err
}

const listLogs = `-- name: ListLogs :many
SELECT id, test_execution_id, case_execution_id, level, message, create_time
FROM logs
WHERE (test_execution_id = ?1)
  -- Cast as text required below since sqlc.narg doesn't work with overridden column type
  AND (CAST(?2 AS TEXT) IS NULL OR id < CAST(?2 AS TEXT))
ORDER BY id DESC
LIMIT ?3
`

type ListLogsParams struct {
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	OffsetID        *string              `json:"offset_id"`
	PageSize        int64                `json:"page_size"`
}

func (q *Queries) ListLogs(ctx context.Context, arg ListLogsParams) ([]*Log, error) {
	rows, err := q.db.QueryContext(ctx, listLogs, arg.TestExecutionID, arg.OffsetID, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Log
	for rows.Next() {
		var i Log
		if err := rows.Scan(
			&i.ID,
			&i.TestExecutionID,
			&i.CaseExecutionID,
			&i.Level,
			&i.Message,
			&i.CreateTime,
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
