package testservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/inmem"
	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
)

func TestService_GetTestExecution(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	scheduled := fake.GenScheduledTestExec(tt.ID)
	te, err := fakes.repo.CreateScheduledTestExecution(ctx, scheduled)
	require.NoError(t, err)

	req := &testsv1.GetTestExecutionRequest{
		TestExecutionId: te.ID.String(),
	}
	res, err := s.GetTestExecution(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	assert.Equal(t, te.Proto(), res.Msg.TestExecution)
	assert.Equal(t, scheduled.Payload, res.Msg.Input.Data)
}

func TestService_ListTestExecutions(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	wantCount := 30
	want := make([]*testsv1.TestExecution, wantCount)

	for i := range wantCount {
		scheduled := fake.GenScheduledTestExec(tt.ID)
		te, err := fakes.repo.CreateScheduledTestExecution(ctx, scheduled)
		require.NoError(t, err)
		want[i] = te.Proto()
	}

	pageSize := 10
	numReqs := wantCount / pageSize

	var got []*testsv1.TestExecution

	var nextPageTkn string

	for i := range numReqs {
		req := &testsv1.ListTestExecutionsRequest{
			TestId:        tt.ID.String(),
			PageSize:      int32(pageSize),
			NextPageToken: nextPageTkn,
		}
		res, err := s.ListTestExecutions(ctx, connect.NewRequest(req))
		require.NoError(t, err)
		if i == numReqs-1 {
			require.Empty(t, res.Msg.NextPageToken)
		} else {
			require.NotEmpty(t, res.Msg.NextPageToken)
			nextPageTkn = res.Msg.NextPageToken
		}
		got = append(got, res.Msg.TestExecutions...)
	}

	assert.Len(t, got, wantCount)
	assert.Equal(t, want, got)
}

func TestService_AckTestExecutionStarted(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	scheduled := fake.GenScheduledTestExec(tt.ID)
	te, err := fakes.repo.CreateScheduledTestExecution(ctx, scheduled)
	require.NoError(t, err)

	req := &testsv1.AckTestExecutionStartedRequest{
		TestExecutionId: te.ID.String(),
		StartTime:       timestamppb.New(time.Now().UTC()),
	}
	res, err := s.AckTestExecutionStarted(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetTestExecution(ctx, te.ID)
	require.NoError(t, err)
	assert.Equal(t, req.StartTime.AsTime(), *ackd.StartTime)
}

func TestService_AckTestExecutionFinished(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	scheduled := fake.GenScheduledTestExec(tt.ID)
	te, err := fakes.repo.CreateScheduledTestExecution(ctx, scheduled)
	require.NoError(t, err)

	req := &testsv1.AckTestExecutionFinishedRequest{
		TestExecutionId: te.ID.String(),
		FinishTime:      timestamppb.New(time.Now().UTC()),
		Error:           ptr.Get("bang"),
	}
	res, err := s.AckTestExecutionFinished(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetTestExecution(ctx, te.ID)
	require.NoError(t, err)
	assert.Equal(t, req.FinishTime.AsTime(), *ackd.FinishTime)
}

func TestService_RetryTestExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	repo := inmem.NewTestRepository(inmem.NewDB())

	// Setup test
	caseErr := "case error: bang"
	baseTest, err := repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	// Setup test/case executions
	testExec := createTestExec(t, ctx, repo, baseTest.ID, &caseErr)
	successCaseExec := createCaseExec(t, ctx, repo, testExec.ID, nil)
	failureCaseExec := createCaseExec(t, ctx, repo, testExec.ID, &caseErr)

	// Setup logs
	testExecLog := fake.GenTestExecLog(testExec.ID)
	err = repo.CreateLog(ctx, testExecLog)
	require.NoError(t, err)
	numCaseLogs := 10
	successCaseLogs := createCaseLogs(t, ctx, repo, testExec.ID, successCaseExec.ID, numCaseLogs)
	failureCaseLogs := createCaseLogs(t, ctx, repo, testExec.ID, failureCaseExec.ID, numCaseLogs)

	svc := New(repo, fake.NewWorkflower(
		fake.WithHistory(
			fake.GenCaseFailureHistory(
				testExec.ID,
				testExecLog.ID,
				successCaseExec.ID,
				failureCaseExec.ID,
			),
		),
	))

	// Retry test execution
	req := &testsv1.RetryTestExecutionRequest{
		TestExecutionId: testExec.ID.String(),
	}
	res, err := svc.RetryTestExecution(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	// Assert test execution record reset
	gotTestExec := res.Msg.TestExecution
	assert.Equal(t, testExec.ID.String(), gotTestExec.Id)
	assert.Equal(t, testExec.TestID.String(), gotTestExec.TestId)
	assert.True(t, gotTestExec.ScheduleTime.AsTime().After(testExec.ScheduleTime))
	assert.Nil(t, gotTestExec.StartTime)
	assert.Nil(t, gotTestExec.FinishTime)
	assert.Nil(t, gotTestExec.Error)

	// Assert test execution log record was not deleted
	gotLog, err := repo.GetLog(ctx, testExecLog.ID)
	assert.NoError(t, err)
	assert.Equal(t, testExecLog, gotLog)

	// Assert success case execution log records were not deleted
	for _, l := range successCaseLogs {
		gotLog, err = repo.GetLog(ctx, l.ID)
		assert.NoError(t, err)
		assert.Equal(t, l, gotLog)
	}

	// Assert success case execution log records were not deleted
	for _, l := range failureCaseLogs {
		_, err = repo.GetLog(ctx, l.ID)
		assert.ErrorIs(t, err, test.ErrorLogNotFound)
	}
}
