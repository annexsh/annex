package testservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func TestService_GetTestExecution(t *testing.T) {
	tests := []struct {
		name  string
		input *test.Payload
	}{
		{
			name:  "without input",
			input: nil,
		},
		{
			name:  "without input",
			input: fake.GenInput(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testExec := fake.GenTestExec(uuid.New())
			testExec.HasInput = tt.input != nil

			r := &RepositoryMock{
				GetTestExecutionFunc: func(ctx context.Context, id test.TestExecutionID) (*test.TestExecution, error) {
					assert.Equal(t, testExec.ID, id)
					return testExec, nil
				},
			}

			if tt.input != nil {
				r.GetTestExecutionInputFunc = func(ctx context.Context, id test.TestExecutionID) (*test.Payload, error) {
					assert.Equal(t, testExec.ID, id)
					return tt.input, nil
				}
			}

			s := Service{repo: r}

			req := &testsv1.GetTestExecutionRequest{
				Context:         "foo",
				TestExecutionId: testExec.ID.String(),
			}

			res, err := s.GetTestExecution(context.Background(), connect.NewRequest(req))
			require.NoError(t, err)

			assert.Equal(t, testExec.Proto(), res.Msg.TestExecution)

			if tt.input == nil {
				assert.Nil(t, res.Msg.Input)
			} else {
				assert.Equal(t, tt.input.Proto(), res.Msg.Input)
			}
		})
	}
}

