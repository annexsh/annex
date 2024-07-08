package inmem

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"github.com/annexsh/annex/internal/ptr"
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

func (e *LogReader) GetLog(_ context.Context, id uuid.UUID) (*test.Log, error) {
	execLog, ok := e.db.execLogs[id]
	if !ok {
		return nil, test.ErrorLogNotFound
	}
	return ptr.Copy(execLog), nil
}

func (e *LogReader) ListLogs(_ context.Context, testExecID test.TestExecutionID) (test.LogList, error) {
	e.db.mu.RLock()
	defer e.db.mu.RUnlock()

	var logs test.LogList
	for _, l := range e.db.execLogs {
		if l.TestExecutionID == testExecID {
			logs = append(logs, ptr.Copy(l))
		}
	}
	slices.SortFunc(logs, func(a, b *test.Log) int {
		if a.CreateTime.Before(b.CreateTime) {
			return -1
		}
		return 1
	})
	return logs, nil
}

type LogWriter struct {
	db *DB
}

func NewLogWriter(db *DB) *LogWriter {
	return &LogWriter{db: db}
}

func (e *LogWriter) CreateLog(_ context.Context, log *test.Log) error {
	e.db.mu.Lock()
	defer e.db.mu.Unlock()

	if _, ok := e.db.testExecs[log.TestExecutionID]; !ok {
		return test.ErrorTestExecutionNotFound
	}

	if log.CaseExecutionID != nil {
		if _, ok := e.db.caseExecs[getCaseExecKey(log.TestExecutionID, *log.CaseExecutionID)]; !ok {
			return test.ErrorCaseExecutionNotFound
		}
	}

	e.db.execLogs[log.ID] = log
	return nil
}

func (e *LogWriter) DeleteLog(_ context.Context, id uuid.UUID) error {
	e.db.mu.Lock()
	defer e.db.mu.Unlock()
	delete(e.db.execLogs, id)
	return nil
}
