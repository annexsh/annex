package postgres

import (
	"context"

	"github.com/annexsh/annex/postgres/sqlc"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
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

func (e *LogReader) GetLog(ctx context.Context, id uuid.V7) (*test.Log, error) {
	execLog, err := e.db.GetLog(ctx, id)
	if err != nil {
		return nil, err
	}
	return marshalLog(execLog), nil
}

func (e *LogReader) ListLogs(ctx context.Context, testExecID test.TestExecutionID, filter test.PageFilter[uuid.V7]) (test.LogList, error) {
	params := sqlc.ListLogsParams{
		TestExecutionID: testExecID,
		PageSize:        int32(filter.Size),
	}
	if filter.OffsetID != nil {
		params.OffsetID = filter.OffsetID
	}

	logs, err := e.db.ListLogs(ctx, params)
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
		CreateTime:      log.CreateTime.UTC(),
	})
}

func (e *LogWriter) DeleteLog(ctx context.Context, id uuid.V7) error {
	return e.db.DeleteLog(ctx, id)
}
