package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/postgres/sqlc"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func newTestDB(t *testing.T) (*DB, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := newPool(ctx, "postgres://postgres:postgres@0.0.0.0:5432/postgres")
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))

	_, err = pool.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName)
	require.NoError(t, err)

	pool.Close()

	pool, err = OpenPool(ctx, "postgres", "postgres", "0.0.0.0:5432", WithMigration())
	require.NoError(t, err)

	return NewDB(pool), pool.Close
}

func createDummyTest(ctx context.Context, t *testing.T, db *DB, hasInput bool) *sqlc.Test {
	contextID := "foo"
	testSuiteID := uuid.New()
	err := db.CreateContext(ctx, contextID)
	require.NoError(t, err)
	_, err = db.CreateTestSuite(ctx, sqlc.CreateTestSuiteParams{
		ID:        testSuiteID,
		ContextID: contextID,
		Name:      "foobar",
	})
	require.NoError(t, err)

	dummyTest, err := db.CreateTest(ctx, sqlc.CreateTestParams{
		ContextID:   contextID,
		TestSuiteID: testSuiteID,
		ID:          uuid.New(),
		Name:        "baz",
		HasInput:    hasInput,
		CreateTime:  time.Now(),
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
