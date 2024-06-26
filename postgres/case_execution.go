package postgres

import (
	"context"

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

func (c *CaseExecutionReader) ListCaseExecutions(ctx context.Context, testExecID test.TestExecutionID) (test.CaseExecutionList, error) {
	// TODO: pagination
	execs, err := c.db.ListCaseExecutions(ctx, testExecID)
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

func (c *CaseExecutionWriter) CreateScheduledCaseExecution(ctx context.Context, scheduled *test.ScheduledCaseExecution) (*test.CaseExecution, error) {
	exec, err := c.db.CreateCaseExecution(ctx, sqlc.CreateCaseExecutionParams{
		ID:              scheduled.ID,
		TestExecutionID: scheduled.TestExecID,
		CaseName:        scheduled.CaseName,
		ScheduleTime:    sqlc.NewTimestamp(scheduled.ScheduleTime),
	})
	if err != nil {
		return nil, err
	}
	return marshalCaseExec(exec), nil
}

func (c *CaseExecutionWriter) UpdateStartedCaseExecution(ctx context.Context, started *test.StartedCaseExecution) (*test.CaseExecution, error) {
	exec, err := c.db.UpdateCaseExecutionStarted(ctx, sqlc.UpdateCaseExecutionStartedParams{
		ID:              started.ID,
		TestExecutionID: started.TestExecutionID,
		StartTime:       sqlc.NewTimestamp(started.StartTime),
	})
	if err != nil {
		return nil, err
	}
	return marshalCaseExec(exec), nil
}

func (c *CaseExecutionWriter) UpdateFinishedCaseExecution(ctx context.Context, finished *test.FinishedCaseExecution) (*test.CaseExecution, error) {
	exec, err := c.db.UpdateCaseExecutionFinished(ctx, sqlc.UpdateCaseExecutionFinishedParams{
		ID:              finished.ID,
		TestExecutionID: finished.TestExecutionID,
		FinishTime:      sqlc.NewTimestamp(finished.FinishTime),
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
