package test

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	TestReadWriter
	TestExecutionReadWriter
	CaseExecutionReadWriter
	ExecutionLogReadWriter
}

type TestReadWriter interface {
	TestReader
	TestWriter
}

type TestReader interface {
	GetTest(ctx context.Context, id uuid.UUID) (*Test, error)
	ListTests(ctx context.Context) (TestList, error)
	GetTestDefaultPayload(ctx context.Context, testID uuid.UUID) (*Payload, error)
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
	GetTestExecutionPayload(ctx context.Context, id TestExecutionID) (*Payload, error)
	ListTestExecutions(ctx context.Context, testID uuid.UUID, filter *TestExecutionListFilter) (TestExecutionList, error)
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

type ExecutionLogReadWriter interface {
	ExecutionLogReader
	ExecutionLogWriter
}

type ExecutionLogReader interface {
	GetExecutionLog(ctx context.Context, id uuid.UUID) (*ExecutionLog, error)
	ListExecutionLogs(ctx context.Context, testExecID TestExecutionID) (ExecutionLogList, error)
}

type ExecutionLogWriter interface {
	CreateExecutionLog(ctx context.Context, log *ExecutionLog) error
	DeleteExecutionLog(ctx context.Context, id uuid.UUID) error
}

type ResetRollback func(ctx context.Context) error
