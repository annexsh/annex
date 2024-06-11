package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

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

func (g *GroupReader) GroupExists(ctx context.Context, contextID string, name string) (bool, error) {
	if err := g.db.GroupExists(ctx, sqlc.GroupExistsParams{
		ContextID: contextID,
		Name:      name,
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

func (g *GroupWriter) CreateGroup(ctx context.Context, contextID string, name string) error {
	return g.db.CreateGroup(ctx, sqlc.CreateGroupParams{
		ContextID: contextID,
		Name:      name,
	})
}
