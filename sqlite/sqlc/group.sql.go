// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: group.sql

package sqlc

import (
	"context"
)

const createGroup = `-- name: CreateGroup :exec
INSERT INTO groups (context_id, id)
VALUES (?, ?)
ON CONFLICT (context_id, id) DO NOTHING
`

type CreateGroupParams struct {
	ContextID string `json:"context_id"`
	ID        string `json:"id"`
}

func (q *Queries) CreateGroup(ctx context.Context, arg CreateGroupParams) error {
	_, err := q.db.ExecContext(ctx, createGroup, arg.ContextID, arg.ID)
	return err
}

const groupExists = `-- name: GroupExists :exec
SELECT context_id, id
FROM groups
WHERE context_id = ?
  AND id = ?
`

type GroupExistsParams struct {
	ContextID string `json:"context_id"`
	ID        string `json:"id"`
}

func (q *Queries) GroupExists(ctx context.Context, arg GroupExistsParams) error {
	_, err := q.db.ExecContext(ctx, groupExists, arg.ContextID, arg.ID)
	return err
}

const listGroups = `-- name: ListGroups :many
SELECT context_id, id
FROM groups
WHERE (context_id = ?1)
  AND (id > COALESCE(?2, ''))
ORDER BY id
LIMIT ?3
`

type ListGroupsParams struct {
	ContextID string  `json:"context_id"`
	OffsetID  *string `json:"offset_id"`
	PageSize  int64   `json:"page_size"`
}

func (q *Queries) ListGroups(ctx context.Context, arg ListGroupsParams) ([]*Group, error) {
	rows, err := q.db.QueryContext(ctx, listGroups, arg.ContextID, arg.OffsetID, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Group
	for rows.Next() {
		var i Group
		if err := rows.Scan(&i.ContextID, &i.ID); err != nil {
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
