package sqlite

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/test"
)

func TestCreateListContexts(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewContextWriter(db)
	r := NewContextReader(db)

	count := 4
	pageSize := 2

	var want []string
	for i := 0; i < count; i++ {
		contextID := fmt.Sprint("context-", i)
		err := w.CreateContext(ctx, contextID)
		require.NoError(t, err)
		want = append(want, contextID)
	}

	got, err := r.ListContexts(ctx, test.PageFilter[string]{Size: count})
	require.NoError(t, err)
	assert.Equal(t, want, got)

	// Page 1
	got1, err := r.ListContexts(ctx, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: nil,
	})
	require.NoError(t, err)
	require.Len(t, got1, pageSize)

	// Page 2
	got2, err := r.ListContexts(ctx, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: &got1[1],
	})
	require.NoError(t, err)
	assert.Len(t, got2, pageSize)

	got = append(got1, got2...)
	assert.Equal(t, want, got)

	// Page 3 (empty)
	got3, err := r.ListContexts(ctx, test.PageFilter[string]{
		Size:     pageSize,
		OffsetID: &got2[1],
	})
	require.NoError(t, err)
	assert.Empty(t, got3)
}

func TestCreateExistingContext(t *testing.T) {
	tests := []struct {
		name      string
		contextID string
		wantErr   bool
	}{
		{
			name:      "create existing default context returns no error",
			contextID: "default",
			wantErr:   false,
		},
		{
			name:      "create any other existing context returns error",
			contextID: "other",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db, closer := newTestDB(t)
			defer closer()
			w := NewContextWriter(db)

			err := w.CreateContext(ctx, tt.contextID)
			require.NoError(t, err)

			err = w.CreateContext(ctx, tt.contextID)
			if tt.wantErr {
				assert.ErrorIs(t, err, test.ErrorContextAlreadyExists)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
