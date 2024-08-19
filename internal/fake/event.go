package fake

import (
	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func GenExecEvent(testExecID test.TestExecutionID) *eventsv1.Event {
	return &eventsv1.Event{
		EventId:         uuid.New().String(),
		TestExecutionId: testExecID.String(),
		Type:            eventsv1.Event_TYPE_CASE_EXECUTION_STARTED,
		Data: &eventsv1.Event_Data{
			Type: eventsv1.Event_Data_TYPE_TEST_EXECUTION,
			Data: &eventsv1.Event_Data_TestExecution{
				TestExecution: GenTestExec(uuid.New()).Proto(),
			},
		},
		CreateTime: timestamppb.Now(),
	}
}
