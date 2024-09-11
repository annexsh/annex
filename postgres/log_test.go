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
	"github.com/annexsh/annex/uuid"
)

func TestCreateGetLog(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)

	defer closer()
	w := NewLogWriter(db)
	r := NewLogReader(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)
	want := fake.GenTestExecLog(dummyTestExec.ID)

	err := w.CreateLog(ctx, want)
	require.NoError(t, err)

	got, err := r.GetLog(ctx, want.ID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestListLogs(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()
	w := NewLogWriter(db)
	r := NewLogReader(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)
	count := 4
	pageSize := 2

	want := make(test.LogList, count)
	for i := count - 1; i >= 0; i-- {
		log := fake.GenTestExecLog(dummyTestExec.ID)
		err := w.CreateLog(ctx, log)
		require.NoError(t, err)
		want[i] = log // add in reverse since we expect order by ascending
	}

	// Page 1
	got1, err := r.ListLogs(ctx, dummyTestExec.ID, test.PageFilter[uuid.V7]{
		Size:     pageSize,
		OffsetID: nil,
	})
	require.NoError(t, err)
	require.Len(t, got1, pageSize)

	// Page 2
	got2, err := r.ListLogs(ctx, dummyTestExec.ID, test.PageFilter[uuid.V7]{
		Size:     pageSize,
		OffsetID: ptr.Get(got1[1].ID),
	})
	require.NoError(t, err)
	assert.Len(t, got2, pageSize)

	got := append(got1, got2...)
	assert.Equal(t, want, got)

	// Page 3 (empty)
	got3, err := r.ListLogs(ctx, dummyTestExec.ID, test.PageFilter[uuid.V7]{
		Size:     pageSize,
		OffsetID: ptr.Get(got2[1].ID),
	})
	require.NoError(t, err)
	assert.Empty(t, got3)
}

func TestDeleteLog(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()
	w := NewLogWriter(db)

	dummyTestExec := createDummyTestExec(ctx, t, db)

	logID := uuid.New()
	err := db.CreateLog(ctx, sqlc.CreateLogParams{
		ID:              logID,
		TestExecutionID: dummyTestExec.ID,
		Level:           "INFO",
		Message:         "Lorem ipsum dolor sit amet,",
		CreateTime:      time.Now().UTC(),
	})
	require.NoError(t, err)

	err = w.DeleteLog(ctx, logID)
	require.NoError(t, err)
	_, err = db.GetLog(ctx, logID)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}
