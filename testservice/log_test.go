package testservice

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
)

func TestService_PublishTestExecutionLog(t *testing.T) {
	tests := []struct {
		name      string
		isCaseLog bool
	}{
		{
			name:      "publish test execution log",
			isCaseLog: false,
		},
		{
			name:      "publish case execution log",
			isCaseLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s, fakes := newService()

			created, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
			require.NoError(t, err)

			te, err := fakes.repo.CreateScheduledTestExecution(ctx, fake.GenScheduledTestExec(created.ID))
			require.NoError(t, err)

			var reqCaseExecID *int32
			var wantCaseExecID *test.CaseExecutionID
			if tt.isCaseLog {
				ce, err := fakes.repo.CreateScheduledCaseExecution(ctx, fake.GenScheduledCaseExec(te.ID))
				require.NoError(t, err)
				wantCaseExecID = &ce.ID
				reqCaseExecID = ptr.Get(wantCaseExecID.Int32())
			}

			req := &testsv1.PublishTestExecutionLogRequest{
				TestExecutionId: te.ID.String(),
				CaseExecutionId: reqCaseExecID,
				Level:           "INFO",
				Message:         "lorem ipsum",
				CreateTime:      timestamppb.Now(),
			}
			res, err := s.PublishTestExecutionLog(ctx, connect.NewRequest(req))
			require.NoError(t, err)
			assert.NotNil(t, res)

			execLogID, err := uuid.Parse(res.Msg.LogId)
			require.NoError(t, err)

			got, err := fakes.repo.GetLog(ctx, execLogID)
			require.NoError(t, err)

			want := &test.Log{
				ID:              execLogID,
				TestExecutionID: te.ID,
				CaseExecutionID: wantCaseExecID,
				Level:           req.Level,
				Message:         req.Message,
				CreateTime:      req.CreateTime.AsTime(),
			}

			assert.Equal(t, got, want)
		})
	}
}

func TestService_ListTestExecutionLogs(t *testing.T) {
	wantNumTestLogs := 15
	wantNumCaseLogs := 15

	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	te, err := fakes.repo.CreateScheduledTestExecution(ctx, fake.GenScheduledTestExec(tt.ID))
	require.NoError(t, err)

	ce, err := fakes.repo.CreateScheduledCaseExecution(ctx, fake.GenScheduledCaseExec(te.ID))
	require.NoError(t, err)

	var want []*testsv1.Log

	for range wantNumTestLogs {
		l := fake.GenTestExecLog(te.ID)
		err = fakes.repo.CreateLog(ctx, l)
		require.NoError(t, err)
		want = append(want, l.Proto())
	}

	for range wantNumCaseLogs {
		l := fake.GenCaseExecLog(te.ID, ce.ID)
		err = fakes.repo.CreateLog(ctx, l)
		require.NoError(t, err)
		want = append(want, l.Proto())
	}

	req := &testsv1.ListTestExecutionLogsRequest{
		TestExecutionId: te.ID.String(),
	}
	res, err := s.ListTestExecutionLogs(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	got := res.Msg.Logs
	assert.Len(t, got, wantNumTestLogs+wantNumCaseLogs)
	assert.Equal(t, want, got)
}
