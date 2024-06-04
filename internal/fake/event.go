package fake

import (
	"time"

	"github.com/google/uuid"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/test"
)

func GenExecEvent(testExecID test.TestExecutionID) *event.ExecutionEvent {
	return &event.ExecutionEvent{
		ID:         uuid.New(),
		TestExecID: testExecID,
		Type:       event.TypeTestExecutionStarted,
		Data: event.Data{
			Type:          event.DataTypeTestExecution,
			TestExecution: GenTestExec(uuid.New()),
		},
		CreateTime: time.Now(),
	}
}
