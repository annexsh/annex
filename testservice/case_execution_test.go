package testservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/uuid"
)

func TestService_AckCaseExecutionScheduled(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	te, err := fakes.repo.CreateScheduledTestExecution(ctx, fake.GenScheduledTestExec(tt.ID))
	require.NoError(t, err)

	caseID := fake.GenCaseID()
	req := &testsv1.AckCaseExecutionScheduledRequest{
		TestExecutionId: te.ID.String(),
		CaseExecutionId: caseID.Int32(),
		CaseName:        uuid.NewString(),
		ScheduleTime:    timestamppb.New(time.Now().UTC()),
	}
	res, err := s.AckCaseExecutionScheduled(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetCaseExecution(ctx, te.ID, caseID)
	require.NoError(t, err)
	assert.Equal(t, req.ScheduleTime.AsTime(), ackd.ScheduleTime)
}

func TestService_AckCaseExecutionStarted(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	testExec := fake.GenScheduledTestExec(tt.ID)
	te, err := fakes.repo.CreateScheduledTestExecution(ctx, testExec)
	require.NoError(t, err)

	scheduled := fake.GenScheduledCaseExec(te.ID)
	caseExec, err := fakes.repo.CreateScheduledCaseExecution(ctx, scheduled)
	require.NoError(t, err)

	req := &testsv1.AckCaseExecutionStartedRequest{
		TestExecutionId: caseExec.TestExecutionID.String(),
		CaseExecutionId: caseExec.ID.Int32(),
		StartTime:       timestamppb.New(time.Now().UTC()),
	}
	res, err := s.AckCaseExecutionStarted(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetCaseExecution(ctx, te.ID, caseExec.ID)
	require.NoError(t, err)
	assert.Equal(t, req.StartTime.AsTime(), *ackd.StartTime)
}

func TestService_AckCaseExecutionFinished(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	testExec := fake.GenScheduledTestExec(tt.ID)
	te, err := fakes.repo.CreateScheduledTestExecution(ctx, testExec)
	require.NoError(t, err)

	scheduled := fake.GenScheduledCaseExec(te.ID)
	caseExec, err := fakes.repo.CreateScheduledCaseExecution(ctx, scheduled)
	require.NoError(t, err)

	req := &testsv1.AckCaseExecutionFinishedRequest{
		TestExecutionId: caseExec.TestExecutionID.String(),
		CaseExecutionId: caseExec.ID.Int32(),
		FinishTime:      timestamppb.New(time.Now().UTC()),
		Error:           ptr.Get("bang"),
	}
	res, err := s.AckCaseExecutionFinished(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetCaseExecution(ctx, te.ID, caseExec.ID)
	require.NoError(t, err)
	assert.Equal(t, req.FinishTime.AsTime(), *ackd.FinishTime)
	assert.Equal(t, req.Error, ackd.Error)
}

func TestService_ListTestCaseExecutions(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	te, err := fakes.repo.CreateScheduledTestExecution(ctx, fake.GenScheduledTestExec(tt.ID))
	require.NoError(t, err)

	wantCount := 30
	want := make([]*testsv1.CaseExecution, wantCount)

	for i := range wantCount {
		scheduled := fake.GenScheduledCaseExec(te.ID)
		ce, err := fakes.repo.CreateScheduledCaseExecution(ctx, scheduled)
		require.NoError(t, err)
		want[i] = ce.Proto()
	}

	req := &testsv1.ListCaseExecutionsRequest{
		TestExecutionId: te.ID.String(),
	}
	res, err := s.ListCaseExecutions(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	got := res.Msg.CaseExecutions
	assert.Len(t, got, wantCount)
	assert.Equal(t, want, got)
}
