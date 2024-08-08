package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/annexsh/annex/postgres/sqlc"

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

func (t *TestReader) GetTest(ctx context.Context, id uuid.UUID) (*test.Test, error) {
	tt, err := t.db.GetTest(ctx, id)
	if err != nil {
		return nil, err
	}
	return marshalTest(tt), nil
}

func (t *TestReader) ListTests(ctx context.Context, contextID string, groupID string) (test.TestList, error) {
	// TODO: pagination
	tests, err := t.db.ListTests(ctx, sqlc.ListTestsParams{
		ContextID: contextID,
		GroupID:   groupID,
	})
	if err != nil {
		return nil, err
	}
	return marshalTests(tests), nil
}

func (t *TestReader) GetTestDefaultInput(ctx context.Context, testID uuid.UUID) (*test.Payload, error) {
	payload, err := t.db.GetTestDefaultInput(ctx, testID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func (t *TestWriter) CreateTest(ctx context.Context, definition *test.TestDefinition) (*test.Test, error) {
	var tt *test.Test

	if err := t.db.ExecuteTx(ctx, func(querier sqlc.Querier) error {
		created, err := createTest(ctx, querier, definition)
		if err != nil {
			return err
		}
		tt = marshalTest(created)
		return nil
	}); err != nil {
		return nil, err
	}

	return tt, nil
}

func (t *TestWriter) CreateTests(ctx context.Context, definitions ...*test.TestDefinition) (test.TestList, error) {
	tests := make(test.TestList, len(definitions))

	if err := t.db.ExecuteTx(ctx, func(querier sqlc.Querier) error {
		for i, def := range definitions {
			created, err := createTest(ctx, querier, def)
			if err != nil {
				return err
			}
			tests[i] = marshalTest(created)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return tests, nil
}

func createTest(ctx context.Context, querier sqlc.Querier, definition *test.TestDefinition) (*sqlc.Test, error) {
	created, err := querier.CreateTest(ctx, sqlc.CreateTestParams{
		ContextID: definition.ContextID,
		GroupID:   definition.GroupID,
		ID:        definition.TestID,
		Name:      definition.Name,
		HasInput:  definition.DefaultInput != nil,
	})
	if err != nil {
		return nil, err
	}
	if created.HasInput {
		if err = querier.CreateTestDefaultInput(ctx, sqlc.CreateTestDefaultInputParams{
			TestID: created.ID,
			Data:   definition.DefaultInput.Data,
		}); err != nil {
			return nil, err
		}
	}
	return created, nil
}
