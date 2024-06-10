package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/test"
)

var (
	_ test.LogReader = (*LogReader)(nil)
	_ test.LogWriter = (*LogWriter)(nil)
)

type LogReader struct {
	db *DB
}

func NewLogReader(db *DB) *LogReader {
	return &LogReader{db: db}
}

func (e *LogReader) GetLog(ctx context.Context, id uuid.UUID) (*test.Log, error) {
	execLog, err := e.db.GetLog(ctx, id)
	if err != nil {
		return nil, err
	}
	return marshalExecLog(execLog), nil
}

func (e *LogReader) ListLogs(ctx context.Context, testExecID test.TestExecutionID) (test.LogList, error) {
	// TODO: pagination
	logs, err := e.db.ListLogs(ctx, testExecID)
	if err != nil {
		return nil, err
	}
	return marshalExecLogs(logs), nil
}

type LogWriter struct {
	db *DB
}

func NewLogWriter(db *DB) *LogWriter {
	return &LogWriter{db: db}
}

func (e *LogWriter) CreateLog(ctx context.Context, log *test.Log) error {
	return e.db.CreateLog(ctx, sqlc.CreateLogParams{
		ID:              log.ID,
		TestExecutionID: log.TestExecutionID,
		CaseExecutionID: log.CaseExecutionID,
		Level:           log.Level,
		Message:         log.Message,
		CreateTime:      sqlc.NewTimestamp(log.CreateTime),
	})
}

func (e *LogWriter) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return e.db.DeleteLog(ctx, id)
}
