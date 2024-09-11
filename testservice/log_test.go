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
