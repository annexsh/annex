package postgres

import (
	"context"

	"github.com/annexsh/annex/postgres/sqlc"
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
		PageSize:  int32(filter.Size),
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
