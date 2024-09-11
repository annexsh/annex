package sqlite

import (
	"context"

	"github.com/annexsh/annex/sqlite/sqlc"
	"github.com/annexsh/annex/test"
)

var (
	_ test.ContextReader = (*ContextReader)(nil)
	_ test.ContextWriter = (*ContextWriter)(nil)
)

type ContextReader struct {
	db *DB
}

func NewContextReader(db *DB) *ContextReader {
	return &ContextReader{db: db}
}

func (c *ContextReader) ListContexts(ctx context.Context, filter test.PageFilter[string]) ([]string, error) {
	return c.db.ListContexts(ctx, sqlc.ListContextsParams{
		PageSize: int64(filter.Size),
		OffsetID: filter.OffsetID,
	})
}

type ContextWriter struct {
	db *DB
}

func NewContextWriter(db *DB) *ContextWriter {
	return &ContextWriter{db: db}
}

func (c *ContextWriter) CreateContext(ctx context.Context, id string) error {
	if err := c.db.CreateContext(ctx, id); err != nil {
		if err.Error() == "constraint failed: UNIQUE constraint failed: contexts.id (1555)" {
			if id == "default" {
				return nil
			}
			return test.ErrorContextAlreadyExists
		}
		return err
	}

	return nil
}
