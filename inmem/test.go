package inmem

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
)

var (
	_ test.TestReader = (*TestReader)(nil)
	_ test.TestWriter = (*TestWriter)(nil)
)

type TestReader struct {
	db *DB
}

func NewTestReader(db *DB) *TestReader {
	return &TestReader{db: db}
}

func (t *TestReader) GetTest(_ context.Context, id uuid.UUID) (*test.Test, error) {
	t.db.mu.RLock()
	defer t.db.mu.RUnlock()

	tt, ok := t.db.tests[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return ptr.Copy(tt), nil
}

func (t *TestReader) ListTests(_ context.Context) (test.TestList, error) {
	t.db.mu.RLock()
	defer t.db.mu.RUnlock()

	i := 0
	tests := make(test.TestList, len(t.db.tests))
	for _, tt := range t.db.tests {
		tests[i] = ptr.Copy(tt)
		i++
	}

	slices.SortFunc(tests, func(a, b *test.Test) int {
		if a.CreateTime.Before(b.CreateTime) || a.CreateTime.Equal(b.CreateTime) && a.ID.String() < b.ID.String() {
			return -1
		}
		return 1
	})

	return tests, nil
}

func (t *TestReader) GetTestDefaultInput(_ context.Context, testID uuid.UUID) (*test.Payload, error) {
	t.db.mu.RLock()
	defer t.db.mu.RUnlock()

	p, ok := t.db.defaultInputs[testID]
	if !ok {
		return nil, errors.New("not found")
	}
	return ptr.Copy(p), nil
}

type TestWriter struct {
	db *DB
}

func NewTestWriter(db *DB) *TestWriter {
	return &TestWriter{db: db}
}

func (t *TestWriter) CreateTest(_ context.Context, definition *test.TestDefinition) (*test.Test, error) {
	t.db.mu.Lock()
	defer t.db.mu.Unlock()
	tt := t.createTestUnsafe(definition)
	return ptr.Copy(tt), nil
}

func (t *TestWriter) CreateTests(_ context.Context, definitions ...*test.TestDefinition) (test.TestList, error) {
	t.db.mu.Lock()
	defer t.db.mu.Unlock()

	tests := make(test.TestList, len(definitions))
	i := 0
	for _, def := range definitions {
		tt := t.createTestUnsafe(def)
		tests[i] = ptr.Copy(tt)
		i++
	}
	return tests, nil
}

func (t *TestWriter) createTestUnsafe(definition *test.TestDefinition) *test.Test {
	for _, tt := range t.db.tests {
		if tt.Context == tt.Context && tt.Group == definition.Group && tt.Name == definition.Name {
			definition.TestID = tt.ID
		}
	}
	tt := &test.Test{
		ID:         definition.TestID,
		Context:    definition.Context,
		Group:      definition.Group,
		Name:       definition.Name,
		HasInput:   definition.DefaultInput != nil,
		CreateTime: time.Now().UTC(),
	}
	t.db.tests[tt.ID] = tt
	t.db.defaultInputs[tt.ID] = definition.DefaultInput
	return tt
}
