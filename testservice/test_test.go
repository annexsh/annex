package testservice

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func TestService_RegisterTests(t *testing.T) {
	defs := []*testsv1.TestDefinition{
		{Name: "test-1", DefaultInput: nil},
		{Name: "test-2", DefaultInput: fake.GenDefaultInput().Proto()},
	}

	req := &testsv1.RegisterTestsRequest{
		Context:     uuid.NewString(),
		Group:       uuid.NewString(),
		Definitions: defs,
	}

	r := &RepositoryMock{
		CreateGroupFunc: func(ctx context.Context, contextID string, groupID string) error {
			assert.Equal(t, req.Context, contextID)
			assert.Equal(t, req.Group, groupID)
			return nil
		},
	}

	r.CreateTestFunc = func(ctx context.Context, def *test.Test) (*test.Test, error) {
		assert.False(t, def.ID.Empty())
		assert.False(t, def.CreateTime.IsZero())
		i := len(r.CreateTestCalls()) - 1
		wantTest := &test.Test{
			ContextID:  req.Context,
			GroupID:    req.Group,
			ID:         def.ID,
			Name:       req.Definitions[i].Name,
			HasInput:   req.Definitions[i].DefaultInput != nil,
			CreateTime: def.CreateTime,
		}
		assert.Equal(t, wantTest, def)
		return def, nil
	}

	r.CreateTestDefaultInputFunc = func(ctx context.Context, testID uuid.V7, defaultInput *test.Payload) error {
		assert.False(t, testID.Empty())
		i := len(r.CreateTestCalls()) - 1
		assert.Equal(t, req.Definitions[i].DefaultInput, defaultInput.Proto())
		return nil
	}

	r.ExecuteTxFunc = func(ctx context.Context, query func(repo test.Repository) error) error {
		return query(r)
	}

	s := Service{repo: r}

	res, err := s.RegisterTests(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	require.Len(t, res.Msg.Tests, len(req.Definitions))

	for i, gotTest := range res.Msg.Tests {
		assert.NotEmpty(t, gotTest.Id)
		assert.Equal(t, req.Context, gotTest.Context)
		assert.Equal(t, req.Group, gotTest.Group)
		assert.Equal(t, req.Definitions[i].Name, gotTest.Name)
		assert.Equal(t, req.Definitions[i].DefaultInput != nil, gotTest.HasInput)
		assert.NotEmpty(t, gotTest.CreateTime)
	}
}

func TestService_RegisterTests_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.RegisterTestsRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.RegisterTestsRequest{
				Context:     "",
				Group:       "foo",
				Definitions: []*testsv1.TestDefinition{{Name: "bar"}},
			},
			wantFieldViolation: wantBlankContextFieldViolation,
		},
		{
			name: "blank name",
			req: &testsv1.RegisterTestsRequest{
				Context:     "foo",
				Group:       "",
				Definitions: []*testsv1.TestDefinition{{Name: "bar"}},
			},
			wantFieldViolation: wantBlankGroupFieldViolation,
		},
		{
			name: "empty definitions",
			req: &testsv1.RegisterTestsRequest{
				Context:     "foo",
				Group:       "bar",
				Definitions: []*testsv1.TestDefinition{},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "definitions",
				Description: "Definitions must not be empty",
			},
		},
		{
			name: "definitions length greater than max",
			req: &testsv1.RegisterTestsRequest{
				Context:     "foo",
				Group:       "bar",
				Definitions: genTestDefinitions(maxRegisterTests + 1),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "definitions",
				Description: fmt.Sprintf("Definitions must not exceed a length of %d per request", maxRegisterTests),
			},
		},
		{
			name: "default input data empty",
			req: &testsv1.RegisterTestsRequest{
				Context: "foo",
				Group:   "bar",
				Definitions: []*testsv1.TestDefinition{
					{
						Name: "baz",
						DefaultInput: &testsv1.Payload{
							Data:     nil,
							Metadata: map[string][]byte{"encoding": []byte(converter.MetadataEncodingJSON)},
						},
					},
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "definitions[0].default_input.data",
				Description: "Data can't be empty",
			},
		},
		{
			name: "default input missing json encoding metadata",
			req: &testsv1.RegisterTestsRequest{
				Context: "foo",
				Group:   "bar",
				Definitions: []*testsv1.TestDefinition{
					{
						Name: "baz",
						DefaultInput: &testsv1.Payload{
							Data: []byte("qux"),
						},
					},
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "definitions[0].default_input.metadata",
				Description: "Metadata encoding must be json/plain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.RegisterTests(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_GetTest(t *testing.T) {
	testID := uuid.New()
	tt := fake.GenTest()

	r := &RepositoryMock{
		GetTestFunc: func(ctx context.Context, id uuid.V7) (*test.Test, error) {
			assert.Equal(t, testID, id)
			return tt, nil
		},
	}

	s := Service{repo: r}

	req := &testsv1.GetTestRequest{
		Context: "foo",
		TestId:  testID.String(),
	}
	res, err := s.GetTest(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, tt.Proto(), res.Msg.Test)
}

func TestService_GetTest_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.GetTestRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.GetTestRequest{
				Context: "",
				TestId:  uuid.NewString(),
			},
			wantFieldViolation: wantBlankContextFieldViolation,
		},
		{
			name: "blank test id",
			req: &testsv1.GetTestRequest{
				Context: "foo",
				TestId:  "",
			},
			wantFieldViolation: wantBlankTestIDFieldViolation,
		},
		{
			name: "test id not a uuid",
			req: &testsv1.GetTestRequest{
				Context: "foo",
				TestId:  "bar",
			},
			wantFieldViolation: wantTestIDNotUUIDFieldViolation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.GetTest(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_GetDefaultInput(t *testing.T) {
	testID := uuid.New()
	input := fake.GenInput()

	r := &RepositoryMock{
		GetTestDefaultInputFunc: func(ctx context.Context, gotTestID uuid.V7) (*test.Payload, error) {
			assert.Equal(t, testID, gotTestID)
			return input, nil
		},
	}

	s := Service{repo: r}

	req := &testsv1.GetTestDefaultInputRequest{
		Context: "foo",
		TestId:  testID.String(),
	}
	res, err := s.GetTestDefaultInput(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)

	want := string(input.Data)
	assert.Equal(t, want, res.Msg.DefaultInput)
}

func TestService_GetDefaultInput_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.GetTestDefaultInputRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.GetTestDefaultInputRequest{
				Context: "",
				TestId:  uuid.NewString(),
			},
			wantFieldViolation: wantBlankContextFieldViolation,
		},
		{
			name: "blank test id",
			req: &testsv1.GetTestDefaultInputRequest{
				Context: "foo",
				TestId:  "",
			},
			wantFieldViolation: wantBlankTestIDFieldViolation,
		},
		{
			name: "test id not a uuid",
			req: &testsv1.GetTestDefaultInputRequest{
				Context: "foo",
				TestId:  "bar",
			},
			wantFieldViolation: wantTestIDNotUUIDFieldViolation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.GetTestDefaultInput(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_ListTests(t *testing.T) {
	pageSize := 2
	contextID := "foo"
	groupID := "bar"
	wantPage1 := test.TestList{
		fake.GenTest(fake.WithContextID(contextID), fake.WithGroupID(groupID)),
		fake.GenTest(fake.WithContextID(contextID), fake.WithGroupID(groupID)),
	}
	wantPage2 := test.TestList{
		fake.GenTest(fake.WithContextID(contextID), fake.WithGroupID(groupID)),
	}

	r := new(RepositoryMock)
	r.ListTestsFunc = func(ctx context.Context, gotContextID string, gotGroupID string, filter test.PageFilter[uuid.V7]) (test.TestList, error) {
		assert.Equal(t, contextID, gotContextID)
		assert.Equal(t, groupID, gotGroupID)
		assert.Equal(t, pageSize, filter.Size)

		switch len(r.ListTestsCalls()) {
		case 1:
			assert.Nil(t, filter.OffsetID)
			return wantPage1, nil
		case 2:
			assert.Equal(t, wantPage1[pageSize-1].ID, *filter.OffsetID)
			return wantPage2, nil
		default:
			panic("unexpected list invocation")
		}
	}

	s := Service{repo: r}

	req := &testsv1.ListTestsRequest{
		Context:  contextID,
		Group:    groupID,
		PageSize: int32(pageSize),
	}
	res, err := s.ListTests(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage1.Proto(), res.Msg.Tests)
	assert.NotEmpty(t, res.Msg.NextPageToken)

	req.NextPageToken = res.Msg.NextPageToken
	res, err = s.ListTests(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage2.Proto(), res.Msg.Tests)
	assert.Empty(t, res.Msg.NextPageToken)
}

func TestService_ListTests_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.ListTestsRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.ListTestsRequest{
				Context:  "",
				Group:    "foo",
				PageSize: 1,
			},
			wantFieldViolation: wantBlankContextFieldViolation,
		},
		{
			name: "blank group",
			req: &testsv1.ListTestsRequest{
				Context:  "foo",
				Group:    "",
				PageSize: 1,
			},
			wantFieldViolation: wantBlankGroupFieldViolation,
		},
		{
			name: "page size less than 0",
			req: &testsv1.ListTestsRequest{
				Context:  "foo",
				Group:    "bar",
				PageSize: int32(-1),
			},
			wantFieldViolation: wantPageSizeFieldViolation,
		},
		{
			name: "page size greater than max",
			req: &testsv1.ListTestsRequest{
				Context:  "foo",
				Group:    "bar",
				PageSize: maxPageSize + 1,
			},
			wantFieldViolation: wantPageSizeFieldViolation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.ListTests(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_ExecuteTest(t *testing.T) {
	input := fake.GenInput()
	tt := fake.GenTest(fake.WithHasInput(true))

	var gotTestExec *test.TestExecution

	r := &RepositoryMock{
		GetTestFunc: func(ctx context.Context, testID uuid.V7) (*test.Test, error) {
			assert.Equal(t, tt.ID, testID)
			return tt, nil
		},
		CreateTestExecutionScheduledFunc: func(ctx context.Context, scheduled *test.ScheduledTestExecution) (*test.TestExecution, error) {
			assert.Equal(t, tt.ID, scheduled.TestID)
			assert.False(t, scheduled.ID.Empty())
			assert.Equal(t, tt.HasInput, scheduled.HasInput)
			assert.False(t, scheduled.ScheduleTime.IsZero())
			gotTestExec = &test.TestExecution{
				ID:           scheduled.ID,
				TestID:       scheduled.TestID,
				HasInput:     scheduled.HasInput,
				ScheduleTime: scheduled.ScheduleTime,
			}
			return gotTestExec, nil
		},
		CreateTestExecutionInputFunc: func(ctx context.Context, testExecID test.TestExecutionID, gotInput *test.Payload) error {
			assert.Equal(t, gotTestExec.ID, testExecID)
			assert.Equal(t, input, gotInput)
			return nil
		},
	}

	r.ExecuteTxFunc = func(ctx context.Context, query func(repo test.Repository) error) error {
		return query(r)
	}

	p := &PublisherMock{
		PublishFunc: func(testExecID string, e *eventsv1.Event) error {
			assert.Equal(t, gotTestExec.ID.String(), testExecID)
			assert.Equal(t, gotTestExec.ID.String(), e.TestExecutionId)
			assert.Equal(t, eventsv1.Event_TYPE_TEST_EXECUTION_SCHEDULED, e.Type)
			assert.NotEmpty(t, e.EventId)
			assert.True(t, e.CreateTime.IsValid())
			assert.Equal(t, eventsv1.Event_Data_TYPE_TEST_EXECUTION, e.Data.Type)
			assert.Equal(t, gotTestExec.Proto(), e.Data.GetTestExecution())
			return nil
		},
	}

	w := &WorkflowerMock{
		ExecuteWorkflowFunc: func(ctx context.Context, options client.StartWorkflowOptions, workflow any, args ...any) (client.WorkflowRun, error) {
			want := newStartWorkflowOpts(gotTestExec.ID.WorkflowID(), tt.ContextID, tt.GroupID)
			assert.Equal(t, want, options)
			assert.Equal(t, tt.Name, workflow)
			assert.Len(t, args, 1)
			assert.Equal(t, input.Proto(), args[0])
			return nil, nil
		},
	}

	s := Service{
		repo:     r,
		executor: newExecutor(r, p, w, log.NewNopLogger()),
	}

	req := &testsv1.ExecuteTestRequest{
		Context: tt.ContextID,
		TestId:  tt.ID.String(),
		Input:   input.Proto(),
	}

	res, err := s.ExecuteTest(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, gotTestExec.Proto(), res.Msg.TestExecution)
}

func TestService_ExecuteTest_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.ExecuteTestRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.ExecuteTestRequest{
				Context: "",
				TestId:  uuid.NewString(),
			},
			wantFieldViolation: wantBlankContextFieldViolation,
		},
		{
			name: "blank test id",
			req: &testsv1.ExecuteTestRequest{
				Context: "foo",
				TestId:  "",
			},
			wantFieldViolation: wantBlankTestIDFieldViolation,
		},
		{
			name: "test id not a uuid",
			req: &testsv1.ExecuteTestRequest{
				Context: "foo",
				TestId:  "bar",
			},
			wantFieldViolation: wantTestIDNotUUIDFieldViolation,
		},
		{
			name: "default input data empty",
			req: &testsv1.ExecuteTestRequest{
				Context: "foo",
				TestId:  uuid.NewString(),
				Input: &testsv1.Payload{
					Data:     nil,
					Metadata: map[string][]byte{"encoding": []byte(converter.MetadataEncodingJSON)},
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "input.data",
				Description: "Data can't be empty",
			},
		},
		{
			name: "default input missing json encoding metadata",
			req: &testsv1.ExecuteTestRequest{
				Context: "foo",
				TestId:  uuid.NewString(),
				Input: &testsv1.Payload{
					Data: []byte("qux"),
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "input.metadata",
				Description: "Metadata encoding must be json/plain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.ExecuteTest(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func genTestDefinitions(count int) []*testsv1.TestDefinition {
	defs := make([]*testsv1.TestDefinition, count)
	for i := 0; i < count; i++ {
		defs[i] = &testsv1.TestDefinition{
			Name:         strconv.Itoa(i),
			DefaultInput: fake.GenDefaultInput().Proto(),
		}
	}
	return defs
}
