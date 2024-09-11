package sqlite

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/annexsh/annex/sqlite/sqlc"
	"github.com/annexsh/annex/test"
)

var (
	_ test.GroupReader = (*GroupReader)(nil)
	_ test.GroupWriter = (*GroupWriter)(nil)
)

type GroupReader struct {
	db *DB
}

func NewGroupReader(db *DB) *GroupReader {
	return &GroupReader{db: db}
}

func (g *GroupReader) ListGroups(ctx context.Context, contextID string, filter test.PageFilter[string]) ([]string, error) {
	groups, err := g.db.ListGroups(ctx, sqlc.ListGroupsParams{
		ContextID: contextID,
		PageSize:  int64(filter.Size),
		OffsetID:  filter.OffsetID,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(groups))
	for i := range groups {
		ids[i] = groups[i].ID
	}

	return ids, nil
}

func (g *GroupReader) GroupExists(ctx context.Context, contextID string, groupID string) (bool, error) {
	if err := g.db.GroupExists(ctx, sqlc.GroupExistsParams{
		ContextID: contextID,
		ID:        groupID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type GroupWriter struct {
	db *DB
}

func NewGroupWriter(db *DB) *GroupWriter {
	return &GroupWriter{db: db}
}

func (g *GroupWriter) CreateGroup(ctx context.Context, contextID string, groupID string) error {
	return g.db.CreateGroup(ctx, sqlc.CreateGroupParams{
		ContextID: contextID,
		ID:        groupID,
	})
}
