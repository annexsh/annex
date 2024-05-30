package test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	testWorkflowPrefix = "test.workflow."
	caseActivityPrefix = "case.activity."
)

type TestExecutionID struct {
	uuid.UUID
}

func (i TestExecutionID) WorkflowID() string {
	return testWorkflowPrefix + i.String()
}

func (i TestExecutionID) IsEmpty() bool {
	return i.UUID == uuid.Nil
}

func NewTestExecutionID() TestExecutionID {
	return TestExecutionID{uuid.New()}
}

func ParseTestExecutionID(id string) (TestExecutionID, error) {
	texid, err := uuid.Parse(id)
	if err != nil {
		return TestExecutionID{}, fmt.Errorf("test execution id is not a valid uuid: %w", err)
	}
	return TestExecutionID{texid}, nil
}

func ParseTestWorkflowID(workflowID string) (TestExecutionID, error) {
	if idStr, ok := trimPrefix(workflowID, testWorkflowPrefix); ok {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return TestExecutionID{}, fmt.Errorf("test execution id is not a valid uuid: %w", err)
		}
		return TestExecutionID{id}, nil
	}
	return TestExecutionID{}, ErrorNotTestExecution
}

type CaseExecutionID int32

func (i CaseExecutionID) Int32() int32 {
	return int32(i)
}

func (i CaseExecutionID) String() string {
	return strconv.Itoa(int(i))
}

func (i CaseExecutionID) ActivityID() string {
	return caseActivityPrefix + i.String()
}

func (i CaseExecutionID) IsEmpty() bool {
	return i == 0
}

func ParseCaseActivityID(activityID string) (CaseExecutionID, error) {
	if idStr, ok := trimPrefix(activityID, caseActivityPrefix); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("activity execution id is not a valid int32: %w", err)
		}
		return CaseExecutionID(id), nil
	}
	return 0, ErrorNotCaseExecution
}

func trimPrefix(s string, prefix string) (string, bool) {
	parts := strings.Split(s, prefix)
	if len(parts) != 2 {
		return "", false
	}
	return parts[1], true
}
