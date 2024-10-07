// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"context"

	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

type Querier interface {
	CreateCaseExecutionScheduled(ctx context.Context, arg CreateCaseExecutionScheduledParams) (*CaseExecution, error)
	CreateContext(ctx context.Context, id string) error
	CreateLog(ctx context.Context, arg CreateLogParams) error
	CreateTest(ctx context.Context, arg CreateTestParams) (*Test, error)
	CreateTestDefaultInput(ctx context.Context, arg CreateTestDefaultInputParams) error
	CreateTestExecutionInput(ctx context.Context, arg CreateTestExecutionInputParams) error
	CreateTestExecutionScheduled(ctx context.Context, arg CreateTestExecutionScheduledParams) (*TestExecution, error)
	CreateTestSuite(ctx context.Context, arg CreateTestSuiteParams) (uuid.V7, error)
	DeleteCaseExecution(ctx context.Context, arg DeleteCaseExecutionParams) error
	DeleteLog(ctx context.Context, id uuid.V7) error
	DeleteTest(ctx context.Context, id uuid.V7) error
	GetCaseExecution(ctx context.Context, arg GetCaseExecutionParams) (*CaseExecution, error)
	GetLog(ctx context.Context, id uuid.V7) (*Log, error)
	GetTest(ctx context.Context, id uuid.V7) (*Test, error)
	GetTestByName(ctx context.Context, arg GetTestByNameParams) (*Test, error)
	GetTestDefaultInput(ctx context.Context, testID string) (*TestDefaultInput, error)
	GetTestExecution(ctx context.Context, id test.TestExecutionID) (*TestExecution, error)
	GetTestExecutionInput(ctx context.Context, testExecutionID test.TestExecutionID) (*TestExecutionInput, error)
	GetTestSuiteVersion(ctx context.Context, arg GetTestSuiteVersionParams) (string, error)
	ListCaseExecutions(ctx context.Context, arg ListCaseExecutionsParams) ([]*CaseExecution, error)
	ListContexts(ctx context.Context, arg ListContextsParams) ([]string, error)
	ListLogs(ctx context.Context, arg ListLogsParams) ([]*Log, error)
	ListTestExecutions(ctx context.Context, arg ListTestExecutionsParams) ([]*TestExecution, error)
	ListTestSuites(ctx context.Context, arg ListTestSuitesParams) ([]*TestSuite, error)
	ListTests(ctx context.Context, arg ListTestsParams) ([]*Test, error)
	ResetTestExecution(ctx context.Context, arg ResetTestExecutionParams) (*TestExecution, error)
	SetTestSuiteVersion(ctx context.Context, arg SetTestSuiteVersionParams) error
	UpdateCaseExecutionFinished(ctx context.Context, arg UpdateCaseExecutionFinishedParams) (*CaseExecution, error)
	UpdateCaseExecutionStarted(ctx context.Context, arg UpdateCaseExecutionStartedParams) (*CaseExecution, error)
	UpdateTestExecutionFinished(ctx context.Context, arg UpdateTestExecutionFinishedParams) (*TestExecution, error)
	UpdateTestExecutionStarted(ctx context.Context, arg UpdateTestExecutionStartedParams) (*TestExecution, error)
}

var _ Querier = (*Queries)(nil)
