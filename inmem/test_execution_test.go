package inmem

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func TestTestExecutionReader_GetTestExecution(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewTestExecutionReader(db)

	want := fake.GenTestExec(uuid.New())
	db.testExecs[want.ID] = want

	got, err := r.GetTestExecution(ctx, want.ID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestTestExecutionReader_GetTestExecutionPayload(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewTestExecutionReader(db)

	testExecID := test.NewTestExecutionID()
	want := []byte{1, 2, 3}
	db.testExecPayloads[testExecID] = want

	got, err := r.GetTestExecutionInput(ctx, testExecID)
	require.NoError(t, err)
	assert.Equal(t, want, got.Data)
}

func TestTestExecutionReader_ListTestExecutions(t *testing.T) {
	// TODO
}

func TestTestExecutionWriter_CreateScheduledTestExecution(t *testing.T) {
	tests := []struct {
		name         string
		existingTest bool
		wantErr      error
	}{
		{
			name:         "create success",
			existingTest: true,
		},
		{
			name:         "test not found error",
			existingTest: false,
			wantErr:      test.ErrorTestNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewTestExecutionWriter(db)

			testExec := fake.GenTest()
			if tt.existingTest {
				db.tests[testExec.ID] = testExec
			}

			sched := fake.GenScheduledTestExec(testExec.ID)
			got, err := w.CreateScheduledTestExecution(ctx, sched)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, sched.ID, got.ID)
			assert.Equal(t, sched.TestID, got.TestID)
			assert.Equal(t, sched.Payload != nil, got.HasInput)
			assert.Equal(t, sched.ScheduleTime, got.ScheduleTime)
			assert.Nil(t, got.StartTime)
			assert.Nil(t, got.FinishTime)
			assert.Nil(t, got.Error)
		})
	}
}

func TestTestExecutionWriter_UpdateStartedTestExecution(t *testing.T) {
	tests := []struct {
		name         string
		existingTest bool
		wantErr      error
	}{
		{
			name:         "update success",
			existingTest: true,
		},
		{
			name:         "test execution not found error",
			existingTest: false,
			wantErr:      test.ErrorTestExecutionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewTestExecutionWriter(db)

			existing := fake.GenTestExec(uuid.New())
			existing.StartTime = nil
			existing.FinishTime = nil
			existing.Error = nil

			if tt.existingTest {
				db.testExecs[existing.ID] = existing
			}

			started := fake.GenStartedTestExec(existing.ID)
			got, err := w.UpdateStartedTestExecution(ctx, started)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, started.ID, got.ID)
			assert.Equal(t, existing.TestID, got.TestID)
			assert.Equal(t, existing.HasInput, got.HasInput)
			assert.Equal(t, existing.ScheduleTime, got.ScheduleTime)
			assert.Equal(t, started.StartTime, *got.StartTime)
			assert.Nil(t, got.FinishTime)
			assert.Nil(t, got.Error)
		})
	}
}

func TestTestExecutionWriter_UpdateFinishedTestExecution(t *testing.T) {
	tests := []struct {
		name         string
		existingTest bool
		wantErr      error
	}{
		{
			name:         "update success",
			existingTest: true,
		},
		{
			name:         "test execution not found error",
			existingTest: false,
			wantErr:      test.ErrorTestExecutionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewTestExecutionWriter(db)

			existing := fake.GenTestExec(uuid.New())
			existing.FinishTime = nil
			existing.Error = nil

			if tt.existingTest {
				db.testExecs[existing.ID] = existing
			}

			finished := fake.GenFinishedTestExec(existing.ID, ptr.Get("bang"))
			got, err := w.UpdateFinishedTestExecution(ctx, finished)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, finished.ID, got.ID)
			assert.Equal(t, existing.TestID, got.TestID)
			assert.Equal(t, existing.HasInput, got.HasInput)
			assert.Equal(t, existing.ScheduleTime, got.ScheduleTime)
			assert.Equal(t, finished.FinishTime, *got.FinishTime)
			assert.Equal(t, finished.Error, got.Error)
		})
	}
}

func TestTestExecutionWriter_ResetTestExecution(t *testing.T) {
	tests := []struct {
		name             string
		existingTest     bool
		existingTestExec bool
		wantErr          error
	}{
		{
			name:             "reset success",
			existingTest:     true,
			existingTestExec: true,
		},
		{
			name:         "test exec not found error",
			existingTest: true,
			wantErr:      test.ErrorTestExecutionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewTestExecutionWriter(db)

			testID := uuid.New()
			if tt.existingTest {
				db.tests[testID] = fake.GenTest()
			}

			numValidCaseExecs := 3
			numValidLogsPerExec := 10
			numStaleCaseExecs := 3
			numStaleLogsPerExec := 10

			existing := fake.GenTestExec(testID)

			var staleCaseExecIDs []test.CaseExecutionID
			var staleLogIDs []uuid.V7

			if tt.existingTestExec {
				db.testExecs[existing.ID] = existing

				// Gen valid test execution logs
				for range numValidLogsPerExec {
					validLog := fake.GenTestExecLog(existing.ID)
					db.execLogs[validLog.ID] = validLog
				}

				// Gen stale test execution logs
				for range numStaleLogsPerExec {
					staleLog := fake.GenTestExecLog(existing.ID)
					db.execLogs[staleLog.ID] = staleLog
					staleLogIDs = append(staleLogIDs, staleLog.ID)
				}

				// Create valid case executions
				for range numValidCaseExecs {
					caseExec := fake.GenCaseExec(existing.ID)
					db.caseExecs[getCaseExecKey(existing.ID, caseExec.ID)] = caseExec
					// Gen stale case execution logs
					for range numValidLogsPerExec {
						validLog := fake.GenCaseExecLog(existing.ID, caseExec.ID)
						db.execLogs[validLog.ID] = validLog
					}
				}

				// Create stale case executions
				for range numStaleCaseExecs {
					caseExec := fake.GenCaseExec(existing.ID)
					staleCaseExecIDs = append(staleCaseExecIDs, caseExec.ID)
					db.caseExecs[getCaseExecKey(existing.ID, caseExec.ID)] = caseExec
					// Gen stale case execution logs
					for range numStaleLogsPerExec {
						staleLog := fake.GenCaseExecLog(existing.ID, caseExec.ID)
						db.execLogs[staleLog.ID] = staleLog
						staleLogIDs = append(staleLogIDs, staleLog.ID)
					}
				}
			}

			reset := &test.ResetTestExecution{
				ID:                  existing.ID,
				ResetTime:           time.Now().UTC(),
				StaleCaseExecutions: staleCaseExecIDs,
				StaleLogs:           staleLogIDs,
			}

			got, rollback, err := w.ResetTestExecution(ctx, reset)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, rollback)
			assert.Equal(t, reset.ID, got.ID)
			assert.Equal(t, existing.TestID, got.TestID)
			assert.Equal(t, existing.HasInput, got.HasInput)
			assert.Equal(t, reset.ResetTime, got.ScheduleTime)
			assert.Nil(t, got.StartTime)
			assert.Nil(t, got.FinishTime)
			assert.Nil(t, got.Error)

			assert.Len(t, db.caseExecs, numValidCaseExecs)
			for _, staleID := range staleCaseExecIDs {
				assert.NotContains(t, db.caseExecs, staleID)
			}

			numExecs := numValidCaseExecs + 1 // +1 is the test execution itself
			assert.Len(t, db.execLogs, numValidLogsPerExec*numExecs)
			for _, staleID := range staleLogIDs {
				assert.NotContains(t, db.execLogs, staleID)
			}
		})
	}
}
