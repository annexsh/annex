package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/postgres/sqlc"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

var (
	_ test.TestExecutionReader = (*TestExecutionReader)(nil)
	_ test.TestExecutionWriter = (*TestExecutionWriter)(nil)
)

type TestExecutionReader struct {
	db *DB
}

func NewTestExecutionReader(db *DB) *TestExecutionReader {
	return &TestExecutionReader{db: db}
}

func (t *TestExecutionReader) GetTestExecution(ctx context.Context, id test.TestExecutionID) (*test.TestExecution, error) {
	exec, err := t.db.GetTestExecution(ctx, id)
	if err != nil {
		return nil, err
	}
	return marshalTestExec(exec), nil
}

func (t *TestExecutionReader) GetTestExecutionInput(ctx context.Context, id test.TestExecutionID) (*test.Payload, error) {
	payload, err := t.db.GetTestExecutionInput(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, test.ErrorTestExecutionPayloadNotFound
		}
		return nil, err
	}
	return marshalTestExecInput(payload), nil
}

func (t *TestExecutionReader) ListTestExecutions(ctx context.Context, testID uuid.V7, filter test.PageFilter[test.TestExecutionID]) (test.TestExecutionList, error) {
	params := sqlc.ListTestExecutionsParams{
		TestID:   testID,
		PageSize: int32(filter.Size),
	}
	if filter.OffsetID != nil {
		params.OffsetID = &filter.OffsetID.V7
	}

	execs, err := t.db.ListTestExecutions(ctx, params)
	if err != nil {
		return nil, err
	}

	return marshalTestExecs(execs), nil
}

type TestExecutionWriter struct {
	db *DB
}

func NewTestExecutionWriter(db *DB) *TestExecutionWriter {
	return &TestExecutionWriter{db: db}
}

func (t *TestExecutionWriter) CreateTestExecutionInput(ctx context.Context, testExecID test.TestExecutionID, input *test.Payload) error {
	return t.db.CreateTestExecutionInput(ctx, sqlc.CreateTestExecutionInputParams{
		TestExecutionID: testExecID,
		Data:            input.Data,
	})
}

func (t *TestExecutionWriter) CreateTestExecutionScheduled(ctx context.Context, scheduled *test.ScheduledTestExecution) (*test.TestExecution, error) {
	testExec, err := t.db.CreateTestExecutionScheduled(ctx, sqlc.CreateTestExecutionScheduledParams{
		ID:           scheduled.ID,
		TestID:       scheduled.TestID,
		HasInput:     scheduled.HasInput,
		ScheduleTime: scheduled.ScheduleTime.UTC(),
	})
	if err != nil {
		return nil, err
	}

	return marshalTestExec(testExec), nil
}

func (t *TestExecutionWriter) UpdateTestExecutionStarted(ctx context.Context, started *test.StartedTestExecution) (*test.TestExecution, error) {
	exec, err := t.db.UpdateTestExecutionStarted(ctx, sqlc.UpdateTestExecutionStartedParams{
		ID:        started.ID,
		StartTime: ptr.Get(started.StartTime.UTC()),
	})
	if err != nil {
		return nil, err
	}
	return marshalTestExec(exec), nil
}

func (t *TestExecutionWriter) UpdateTestExecutionFinished(ctx context.Context, finished *test.FinishedTestExecution) (*test.TestExecution, error) {
	exec, err := t.db.UpdateTestExecutionFinished(ctx, sqlc.UpdateTestExecutionFinishedParams{
		ID:         finished.ID,
		FinishTime: ptr.Get(finished.FinishTime.UTC()),
		Error:      finished.Error,
	})
	if err != nil {
		return nil, err
	}
	return marshalTestExec(exec), nil
}

func (t *TestExecutionWriter) ResetTestExecution(ctx context.Context, testExecID test.TestExecutionID, resetTime time.Time) (*test.TestExecution, error) {
	exec, err := t.db.ResetTestExecution(ctx, sqlc.ResetTestExecutionParams{
		ID:        testExecID,
		ResetTime: resetTime,
	})
	if err != nil {
		return nil, err
	}
	return marshalTestExec(exec), nil
}
