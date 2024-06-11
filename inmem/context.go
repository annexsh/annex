package inmem

import (
	"context"

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

func (c *ContextReader) ContextExists(_ context.Context, id string) (bool, error) {
	return c.db.contexts.Contains(id), nil
}

type ContextWriter struct {
	db *DB
}

func NewContextWriter(db *DB) *ContextWriter {
	return &ContextWriter{db: db}
}

func (c *ContextWriter) CreateContext(_ context.Context, id string) error {
	c.db.contexts.Add(id)
	return nil
}
