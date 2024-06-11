package testservice

import (
	"context"
	"testing"
	"time"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	testv1 "github.com/annexsh/annex-proto/gen/go/type/test/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/test"
)

func TestService_RegisterTest(t *testing.T) {
	tests := []struct {
		name         string
		testName     string
		defaultInput *testv1.Payload
	}{
		{
			name:         "create test without payload",
			testName:     uuid.NewString(),
			defaultInput: nil,
		},
		{
			name:         "create test with payload",
			testName:     uuid.NewString(),
			defaultInput: fake.GenDefaultInput().Proto(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testservicev1.RegisterTestsRequest{
				Context: uuid.NewString(),
				Group:   uuid.NewString(),
				Definitions: []*testv1.TestDefinition{
					{
						Name:         tt.testName,
						DefaultInput: tt.defaultInput,
					},
				},
			}
			s, _ := newService()
			res, err := s.RegisterTests(context.Background(), req)
			require.NoError(t, err)
			require.Len(t, res.Tests, len(req.Definitions))

			for _, gotTest := range res.Tests {
				assert.NotEmpty(t, gotTest.Id)
				assert.Equal(t, req.Context, gotTest.Context)
				assert.Equal(t, req.Group, gotTest.Group)
				assert.Equal(t, tt.testName, gotTest.Name)
				assert.Equal(t, tt.defaultInput != nil, gotTest.HasInput)
				assert.NotEmpty(t, gotTest.CreateTime)
			}
		})
	}
}

func TestService_GetDefaultInput(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	def := fake.GenTestDefinition()
	want := string(def.DefaultInput.Data)

	_, err := fakes.repo.CreateTest(ctx, def)
	require.NoError(t, err)

	res, err := s.GetTestDefaultInput(ctx, &testservicev1.GetTestDefaultInputRequest{
		TestId: def.TestID.String(),
	})
	require.NoError(t, err)

	assert.Equal(t, want, res.DefaultInput)
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
		TestId: tt.ID.String(),
		Input:  fake.GenInput().Proto(),
	})
	require.NoError(t, err)

	testExecID, err := test.ParseTestExecutionID(res.TestExecution.Id)
	require.NoError(t, err)

	wr := fakes.workflower.GetWorkflow(ctx, testExecID.WorkflowID(), "")
	err = wr.Get(ctx, nil)
	require.NoError(t, err)
}
