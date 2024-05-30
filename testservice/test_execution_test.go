package testservice

import (
	"context"
	"testing"
	"time"

	testservicev1 "github.com/annexhq/annex-proto/gen/go/rpc/testservice/v1"
	testv1 "github.com/annexhq/annex-proto/gen/go/type/test/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexhq/annex/inmem"
	"github.com/annexhq/annex/internal/fake"
	"github.com/annexhq/annex/internal/ptr"
	"github.com/annexhq/annex/test"
)

func TestService_GetTestExecution(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	scheduled := fake.GenScheduledTestExec(tt.ID)
	te, err := fakes.repo.CreateScheduledTestExecution(ctx, scheduled)
	require.NoError(t, err)

	res, err := s.GetTestExecution(ctx, &testservicev1.GetTestExecutionRequest{
		Id: te.ID.String(),
	})
	require.NoError(t, err)

	assert.Equal(t, te.Proto(), res.TestExecution)
	assert.Equal(t, scheduled.Payload, res.Input.Data)
}

func TestService_ListTestExecutions(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	wantCount := 30
	want := make([]*testv1.TestExecution, wantCount)

	for i := range wantCount {
		scheduled := fake.GenScheduledTestExec(tt.ID)
		te, err := fakes.repo.CreateScheduledTestExecution(ctx, scheduled)
		require.NoError(t, err)
		want[i] = te.Proto()
	}

	pageSize := 10
	numReqs := wantCount / pageSize

	var got []*testv1.TestExecution

	var nextPageTkn string

	for i := range numReqs {
		res, err := s.ListTestExecutions(ctx, &testservicev1.ListTestExecutionsRequest{
			TestId:        tt.ID.String(),
			PageSize:      int32(pageSize),
			NextPageToken: nextPageTkn,
		})
		require.NoError(t, err)
		if i == numReqs-1 {
			require.Empty(t, res.NextPageToken)
		} else {
			require.NotEmpty(t, res.NextPageToken)
			nextPageTkn = res.NextPageToken
		}
		got = append(got, res.TestExecutions...)
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

	req := &testservicev1.AckTestExecutionStartedRequest{
		TestExecId: te.ID.String(),
		StartedAt:  timestamppb.New(time.Now().UTC()),
	}
	res, err := s.AckTestExecutionStarted(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetTestExecution(ctx, te.ID)
	require.NoError(t, err)
	assert.Equal(t, req.StartedAt.AsTime(), *ackd.StartTime)
}

func TestService_AckTestExecutionFinished(t *testing.T) {
	ctx := context.Background()
	s, fakes := newService()

	tt, err := fakes.repo.CreateTest(ctx, fake.GenTestDefinition())
	require.NoError(t, err)

	scheduled := fake.GenScheduledTestExec(tt.ID)
	te, err := fakes.repo.CreateScheduledTestExecution(ctx, scheduled)
	require.NoError(t, err)

	req := &testservicev1.AckTestExecutionFinishedRequest{
		TestExecId: te.ID.String(),
		FinishedAt: timestamppb.New(time.Now().UTC()),
		Error:      ptr.Get("bang"),
	}
	res, err := s.AckTestExecutionFinished(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, res)

	ackd, err := fakes.repo.GetTestExecution(ctx, te.ID)
	require.NoError(t, err)
	assert.Equal(t, req.FinishedAt.AsTime(), *ackd.FinishTime)
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
	err = repo.CreateExecutionLog(ctx, testExecLog)
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
	res, err := svc.RetryTestExecution(ctx, &testservicev1.RetryTestExecutionRequest{
		TestExecId: testExec.ID.String(),
	})
	require.NoError(t, err)

	// Assert test execution record reset
	gotTestExec := res.TestExecution
	assert.Equal(t, testExec.ID.String(), gotTestExec.Id)
	assert.Equal(t, testExec.TestID.String(), gotTestExec.TestId)
	assert.True(t, gotTestExec.ScheduledAt.AsTime().After(testExec.ScheduleTime))
	assert.Nil(t, gotTestExec.StartedAt)
	assert.Nil(t, gotTestExec.FinishedAt)
	assert.Nil(t, gotTestExec.Error)

	// Assert test execution log record was not deleted
	gotLog, err := repo.GetExecutionLog(ctx, testExecLog.ID)
	assert.NoError(t, err)
	assert.Equal(t, testExecLog, gotLog)

	// Assert success case execution log records were not deleted
	for _, l := range successCaseLogs {
		gotLog, err = repo.GetExecutionLog(ctx, l.ID)
		assert.NoError(t, err)
		assert.Equal(t, l, gotLog)
	}

	// Assert success case execution log records were not deleted
	for _, l := range failureCaseLogs {
		_, err = repo.GetExecutionLog(ctx, l.ID)
		assert.ErrorIs(t, err, test.ErrorExecutionLogNotFound)
	}
}
