package inmem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/test"
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

func TestTestReader_GetTestDefaultPayload(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewTestReader(db)

	want := fake.GenDefaultPayload()
	testID := uuid.New()
	db.defaultPayloads[testID] = want

	got, err := r.GetTestDefaultPayload(ctx, testID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestTestReader_ListTests(t *testing.T) {
	ctx := context.Background()
	db := NewDB()
	r := NewTestReader(db)

	count := 30
	want := make(test.TestList, count)

	for i := range count {
		tt := fake.GenTest()
		want[i] = tt
		db.tests[tt.ID] = tt
	}

	got, err := r.ListTests(ctx)
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
	assert.Equal(t, def.Project, got.Project)
	assert.Len(t, got.Runners, 1)
	assert.Equal(t, def.RunnerID, got.Runners[0].ID)
	assert.Equal(t, true, got.Runners[0].IsActive)
	assert.NotEmpty(t, true, got.Runners[0].LastHeartbeatTime)
	assert.Equal(t, def.DefaultPayload != nil, got.HasPayload)
}
