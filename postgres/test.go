package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

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

func (t *TestReader) ListTests(ctx context.Context) (test.TestList, error) {
	// TODO: pagination
	tests, err := t.db.ListTests(ctx)
	if err != nil {
		return nil, err
	}
	return marshalTests(tests), nil
}

func (t *TestReader) GetTestDefaultPayload(ctx context.Context, testID uuid.UUID) (*test.Payload, error) {
	payload, err := t.db.GetTestDefaultPayload(ctx, testID)
	if err != nil {
		return nil, err
	}
	return marshalTestDefaultPayload(payload), nil
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
		ID:                definition.TestID,
		Project:           definition.Project,
		Name:              definition.Name,
		HasPayload:        definition.DefaultPayload != nil,
		RunnerID:          definition.RunnerID,
		RunnerHeartbeatAt: sqlc.NewTimestamp(time.Now()),
	})
	if err != nil {
		return nil, err
	}
	if created.HasPayload {
		if err = querier.CreateTestDefaultPayload(ctx, sqlc.CreateTestDefaultPayloadParams{
			TestID:  created.ID,
			Payload: definition.DefaultPayload.Payload,
			IsZero:  definition.DefaultPayload.IsZero,
		}); err != nil {
			return nil, err
		}
	}
	return created, nil
}
