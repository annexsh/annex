package testservice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/annexsh/annex-proto/go/gen/annex/tests/v1/testsv1connect"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	initialVersion := "foo"
	newVersion := "bar"
	contextID := uuid.NewString()
	testSuiteID := uuid.New()

	existingTest1 := fake.GenTest(fake.WithContextID(contextID), fake.WithTestSuiteID(testSuiteID))
	existingTest2 := fake.GenTest(fake.WithContextID(contextID), fake.WithTestSuiteID(testSuiteID)) // want delete

	existing := test.TestList{
		existingTest1,
		existingTest2,
	}

	reqs := []*testsv1.RegisterTestsRequest{
		{ // updated existing test 1
			Context:     contextID,
			TestSuiteId: testSuiteID.String(),
			Definition: &testsv1.TestDefinition{
				Name:         existingTest1.Name,
				DefaultInput: fake.GenDefaultInput().Proto(),
			},
			Version:  newVersion,
			RunnerId: "bar",
		},
		{ // new test
			Context:     contextID,
			TestSuiteId: testSuiteID.String(),
			Definition: &testsv1.TestDefinition{
				Name:         "test-new",
				DefaultInput: fake.GenDefaultInput().Proto(),
			},
			Version:  newVersion,
			RunnerId: "bar",
		},
	}

	wantDefaultInputs := []string{reqs[0].Definition.DefaultInput.String(), reqs[1].Definition.DefaultInput.String()}

	r := &RepositoryMock{
		GetTestSuiteVersionFunc: func(ctx context.Context, contextID string, id uuid.V7) (string, error) {
			assert.Equal(t, contextID, contextID)
			assert.Equal(t, testSuiteID, id)
			return initialVersion, nil
		},
		ListTestsFunc: func(ctx context.Context, contextID string, id uuid.V7, filter test.PageFilter[uuid.V7]) (test.TestList, error) {
			assert.Equal(t, contextID, contextID)
			assert.Equal(t, testSuiteID, id)
			assert.Equal(t, 50, filter.Size)
			assert.Nil(t, filter.OffsetID)
			return existing, nil
		},
		DeleteTestFunc: func(ctx context.Context, id uuid.V7) error {
			assert.Equal(t, existingTest2.ID, id)
			return nil
		},
	}

	r.DeleteTestFunc = func(ctx context.Context, id uuid.V7) error {
		// Should only delete existing test 2
		assert.Equal(t, 1, len(r.DeleteTestCalls()))
		assert.Equal(t, existingTest2.ID, id)
		return nil
	}

	var createdTestIDs []uuid.V7

	r.CreateTestFunc = func(ctx context.Context, def *test.Test) (*test.Test, error) {
		assert.LessOrEqual(t, len(r.CreateTestCalls()), 2)
		assert.False(t, def.ID.Empty())
		assert.False(t, def.CreateTime.IsZero())

		var reqDef *testsv1.TestDefinition
		for _, req := range reqs {
			if def.Name == req.Definition.Name {
				reqDef = req.Definition
			}
		}
		require.NotNilf(t, reqDef, "matching request definition not found for CreateTest %+v", def)

		wantTest := &test.Test{
			ContextID:   contextID,
			TestSuiteID: testSuiteID,
			ID:          def.ID,
			Name:        reqDef.Name,
			HasInput:    reqDef.DefaultInput != nil,
			CreateTime:  def.CreateTime,
		}
		assert.Equal(t, wantTest, def)

		createdTestIDs = append(createdTestIDs, def.ID)
		return def, nil
	}

	r.CreateTestDefaultInputFunc = func(ctx context.Context, testID uuid.V7, defaultInput *test.Payload) error {
		assert.Contains(t, createdTestIDs, testID)
		assert.Contains(t, wantDefaultInputs, defaultInput.Proto().String())
		return nil
	}

	r.ExecuteTxFunc = func(ctx context.Context, query func(repo test.Repository) error) error {
		return query(r)
	}

	cli, closer := newTestServiceServer(r)
	defer closer()

	stream := cli.RegisterTests(ctx)

	for _, req := range reqs {
		err := stream.Send(req)
		require.NoError(t, err)
	}

	res, err := stream.CloseAndReceive()
	require.NoError(t, err)
	assert.NotNil(t, res.Msg)
}

