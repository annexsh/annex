package event

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/annexsh/annex/test"
)

type Type uint

const (
	TypeUnspecified Type = iota
	TypeTestExecutionScheduled
	TypeTestExecutionStarted
	TypeTestExecutionFinished
	TypeCaseExecutionScheduled
	TypeCaseExecutionStarted
	TypeCaseExecutionFinished
	TypeExecutionLogPublished
)

type DataType uint

const (
	DataTypeUnspecified DataType = iota
	DataTypeNone
	DataTypeTestExecution
	DataTypeCaseExecution
	DataTypeExecutionLog
)

type Data struct {
	Type          DataType
	TestExecution *test.TestExecution
	CaseExecution *test.CaseExecution
	ExecutionLog  *test.ExecutionLog
}

func (d *Data) GetTestExecution() (*test.TestExecution, error) {
	if d.Type != DataTypeTestExecution || d.TestExecution == nil {
		return nil, errors.New("event data does not contain a valid test execution")
	}
	return d.TestExecution, nil
}

func (d *Data) GetCaseExecution() (*test.CaseExecution, error) {
	if d.Type != DataTypeCaseExecution || d.CaseExecution == nil {
		return nil, errors.New("event data does not contain a valid case execution")
	}
	return d.CaseExecution, nil
}

func (d *Data) GetExecutionLog() (*test.ExecutionLog, error) {
	if d.Type != DataTypeExecutionLog || d.ExecutionLog == nil {
		return nil, errors.New("event data does not contain a valid execution log")
	}
	return d.ExecutionLog, nil
}

type ExecutionEvent struct {
	ID         uuid.UUID
	TestExecID test.TestExecutionID
	Type       Type
	Data       Data
	CreateTime time.Time
}

func NewTestExecutionEvent(eventType Type, testExec *test.TestExecution) *ExecutionEvent {
	return &ExecutionEvent{
		ID:         uuid.New(),
		TestExecID: testExec.ID,
		Type:       eventType,
		Data: Data{
			Type:          DataTypeTestExecution,
			TestExecution: testExec,
		},
		CreateTime: time.Now().UTC(),
	}
}

func NewCaseExecutionEvent(eventType Type, caseExec *test.CaseExecution) *ExecutionEvent {
	return &ExecutionEvent{
		ID:         uuid.New(),
		TestExecID: caseExec.TestExecID,
		Type:       eventType,
		Data: Data{
			Type:          DataTypeCaseExecution,
			CaseExecution: caseExec,
		},
		CreateTime: time.Now().UTC(),
	}
}

func NewExecutionLogEvent(eventType Type, execLog *test.ExecutionLog) *ExecutionEvent {
	return &ExecutionEvent{
		ID:         uuid.New(),
		TestExecID: execLog.TestExecID,
		Type:       eventType,
		Data: Data{
			Type:         DataTypeExecutionLog,
			ExecutionLog: execLog,
		},
		CreateTime: time.Now().UTC(),
	}
}
