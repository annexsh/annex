package inmem

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"

	"github.com/annexsh/annex/event"
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

func (t *TestExecutionReader) GetTestExecution(_ context.Context, id test.TestExecutionID) (*test.TestExecution, error) {
	t.db.mu.RLock()
	defer t.db.mu.RUnlock()

	te, ok := t.db.testExecs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return ptr.Copy(te), nil
}

func (t *TestExecutionReader) GetTestExecutionPayload(_ context.Context, id test.TestExecutionID) (*test.Payload, error) {
	t.db.mu.RLock()
	defer t.db.mu.RUnlock()

	p, ok := t.db.testExecPayloads[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &test.Payload{
		Payload: p,
	}, nil
}

func (t *TestExecutionReader) ListTestExecutions(_ context.Context, testID uuid.UUID, filter *test.TestExecutionListFilter) (test.TestExecutionList, error) {
	t.db.mu.RLock()
	defer t.db.mu.RUnlock()

	i := 0
	execs := make(test.TestExecutionList, len(t.db.testExecs))
	for _, te := range t.db.testExecs {
		execs[i] = te
		i++
	}

	slices.SortFunc(execs, func(a, b *test.TestExecution) int {
		if a.ScheduleTime.Before(b.ScheduleTime) || a.ScheduleTime.Equal(b.ScheduleTime) && a.ID.String() < b.ID.String() {
			return -1
		}
		return 1
	})

	var filtered test.TestExecutionList

	for _, te := range execs {
		if te.TestID == testID {
			idstr := te.ID.String()
			if filter.LastScheduleTime != nil {
				// Skip already seen before last schedule time
				if te.ScheduleTime.Before(*filter.LastScheduleTime) {
					continue
				}
				// Skip already seen before last exec ID if schedule times are the same
				if te.ScheduleTime.Equal(*filter.LastScheduleTime) &&
					filter.LastExecID != nil &&
					idstr <= filter.LastExecID.String() {
					continue
				}
			}

			filtered = append(filtered, ptr.Copy(te))
			if uint32(len(filtered)) == filter.PageSize {
				break
			}
		}
	}

	return filtered, nil
}

type TestExecutionWriter struct {
	db *DB
}

func NewTestExecutionWriter(db *DB) *TestExecutionWriter {
	return &TestExecutionWriter{db: db}
}

func (t *TestExecutionWriter) CreateScheduledTestExecution(_ context.Context, scheduled *test.ScheduledTestExecution) (*test.TestExecution, error) {
	t.db.mu.Lock()
	defer t.db.mu.Unlock()

	if _, ok := t.db.tests[scheduled.TestID]; !ok {
		return nil, test.ErrorTestNotFound
	}

	te := &test.TestExecution{
		ID:           scheduled.ID,
		TestID:       scheduled.TestID,
		HasPayload:   scheduled.Payload != nil,
		ScheduleTime: scheduled.ScheduleTime,
	}
	t.db.testExecs[te.ID] = te
	t.db.testExecPayloads[te.ID] = scheduled.Payload
	t.db.events.Publish(event.NewTestExecutionEvent(event.TypeTestExecutionScheduled, te))
	return ptr.Copy(te), nil
}

func (t *TestExecutionWriter) UpdateStartedTestExecution(_ context.Context, started *test.StartedTestExecution) (*test.TestExecution, error) {
	t.db.mu.Lock()
	defer t.db.mu.Unlock()

	te, ok := t.db.testExecs[started.ID]
	if !ok {
		return nil, test.ErrorTestExecutionNotFound
	}
	te.StartTime = &started.StartTime
	t.db.testExecs[te.ID] = te
	t.db.events.Publish(event.NewTestExecutionEvent(event.TypeTestExecutionStarted, te))
	return ptr.Copy(te), nil
}

func (t *TestExecutionWriter) UpdateFinishedTestExecution(_ context.Context, finished *test.FinishedTestExecution) (*test.TestExecution, error) {
	t.db.mu.Lock()
	defer t.db.mu.Unlock()

	te, ok := t.db.testExecs[finished.ID]
	if !ok {
		return nil, test.ErrorTestExecutionNotFound
	}
	te.FinishTime = &finished.FinishTime
	te.Error = finished.Error
	t.db.testExecs[te.ID] = te
	t.db.events.Publish(event.NewTestExecutionEvent(event.TypeTestExecutionFinished, te))
	return ptr.Copy(te), nil
}

func (t *TestExecutionWriter) ResetTestExecution(_ context.Context, reset *test.ResetTestExecution) (*test.TestExecution, test.ResetRollback, error) {
	t.db.mu.Lock()
	defer t.db.mu.Unlock()

	te, ok := t.db.testExecs[reset.ID]
	if !ok {
		return nil, nil, test.ErrorTestExecutionNotFound
	}

	for _, caseExecID := range reset.StaleCaseExecutions {
		delete(t.db.caseExecs, caseExecKey(te.ID, caseExecID))
	}

	for _, logID := range reset.StaleLogs {
		delete(t.db.execLogs, logID)
	}

	te.ScheduleTime = reset.ResetTime
	te.StartTime = nil
	te.FinishTime = nil
	te.Error = nil

	t.db.testExecs[te.ID] = te
	t.db.events.Publish(event.NewTestExecutionEvent(event.TypeTestExecutionScheduled, te))

	rollback := func(ctx context.Context) error {
		// no-op since in-mem is not intended for prod use anyway
		return nil
	}
	return ptr.Copy(te), rollback, nil
}
