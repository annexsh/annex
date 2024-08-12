package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
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
		return nil, err
	}
	return marshalTestExecPayload(payload), nil
}

func (t *TestExecutionReader) ListTestExecutions(ctx context.Context, testID uuid.UUID, filter *test.TestExecutionListFilter) (test.TestExecutionList, error) {
	params := sqlc.ListTestExecutionsParams{
		TestID:           testID,
		LastScheduleTime: filter.LastScheduleTime,
		LastExecID:       filter.LastTestExecutionID,
	}
	if filter.PageSize > 0 {
		params.PageSize = ptr.Get(int32(filter.PageSize))
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

func (t *TestExecutionWriter) CreateScheduledTestExecution(ctx context.Context, scheduled *test.ScheduledTestExecution) (*test.TestExecution, error) {
	var testExec *sqlc.TestExecution

	err := t.db.ExecuteTx(ctx, func(querier sqlc.Querier) error {
		exec, err := t.db.CreateTestExecution(ctx, sqlc.CreateTestExecutionParams{
			ID:           scheduled.ID,
			TestID:       scheduled.TestID,
			HasInput:     scheduled.Payload != nil,
			ScheduleTime: scheduled.ScheduleTime.UTC(),
		})
		if err != nil {
			return err
		}
		if exec.HasInput {
			if err = querier.CreateTestExecutionInput(ctx, sqlc.CreateTestExecutionInputParams{
				TestExecutionID: exec.ID,
				Data:            scheduled.Payload,
			}); err != nil {
				return err
			}
		}
		testExec = exec
		return nil
	})
	if err != nil {
		return nil, err
	}

	return marshalTestExec(testExec), nil
}

func (t *TestExecutionWriter) UpdateStartedTestExecution(ctx context.Context, started *test.StartedTestExecution) (*test.TestExecution, error) {
	exec, err := t.db.UpdateTestExecutionStarted(ctx, sqlc.UpdateTestExecutionStartedParams{
		ID:        started.ID,
		StartTime: ptr.Get(started.StartTime.UTC()),
	})
	if err != nil {
		return nil, err
	}
	return marshalTestExec(exec), nil
}

func (t *TestExecutionWriter) UpdateFinishedTestExecution(ctx context.Context, finished *test.FinishedTestExecution) (*test.TestExecution, error) {
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

func (t *TestExecutionWriter) ResetTestExecution(ctx context.Context, reset *test.ResetTestExecution) (*test.TestExecution, test.ResetRollback, error) {
	existing, err := t.db.GetTestExecution(ctx, reset.ID)
	if err != nil {
		return nil, nil, err
	}

	tx, querier, err := t.db.WithTx(ctx)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); err != nil {
				err = fmt.Errorf("%w; failed to rollback reset workflow transaction: %w", err, rbErr)
			}
		}
	}()

	for _, caseExecID := range reset.StaleCaseExecutions {
		if err = querier.DeleteCaseExecution(ctx, sqlc.DeleteCaseExecutionParams{
			ID:              caseExecID,
			TestExecutionID: reset.ID,
		}); err != nil {
			return nil, nil, err
		}
	}

	for _, logID := range reset.StaleLogs {
		if err = querier.DeleteLog(ctx, logID); err != nil {
			return nil, nil, err
		}
	}

	// CreateTestExecution is idempotent. On conflict, it resets the existing
	// workflow to a new scheduled state matching the params below.
	testExec, err := querier.CreateTestExecution(ctx, sqlc.CreateTestExecutionParams{
		ID:           reset.ID,
		TestID:       existing.TestID,
		HasInput:     existing.HasInput,
		ScheduleTime: reset.ResetTime.UTC(),
	})
	if err != nil {
		return nil, nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return marshalTestExec(testExec), tx.Rollback, nil
}
