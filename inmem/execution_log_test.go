package inmem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexhq/annex/internal/fake"
	"github.com/annexhq/annex/test"
)

func TestExecutionLogReader_GetExecutionLog(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewExecutionLogReader(db)

	want := fake.GenTestExecLog(test.NewTestExecutionID())
	db.execLogs[want.ID] = want

	got, err := r.GetExecutionLog(ctx, want.ID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestExecutionLogReader_ListExecutionLogs(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewExecutionLogReader(db)

	count := 30
	want := make(test.ExecutionLogList, count)
	testExecID := test.NewTestExecutionID()

	for i := range count {
		l := fake.GenTestExecLog(testExecID)
		want[i] = l
		db.execLogs[l.ID] = l
	}

	got, err := r.ListExecutionLogs(ctx, testExecID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestExecutionLogWriter_CreateExecutionLog(t *testing.T) {
	tests := []struct {
		name             string
		existingTestExec bool
		wantErr          string
	}{
		{
			name:             "create success",
			existingTestExec: true,
		},
		{
			name:             "test execution not found error",
			existingTestExec: false,
			wantErr:          "test execution not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewDB()
			w := NewExecutionLogWriter(db)

			testExec := fake.GenTestExec(uuid.New())
			if tt.existingTestExec {
				db.testExecs[testExec.ID] = testExec
			}

			want := fake.GenTestExecLog(testExec.ID)

			err := w.CreateExecutionLog(ctx, want)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			got, ok := db.execLogs[want.ID]
			require.True(t, ok)
			assert.Equal(t, want, got)
		})
	}
}

func TestExecutionLogWriter_DeleteExecutionLog(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	w := NewExecutionLogWriter(db)

	want := fake.GenTestExecLog(test.NewTestExecutionID())
	db.execLogs[want.ID] = want
	assert.NotEmpty(t, db.execLogs)

	err := w.DeleteExecutionLog(ctx, want.ID)
	require.NoError(t, err)
	assert.Empty(t, db.execLogs)
}
