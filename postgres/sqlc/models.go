// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package sqlc

import (
	"github.com/annexsh/annex/test"
	"github.com/google/uuid"
)

type CaseExecution struct {
	ID              test.CaseExecutionID `json:"id"`
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	CaseName        string               `json:"case_name"`
	ScheduleTime    Timestamp            `json:"schedule_time"`
	StartTime       Timestamp            `json:"start_time"`
	FinishTime      Timestamp            `json:"finish_time"`
	Error           *string              `json:"error"`
}

type Context struct {
	ID string `json:"id"`
}

type Group struct {
	ContextID string `json:"context_id"`
	ID        string `json:"id"`
}

type Log struct {
	ID              uuid.UUID             `json:"id"`
	TestExecutionID test.TestExecutionID  `json:"test_execution_id"`
	CaseExecutionID *test.CaseExecutionID `json:"case_execution_id"`
	Level           string                `json:"level"`
	Message         string                `json:"message"`
	CreateTime      Timestamp             `json:"create_time"`
}

type Test struct {
	ContextID  string    `json:"context_id"`
	GroupID    string    `json:"group_id"`
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	HasInput   bool      `json:"has_input"`
	CreateTime Timestamp `json:"create_time"`
}

type TestDefaultInput struct {
	TestID uuid.UUID `json:"test_id"`
	Data   []byte    `json:"data"`
}

type TestExecution struct {
	ID           test.TestExecutionID `json:"id"`
	TestID       uuid.UUID            `json:"test_id"`
	HasInput     bool                 `json:"has_input"`
	ScheduleTime Timestamp            `json:"schedule_time"`
	StartTime    Timestamp            `json:"start_time"`
	FinishTime   Timestamp            `json:"finish_time"`
	Error        *string              `json:"error"`
}

type TestExecutionInput struct {
	TestExecutionID test.TestExecutionID `json:"test_execution_id"`
	Data            []byte               `json:"data"`
}
