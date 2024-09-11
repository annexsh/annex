package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/sqlite/sqlc"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func newTestDB(t *testing.T) (*DB, func()) {
	sqldb, err := Open(WithMigration())
	require.NoError(t, err)
	return NewDB(sqldb), func() {
		require.NoError(t, sqldb.Close())
	}
}

func createDummyTest(ctx context.Context, t *testing.T, db *DB, hasInput bool) *sqlc.Test {
	contextID := "foo"
	groupID := "bar"
	err := db.CreateContext(ctx, contextID)
	require.NoError(t, err)
	err = db.CreateGroup(ctx, sqlc.CreateGroupParams{ContextID: contextID, ID: groupID})
	require.NoError(t, err)

	dummyTest, err := db.CreateTest(ctx, sqlc.CreateTestParams{
		ContextID:  contextID,
		GroupID:    groupID,
		ID:         uuid.New(),
		Name:       "baz",
		HasInput:   hasInput,
		CreateTime: time.Now(),
	})
	require.NoError(t, err)
	return dummyTest
}

func createDummyTestExec(ctx context.Context, t *testing.T, db *DB) *sqlc.TestExecution {
	dummyTest := createDummyTest(ctx, t, db, false)
	dummyTestExec, err := db.CreateTestExecutionScheduled(ctx, sqlc.CreateTestExecutionScheduledParams{
		ID:           test.NewTestExecutionID(),
		TestID:       dummyTest.ID,
		HasInput:     false,
		ScheduleTime: time.Now(),
	})
	require.NoError(t, err)
	return dummyTestExec
}
