package test

import (
	"context"

	"github.com/annexsh/annex/uuid"
)

type Repository interface {
	ContextReadWriter
	GroupReadWriter
	TestReadWriter
	TestExecutionReadWriter
	CaseExecutionReadWriter
	LogReadWriter
}

type ContextReadWriter interface {
	ContextReader
	ContextWriter
}

type ContextReader interface {
	ListContexts(ctx context.Context) ([]string, error)
	ContextExists(ctx context.Context, id string) (bool, error)
}

type ContextWriter interface {
	CreateContext(ctx context.Context, id string) error
}

type GroupReadWriter interface {
	GroupReader
	GroupWriter
}

type GroupReader interface {
	ListGroups(ctx context.Context, contextID string) ([]string, error)
	GroupExists(ctx context.Context, contextID string, groupID string) (bool, error)
}

type GroupWriter interface {
	CreateGroup(ctx context.Context, contextID string, groupID string) error
}

type TestReadWriter interface {
	TestReader
	TestWriter
}

type TestReader interface {
	GetTest(ctx context.Context, id uuid.V7) (*Test, error)
	ListTests(ctx context.Context, contextID string, groupID string) (TestList, error)
	GetTestDefaultInput(ctx context.Context, testID uuid.V7) (*Payload, error)
}

type TestWriter interface {
	CreateTest(ctx context.Context, test *TestDefinition) (*Test, error)
	CreateTests(ctx context.Context, tests ...*TestDefinition) (TestList, error)
}

type TestExecutionReadWriter interface {
	TestExecutionReader
	TestExecutionWriter
}

type TestExecutionReader interface {
	GetTestExecution(ctx context.Context, id TestExecutionID) (*TestExecution, error)
	GetTestExecutionInput(ctx context.Context, id TestExecutionID) (*Payload, error)
	ListTestExecutions(ctx context.Context, testID uuid.V7, filter *TestExecutionListFilter) (TestExecutionList, error)
}

type TestExecutionWriter interface {
	CreateScheduledTestExecution(ctx context.Context, scheduled *ScheduledTestExecution) (*TestExecution, error)
	UpdateStartedTestExecution(ctx context.Context, started *StartedTestExecution) (*TestExecution, error)
	UpdateFinishedTestExecution(ctx context.Context, finished *FinishedTestExecution) (*TestExecution, error)
	ResetTestExecution(ctx context.Context, reset *ResetTestExecution) (*TestExecution, ResetRollback, error)
}

type CaseExecutionReadWriter interface {
	CaseExecutionReader
	CaseExecutionWriter
}

type CaseExecutionReader interface {
	GetCaseExecution(ctx context.Context, testExecID TestExecutionID, caseExecID CaseExecutionID) (*CaseExecution, error)
	ListCaseExecutions(ctx context.Context, testExecID TestExecutionID) (CaseExecutionList, error)
}

type CaseExecutionWriter interface {
	CreateScheduledCaseExecution(ctx context.Context, scheduled *ScheduledCaseExecution) (*CaseExecution, error)
	UpdateStartedCaseExecution(ctx context.Context, started *StartedCaseExecution) (*CaseExecution, error)
	UpdateFinishedCaseExecution(ctx context.Context, finished *FinishedCaseExecution) (*CaseExecution, error)
	DeleteCaseExecution(ctx context.Context, testExecID TestExecutionID, id CaseExecutionID) error
}

type LogReadWriter interface {
	LogReader
	LogWriter
}

type LogReader interface {
	GetLog(ctx context.Context, id uuid.V7) (*Log, error)
	ListLogs(ctx context.Context, testExecID TestExecutionID) (LogList, error)
}

type LogWriter interface {
	CreateLog(ctx context.Context, log *Log) error
	DeleteLog(ctx context.Context, id uuid.V7) error
}

type ResetRollback func(ctx context.Context) error
