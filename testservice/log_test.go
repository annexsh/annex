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

func TestService_PublishTestExecutionLog(t *testing.T) {
	ctx := context.Background()

	wantLog := &test.Log{
		ID:              uuid.V7{}, // injected by create log mock
		TestExecutionID: test.NewTestExecutionID(),
		CaseExecutionID: ptr.Get(fake.GenCaseID()),
		Level:           "INFO",
		Message:         "foo bar",
		CreateTime:      time.Now().UTC(),
	}

	r := &RepositoryMock{
		CreateLogFunc: func(ctx context.Context, log *test.Log) error {
			assert.False(t, log.ID.Empty())
			wantLog.ID = log.ID
			assert.Equal(t, wantLog, log)
			return nil
		},
	}
	r.ExecuteTxFunc = func(ctx context.Context, query func(repo test.Repository) error) error {
		return query(r)
	}

	p := &PublisherMock{
		PublishFunc: func(testExecID string, e *eventsv1.Event) error {
			assert.Equal(t, wantLog.TestExecutionID.String(), testExecID)
			assert.Equal(t, wantLog.TestExecutionID.String(), e.TestExecutionId)
			assert.Equal(t, eventsv1.Event_TYPE_LOG_PUBLISHED, e.Type)
			assert.NotEmpty(t, e.EventId)
			assert.True(t, e.CreateTime.IsValid())
			assert.Equal(t, eventsv1.Event_Data_TYPE_LOG, e.Data.Type)
			assert.Equal(t, wantLog.Proto(), e.Data.GetLog())
			return nil
		},
	}

	s := Service{repo: r, eventPub: p}

	req := &testsv1.PublishLogRequest{
		Context:         "foo",
		TestExecutionId: wantLog.TestExecutionID.String(),
		CaseExecutionId: ptr.Get(wantLog.CaseExecutionID.Int32()),
		Level:           wantLog.Level,
		Message:         wantLog.Message,
		CreateTime:      timestamppb.New(wantLog.CreateTime),
	}
	res, err := s.PublishLog(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestService_PublishLog_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.PublishLogRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.PublishLogRequest{
				Context:         "",
				TestExecutionId: uuid.NewString(),
				Level:           "INFO",
				Message:         "bar",
				CreateTime:      timestamppb.Now(),
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test execution id",
			req: &testsv1.PublishLogRequest{
				Context:         "foo",
				TestExecutionId: "",
				Level:           "INFO",
				Message:         "bar",
				CreateTime:      timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id can't be blank",
			},
		},
		{
			name: "test execution id not a uuid",
			req: &testsv1.PublishLogRequest{
				Context:         "foo",
				TestExecutionId: "bar",
				Level:           "INFO",
				Message:         "baz",
				CreateTime:      timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "test_execution_id",
				Description: "Test execution id must be a v7 UUID",
			},
		},
		{
			name: "negative case execution id",
			req: &testsv1.PublishLogRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				CaseExecutionId: ptr.Get(int32(-1)),
				Level:           "INFO",
				Message:         "baz",
				CreateTime:      timestamppb.Now(),
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "case_execution_id",
				Description: `Case execution id must be greater than "0"`,
			},
		},
		{
			name: "nil create time",
			req: &testsv1.PublishLogRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				Level:           "INFO",
				Message:         "baz",
				CreateTime:      nil,
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "create_time",
				Description: "Create time must not be nil",
			},
		},
		{
			name: "invalid create time",
			req: &testsv1.PublishLogRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				Level:           "INFO",
				Message:         "baz",
				CreateTime:      &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "create_time",
				Description: "Create time must be a valid timestamp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.PublishLog(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_ListTestExecutionLogs(t *testing.T) {
	pageSize := 2
	testExecID := test.NewTestExecutionID()

	wantPage1 := fake.GenTestExecLogs(testExecID, 2)
	wantPage2 := fake.GenTestExecLogs(testExecID, 1)

	r := new(RepositoryMock)
	r.ListLogsFunc = func(ctx context.Context, gotTestExecID test.TestExecutionID, filter test.PageFilter[uuid.V7]) (test.LogList, error) {
		assert.Equal(t, testExecID, gotTestExecID)
		assert.Equal(t, pageSize, filter.Size)

		switch len(r.ListLogsCalls()) {
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

	req := &testsv1.ListTestExecutionLogsRequest{
		Context:         "foo",
		TestExecutionId: testExecID.String(),
		PageSize:        int32(pageSize),
	}
	res, err := s.ListTestExecutionLogs(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage1.Proto(), res.Msg.Logs)
	assert.NotEmpty(t, res.Msg.NextPageToken)

	req.NextPageToken = res.Msg.NextPageToken
	res, err = s.ListTestExecutionLogs(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage2.Proto(), res.Msg.Logs)
	assert.Empty(t, res.Msg.NextPageToken)
}

func TestService_ListTestExecutionLogs_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.ListTestExecutionLogsRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.ListTestExecutionLogsRequest{
				Context:         "",
				TestExecutionId: uuid.NewString(),
				PageSize:        1,
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank test execution id",
			req: &testsv1.ListTestExecutionLogsRequest{
				Context:         "foo",
				TestExecutionId: "",
				PageSize:        1,
			},
			wantFieldViolation: wantBlankTestExecIDFieldViolation(),
		},
		{
			name: "test execution id not a uuid",
			req: &testsv1.ListTestExecutionLogsRequest{
				Context:         "foo",
				TestExecutionId: "bar",
				PageSize:        1,
			},
			wantFieldViolation: wantTestExecIDNotUUIDFieldViolation(),
		},
		{
			name: "page size less than 0",
			req: &testsv1.ListTestExecutionLogsRequest{
				Context:         "foo",
				TestExecutionId: uuid.NewString(),
				PageSize:        int32(-1),
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
		{
			name: "page size greater than max",
			req: &testsv1.ListTestExecutionLogsRequest{
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
			res, err := s.ListTestExecutionLogs(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}
