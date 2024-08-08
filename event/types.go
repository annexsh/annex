package event

import (
	"strconv"
	"time"

	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewTestExecutionEvent(eventType eventsv1.Event_Type, testExec *testsv1.TestExecution) *eventsv1.Event {
	return &eventsv1.Event{
		EventId:         encodeEventID(testExec.Id + "." + eventType.String()).String(),
		TestExecutionId: testExec.Id,
		Type:            eventType,
		Data: &eventsv1.Event_Data{
			Type: eventsv1.Event_Data_TYPE_TEST_EXECUTION,
			Data: &eventsv1.Event_Data_TestExecution{
				TestExecution: testExec,
			},
		},
		CreateTime: timestamppb.New(time.Now().UTC()),
	}
}

func NewCaseExecutionEvent(eventType eventsv1.Event_Type, caseExec *testsv1.CaseExecution) *eventsv1.Event {
	rawID := caseExec.TestExecutionId + "." + strconv.FormatInt(int64(caseExec.Id), 10) + "." + eventType.String()
	return &eventsv1.Event{
		EventId:         encodeEventID(rawID).String(),
		TestExecutionId: caseExec.TestExecutionId,
		Type:            eventType,
		Data: &eventsv1.Event_Data{
			Type: eventsv1.Event_Data_TYPE_CASE_EXECUTION,
			Data: &eventsv1.Event_Data_CaseExecution{
				CaseExecution: caseExec,
			},
		},
		CreateTime: timestamppb.New(time.Now().UTC()),
	}
}

func NewLogEvent(eventType eventsv1.Event_Type, log *testsv1.Log) *eventsv1.Event {
	return &eventsv1.Event{
		EventId:         encodeEventID(log.Id).String(),
		TestExecutionId: log.TestExecutionId,
		Type:            eventType,
		Data: &eventsv1.Event_Data{
			Type: eventsv1.Event_Data_TYPE_LOG,
			Data: &eventsv1.Event_Data_Log{
				Log: log,
			},
		},
		CreateTime: timestamppb.New(time.Now().UTC()),
	}
}

var uuidNameSpaceEvent = uuid.MustParse("4a4da572-d093-4fe4-a60b-d1ddae369894")

func encodeEventID(raw string) uuid.UUID {
	return uuid.NewSHA1(uuidNameSpaceEvent, []byte(raw))
}
