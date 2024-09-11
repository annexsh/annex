package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/sqlite/sqlc"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
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

func (t *TestReader) GetTest(ctx context.Context, id uuid.V7) (*test.Test, error) {
	tt, err := t.db.GetTest(ctx, id)
	if err != nil {
		return nil, err
	}
	return marshalTest(tt), nil
}

func (t *TestReader) ListTests(ctx context.Context, contextID string, groupID string, filter test.PageFilter[uuid.V7]) (test.TestList, error) {
	params := sqlc.ListTestsParams{
		ContextID: contextID,
		GroupID:   groupID,
		PageSize:  int64(filter.Size),
	}
	if filter.OffsetID != nil {
		params.OffsetID = ptr.Get(filter.OffsetID.String())
	}

	tests, err := t.db.ListTests(ctx, params)
	if err != nil {
		return nil, err
	}
	return marshalTests(tests), nil
}

func (t *TestReader) GetTestDefaultInput(ctx context.Context, testID uuid.V7) (*test.Payload, error) {
	payload, err := t.db.GetTestDefaultInput(ctx, testID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, test.ErrorTestPayloadNotFound
		}
		return nil, err
	}
	return marshalTestDefaultInput(payload), nil
}

type TestWriter struct {
	db *DB
}

func NewTestWriter(db *DB) *TestWriter {
	return &TestWriter{db: db}
}

func (t *TestWriter) CreateTest(ctx context.Context, test *test.Test) (*test.Test, error) {
	tt, err := t.db.CreateTest(ctx, sqlc.CreateTestParams{
		ContextID:  test.ContextID,
		GroupID:    test.GroupID,
		ID:         test.ID,
		Name:       test.Name,
		HasInput:   test.HasInput,
		CreateTime: test.CreateTime,
	})
	if err != nil {
		return nil, err
	}
	return marshalTest(tt), nil
}

func (t *TestWriter) CreateTestDefaultInput(ctx context.Context, testID uuid.V7, defaultInput *test.Payload) error {
	return t.db.CreateTestDefaultInput(ctx, sqlc.CreateTestDefaultInputParams{
		TestID: testID.String(),
		Data:   defaultInput.Data,
	})
}
