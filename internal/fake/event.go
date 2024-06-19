package fake

import (
	"time"

	"github.com/google/uuid"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/test"
)

func GenExecEvent(testExecID test.TestExecutionID) *eventservice.ExecutionEvent {
	return &eventservice.ExecutionEvent{
		ID:         uuid.New(),
		TestExecID: testExecID,
		Type:       eventservice.TypeTestExecutionStarted,
		Data: eventservice.Data{
			Type:          eventservice.DataTypeTestExecution,
			TestExecution: GenTestExec(uuid.New()),
		},
		CreateTime: time.Now(),
	}
}
