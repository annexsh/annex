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
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func TestService_ListCaseExecutions(t *testing.T) {
	pageSize := 2
	testExecID := test.NewTestExecutionID()

	wantPage1 := test.CaseExecutionList{fake.GenCaseExec(testExecID), fake.GenCaseExec(testExecID)}
	wantPage2 := test.CaseExecutionList{fake.GenCaseExec(testExecID)}

	r := new(RepositoryMock)
	r.ListCaseExecutionsFunc = func(ctx context.Context, gotTestExecID test.TestExecutionID, filter test.PageFilter[test.CaseExecutionID]) (test.CaseExecutionList, error) {
		assert.Equal(t, testExecID, gotTestExecID)
		assert.Equal(t, pageSize, filter.Size)

		switch len(r.ListCaseExecutionsCalls()) {
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

	req := &testsv1.ListCaseExecutionsRequest{
		Context:         "foo",
		TestExecutionId: testExecID.String(),
		PageSize:        int32(pageSize),
	}
	res, err := s.ListCaseExecutions(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage1.Proto(), res.Msg.CaseExecutions)
	assert.NotEmpty(t, res.Msg.NextPageToken)

	req.NextPageToken = res.Msg.NextPageToken
	res, err = s.ListCaseExecutions(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage2.Proto(), res.Msg.CaseExecutions)
	assert.Empty(t, res.Msg.NextPageToken)
}

func TestService_ListCaseExecutions_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.ListCaseExecutionsRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.ListCaseExecutionsRequest{
				Context:         "",
				TestExecutionId: uuid.NewString(),
				PageSize:        1,
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test execution id",
			req: &testsv1.ListCaseExecutionsRequest{
				Context:         "foo",
				TestExecutionId: "",
				PageSize:        1,
			},
			wantFieldViolation: wantBlankTestExecIDFieldViolation(),
		},
		{
			name: "test execution id not a uuid",
			req: &testsv1.ListCaseExecutionsRequest{
				Context:         "foo",
				TestExecutionId: "bar",
				PageSize:        1,
			},
			wantFieldViolation: wantTestExecIDNotUUIDFieldViolation(),
		},
		{
			name: "page size less than 0",
			req: &testsv1.ListCaseExecutionsRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				PageSize:        int32(-1),
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
		{
			name: "page size greater than max",
			req: &testsv1.ListCaseExecutionsRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				PageSize:        maxPageSize + 1,
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.ListCaseExecutions(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_AckCaseExecutionScheduled(t *testing.T) {
	wantCaseExec := &test.CaseExecution{
		ID:              fake.GenCaseID(),
		TestExecutionID: test.NewTestExecutionID(),
		CaseName:        "foo",
		ScheduleTime:    time.Now().UTC(),
	}

	r := &RepositoryMock{
		CreateCaseExecutionScheduledFunc: func(ctx context.Context, scheduled *test.ScheduledCaseExecution) (*test.CaseExecution, error) {
			assert.Equal(t, wantCaseExec.ID, scheduled.ID)
			assert.Equal(t, wantCaseExec.TestExecutionID, scheduled.TestExecutionID)
			assert.Equal(t, wantCaseExec.CaseName, scheduled.CaseName)
			assert.Equal(t, wantCaseExec.ScheduleTime, scheduled.ScheduleTime)
			return wantCaseExec, nil
		},
	}

	p := &PublisherMock{
		PublishFunc: func(testExecID string, e *eventsv1.Event) error {
			assert.Equal(t, wantCaseExec.TestExecutionID.String(), testExecID)
			assert.Equal(t, wantCaseExec.TestExecutionID.String(), e.TestExecutionId)
			assert.Equal(t, eventsv1.Event_TYPE_CASE_EXECUTION_SCHEDULED, e.Type)
			assert.NotEmpty(t, e.EventId)
			assert.True(t, e.CreateTime.IsValid())
			assert.Equal(t, eventsv1.Event_Data_TYPE_CASE_EXECUTION, e.Data.Type)
			assert.Equal(t, wantCaseExec.Proto(), e.Data.GetCaseExecution())
			return nil
		},
	}

	s := Service{repo: r, eventPub: p}

	req := &testsv1.AckCaseExecutionScheduledRequest{
		Context:         "foo",
		TestExecutionId: wantCaseExec.TestExecutionID.String(),
		CaseExecutionId: wantCaseExec.ID.Int32(),
		CaseName:        wantCaseExec.CaseName,
		ScheduleTime:    timestamppb.New(wantCaseExec.ScheduleTime),
	}

	res, err := s.AckCaseExecutionScheduled(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestService_AckCaseExecutionScheduled_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.AckCaseExecutionScheduledRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.AckCaseExecutionScheduledRequest{
				Context:         "",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				CaseName:        "foo",
				ScheduleTime:    timestamppb.Now(),
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test execution id",
			req: &testsv1.AckCaseExecutionScheduledRequest{
				Context:         "foo",
				TestExecutionId: "",
				CaseExecutionId: 1,
				CaseName:        "bar",
				ScheduleTime:    timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id can't be blank",
			},
		},
		{
			name: "test execution id not a uuid",
			req: &testsv1.AckCaseExecutionScheduledRequest{
				Context:         "foo",
				TestExecutionId: "bar",
				CaseExecutionId: 1,
				CaseName:        "baz",
				ScheduleTime:    timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id must be a v7 UUID",
			},
		},
		{
			name: "negative case execution id",
			req: &testsv1.AckCaseExecutionScheduledRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: -1,
				CaseName:        "bar",
				ScheduleTime:    timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "case_execution_id",
				Description: `Case execution id must be greater than "0"`,
			},
		},
		{
			name: "blank case name",
			req: &testsv1.AckCaseExecutionScheduledRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				CaseName:        "",
				ScheduleTime:    timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "case_name",
				Description: "Case name can't be blank",
			},
		},
		{
			name: "nil schedule time",
			req: &testsv1.AckCaseExecutionScheduledRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				CaseName:        "bar",
				ScheduleTime:    nil,
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "schedule_time",
				Description: "Schedule time must not be nil",
			},
		},
		{
			name: "invalid start time",
			req: &testsv1.AckCaseExecutionScheduledRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				CaseName:        "bar",
				ScheduleTime:    &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "schedule_time",
				Description: "Schedule time must be a valid timestamp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.AckCaseExecutionScheduled(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_AckCaseExecutionStarted(t *testing.T) {
	wantCaseExec := &test.CaseExecution{
		ID:              fake.GenCaseID(),
		TestExecutionID: test.NewTestExecutionID(),
		CaseName:        "foo",
		ScheduleTime:    time.Now().UTC(),
		StartTime:       ptr.Get(time.Now().UTC()),
	}

	r := &RepositoryMock{
		UpdateCaseExecutionStartedFunc: func(ctx context.Context, started *test.StartedCaseExecution) (*test.CaseExecution, error) {
			assert.Equal(t, wantCaseExec.ID, started.ID)
			assert.Equal(t, wantCaseExec.TestExecutionID, started.TestExecutionID)
			assert.Equal(t, *wantCaseExec.StartTime, started.StartTime)
			return wantCaseExec, nil
		},
	}

	p := &PublisherMock{
		PublishFunc: func(testExecID string, e *eventsv1.Event) error {
			assert.Equal(t, wantCaseExec.TestExecutionID.String(), testExecID)
			assert.Equal(t, wantCaseExec.TestExecutionID.String(), e.TestExecutionId)
			assert.Equal(t, eventsv1.Event_TYPE_CASE_EXECUTION_STARTED, e.Type)
			assert.NotEmpty(t, e.EventId)
			assert.True(t, e.CreateTime.IsValid())
			assert.Equal(t, eventsv1.Event_Data_TYPE_CASE_EXECUTION, e.Data.Type)
			assert.Equal(t, wantCaseExec.Proto(), e.Data.GetCaseExecution())
			return nil
		},
	}

	s := Service{repo: r, eventPub: p}

	req := &testsv1.AckCaseExecutionStartedRequest{
		Context:         "foo",
		TestExecutionId: wantCaseExec.TestExecutionID.String(),
		CaseExecutionId: wantCaseExec.ID.Int32(),
		StartTime:       timestamppb.New(*wantCaseExec.StartTime),
	}

	res, err := s.AckCaseExecutionStarted(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestService_AckCaseExecutionStarted_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.AckCaseExecutionStartedRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.AckCaseExecutionStartedRequest{
				Context:         "",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				StartTime:       timestamppb.Now(),
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test execution id",
			req: &testsv1.AckCaseExecutionStartedRequest{
				Context:         "foo",
				TestExecutionId: "",
				CaseExecutionId: 1,
				StartTime:       timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id can't be blank",
			},
		},
		{
			name: "test execution id not a uuid",
			req: &testsv1.AckCaseExecutionStartedRequest{
				Context:         "foo",
				TestExecutionId: "bar",
				CaseExecutionId: 1,
				StartTime:       timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id must be a v7 UUID",
			},
		},
		{
			name: "negative case execution id",
			req: &testsv1.AckCaseExecutionStartedRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: -1,
				StartTime:       timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "case_execution_id",
				Description: `Case execution id must be greater than "0"`,
			},
		},
		{
			name: "nil start time",
			req: &testsv1.AckCaseExecutionStartedRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				StartTime:       nil,
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "start_time",
				Description: "Start time must not be nil",
			},
		},
		{
			name: "invalid start time",
			req: &testsv1.AckCaseExecutionStartedRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				StartTime:       &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "start_time",
				Description: "Start time must be a valid timestamp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.AckCaseExecutionStarted(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_AckCaseExecutionFinished(t *testing.T) {
	wantCaseExec := &test.CaseExecution{
		ID:              fake.GenCaseID(),
		TestExecutionID: test.NewTestExecutionID(),
		CaseName:        "foo",
		ScheduleTime:    time.Now().UTC(),
		StartTime:       ptr.Get(time.Now().UTC()),
		FinishTime:      ptr.Get(time.Now().UTC()),
		Error:           ptr.Get("bang"),
	}

	r := &RepositoryMock{
		UpdateCaseExecutionStartedFunc: func(ctx context.Context, started *test.StartedCaseExecution) (*test.CaseExecution, error) {
			assert.Equal(t, wantCaseExec.ID, started.ID)
			assert.Equal(t, wantCaseExec.TestExecutionID, started.TestExecutionID)
			assert.Equal(t, *wantCaseExec.StartTime, started.StartTime)
			return wantCaseExec, nil
		},
		UpdateCaseExecutionFinishedFunc: func(ctx context.Context, finished *test.FinishedCaseExecution) (*test.CaseExecution, error) {
			assert.Equal(t, wantCaseExec.ID, finished.ID)
			assert.Equal(t, wantCaseExec.TestExecutionID, finished.TestExecutionID)
			assert.Equal(t, *wantCaseExec.FinishTime, finished.FinishTime)
			assert.Equal(t, wantCaseExec.Error, finished.Error)
			return wantCaseExec, nil
		},
	}

	p := &PublisherMock{
		PublishFunc: func(testExecID string, e *eventsv1.Event) error {
			assert.Equal(t, wantCaseExec.TestExecutionID.String(), testExecID)
			assert.Equal(t, wantCaseExec.TestExecutionID.String(), e.TestExecutionId)
			assert.Equal(t, eventsv1.Event_TYPE_CASE_EXECUTION_FINISHED, e.Type)
			assert.NotEmpty(t, e.EventId)
			assert.True(t, e.CreateTime.IsValid())
			assert.Equal(t, eventsv1.Event_Data_TYPE_CASE_EXECUTION, e.Data.Type)
			assert.Equal(t, wantCaseExec.Proto(), e.Data.GetCaseExecution())
			return nil
		},
	}

	s := Service{repo: r, eventPub: p}

	req := &testsv1.AckCaseExecutionFinishedRequest{
		Context:         "foo",
		TestExecutionId: wantCaseExec.TestExecutionID.String(),
		CaseExecutionId: wantCaseExec.ID.Int32(),
		FinishTime:      timestamppb.New(*wantCaseExec.FinishTime),
		Error:           wantCaseExec.Error,
	}

	res, err := s.AckCaseExecutionFinished(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestService_AckCaseExecutionFinished_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.AckCaseExecutionFinishedRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.AckCaseExecutionFinishedRequest{
				Context:         "",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				FinishTime:      timestamppb.Now(),
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test execution id",
			req: &testsv1.AckCaseExecutionFinishedRequest{
				Context:         "foo",
				TestExecutionId: "",
				CaseExecutionId: 1,
				FinishTime:      timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id can't be blank",
			},
		},
		{
			name: "test execution id not a uuid",
			req: &testsv1.AckCaseExecutionFinishedRequest{
				Context:         "foo",
				TestExecutionId: "bar",
				CaseExecutionId: 1,
				FinishTime:      timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id must be a v7 UUID",
			},
		},
		{
			name: "negative case execution id",
			req: &testsv1.AckCaseExecutionFinishedRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: -1,
				FinishTime:      timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "case_execution_id",
				Description: `Case execution id must be greater than "0"`,
			},
		},
		{
			name: "nil finish time",
			req: &testsv1.AckCaseExecutionFinishedRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				FinishTime:      nil,
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "finish_time",
				Description: "Finish time must not be nil",
			},
		},
		{
			name: "blank error",
			req: &testsv1.AckCaseExecutionFinishedRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: 1,
				FinishTime:      timestamppb.Now(),
				Error:           ptr.Get(""),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "error",
				Description: "Error can't be blank",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.AckCaseExecutionFinished(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}
