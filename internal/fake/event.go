package fake

import (
	"time"

	"github.com/google/uuid"

	"github.com/annexhq/annex/event"
	"github.com/annexhq/annex/test"
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
