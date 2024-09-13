//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/postgres/sqlc"
	"github.com/annexsh/annex/test"
)

func TestCreateGetCaseExecution(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewCaseExecutionWriter(db)
	r := NewCaseExecutionReader(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)
	scheduled := fake.GenScheduledCaseExec(dummyTestExec.ID)

	assertEqual := func(got *test.CaseExecution) {
		assert.Equal(t, scheduled.ID, got.ID)
		assert.Equal(t, scheduled.TestExecutionID, got.TestExecutionID)
		assert.Equal(t, scheduled.CaseName, got.CaseName)
		assert.Equal(t, scheduled.ScheduleTime, got.ScheduleTime)
		assert.Nil(t, got.StartTime)
		assert.Nil(t, got.FinishTime)
		assert.Nil(t, got.Error)
	}

	got, err := w.CreateCaseExecutionScheduled(ctx, scheduled)
	require.NoError(t, err)
	assertEqual(got)

	got, err = r.GetCaseExecution(ctx, got.TestExecutionID, got.ID)
	require.NoError(t, err)
	assertEqual(got)
}

func TestListCaseExecutions(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewCaseExecutionWriter(db)
	r := NewCaseExecutionReader(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)
	count := 4
	pageSize := 2

	want := make(test.CaseExecutionList, count)
	for i := 0; i < count; i++ {
		scheduled := fake.GenScheduledCaseExec(dummyTestExec.ID)
		created, err := w.CreateCaseExecutionScheduled(ctx, scheduled)
		require.NoError(t, err)
		want[i] = created
	}

	// Page 1
	got1, err := r.ListCaseExecutions(ctx, dummyTestExec.ID, test.PageFilter[test.CaseExecutionID]{
		Size:     pageSize,
		OffsetID: nil,
	})
	require.NoError(t, err)
	require.Len(t, got1, pageSize)

	// Page 2
	got2, err := r.ListCaseExecutions(ctx, dummyTestExec.ID, test.PageFilter[test.CaseExecutionID]{
		Size:     pageSize,
		OffsetID: ptr.Get(got1[1].ID),
	})
	require.NoError(t, err)
	assert.Len(t, got2, pageSize)

	got := append(got1, got2...)
	assert.Equal(t, want, got)

	// Page 3 (empty)
	got3, err := r.ListCaseExecutions(ctx, dummyTestExec.ID, test.PageFilter[test.CaseExecutionID]{
		Size:     pageSize,
		OffsetID: ptr.Get(got2[1].ID),
	})
	require.NoError(t, err)
	assert.Empty(t, got3)
}

func TestUpdateStartedCaseExecution(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()
	w := NewCaseExecutionWriter(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)
	created, err := db.CreateCaseExecutionScheduled(ctx, sqlc.CreateCaseExecutionScheduledParams{
		ID:              1,
		TestExecutionID: dummyTestExec.ID,
		CaseName:        "foo",
		ScheduleTime:    time.Now().UTC(),
	})
	require.NoError(t, err)

	started := &test.StartedCaseExecution{
		ID:              created.ID,
		TestExecutionID: created.TestExecutionID,
		StartTime:       time.Now().UTC(),
	}

	got, err := w.UpdateCaseExecutionStarted(ctx, started)
	require.NoError(t, err)

	assert.Equal(t, started.ID, got.ID)
	assert.Equal(t, started.StartTime, *got.StartTime)
	assert.Nil(t, got.FinishTime)
	assert.Nil(t, got.Error)
}

func TestUpdateFinishedCaseExecution(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()
	w := NewCaseExecutionWriter(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)
	created, err := db.CreateCaseExecutionScheduled(ctx, sqlc.CreateCaseExecutionScheduledParams{
		ID:              1,
		TestExecutionID: dummyTestExec.ID,
		CaseName:        "foo",
		ScheduleTime:    time.Now().UTC(),
	})
	require.NoError(t, err)

	finished := &test.FinishedCaseExecution{
		ID:              created.ID,
		TestExecutionID: created.TestExecutionID,
		FinishTime:      time.Now().UTC(),
		Error:           ptr.Get("bang"),
	}

	got, err := w.UpdateCaseExecutionFinished(ctx, finished)
	require.NoError(t, err)

	assert.Equal(t, finished.ID, got.ID)
	assert.Equal(t, finished.FinishTime, *got.FinishTime)
	assert.Equal(t, finished.Error, got.Error)
	assert.Nil(t, got.StartTime)
}

func TestDeleteCaseExecution(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()
	w := NewCaseExecutionWriter(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)
	created, err := db.CreateCaseExecutionScheduled(ctx, sqlc.CreateCaseExecutionScheduledParams{
		ID:              1,
		TestExecutionID: dummyTestExec.ID,
		CaseName:        "foo",
		ScheduleTime:    time.Now().UTC(),
	})
	require.NoError(t, err)

	err = w.DeleteCaseExecution(ctx, created.TestExecutionID, created.ID)
	require.NoError(t, err)

	_, err = db.GetCaseExecution(ctx, sqlc.GetCaseExecutionParams{
		ID:              created.ID,
		TestExecutionID: created.TestExecutionID,
	})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}
