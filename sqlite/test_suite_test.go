package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/test"
)

func TestCreateListTestSuites(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewTestSuiteWriter(db)
	r := NewTestSuiteReader(db)

	contextID := "default"
	err := db.CreateContext(ctx, contextID)
	require.NoError(t, err)

	count := 4
	pageSize := 2

	var want test.TestSuiteList
	for i := 0; i < count; i++ {
		testSuite := fake.GenTestSuite(contextID)
		_, err = w.CreateTestSuite(ctx, testSuite)
		require.NoError(t, err)
		want = append(want, testSuite)
	}

	// Page 1
	got1, err := r.ListTestSuites(ctx, contextID, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: nil,
	})
	require.NoError(t, err)
	require.Len(t, got1, pageSize)

	// Page 2
	got2, err := r.ListTestSuites(ctx, contextID, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: &got1[1].Name,
	})
	require.NoError(t, err)
	assert.Len(t, got2, pageSize)

	got := append(got1, got2...)
	assert.Equal(t, want, got)

	// Page 3 (empty)
	got3, err := r.ListTestSuites(ctx, contextID, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: &got2[1].Name,
	})
	require.NoError(t, err)
	assert.Empty(t, got3)
}