func TestService_RegisterTests_validation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name               string
		msgs               []*testsv1.RegisterTestsRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			msgs: []*testsv1.RegisterTestsRequest{
				{
					Context:     "",
					TestSuiteId: uuid.NewString(),
					Definition:  &testsv1.TestDefinition{Name: "bar"},
					Version:     "v2",
					RunnerId:    "baz",
				},
			},
			wantFieldViolation: wantBlankContextFieldViolation(0),
		},
		{
			name: "blank test suit id",
			msgs: []*testsv1.RegisterTestsRequest{
				{
					Context:     "foo",
					TestSuiteId: "",
					Definition:  &testsv1.TestDefinition{Name: "bar"},
					Version:     "v2",
					RunnerId:    "baz",
				},
			},
			wantFieldViolation: wantBlankTestSuiteFieldViolation(0),
		},
		{
			name: "nil definition",
			msgs: []*testsv1.RegisterTestsRequest{
				{
					Context:     "foo",
					TestSuiteId: uuid.NewString(),
					Definition:  nil,
					Version:     "v2",
					RunnerId:    "baz",
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "stream[0].definition",
				Description: "Definition must not be nil",
			},
		},
		{
			name: "blank definition name",
			msgs: []*testsv1.RegisterTestsRequest{
				{
					Context:     "foo",
					TestSuiteId: uuid.NewString(),
					Definition:  &testsv1.TestDefinition{Name: ""},
					Version:     "v2",
					RunnerId:    "baz",
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "stream[0].definition.name",
				Description: "Name can't be blank",
			},
		},
		{
			name: "definition default input data empty",
			msgs: []*testsv1.RegisterTestsRequest{
				{
					Context:     "foo",
					TestSuiteId: uuid.NewString(),
					Definition: &testsv1.TestDefinition{
						Name: "baz",
						DefaultInput: &testsv1.Payload{
							Data:     nil,
							Metadata: map[string][]byte{"encoding": []byte(converter.MetadataEncodingJSON)},
						},
					},
					Version:  "v2",
					RunnerId: "baz",
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "stream[0].definition.default_input.data",
				Description: "Data can't be empty",
			},
		},
		{
			name: "definition default input missing json encoding metadata",
			msgs: []*testsv1.RegisterTestsRequest{
				{
					Context:     "foo",
					TestSuiteId: uuid.NewString(),
					Definition: &testsv1.TestDefinition{
						Name: "baz",
						DefaultInput: &testsv1.Payload{
							Data: []byte("qux"),
						},
					},
					Version:  "v2",
					RunnerId: "baz",
				},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "stream[0].definition.default_input.metadata",
				Description: "Metadata encoding must be json/plain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RepositoryMock{
				GetTestSuiteVersionFunc: func(ctx context.Context, contextID string, id uuid.V7) (string, error) {
					return "v1", nil
				},
				ListTestsFunc: func(ctx context.Context, contextID string, testSuiteID uuid.V7, filter test.PageFilter[uuid.V7]) (test.TestList, error) {
					return nil, nil
				},
			}
			r.ExecuteTxFunc = func(ctx context.Context, query func(repo test.Repository) error) error {
				return query(r)
			}

			cli, closer := newTestServiceServer(r)
			defer closer()

			stream := cli.RegisterTests(ctx)

			for _, req := range tt.msgs {
				err := stream.Send(req)
				require.NoError(t, err)
			}

			res, err := stream.CloseAndReceive()
			require.Nil(t, res)
			t.Log(err.Error())
			assertInvalidRequest(t, err, tt.wantFieldViolation, true)
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
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test id",
			req: &testsv1.GetTestRequest{
				Context: "foo",
				TestId:  "",
			},
			wantFieldViolation: wantBlankTestIDFieldViolation(),
		},
		{
			name: "test id not a uuid",
			req: &testsv1.GetTestRequest{
				Context: "foo",
				TestId:  "bar",
			},
			wantFieldViolation: wantTestIDNotUUIDFieldViolation(),
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
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test id",
			req: &testsv1.GetTestDefaultInputRequest{
				Context: "foo",
				TestId:  "",
			},
			wantFieldViolation: wantBlankTestIDFieldViolation(),
		},
		{
			name: "test id not a uuid",
			req: &testsv1.GetTestDefaultInputRequest{
				Context: "foo",
				TestId:  "bar",
			},
			wantFieldViolation: wantTestIDNotUUIDFieldViolation(),
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
	testSuiteID := uuid.New()
	wantPage1 := test.TestList{
		fake.GenTest(fake.WithContextID(contextID), fake.WithTestSuiteID(testSuiteID)),
		fake.GenTest(fake.WithContextID(contextID), fake.WithTestSuiteID(testSuiteID)),
	}
	wantPage2 := test.TestList{
		fake.GenTest(fake.WithContextID(contextID), fake.WithTestSuiteID(testSuiteID)),
	}

	r := new(RepositoryMock)
	r.ListTestsFunc = func(ctx context.Context, gotContextID string, gotTestSuiteID uuid.V7, filter test.PageFilter[uuid.V7]) (test.TestList, error) {
		assert.Equal(t, contextID, gotContextID)
		assert.Equal(t, testSuiteID, gotTestSuiteID)
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
		Context:     contextID,
		TestSuiteId: testSuiteID.String(),
		PageSize:    int32(pageSize),
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
				Context:     "",
				TestSuiteId: uuid.NewString(),
				PageSize:    1,
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test suite id",
			req: &testsv1.ListTestsRequest{
				Context:     "foo",
				TestSuiteId: "",
				PageSize:    1,
			},
			wantFieldViolation: wantBlankTestSuiteFieldViolation(),
		},
		{
			name: "page size less than 0",
			req: &testsv1.ListTestsRequest{
				Context:     "foo",
				TestSuiteId: uuid.NewString(),
				PageSize:    int32(-1),
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
		{
			name: "page size greater than max",
			req: &testsv1.ListTestsRequest{
				Context:     "foo",
				TestSuiteId: uuid.NewString(),
				PageSize:    maxPageSize + 1,
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
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
			want := newStartWorkflowOpts(gotTestExec.ID.WorkflowID(), tt.ContextID, tt.TestSuiteID)
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
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test id",
			req: &testsv1.ExecuteTestRequest{
				Context: "foo",
				TestId:  "",
			},
			wantFieldViolation: wantBlankTestIDFieldViolation(),
		},
		{
			name: "test id not a uuid",
			req: &testsv1.ExecuteTestRequest{
				Context: "foo",
				TestId:  "bar",
			},
			wantFieldViolation: wantTestIDNotUUIDFieldViolation(),
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

func newTestServiceServer(repo test.Repository) (testsv1connect.TestServiceClient, func()) {
	s := &Service{repo: repo}
	mux := http.NewServeMux()
	mux.Handle(testsv1connect.NewTestServiceHandler(s))
	srv := httptest.NewUnstartedServer(mux)
	srv.EnableHTTP2 = true
	srv.Start()
	return testsv1connect.NewTestServiceClient(srv.Client(), srv.URL, connect.WithGRPC()), srv.Close
}
