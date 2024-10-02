package test

import (
	"context"
	"time"

	"github.com/annexsh/annex/uuid"
)

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Repository interface {
	ContextReadWriter
	TestSuiteReadWriter
	TestReadWriter
	TestExecutionReadWriter
	CaseExecutionReadWriter
	LogReadWriter
	WithTx(ctx context.Context) (Repository, Tx, error)
	ExecuteTx(ctx context.Context, query func(repo Repository) error) error
}

type ContextReadWriter interface {
	ContextReader
	ContextWriter
}

type ContextReader interface {
	ListContexts(ctx context.Context, filter PageFilter[string]) ([]string, error)
}

type ContextWriter interface {
	CreateContext(ctx context.Context, id string) error
}

type TestSuiteReadWriter interface {
	TestSuiteReader
	TestSuiteWriter
}

type TestSuiteReader interface {
	ListTestSuites(ctx context.Context, contextID string, filter PageFilter[string]) (TestSuiteList, error)
	GetTestSuiteVersion(ctx context.Context, contextID string, id uuid.V7) (string, error)
}

type TestSuiteWriter interface {
	CreateTestSuite(ctx context.Context, testSuite *TestSuite) (uuid.V7, error)
}

type TestReadWriter interface {
	TestReader
	TestWriter
}

type TestReader interface {
	GetTest(ctx context.Context, id uuid.V7) (*Test, error)
	ListTests(ctx context.Context, contextID string, testSuiteID uuid.V7, filter PageFilter[uuid.V7]) (TestList, error)
	GetTestDefaultInput(ctx context.Context, testID uuid.V7) (*Payload, error)
}

type TestWriter interface {
	CreateTest(ctx context.Context, test *Test) (*Test, error)
	DeleteTest(ctx context.Context, id uuid.V7) error
	CreateTestDefaultInput(ctx context.Context, testID uuid.V7, defaultInput *Payload) error
}

type TestExecutionReadWriter interface {
	TestExecutionReader
	TestExecutionWriter
}

type TestExecutionReader interface {
	GetTestExecution(ctx context.Context, id TestExecutionID) (*TestExecution, error)
	GetTestExecutionInput(ctx context.Context, id TestExecutionID) (*Payload, error)
	ListTestExecutions(ctx context.Context, testID uuid.V7, filter PageFilter[TestExecutionID]) (TestExecutionList, error)
}

type TestExecutionWriter interface {
	CreateTestExecutionScheduled(ctx context.Context, scheduled *ScheduledTestExecution) (*TestExecution, error)
	CreateTestExecutionInput(ctx context.Context, testExecID TestExecutionID, input *Payload) error
	UpdateTestExecutionStarted(ctx context.Context, started *StartedTestExecution) (*TestExecution, error)
	UpdateTestExecutionFinished(ctx context.Context, finished *FinishedTestExecution) (*TestExecution, error)
	ResetTestExecution(ctx context.Context, testExecID TestExecutionID, resetTime time.Time) (*TestExecution, error)
}

type CaseExecutionReadWriter interface {
	CaseExecutionReader
	CaseExecutionWriter
}

type CaseExecutionReader interface {
	GetCaseExecution(ctx context.Context, testExecID TestExecutionID, caseExecID CaseExecutionID) (*CaseExecution, error)
	ListCaseExecutions(ctx context.Context, testExecID TestExecutionID, filter PageFilter[CaseExecutionID]) (CaseExecutionList, error)
}

type CaseExecutionWriter interface {
	CreateCaseExecutionScheduled(ctx context.Context, scheduled *ScheduledCaseExecution) (*CaseExecution, error)
	UpdateCaseExecutionStarted(ctx context.Context, started *StartedCaseExecution) (*CaseExecution, error)
	UpdateCaseExecutionFinished(ctx context.Context, finished *FinishedCaseExecution) (*CaseExecution, error)
	DeleteCaseExecution(ctx context.Context, testExecID TestExecutionID, id CaseExecutionID) error
}

type LogReadWriter interface {
	LogReader
	LogWriter
}

type LogReader interface {
	GetLog(ctx context.Context, id uuid.V7) (*Log, error)
	ListLogs(ctx context.Context, testExecID TestExecutionID, filter PageFilter[uuid.V7]) (LogList, error)
}

type LogWriter interface {
	CreateLog(ctx context.Context, log *Log) error
	DeleteLog(ctx context.Context, id uuid.V7) error
}

type ResetRollback func(ctx context.Context) error
