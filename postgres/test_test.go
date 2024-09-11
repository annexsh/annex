package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/postgres/sqlc"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func TestCreateGetTest(t *testing.T) {
	tests := []struct {
		name         string
		defaultInput *test.Payload
	}{
		{
			name:         "no default input",
			defaultInput: nil,
		},
		{
			name:         "has default input",
			defaultInput: fake.GenDefaultInput(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db, closer := newTestDB(t)
			defer closer()

			w := NewTestWriter(db)
			r := NewTestReader(db)

			contextID := "foo"
			groupID := "bar"

			err := db.CreateContext(ctx, contextID)
			require.NoError(t, err)
			err = db.CreateGroup(ctx, sqlc.CreateGroupParams{ContextID: contextID, ID: groupID})
			require.NoError(t, err)

			def := fake.GenTest(fake.WithContextID(contextID), fake.WithGroupID(groupID))

			assertEqual := func(got *test.Test) {
				assert.Equal(t, def.ID, got.ID)
				assert.Equal(t, def.ContextID, got.ContextID)
				assert.Equal(t, def.GroupID, got.GroupID)
				assert.Equal(t, def.Name, got.Name)
				assert.Equal(t, def.HasInput, got.HasInput)
				assert.Equal(t, def.CreateTime, got.CreateTime)
			}

			got, err := w.CreateTest(ctx, def)
			require.NoError(t, err)
			assertEqual(got)

			got, err = r.GetTest(ctx, got.ID)
			require.NoError(t, err)
			assertEqual(got)

			if tt.defaultInput != nil {
				err = w.CreateTestDefaultInput(ctx, got.ID, tt.defaultInput)
				require.NoError(t, err)
			}

			gotDefInput, err := r.GetTestDefaultInput(ctx, got.ID)
			if tt.defaultInput != nil {
				require.NoError(t, err)
				assert.Equal(t, tt.defaultInput, gotDefInput)
			} else {
				assert.ErrorIs(t, err, test.ErrorTestPayloadNotFound)
				assert.Nil(t, gotDefInput)
			}
		})
	}
}

func TestListTests(t *testing.T) {
	ctx := context.Background()
	db, closer := newTestDB(t)
	defer closer()

	w := NewTestWriter(db)
	r := NewTestReader(db)

	contextID := "foo"
	groupID := "bar"
	count := 4
	pageSize := 2

	err := db.CreateContext(ctx, contextID)
	require.NoError(t, err)
	err = db.CreateGroup(ctx, sqlc.CreateGroupParams{ContextID: contextID, ID: groupID})
	require.NoError(t, err)

	want := make(test.TestList, count)
	for i := count - 1; i >= 0; i-- {
		tt := fake.GenTest(fake.WithContextID(contextID), fake.WithGroupID(groupID))
		created, err := w.CreateTest(ctx, tt)
		require.NoError(t, err)
		want[i] = created // add in reverse since we expect order by ascending
	}

	// Page 1
	got1, err := r.ListTests(ctx, contextID, groupID, test.PageFilter[uuid.V7]{
		Size:     pageSize,
		OffsetID: nil,
	})
	require.NoError(t, err)
	require.Len(t, got1, pageSize)

	// Page 2
	got2, err := r.ListTests(ctx, contextID, groupID, test.PageFilter[uuid.V7]{
		Size:     pageSize,
		OffsetID: ptr.Get(got1[1].ID),
	})
	require.NoError(t, err)
	assert.Len(t, got2, pageSize)

	got := append(got1, got2...)
	assert.Equal(t, want, got)

	// Page 3 (empty)
	got3, err := r.ListTests(ctx, contextID, groupID, test.PageFilter[uuid.V7]{
		Size:     pageSize,
		OffsetID: ptr.Get(got2[1].ID),
	})
	require.NoError(t, err)
	assert.Empty(t, got3)
}
