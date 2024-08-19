package test

import (
	"time"

	"github.com/annexsh/annex/uuid"
)

type TestDefinition struct {
	ContextID    string
	GroupID      string
	TestID       uuid.V7
	Name         string
	DefaultInput *Payload
}

type Test struct {
	ContextID  string
	GroupID    string
	ID         uuid.V7
	Name       string
	HasInput   bool
	CreateTime time.Time
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

type ResetTestExecution struct {
	ID                  TestExecutionID
	ResetTime           time.Time
	StaleCaseExecutions []CaseExecutionID
	StaleLogs           []uuid.V7
}

type ScheduledTestExecution struct {
	ID           TestExecutionID
	TestID       uuid.V7
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
	LastScheduleTime    *time.Time // required when listing next page
	LastTestExecutionID *uuid.V7   // required when listing next page
	PageSize            uint32
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
	ID           CaseExecutionID
	TestExecID   TestExecutionID
	CaseName     string
	ScheduleTime time.Time
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
