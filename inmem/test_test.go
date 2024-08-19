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

func TestTestReader_GetTest(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewTestReader(db)

	want := fake.GenTest()
	db.tests[want.ID] = want

	got, err := r.GetTest(ctx, want.ID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestTestReader_GetTestDefaultInput(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewTestReader(db)

	want := fake.GenDefaultInput()
	testID := uuid.New()
	db.defaultInputs[testID] = want

	got, err := r.GetTestDefaultInput(ctx, testID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestTestReader_ListTests(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewTestReader(db)

	contextID := uuid.NewString()
	groupID := uuid.NewString()

	count := 30
	want := make(test.TestList, count)

	for i := range count {
		tt := fake.GenTest(fake.WithContextID(contextID), fake.WithGroupID(groupID))
		want[i] = tt
		db.tests[tt.ID] = tt
	}

	got, err := r.ListTests(ctx, contextID, groupID)
	require.NoError(t, err)
	assert.Len(t, got, count)
	require.Equal(t, want, got)
}

func TestTestWriter_CreateTest(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	w := NewTestWriter(db)

	def := fake.GenTestDefinition()
	got, err := w.CreateTest(ctx, def)
	require.NoError(t, err)
	assertCreatedTest(t, def, got)
}

func TestTestWriter_CreateTests(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	w := NewTestWriter(db)

	count := 30
	defs := make([]*test.TestDefinition, count)
	for i := range count {
		defs[i] = fake.GenTestDefinition()
	}

	gotList, err := w.CreateTests(ctx, defs...)
	require.NoError(t, err)

	for i, got := range gotList {
		assertCreatedTest(t, defs[i], got)
	}
}

func assertCreatedTest(t *testing.T, def *test.TestDefinition, got *test.Test) {
	assert.Equal(t, def.TestID, got.ID)
	assert.Equal(t, def.Name, got.Name)
	assert.Equal(t, def.GroupID, got.GroupID)
	assert.Equal(t, def.DefaultInput != nil, got.HasInput)
}
