package testservice

import (
	"context"
	"testing"
	"time"

	testservicev1 "github.com/annexhq/annex-proto/gen/go/rpc/testservice/v1"
	testv1 "github.com/annexhq/annex-proto/gen/go/type/test/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexhq/annex/internal/fake"
	"github.com/annexhq/annex/test"
)

func TestService_RegisterTest(t *testing.T) {
	tests := []struct {
		name           string
		testName       string
		project        string
		defaultPayload *testv1.Payload
	}{
		{
			name:           "create test without payload",
			testName:       uuid.NewString(),
			project:        uuid.NewString(),
			defaultPayload: nil,
		},
		{
			name:           "create test with payload",
			testName:       uuid.NewString(),
			project:        uuid.NewString(),
			defaultPayload: fake.GenDefaultPayload().Proto(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testservicev1.RegisterTestsRequest{
				RunnerId: uuid.NewString(),
				Definitions: []*testv1.TestDefinition{
					{
						Project:        tt.project,
						Name:           tt.testName,
						DefaultPayload: tt.defaultPayload,
					},
				},
			}
			s, _ := newService()
			res, err := s.RegisterTests(context.Background(), req)
			require.NoError(t, err)
			require.Len(t, res.Registered, len(req.Definitions))

			for _, gotTest := range res.Registered {
				assert.NotEmpty(t, gotTest.Id)
				assert.Equal(t, tt.project, gotTest.Project)
				assert.Equal(t, tt.testName, gotTest.Name)
				assert.Equal(t, tt.defaultPayload != nil, gotTest.HasPayload)
				assert.Equal(t, req.RunnerId, gotTest.LastAvailable.Id)
				assert.NotEmpty(t, gotTest.LastAvailable.LastHeartbeat)
				assert.True(t, gotTest.LastAvailable.IsActive)
				assert.NotEmpty(t, gotTest.CreatedAt)
			}
		})
	}
}

func TestService_GetDefaultPayload(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	def := fake.GenTestDefinition()
	want := string(def.DefaultPayload.Payload)

	_, err := fakes.repo.CreateTest(ctx, def)
	require.NoError(t, err)

	res, err := s.GetTestDefaultPayload(ctx, &testservicev1.GetTestDefaultPayloadRequest{
		TestId: def.TestID.String(),
	})
	require.NoError(t, err)

	assert.Equal(t, want, res.DefaultPayload)
}

func TestService_ListTests(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	wantCount := 30
	want := make([]*testv1.Test, wantCount)

	for i := range wantCount {
		def := fake.GenTestDefinition()
		tt, err := fakes.repo.CreateTest(ctx, def)
		require.NoError(t, err)
		want[i] = tt.Proto()
	}

	res, err := s.ListTests(ctx, &testservicev1.ListTestsRequest{
		PageSize:      0, // TODO: pagination
		NextPageToken: "",
	})
	require.NoError(t, err)

	got := res.Tests
	assert.Len(t, got, wantCount)
	require.Equal(t, want, got)
}

func TestService_ExecuteTest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	s, fakes := newService()

	def := fake.GenTestDefinition()
	def.Name = fake.WorkflowName

	tt, err := fakes.repo.CreateTest(ctx, def)
	require.NoError(t, err)

	res, err := s.ExecuteTest(ctx, &testservicev1.ExecuteTestRequest{
		TestId:  tt.ID.String(),
		Payload: fake.GenPayload().Proto(),
	})
	require.NoError(t, err)

	testExecID, err := test.ParseTestExecutionID(res.TestExecution.Id)
	require.NoError(t, err)

	wr := fakes.workflower.GetWorkflow(ctx, testExecID.WorkflowID(), "")
	err = wr.Get(ctx, nil)
	require.NoError(t, err)
}
