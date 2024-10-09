// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: test_suite.sql

package sqlc

import (
	"context"

	"github.com/annexsh/annex/uuid"
)

const createTestSuite = `-- name: CreateTestSuite :one
INSERT INTO test_suites (id, context_id, name, description)
VALUES (?, ?, ?, ?)
ON CONFLICT (context_id, name) DO UPDATE
    SET description = excluded.description
RETURNING id
`

type CreateTestSuiteParams struct {
	ID          uuid.V7 `json:"id"`
	ContextID   string  `json:"context_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (q *Queries) CreateTestSuite(ctx context.Context, arg CreateTestSuiteParams) (uuid.V7, error) {
	row := q.db.QueryRowContext(ctx, createTestSuite,
		arg.ID,
		arg.ContextID,
		arg.Name,
		arg.Description,
	)
	var id uuid.V7
	err := row.Scan(&id)
	return id, err
}

const getTestSuiteVersion = `-- name: GetTestSuiteVersion :one
SELECT version
FROM test_suite_registrations
WHERE context_id = ?
  AND test_suite_id = ?
`

type GetTestSuiteVersionParams struct {
	ContextID   string `json:"context_id"`
	TestSuiteID string `json:"test_suite_id"`
}

func (q *Queries) GetTestSuiteVersion(ctx context.Context, arg GetTestSuiteVersionParams) (string, error) {
	row := q.db.QueryRowContext(ctx, getTestSuiteVersion, arg.ContextID, arg.TestSuiteID)
	var version string
	err := row.Scan(&version)
	return version, err
}

const listTestSuites = `-- name: ListTestSuites :many
;

SELECT id, context_id, name, description
FROM test_suites
WHERE (context_id = ?1)
  AND (name > COALESCE(?2, ''))
ORDER BY name
LIMIT ?3
`

type ListTestSuitesParams struct {
	ContextID string  `json:"context_id"`
	OffsetID  *string `json:"offset_id"`
	PageSize  int64   `json:"page_size"`
}

func (q *Queries) ListTestSuites(ctx context.Context, arg ListTestSuitesParams) ([]*TestSuite, error) {
	rows, err := q.db.QueryContext(ctx, listTestSuites, arg.ContextID, arg.OffsetID, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*TestSuite
	for rows.Next() {
		var i TestSuite
		if err := rows.Scan(
			&i.ID,
			&i.ContextID,
			&i.Name,
			&i.Description,
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

const setTestSuiteVersion = `-- name: SetTestSuiteVersion :exec
INSERT INTO test_suite_registrations (context_id, test_suite_id, runner_id, version)
VALUES (?, ?, ?, ?)
ON CONFLICT (context_id, test_suite_id) DO UPDATE
    SET runner_id = excluded.runner_id,
        version   = excluded.version
`

type SetTestSuiteVersionParams struct {
	ContextID   string `json:"context_id"`
	TestSuiteID string `json:"test_suite_id"`
	RunnerID    string `json:"runner_id"`
	Version     string `json:"version"`
}

func (q *Queries) SetTestSuiteVersion(ctx context.Context, arg SetTestSuiteVersionParams) error {
	_, err := q.db.ExecContext(ctx, setTestSuiteVersion,
		arg.ContextID,
		arg.TestSuiteID,
		arg.RunnerID,
		arg.Version,
	)
	return err
}