package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/test"
)

var (
	_ test.ExecutionLogReader = (*ExecutionLogReader)(nil)
	_ test.ExecutionLogWriter = (*ExecutionLogWriter)(nil)
)

type ExecutionLogReader struct {
	db *DB
}

func NewExecutionLogReader(db *DB) *ExecutionLogReader {
	return &ExecutionLogReader{db: db}
}

func (e *ExecutionLogReader) GetExecutionLog(ctx context.Context, id uuid.UUID) (*test.ExecutionLog, error) {
	execLog, err := e.db.GetExecutionLog(ctx, id)
	if err != nil {
		return nil, err
	}
	return marshalExecLog(execLog), nil
}

func (e *ExecutionLogReader) ListExecutionLogs(ctx context.Context, testExecID test.TestExecutionID) (test.ExecutionLogList, error) {
	// TODO: pagination
	logs, err := e.db.ListTestExecutionLogs(ctx, testExecID)
	if err != nil {
		return nil, err
	}
	return marshalExecLogs(logs), nil
}

type ExecutionLogWriter struct {
	db *DB
}

func NewExecutionLogWriter(db *DB) *ExecutionLogWriter {
	return &ExecutionLogWriter{db: db}
}

func (e *ExecutionLogWriter) CreateExecutionLog(ctx context.Context, log *test.ExecutionLog) error {
	return e.db.CreateLog(ctx, sqlc.CreateLogParams{
		ID:         log.ID,
		TestExecID: log.TestExecID,
		CaseExecID: log.CaseExecID,
		Level:      log.Level,
		Message:    log.Message,
		CreatedAt:  sqlc.NewTimestamp(log.CreateTime),
	})
}

func (e *ExecutionLogWriter) DeleteExecutionLog(ctx context.Context, id uuid.UUID) error {
	return e.db.DeleteLog(ctx, id)
}
