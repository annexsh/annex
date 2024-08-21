package inmem

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func TestLogReader_GetLog(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewLogReader(db)

	want := fake.GenTestExecLog(test.NewTestExecutionID())
	db.execLogs[want.ID] = want

	got, err := r.GetLog(ctx, want.ID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestLogReader_ListLogs(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewLogReader(db)

	count := 30
	want := make(test.LogList, count)
	testExecID := test.NewTestExecutionID()

	for i := range count {
		l := fake.GenTestExecLog(testExecID)
		want[i] = l
		db.execLogs[l.ID] = l
	}

	got, err := r.ListLogs(ctx, testExecID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestLogWriter_CreateLog(t *testing.T) {
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
			w := NewLogWriter(db)

			testExec := fake.GenTestExec(uuid.New())
			if tt.existingTestExec {
				db.testExecs[testExec.ID] = testExec
			}

			want := fake.GenTestExecLog(testExec.ID)

			err := w.CreateLog(ctx, want)
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

func TestLogWriter_DeleteLog(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	w := NewLogWriter(db)

	want := fake.GenTestExecLog(test.NewTestExecutionID())
	db.execLogs[want.ID] = want
	assert.NotEmpty(t, db.execLogs)

	err := w.DeleteLog(ctx, want.ID)
	require.NoError(t, err)
	assert.Empty(t, db.execLogs)
}
