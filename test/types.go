package test

import (
	"time"

	"github.com/google/uuid"
)

type TestDefinition struct {
	TestID         uuid.UUID
	Project        string
	Name           string
	DefaultPayload *Payload
	RunnerID       string
}

type Test struct {
	ID         uuid.UUID
	Project    string
	Name       string
	HasPayload bool
	Runners    []*TestRunner
	CreateTime time.Time
}

type TestRunner struct {
	ID                string
	LastHeartbeatTime time.Time
	IsActive          bool
}

type TestList []*Test

type Payload struct {
	Metadata map[string][]byte
	Payload  []byte
	IsZero   bool
}

type TestExecution struct {
	ID           TestExecutionID
	TestID       uuid.UUID
	HasPayload   bool
	ScheduleTime time.Time
	StartTime    *time.Time
	FinishTime   *time.Time
	Error        *string
}

type TestExecutionList []*TestExecution

type ScheduledTestExecution struct {
	ID           TestExecutionID
	TestID       uuid.UUID
	Payload      []byte
	ScheduleTime time.Time
}

type StartedTestExecution struct {
	ID        TestExecutionID
	StartTime time.Time
}

type FinishedTestExecution struct {
	ID         TestExecutionID
	FinishTime time.Time
	Error      *string
}

type TestExecutionListFilter struct {
	LastScheduleTime *time.Time // required when listing next page
	LastExecID       *uuid.UUID // required when listing next page
	PageSize         uint32
}

type CaseExecution struct {
	ID           CaseExecutionID
	TestExecID   TestExecutionID
	CaseName     string
	ScheduleTime time.Time
	StartTime    *time.Time
	FinishTime   *time.Time
	Error        *string
}

type CaseExecutionList []*CaseExecution

type ScheduledCaseExecution struct {
	ID           CaseExecutionID
	TestExecID   TestExecutionID
	CaseName     string
	ScheduleTime time.Time
}

type StartedCaseExecution struct {
	ID         CaseExecutionID
	TestExecID TestExecutionID
	StartTime  time.Time
}

type FinishedCaseExecution struct {
	ID         CaseExecutionID
	TestExecID TestExecutionID
	FinishTime time.Time
	Error      *string
}

type ExecutionLog struct {
	ID         uuid.UUID
	TestExecID TestExecutionID
	CaseExecID *CaseExecutionID
	Level      string
	Message    string
	CreateTime time.Time
}

type ExecutionLogList []*ExecutionLog

type ResetTestExecution struct {
	ID                  TestExecutionID
	ResetTime           time.Time
	StaleCaseExecutions []CaseExecutionID
	StaleLogs           []uuid.UUID
}
