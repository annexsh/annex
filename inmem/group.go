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

func (g *GroupReader) ListGroups(_ context.Context, contextID string) ([]string, error) {
	var groups []string
	for group := range g.db.groups.Iter() {
		if group.contextID == contextID {
			groups = append(groups, group.groupID)
		}
	}
	return groups, nil
}

func (g *GroupReader) GroupExists(_ context.Context, contextID string, groupID string) (bool, error) {
	return g.db.groups.Contains(getGroupKey(contextID, groupID)), nil
}

type GroupWriter struct {
	db *DB
}

func NewGroupWriter(db *DB) *GroupWriter {
	return &GroupWriter{db: db}
}

func (g *GroupWriter) CreateGroup(_ context.Context, contextID string, groupID string) error {
	g.db.groups.Add(getGroupKey(contextID, groupID))
	return nil
}

type groupKey struct {
	contextID string
	groupID   string
}

func getGroupKey(contextID string, groupID string) groupKey {
	return groupKey{
		contextID: contextID,
		groupID:   groupID,
	}
}
