// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: test.sql

package sqlc

import (
	"context"

	"github.com/annexsh/annex/uuid"
)

const createTest = `-- name: CreateTest :one
INSERT INTO tests (context_id, group_id, id, name, has_input)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (context_id, group_id, name) DO UPDATE
    SET has_input   = excluded.has_input,
        create_time = now()
RETURNING context_id, group_id, id, name, has_input, create_time
`

type CreateTestParams struct {
	ContextID string  `json:"context_id"`
	GroupID   string  `json:"group_id"`
	ID        uuid.V7 `json:"id"`
	Name      string  `json:"name"`
	HasInput  bool    `json:"has_input"`
}

func (q *Queries) CreateTest(ctx context.Context, arg CreateTestParams) (*Test, error) {
	row := q.db.QueryRow(ctx, createTest,
		arg.ContextID,
		arg.GroupID,
		arg.ID,
		arg.Name,
		arg.HasInput,
	)
	var i Test
	err := row.Scan(
		&i.ContextID,
		&i.GroupID,
		&i.ID,
		&i.Name,
		&i.HasInput,
		&i.CreateTime,
	)
	return &i, err
}

const createTestDefaultInput = `-- name: CreateTestDefaultInput :exec
INSERT INTO test_default_inputs (test_id, data)
VALUES ($1, $2)
ON CONFLICT (test_id) DO UPDATE
    SET data = excluded.data
`

type CreateTestDefaultInputParams struct {
	TestID uuid.V7 `json:"test_id"`
	Data   []byte  `json:"data"`
}

func (q *Queries) CreateTestDefaultInput(ctx context.Context, arg CreateTestDefaultInputParams) error {
	_, err := q.db.Exec(ctx, createTestDefaultInput, arg.TestID, arg.Data)
	return err
}

const getTest = `-- name: GetTest :one
SELECT context_id, group_id, id, name, has_input, create_time
FROM tests
WHERE id = $1
`

func (q *Queries) GetTest(ctx context.Context, id uuid.V7) (*Test, error) {
	row := q.db.QueryRow(ctx, getTest, id)
	var i Test
	err := row.Scan(
		&i.ContextID,
		&i.GroupID,
		&i.ID,
		&i.Name,
		&i.HasInput,
		&i.CreateTime,
	)
	return &i, err
}

const getTestByName = `-- name: GetTestByName :one
SELECT context_id, group_id, id, name, has_input, create_time
FROM tests
WHERE name = $1
  AND group_id = $2
`

type GetTestByNameParams struct {
	Name    string `json:"name"`
	GroupID string `json:"group_id"`
}

func (q *Queries) GetTestByName(ctx context.Context, arg GetTestByNameParams) (*Test, error) {
	row := q.db.QueryRow(ctx, getTestByName, arg.Name, arg.GroupID)
	var i Test
	err := row.Scan(
		&i.ContextID,
		&i.GroupID,
		&i.ID,
		&i.Name,
		&i.HasInput,
		&i.CreateTime,
	)
	return &i, err
}

const getTestDefaultInput = `-- name: GetTestDefaultInput :one
SELECT test_id, data
FROM test_default_inputs
WHERE test_id = $1
`

func (q *Queries) GetTestDefaultInput(ctx context.Context, testID uuid.V7) (*TestDefaultInput, error) {
	row := q.db.QueryRow(ctx, getTestDefaultInput, testID)
	var i TestDefaultInput
	err := row.Scan(&i.TestID, &i.Data)
	return &i, err
}

const listTests = `-- name: ListTests :many
SELECT context_id, group_id, id, name, has_input, create_time
FROM tests
WHERE context_id = $1 AND group_id = $2
`

type ListTestsParams struct {
	ContextID string `json:"context_id"`
	GroupID   string `json:"group_id"`
}

func (q *Queries) ListTests(ctx context.Context, arg ListTestsParams) ([]*Test, error) {
	rows, err := q.db.Query(ctx, listTests, arg.ContextID, arg.GroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Test
	for rows.Next() {
		var i Test
		if err := rows.Scan(
			&i.ContextID,
			&i.GroupID,
			&i.ID,
			&i.Name,
			&i.HasInput,
			&i.CreateTime,
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
