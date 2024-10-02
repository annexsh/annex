package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/annexsh/annex/postgres/sqlc"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

var (
	_ test.TestSuiteReader = (*TestSuiteReader)(nil)
	_ test.TestSuiteWriter = (*TestSuiteWriter)(nil)
)

type TestSuiteReader struct {
	db *DB
}

func NewTestSuiteReader(db *DB) *TestSuiteReader {
	return &TestSuiteReader{db: db}
}

func (t *TestSuiteReader) ListTestSuites(ctx context.Context, contextID string, filter test.PageFilter[string]) (test.TestSuiteList, error) {
	testSuites, err := t.db.ListTestSuites(ctx, sqlc.ListTestSuitesParams{
		ContextID: contextID,
		PageSize:  int32(filter.Size),
		OffsetID:  filter.OffsetID,
	})
	if err != nil {
		return nil, err
	}

	return marshalTestSuites(testSuites), nil
}

func (t *TestSuiteReader) GetTestSuiteVersion(ctx context.Context, contextID string, id uuid.V7) (string, error) {
	version, err := t.db.GetTestSuiteVersion(ctx, sqlc.GetTestSuiteVersionParams{
		ContextID:   contextID,
		TestSuiteID: id,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", test.ErrorTestSuiteNotFound
	}
	return version, nil
}

type TestSuiteWriter struct {
	db *DB
}

func NewTestSuiteWriter(db *DB) *TestSuiteWriter {
	return &TestSuiteWriter{db: db}
}

func (t *TestSuiteWriter) CreateTestSuite(ctx context.Context, testSuite *test.TestSuite) (uuid.V7, error) {
	return t.db.CreateTestSuite(ctx, sqlc.CreateTestSuiteParams{
		ID:          testSuite.ID,
		ContextID:   testSuite.ContextID,
		Name:        testSuite.Name,
		Description: testSuite.Description,
	})
}
