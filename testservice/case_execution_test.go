package testservice

import (
	"context"
	"testing"
	"time"

	testservicev1 "github.com/annexhq/annex-proto/gen/go/rpc/testservice/v1"
	testv1 "github.com/annexhq/annex-proto/gen/go/type/test/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexhq/annex/internal/fake"
	"github.com/annexhq/annex/internal/ptr"
)

func TestService_AckCaseExecutionScheduled(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	te, err := fakes.repo.CreateScheduledTestExecution(ctx, fake.GenScheduledTestExec(tt.ID))
	require.NoError(t, err)

	caseID := fake.GenCaseID()
	req := &testservicev1.AckCaseExecutionScheduledRequest{
		Id:          caseID.Int32(),
		TestExecId:  te.ID.String(),
		CaseName:    uuid.NewString(),
		ScheduledAt: timestamppb.New(time.Now().UTC()),
	}
	res, err := s.AckCaseExecutionScheduled(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetCaseExecution(ctx, te.ID, caseID)
	require.NoError(t, err)
	assert.Equal(t, req.ScheduledAt.AsTime(), ackd.ScheduleTime)
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

	req := &testservicev1.AckCaseExecutionStartedRequest{
		Id:         caseExec.ID.Int32(),
		TestExecId: caseExec.TestExecID.String(),
		StartedAt:  timestamppb.New(time.Now().UTC()),
	}
	res, err := s.AckCaseExecutionStarted(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetCaseExecution(ctx, te.ID, caseExec.ID)
	require.NoError(t, err)
	assert.Equal(t, req.StartedAt.AsTime(), *ackd.StartTime)
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

	req := &testservicev1.AckCaseExecutionFinishedRequest{
		Id:         caseExec.ID.Int32(),
		TestExecId: caseExec.TestExecID.String(),
		FinishedAt: timestamppb.New(time.Now().UTC()),
		Error:      ptr.Get("bang"),
	}
	res, err := s.AckCaseExecutionFinished(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetCaseExecution(ctx, te.ID, caseExec.ID)
	require.NoError(t, err)
	assert.Equal(t, req.FinishedAt.AsTime(), *ackd.FinishTime)
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
	want := make([]*testv1.CaseExecution, wantCount)

	for i := range wantCount {
		scheduled := fake.GenScheduledCaseExec(te.ID)
		ce, err := fakes.repo.CreateScheduledCaseExecution(ctx, scheduled)
		require.NoError(t, err)
		want[i] = ce.Proto()
	}

	res, err := s.ListTestCaseExecutions(ctx, &testservicev1.ListTestCaseExecutionsRequest{
		TestExecId: te.ID.String(),
	})
	require.NoError(t, err)

	got := res.CaseExecutions
	assert.Len(t, got, wantCount)
	assert.Equal(t, want, got)
}
