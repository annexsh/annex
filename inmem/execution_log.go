package inmem

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/ptr"
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

func (e *ExecutionLogReader) GetExecutionLog(_ context.Context, id uuid.UUID) (*test.ExecutionLog, error) {
	execLog, ok := e.db.execLogs[id]
	if !ok {
		return nil, test.ErrorExecutionLogNotFound
	}
	return ptr.Copy(execLog), nil
}

func (e *ExecutionLogReader) ListExecutionLogs(_ context.Context, testExecID test.TestExecutionID) (test.ExecutionLogList, error) {
	e.db.mu.RLock()
	defer e.db.mu.RUnlock()

	var logs test.ExecutionLogList
	for _, l := range e.db.execLogs {
		if l.TestExecID == testExecID {
			logs = append(logs, ptr.Copy(l))
		}
	}
	slices.SortFunc(logs, func(a, b *test.ExecutionLog) int {
		if a.CreateTime.Before(b.CreateTime) {
			return -1
		}
		return 1
	})
	return logs, nil
}

type ExecutionLogWriter struct {
	db *DB
}

func NewExecutionLogWriter(db *DB) *ExecutionLogWriter {
	return &ExecutionLogWriter{db: db}
}

func (e *ExecutionLogWriter) CreateExecutionLog(_ context.Context, log *test.ExecutionLog) error {
	e.db.mu.Lock()
	defer e.db.mu.Unlock()

	if _, ok := e.db.testExecs[log.TestExecID]; !ok {
		return test.ErrorTestExecutionNotFound
	}

	if log.CaseExecID != nil {
		if _, ok := e.db.caseExecs[caseExecKey(log.TestExecID, *log.CaseExecID)]; !ok {
			return test.ErrorCaseExecutionNotFound
		}
	}

	e.db.execLogs[log.ID] = log
	e.db.events.Publish(event.NewExecutionLogEvent(event.TypeExecutionLogPublished, log))
	return nil
}

func (e *ExecutionLogWriter) DeleteExecutionLog(_ context.Context, id uuid.UUID) error {
	e.db.mu.Lock()
	defer e.db.mu.Unlock()
	delete(e.db.execLogs, id)
	return nil
}
