package inmem

import (
	"context"
	"slices"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/ptr"
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

func (c *CaseExecutionReader) GetCaseExecution(_ context.Context, testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) (*test.CaseExecution, error) {
	c.db.mu.RLock()
	defer c.db.mu.RUnlock()
	ce, ok := c.db.caseExecs[caseExecKey(testExecID, caseExecID)]
	if !ok {
		return nil, test.ErrorCaseExecutionNotFound
	}
	return ptr.Copy(ce), nil
}

func (c *CaseExecutionReader) ListCaseExecutions(_ context.Context, testExecID test.TestExecutionID) (test.CaseExecutionList, error) {
	c.db.mu.RLock()
	defer c.db.mu.RUnlock()

	var execs test.CaseExecutionList
	for _, ce := range c.db.caseExecs {
		if ce.TestExecutionID == testExecID {
			execs = append(execs, ptr.Copy(ce))
		}
	}
	slices.SortFunc(execs, func(a, b *test.CaseExecution) int {
		if a.ScheduleTime.Before(b.ScheduleTime) || a.ScheduleTime.Equal(b.ScheduleTime) && a.ID.String() < b.ID.String() {
			return -1
		}
		return 1
	})

	return execs, nil
}

type CaseExecutionWriter struct {
	db *DB
}

func NewCaseExecutionWriter(db *DB) *CaseExecutionWriter {
	return &CaseExecutionWriter{db: db}
}

func (c *CaseExecutionWriter) CreateScheduledCaseExecution(_ context.Context, scheduled *test.ScheduledCaseExecution) (*test.CaseExecution, error) {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	if _, ok := c.db.testExecs[scheduled.TestExecID]; !ok {
		return nil, test.ErrorTestExecutionNotFound
	}

	ce := &test.CaseExecution{
		ID:              scheduled.ID,
		TestExecutionID: scheduled.TestExecID,
		CaseName:        scheduled.CaseName,
		ScheduleTime:    scheduled.ScheduleTime,
	}
	c.db.caseExecs[caseExecKey(ce.TestExecutionID, ce.ID)] = ce
	c.db.events.Publish(event.NewCaseExecutionEvent(event.TypeCaseExecutionScheduled, ce))
	return ptr.Copy(ce), nil
}

func (c *CaseExecutionWriter) UpdateStartedCaseExecution(_ context.Context, started *test.StartedCaseExecution) (*test.CaseExecution, error) {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	key := caseExecKey(started.TestExecutionID, started.ID)
	ce, ok := c.db.caseExecs[key]
	if !ok {
		return nil, test.ErrorCaseExecutionNotFound
	}
	ce.StartTime = &started.StartTime
	c.db.caseExecs[key] = ce
	c.db.events.Publish(event.NewCaseExecutionEvent(event.TypeCaseExecutionStarted, ce))
	return ptr.Copy(ce), nil
}

func (c *CaseExecutionWriter) UpdateFinishedCaseExecution(_ context.Context, finished *test.FinishedCaseExecution) (*test.CaseExecution, error) {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	key := caseExecKey(finished.TestExecutionID, finished.ID)
	ce, ok := c.db.caseExecs[key]
	if !ok {
		return nil, test.ErrorCaseExecutionNotFound
	}
	ce.FinishTime = &finished.FinishTime
	ce.Error = finished.Error
	c.db.caseExecs[key] = ce
	c.db.events.Publish(event.NewCaseExecutionEvent(event.TypeCaseExecutionFinished, ce))
	return ptr.Copy(ce), nil
}

func (c *CaseExecutionWriter) DeleteCaseExecution(_ context.Context, testExecID test.TestExecutionID, id test.CaseExecutionID) error {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()
	delete(c.db.caseExecs, caseExecKey(testExecID, id))
	return nil
}

func caseExecKey(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) string {
	return testExecID.String() + "." + caseExecID.String()
}
