package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

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

func (c *ContextReader) ContextExists(ctx context.Context, id string) (bool, error) {
	if err := c.db.ContextExists(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type ContextWriter struct {
	db *DB
}

func NewContextWriter(db *DB) *ContextWriter {
	return &ContextWriter{db: db}
}

func (c *ContextWriter) CreateContext(ctx context.Context, id string) error {
	return c.db.CreateContext(ctx, id)
}
