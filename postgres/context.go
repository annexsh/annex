package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/annexsh/annex/postgres/sqlc"
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
		OffsetID: filter.OffsetID,
		PageSize: int32(filter.Size),
	})
}

type ContextWriter struct {
	db *DB
}

func NewContextWriter(db *DB) *ContextWriter {
	return &ContextWriter{db: db}
}

func (c *ContextWriter) CreateContext(ctx context.Context, id string) error {
	err := c.db.CreateContext(ctx, id)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		if id == "default" {
			return nil
		}
		return test.ErrorContextAlreadyExists
	}

	return err
}
