package inmem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
)

func TestCaseExecutionReader_GetCaseExecution(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewCaseExecutionReader(db)

	want := fake.GenCaseExec(test.NewTestExecutionID())
	db.caseExecs[getCaseExecKey(want.TestExecutionID, want.ID)] = want

	got, err := r.GetCaseExecution(ctx, want.TestExecutionID, want.ID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCaseExecutionReader_ListCaseExecutions(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewCaseExecutionReader(db)

	count := 30
	want := make(test.CaseExecutionList, count)
	testExecID := test.NewTestExecutionID()

	for i := range count {
		ce := fake.GenCaseExec(testExecID)
		want[i] = ce
		db.caseExecs[getCaseExecKey(ce.TestExecutionID, ce.ID)] = ce
	}

	got, err := r.ListCaseExecutions(ctx, testExecID)
	require.NoError(t, err)
	assert.Len(t, got, count)
	require.Equal(t, want, got)
}

func TestCaseExecutionWriter_CreateScheduledCaseExecution(t *testing.T) {
	tests := []struct {
		name             string
		existingTestExec bool
		wantErr          error
	}{
		{
			name:             "create success",
			existingTestExec: true,
		},
		{
			name:             "test execution not found error",
			existingTestExec: false,
			wantErr:          test.ErrorTestExecutionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewCaseExecutionWriter(db)

			testExec := fake.GenTestExec(uuid.New())
			if tt.existingTestExec {
				db.testExecs[testExec.ID] = testExec
			}

			sched := fake.GenScheduledCaseExec(testExec.ID)

			got, err := w.CreateScheduledCaseExecution(ctx, sched)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, sched.ID, got.ID)
			assert.Equal(t, sched.TestExecID, got.TestExecutionID)
			assert.Equal(t, sched.CaseName, got.CaseName)
			assert.Equal(t, sched.ScheduleTime, got.ScheduleTime)
			assert.Nil(t, got.StartTime)
			assert.Nil(t, got.FinishTime)
			assert.Nil(t, got.Error)
		})
	}
}

func TestCaseExecutionWriter_UpdateStartedCaseExecution(t *testing.T) {
	tests := []struct {
		name         string
		existingCase bool
		wantErr      error
	}{
		{
			name:         "update success",
			existingCase: true,
		},
		{
			name:         "case execution not found error",
			existingCase: false,
			wantErr:      test.ErrorCaseExecutionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewCaseExecutionWriter(db)

			existing := fake.GenCaseExec(test.NewTestExecutionID())
			existing.StartTime = nil
			existing.FinishTime = nil
			existing.Error = nil
			dbKey := getCaseExecKey(existing.TestExecutionID, existing.ID)

			if tt.existingCase {
				db.caseExecs[dbKey] = existing
			}

			started := fake.GenStartedCaseExec(existing.TestExecutionID, existing.ID)
			got, err := w.UpdateStartedCaseExecution(ctx, started)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, started.ID, got.ID)
			assert.Equal(t, existing.TestExecutionID, got.TestExecutionID)
			assert.Equal(t, existing.CaseName, got.CaseName)
			assert.Equal(t, existing.ScheduleTime, got.ScheduleTime)
			assert.Equal(t, started.StartTime, *got.StartTime)
			assert.Nil(t, got.FinishTime)
			assert.Nil(t, got.Error)
		})
	}
}

func TestCaseExecutionWriter_UpdateFinishedCaseExecution(t *testing.T) {
	tests := []struct {
		name         string
		existingCase bool
		wantErr      error
	}{
		{
			name:         "update success",
			existingCase: true,
		},
		{
			name:         "case execution not found error",
			existingCase: false,
			wantErr:      test.ErrorCaseExecutionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewCaseExecutionWriter(db)

			existing := fake.GenCaseExec(test.NewTestExecutionID())
			existing.FinishTime = nil
			existing.Error = nil
			dbKey := getCaseExecKey(existing.TestExecutionID, existing.ID)

			if tt.existingCase {
				db.caseExecs[dbKey] = existing
			}

			finished := fake.GenFinishedCaseExec(existing.TestExecutionID, existing.ID, ptr.Get("bang"))
			got, err := w.UpdateFinishedCaseExecution(ctx, finished)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, finished.ID, got.ID)
			assert.Equal(t, existing.TestExecutionID, got.TestExecutionID)
			assert.Equal(t, existing.CaseName, got.CaseName)
			assert.Equal(t, existing.ScheduleTime, got.ScheduleTime)
			assert.Equal(t, existing.StartTime, got.StartTime)
			assert.Equal(t, finished.FinishTime, *got.FinishTime)
			assert.Equal(t, finished.Error, got.Error)
		})
	}
}

func TestCaseExecutionWriter_DeleteCaseExecution(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	w := NewCaseExecutionWriter(db)

	want := fake.GenCaseExec(test.NewTestExecutionID())
	db.caseExecs[getCaseExecKey(want.TestExecutionID, want.ID)] = want
	assert.NotEmpty(t, db.caseExecs)

	err := w.DeleteCaseExecution(ctx, want.TestExecutionID, want.ID)
	require.NoError(t, err)
	assert.Empty(t, db.caseExecs)
}
