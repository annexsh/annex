//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/postgres/sqlc"
	"github.com/annexsh/annex/test"
)

func TestCreateGetTestExecution(t *testing.T) {
	tests := []struct {
		name  string
		input *test.Payload
	}{
		{
			name:  "no input",
			input: nil,
		},
		{
			name:  "has input",
			input: fake.GenInput(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db, closer := newTestDB(t)
			defer closer()

			w := NewTestExecutionWriter(db)
			r := NewTestExecutionReader(db)

			dummyTest := createDummyTest(ctx, t, db, tt.input != nil)
			scheduled := fake.GenScheduledTestExec(dummyTest.ID)
			scheduled.HasInput = tt.input != nil

			assertEqual := func(got *test.TestExecution) {
				assert.Equal(t, scheduled.ID, got.ID)
				assert.Equal(t, scheduled.TestID, got.TestID)
				assert.Equal(t, scheduled.ScheduleTime, got.ScheduleTime)
				assert.Equal(t, scheduled.HasInput, got.HasInput)
				assert.Nil(t, got.StartTime)
				assert.Nil(t, got.FinishTime)
				assert.Nil(t, got.Error)
			}

			got, err := w.CreateTestExecutionScheduled(ctx, scheduled)
			require.NoError(t, err)
			assertEqual(got)

			got, err = r.GetTestExecution(ctx, got.ID)
			require.NoError(t, err)
			assertEqual(got)

			if tt.input != nil {
				err = w.CreateTestExecutionInput(ctx, got.ID, tt.input)
				require.NoError(t, err)
			}

			gotInput, err := r.GetTestExecutionInput(ctx, got.ID)
			if tt.input != nil {
				require.NoError(t, err)
				assert.Equal(t, tt.input, gotInput)
			} else {
				assert.ErrorIs(t, err, test.ErrorTestExecutionPayloadNotFound)
				assert.Nil(t, gotInput)
			}
		})
	}
}

func TestListTestExecutions(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewTestExecutionWriter(db)
	r := NewTestExecutionReader(db)

	dummyTest := createDummyTest(ctx, t, db, true)
	count := 4
	pageSize := 2

	want := make(test.TestExecutionList, count)
	for i := count - 1; i >= 0; i-- {
		scheduled := fake.GenScheduledTestExec(dummyTest.ID)
		created, err := w.CreateTestExecutionScheduled(ctx, scheduled)
		require.NoError(t, err)
		want[i] = created // add in reverse since we expect order by ascending
	}

	// Page 1
	got1, err := r.ListTestExecutions(ctx, dummyTest.ID, test.PageFilter[test.TestExecutionID]{
		Size:     pageSize,
		OffsetID: nil,
	})
	require.NoError(t, err)
	require.Len(t, got1, pageSize)

	// Page 2
	got2, err := r.ListTestExecutions(ctx, dummyTest.ID, test.PageFilter[test.TestExecutionID]{
		Size:     pageSize,
		OffsetID: ptr.Get(got1[1].ID),
	})
	require.NoError(t, err)
	assert.Len(t, got2, pageSize)

	got := append(got1, got2...)
	assert.Equal(t, want, got)

	// Page 3 (empty)
	got3, err := r.ListTestExecutions(ctx, dummyTest.ID, test.PageFilter[test.TestExecutionID]{
		Size:     pageSize,
		OffsetID: ptr.Get(got2[1].ID),
	})
	require.NoError(t, err)
	assert.Empty(t, got3)
}

func TestUpdateStartedTestExecution(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewTestExecutionWriter(db)

	dummyTest := createDummyTest(ctx, t, db, true)
	created, err := db.CreateTestExecutionScheduled(ctx, sqlc.CreateTestExecutionScheduledParams{
		ID:           test.NewTestExecutionID(),
		TestID:       dummyTest.ID,
		HasInput:     false,
		ScheduleTime: time.Now(),
	})
	require.NoError(t, err)

	started := &test.StartedTestExecution{
		ID:        created.ID,
		StartTime: time.Now().UTC(),
	}

	got, err := w.UpdateTestExecutionStarted(ctx, started)
	require.NoError(t, err)

	assert.Equal(t, started.ID, got.ID)
	assert.Equal(t, started.StartTime, *got.StartTime)
	assert.Nil(t, got.FinishTime)
	assert.Nil(t, got.Error)
}

func TestUpdateFinishedTestExecution(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewTestExecutionWriter(db)

	dummyTest := createDummyTest(ctx, t, db, true)
	created, err := db.CreateTestExecutionScheduled(ctx, sqlc.CreateTestExecutionScheduledParams{
		ID:           test.NewTestExecutionID(),
		TestID:       dummyTest.ID,
		HasInput:     false,
		ScheduleTime: time.Now(),
	})
	require.NoError(t, err)

	finished := &test.FinishedTestExecution{
		ID:         created.ID,
		FinishTime: time.Now().UTC(),
		Error:      ptr.Get("bang"),
	}

	got, err := w.UpdateTestExecutionFinished(ctx, finished)
	require.NoError(t, err)

	assert.Equal(t, finished.ID, got.ID)
	assert.Equal(t, finished.FinishTime, *got.FinishTime)
	assert.Equal(t, finished.Error, got.Error)
	assert.Nil(t, got.StartTime)
}

func TestResetTestExecution(t *testing.T) {

}
