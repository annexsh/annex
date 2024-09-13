//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/test"
)

func TestCreateListGroups(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewGroupWriter(db)
	r := NewGroupReader(db)

	contextID := "default"
	err := db.CreateContext(ctx, contextID)
	require.NoError(t, err)

	count := 4
	pageSize := 2

	var want []string
	for i := 0; i < count; i++ {
		groupID := fmt.Sprint("group-", i)
		err = w.CreateGroup(ctx, contextID, groupID)
		require.NoError(t, err)
		want = append(want, groupID)
	}

	got, err := r.ListGroups(ctx, contextID, test.PageFilter[string]{Size: count})
	require.NoError(t, err)
	assert.Equal(t, want, got)

	// Page 1
	got1, err := r.ListGroups(ctx, contextID, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: nil,
	})
	require.NoError(t, err)
	require.Len(t, got1, pageSize)

	// Page 2
	got2, err := r.ListGroups(ctx, contextID, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: &got1[1],
	})
	require.NoError(t, err)
	assert.Len(t, got2, pageSize)

	got = append(got1, got2...)
	assert.Equal(t, want, got)

	// Page 3 (empty)
	got3, err := r.ListGroups(ctx, contextID, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: &got2[1],
	})
	require.NoError(t, err)
	assert.Empty(t, got3)
}
