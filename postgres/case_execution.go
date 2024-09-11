package postgres

import (
	"context"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/test"
)

var (
	_ test.CaseExecutionReader = (*CaseExecutionReader)(nil)
	_ test.CaseExecutionWriter = (*CaseExecutionWriter)(nil)
)

type CaseExecutionReader struct {
	db *DB
}

func NewCaseExecutionReader(db *DB) *CaseExecutionReader {
	return &CaseExecutionReader{db: db}
}

func (c *CaseExecutionReader) GetCaseExecution(ctx context.Context, testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) (*test.CaseExecution, error) {
	caseExec, err := c.db.GetCaseExecution(ctx, sqlc.GetCaseExecutionParams{
		ID:              caseExecID,
		TestExecutionID: testExecID,
	})
	if err != nil {
		return nil, err
	}
	return marshalCaseExec(caseExec), nil
}

func (c *CaseExecutionReader) ListCaseExecutions(ctx context.Context, testExecID test.TestExecutionID, filter test.PageFilter[test.CaseExecutionID]) (test.CaseExecutionList, error) {
	params := sqlc.ListCaseExecutionsParams{
		TestExecutionID: testExecID,
		PageSize:        int32(filter.Size),
	}
	if filter.OffsetID != nil {
		params.OffsetID = ptr.Get(int32(*filter.OffsetID))
	}

	execs, err := c.db.ListCaseExecutions(ctx, params)
	if err != nil {
		return nil, err
	}
	return marshalCaseExecs(execs), nil
}

type CaseExecutionWriter struct {
	db *DB
}

func NewCaseExecutionWriter(db *DB) *CaseExecutionWriter {
	return &CaseExecutionWriter{db: db}
}

func (c *CaseExecutionWriter) CreateCaseExecutionScheduled(ctx context.Context, scheduled *test.ScheduledCaseExecution) (*test.CaseExecution, error) {
	exec, err := c.db.CreateCaseExecutionScheduled(ctx, sqlc.CreateCaseExecutionScheduledParams{
		ID:              scheduled.ID,
		TestExecutionID: scheduled.TestExecutionID,
		CaseName:        scheduled.CaseName,
		ScheduleTime:    scheduled.ScheduleTime.UTC(),
	})
	if err != nil {
		return nil, err
	}
	return marshalCaseExec(exec), nil
}

func (c *CaseExecutionWriter) UpdateCaseExecutionStarted(ctx context.Context, started *test.StartedCaseExecution) (*test.CaseExecution, error) {
	exec, err := c.db.UpdateCaseExecutionStarted(ctx, sqlc.UpdateCaseExecutionStartedParams{
		ID:              started.ID,
		TestExecutionID: started.TestExecutionID,
		StartTime:       ptr.Get(started.StartTime.UTC()),
	})
	if err != nil {
		return nil, err
	}
	return marshalCaseExec(exec), nil
}

func (c *CaseExecutionWriter) UpdateCaseExecutionFinished(ctx context.Context, finished *test.FinishedCaseExecution) (*test.CaseExecution, error) {
	exec, err := c.db.UpdateCaseExecutionFinished(ctx, sqlc.UpdateCaseExecutionFinishedParams{
		ID:              finished.ID,
		TestExecutionID: finished.TestExecutionID,
		FinishTime:      ptr.Get(finished.FinishTime.UTC()),
		Error:           finished.Error,
	})
	if err != nil {
		return nil, err
	}
	return marshalCaseExec(exec), nil
}

func (c *CaseExecutionWriter) DeleteCaseExecution(ctx context.Context, testExecID test.TestExecutionID, id test.CaseExecutionID) error {
	return c.db.DeleteCaseExecution(ctx, sqlc.DeleteCaseExecutionParams{
		ID:              id,
		TestExecutionID: testExecID,
	})
}
