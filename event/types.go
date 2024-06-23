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
	TypeLogPublished
)

type DataType uint

const (
	DataTypeUnspecified DataType = iota
	DataTypeNone
	DataTypeTestExecution
	DataTypeCaseExecution
	DataTypeLog
)

type Data struct {
	Type          DataType
	TestExecution *test.TestExecution
	CaseExecution *test.CaseExecution
	Log           *test.Log
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

func (d *Data) GetLog() (*test.Log, error) {
	if d.Type != DataTypeLog || d.Log == nil {
		return nil, errors.New("event data does not contain a valid execution log")
	}
	return d.Log, nil
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
		ID:         getTestExecEventID(testExec, eventType),
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
		ID:         getCaseExecEventID(caseExec, eventType),
		TestExecID: caseExec.TestExecutionID,
		Type:       eventType,
		Data: Data{
			Type:          DataTypeCaseExecution,
			CaseExecution: caseExec,
		},
		CreateTime: time.Now().UTC(),
	}
}

func NewLogEvent(eventType Type, log *test.Log) *ExecutionEvent {
	return &ExecutionEvent{
		ID:         getGetLogEventID(log.ID),
		TestExecID: log.TestExecutionID,
		Type:       eventType,
		Data: Data{
			Type: DataTypeLog,
			Log:  log,
		},
		CreateTime: time.Now().UTC(),
	}
}

var uuidNameSpaceEvent = uuid.MustParse("4a4da572-d093-4fe4-a60b-d1ddae369894")

func getTestExecEventID(testExec *test.TestExecution, eventType Type) uuid.UUID {
	raw := testExec.ID.String() + "." + eventType.Proto().String()
	return uuid.NewSHA1(uuidNameSpaceEvent, []byte(raw))
}

func getCaseExecEventID(caseExec *test.CaseExecution, eventType Type) uuid.UUID {
	raw := caseExec.TestExecutionID.String() + "." + caseExec.ID.String() + "." + eventType.Proto().String()
	return uuid.NewSHA1(uuidNameSpaceEvent, []byte(raw))
}

func getGetLogEventID(logID uuid.UUID) uuid.UUID {
	return logID
}
