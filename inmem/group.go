package inmem

import (
	"context"

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

func (g *GroupReader) GroupExists(_ context.Context, contextID string, name string) (bool, error) {
	return g.db.groups.Contains(getGroupKey(contextID, name)), nil
}

type GroupWriter struct {
	db *DB
}

func NewGroupWriter(db *DB) *GroupWriter {
	return &GroupWriter{db: db}
}

func (g *GroupWriter) CreateGroup(_ context.Context, contextID string, name string) error {
	g.db.groups.Add(getGroupKey(contextID, name))
	return nil
}

type groupKey struct {
	contextID string
	name      string
}

func getGroupKey(contextID string, name string) groupKey {
	return groupKey{
		contextID: contextID,
		name:      name,
	}
}
