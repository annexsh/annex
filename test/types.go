package test

import (
	"time"

	"github.com/annexsh/annex/uuid"
)

type TestSuite struct {
	ID          uuid.V7
	ContextID   string
	Name        string
	Description *string
}

type TestSuiteList []*TestSuite

type TestSuiteRunner struct {
	ID             string
	LastAccessTime time.Time
}

type TestSuiteRunnerList []*TestSuiteRunner

type Test struct {
	ContextID   string
	TestSuiteID uuid.V7
	ID          uuid.V7
	Name        string
	HasInput    bool
	CreateTime  time.Time
}

type TestList []*Test

type Payload struct {
	Metadata map[string][]byte
	Data     []byte
}

type TestExecution struct {
	ID           TestExecutionID
	TestID       uuid.V7
	HasInput     bool
	ScheduleTime time.Time
	StartTime    *time.Time
	FinishTime   *time.Time
	Error        *string
}

type TestExecutionList []*TestExecution

type ScheduledTestExecution struct {
	ID           TestExecutionID
	TestID       uuid.V7
	HasInput     bool
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

type CaseExecution struct {
	ID              CaseExecutionID
	TestExecutionID TestExecutionID
	CaseName        string
	ScheduleTime    time.Time
	StartTime       *time.Time
	FinishTime      *time.Time
	Error           *string
}

type CaseExecutionList []*CaseExecution

type ScheduledCaseExecution struct {
	ID              CaseExecutionID
	TestExecutionID TestExecutionID
	CaseName        string
	ScheduleTime    time.Time
}

type StartedCaseExecution struct {
	ID              CaseExecutionID
	TestExecutionID TestExecutionID
	StartTime       time.Time
}

type FinishedCaseExecution struct {
	ID              CaseExecutionID
	TestExecutionID TestExecutionID
	FinishTime      time.Time
	Error           *string
}

type Log struct {
	ID              uuid.V7
	TestExecutionID TestExecutionID
	CaseExecutionID *CaseExecutionID
	Level           string
	Message         string
	CreateTime      time.Time
}

type LogList []*Log

type Identifier interface {
	string | uuid.V7 | TestExecutionID | CaseExecutionID
}

type PageFilter[T Identifier] struct {
	Size     int
	OffsetID *T
}
