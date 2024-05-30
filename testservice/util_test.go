package testservice

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/annexhq/annex/inmem"
	"github.com/annexhq/annex/internal/fake"
	"github.com/annexhq/annex/test"
)

type fakeDeps struct {
	repo       test.Repository
	workflower Workflower
}

func newService() (*Service, *fakeDeps) {
	repo := inmem.NewTestRepository(inmem.NewDB())
	workflower := fake.NewWorkflower()
	defer workflower.Cleanup()
	return New(repo, workflower), &fakeDeps{
		repo:       repo,
		workflower: workflower,
	}
}

func createTestExec(t *testing.T, ctx context.Context, repo test.Repository, testID uuid.UUID, failure *string) *test.TestExecution {
	te, err := repo.CreateScheduledTestExecution(ctx, fake.GenScheduledTestExec(testID))
	require.NoError(t, err)
	te, err = repo.UpdateStartedTestExecution(ctx, fake.GenStartedTestExec(te.ID))
	require.NoError(t, err)
	te, err = repo.UpdateFinishedTestExecution(ctx, fake.GenFinishedTestExec(te.ID, failure))
	require.NoError(t, err)
	return te
}

func createCaseExec(t *testing.T, ctx context.Context, repo test.Repository, testExecID test.TestExecutionID, failure *string) *test.CaseExecution {
	ce, err := repo.CreateScheduledCaseExecution(ctx, fake.GenScheduledCaseExec(testExecID))
	require.NoError(t, err)
	ce, err = repo.UpdateStartedCaseExecution(ctx, fake.GenStartedCaseExec(testExecID, ce.ID))
	require.NoError(t, err)
	ce, err = repo.UpdateFinishedCaseExecution(ctx, fake.GenFinishedCaseExec(testExecID, ce.ID, failure))
	require.NoError(t, err)
	return ce
}

func createCaseLogs(t *testing.T, ctx context.Context, repo test.Repository, testExecID test.TestExecutionID, caseExecID test.CaseExecutionID, count int) test.ExecutionLogList {
	logs := fake.GenCaseExecLogs(count, testExecID, caseExecID)
	for _, log := range logs {
		err := repo.CreateExecutionLog(ctx, log)
		require.NoError(t, err)
	}
	return logs
}