func TestService_ListTestExecutions(t *testing.T) {
	pageSize := 2
	testID := uuid.New()

	wantPage1 := test.TestExecutionList{fake.GenTestExec(testID), fake.GenTestExec(testID)}
	wantPage2 := test.TestExecutionList{fake.GenTestExec(testID)}

	r := new(RepositoryMock)
	r.ListTestExecutionsFunc = func(ctx context.Context, gotTestID uuid.V7, filter test.PageFilter[test.TestExecutionID]) (test.TestExecutionList, error) {
		assert.Equal(t, testID, gotTestID)
		assert.Equal(t, pageSize, filter.Size)

		switch len(r.ListTestExecutionsCalls()) {
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

	req := &testsv1.ListTestExecutionsRequest{
		Context:  "foo",
		TestId:   testID.String(),
		PageSize: int32(pageSize),
	}
	res, err := s.ListTestExecutions(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage1.Proto(), res.Msg.TestExecutions)
	assert.NotEmpty(t, res.Msg.NextPageToken)

	req.NextPageToken = res.Msg.NextPageToken
	res, err = s.ListTestExecutions(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage2.Proto(), res.Msg.TestExecutions)
	assert.Empty(t, res.Msg.NextPageToken)
}

func TestService_AckTestExecutionStarted(t *testing.T) {
	wantTestExec := &test.TestExecution{
		ID:           test.NewTestExecutionID(),
		TestID:       uuid.New(),
		HasInput:     true,
		ScheduleTime: time.Now().UTC(),
		StartTime:    ptr.Get(time.Now().UTC()),
	}

	r := &RepositoryMock{
		UpdateTestExecutionStartedFunc: func(ctx context.Context, started *test.StartedTestExecution) (*test.TestExecution, error) {
			assert.Equal(t, wantTestExec.ID, started.ID)
			assert.Equal(t, *wantTestExec.StartTime, started.StartTime)
			return wantTestExec, nil
		},
	}

	p := &PublisherMock{
		PublishFunc: func(testExecID string, e *eventsv1.Event) error {
			assert.Equal(t, wantTestExec.ID.String(), testExecID)
			assert.Equal(t, wantTestExec.ID.String(), e.TestExecutionId)
			assert.Equal(t, eventsv1.Event_TYPE_TEST_EXECUTION_STARTED, e.Type)
			assert.NotEmpty(t, e.EventId)
			assert.True(t, e.CreateTime.IsValid())
			assert.Equal(t, eventsv1.Event_Data_TYPE_TEST_EXECUTION, e.Data.Type)
			assert.Equal(t, wantTestExec.Proto(), e.Data.GetTestExecution())
			return nil
		},
	}

	s := Service{repo: r, eventPub: p}

	req := &testsv1.AckTestExecutionStartedRequest{
		Context:         "foo",
		TestExecutionId: wantTestExec.ID.String(),
		StartTime:       timestamppb.New(*wantTestExec.StartTime),
	}

	res, err := s.AckTestExecutionStarted(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestService_AckTestExecutionFinished(t *testing.T) {
	wantTestExec := &test.TestExecution{
		ID:           test.NewTestExecutionID(),
		TestID:       uuid.New(),
		HasInput:     true,
		ScheduleTime: time.Now().UTC(),
		StartTime:    ptr.Get(time.Now().UTC()),
		FinishTime:   ptr.Get(time.Now().UTC()),
		Error:        ptr.Get("bang"),
	}

	r := &RepositoryMock{
		UpdateTestExecutionFinishedFunc: func(ctx context.Context, finished *test.FinishedTestExecution) (*test.TestExecution, error) {
			assert.Equal(t, wantTestExec.ID, finished.ID)
			assert.Equal(t, *wantTestExec.FinishTime, finished.FinishTime)
			assert.Equal(t, wantTestExec.Error, finished.Error)
			wantTestExec.FinishTime = &finished.FinishTime
			wantTestExec.Error = finished.Error
			return wantTestExec, nil
		},
	}

	p := &PublisherMock{
		PublishFunc: func(testExecID string, e *eventsv1.Event) error {
			assert.Equal(t, wantTestExec.ID.String(), testExecID)
			assert.Equal(t, wantTestExec.ID.String(), e.TestExecutionId)
			assert.Equal(t, eventsv1.Event_TYPE_TEST_EXECUTION_FINISHED, e.Type)
			assert.NotEmpty(t, e.EventId)
			assert.True(t, e.CreateTime.IsValid())
			assert.Equal(t, eventsv1.Event_Data_TYPE_TEST_EXECUTION, e.Data.Type)
			assert.Equal(t, wantTestExec.Proto(), e.Data.GetTestExecution())
			return nil
		},
	}

	s := Service{repo: r, eventPub: p}

	req := &testsv1.AckTestExecutionFinishedRequest{
		Context:         "foo",
		TestExecutionId: wantTestExec.ID.String(),
		FinishTime:      timestamppb.New(*wantTestExec.FinishTime),
		Error:           wantTestExec.Error,
	}

	res, err := s.AckTestExecutionFinished(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestService_RetryTestExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	testExec := fake.GenTestExec(uuid.New())
	testExecLog := fake.GenTestExecLog(testExec.ID)
	wantResetTestExec := fake.GenResetTestExec(testExec)

	successCaseExec := fake.GenCaseExec(testExec.ID)
	failureCaseExec := fake.GenCaseExec(testExec.ID)
	failureCaseExec.Error = ptr.Get("case error: bang")
	successCaseLogs := fake.GenCaseExecLogs(testExec.ID, successCaseExec.ID, 10)
	failureCaseLogs := fake.GenCaseExecLogs(testExec.ID, failureCaseExec.ID, 10)

	failureCaseExecLogIDs := make([]uuid.V7, len(failureCaseLogs))
	for i, log := range failureCaseLogs {
		failureCaseExecLogIDs[i] = log.ID
	}

	his := fake.GenCaseFailureHistory(testExec.ID, testExecLog.ID, successCaseExec.ID, failureCaseExec.ID)
	hisEventResetID := int64(11) // fake history sets last successful case at event #11

	r := &RepositoryMock{
		GetTestExecutionFunc: func(ctx context.Context, id test.TestExecutionID) (*test.TestExecution, error) {
			assert.Equal(t, testExec.ID, id)
			return testExec, nil
		},
		ListCaseExecutionsFunc: func(ctx context.Context, testExecID test.TestExecutionID, filter test.PageFilter[test.CaseExecutionID]) (test.CaseExecutionList, error) {
			assert.Equal(t, testExec.ID, testExecID)
			return test.CaseExecutionList{successCaseExec, failureCaseExec}, nil
		},
		ListLogsFunc: func(ctx context.Context, testExecID test.TestExecutionID, filter test.PageFilter[uuid.V7]) (test.LogList, error) {
			assert.Equal(t, testExec.ID, testExecID)
			return append(successCaseLogs, failureCaseLogs...), nil
		},
		DeleteCaseExecutionFunc: func(ctx context.Context, testExecID test.TestExecutionID, id test.CaseExecutionID) error {
			assert.Equal(t, testExec.ID, testExecID)
			assert.Equal(t, failureCaseExec.ID, id)
			return nil
		},
		DeleteLogFunc: func(ctx context.Context, id uuid.V7) error {
			assert.Contains(t, failureCaseExecLogIDs, id)
			return nil
		},
		ResetTestExecutionFunc: func(ctx context.Context, testExecID test.TestExecutionID, resetTime time.Time) (*test.TestExecution, error) {
			assert.Equal(t, testExec.ID, testExecID)
			assert.False(t, resetTime.IsZero())
			return wantResetTestExec, nil
		},
	}
	r.ExecuteTxFunc = func(ctx context.Context, query func(repo test.Repository) error) error {
		return query(r)
	}

	w := &WorkflowerMock{
		GetWorkflowHistoryFunc: func(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType enums.HistoryEventFilterType) client.HistoryEventIterator {
			assert.Equal(t, testExec.ID.WorkflowID(), workflowID)
			assert.Equal(t, runID, "")
			assert.Equal(t, isLongPoll, false)
			assert.Equal(t, filterType, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
			return fake.NewHistoryEventIterator(his)
		},
		ResetWorkflowExecutionFunc: func(ctx context.Context, req *workflowservice.ResetWorkflowExecutionRequest) (*workflowservice.ResetWorkflowExecutionResponse, error) {
			assert.NotEmpty(t, req.Namespace)
			assert.Equal(t, testExec.ID.WorkflowID(), req.WorkflowExecution.WorkflowId)
			assert.Equal(t, retryReason, req.Reason)
			assert.Equal(t, hisEventResetID, req.WorkflowTaskFinishEventId)
			return nil, nil
		},
	}

	// Setup logs

	svc := New(r, fake.NewPubSub(), w)

	// Retry test execution
	req := &testsv1.RetryTestExecutionRequest{
		TestExecutionId: testExec.ID.String(),
	}
	res, err := svc.RetryTestExecution(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	// Assert test execution record reset
	gotTestExec := res.Msg.TestExecution

	assert.Equal(t, wantResetTestExec.ID.String(), gotTestExec.Id)
	assert.Equal(t, wantResetTestExec.TestID.String(), gotTestExec.TestId)
	assert.Equal(t, gotTestExec.ScheduleTime.AsTime(), wantResetTestExec.ScheduleTime)
	assert.Nil(t, gotTestExec.StartTime)
	assert.Nil(t, gotTestExec.FinishTime)
	assert.Nil(t, gotTestExec.Error)
}
