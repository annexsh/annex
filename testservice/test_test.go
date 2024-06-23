package testservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"
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
		defaultInput *testsv1.Payload
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
			req := &testsv1.RegisterTestsRequest{
				Context: uuid.NewString(),
				Group:   uuid.NewString(),
				Definitions: []*testsv1.TestDefinition{
					{
						Name:         tt.testName,
						DefaultInput: tt.defaultInput,
					},
				},
			}
			s, _ := newService()
			res, err := s.RegisterTests(context.Background(), connect.NewRequest(req))
			require.NoError(t, err)

			require.Len(t, res.Msg.Tests, len(req.Definitions))

			for _, gotTest := range res.Msg.Tests {
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

	req := &testsv1.GetTestDefaultInputRequest{
		TestId: def.TestID.String(),
	}
	res, err := s.GetTestDefaultInput(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	assert.Equal(t, want, res.Msg.DefaultInput)
}

func TestService_ListTests(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	wantCount := 30
	want := make([]*testsv1.Test, wantCount)

	contextID := "test-context"
	groupID := "test-group"
	err := fakes.repo.CreateContext(ctx, contextID)
	require.NoError(t, err)
	err = fakes.repo.CreateGroup(ctx, contextID, groupID)
	require.NoError(t, err)

	for i := range wantCount {
		def := fake.GenTestDefinition(fake.WithContextID(contextID), fake.WithGroupID(groupID))
		tt, err := fakes.repo.CreateTest(ctx, def)
		require.NoError(t, err)
		want[i] = tt.Proto()
	}

	req := &testsv1.ListTestsRequest{
		Context:       contextID,
		Group:         groupID,
		PageSize:      0, // TODO: pagination
		NextPageToken: "",
	}
	res, err := s.ListTests(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	got := res.Msg.Tests
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

	req := &testsv1.ExecuteTestRequest{
		TestId: tt.ID.String(),
		Input:  fake.GenInput().Proto(),
	}
	res, err := s.ExecuteTest(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	testExecID, err := test.ParseTestExecutionID(res.Msg.TestExecution.Id)
	require.NoError(t, err)

	wr := fakes.workflower.GetWorkflow(ctx, testExecID.WorkflowID(), "")
	err = wr.Get(ctx, nil)
	require.NoError(t, err)
}
