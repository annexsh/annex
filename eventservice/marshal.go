package eventservice

import (
	eventsv1 "github.com/annexsh/annex-proto/gen/go/annex/events/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var typeProto = map[Type]eventsv1.Event_Type{
	TypeUnspecified:            eventsv1.Event_TYPE_UNSPECIFIED,
	TypeTestExecutionScheduled: eventsv1.Event_TYPE_TEST_EXECUTION_SCHEDULED,
	TypeTestExecutionStarted:   eventsv1.Event_TYPE_TEST_EXECUTION_STARTED,
	TypeTestExecutionFinished:  eventsv1.Event_TYPE_TEST_EXECUTION_FINISHED,
	TypeCaseExecutionScheduled: eventsv1.Event_TYPE_CASE_EXECUTION_SCHEDULED,
	TypeCaseExecutionStarted:   eventsv1.Event_TYPE_CASE_EXECUTION_STARTED,
	TypeCaseExecutionFinished:  eventsv1.Event_TYPE_CASE_EXECUTION_FINISHED,
	TypeLogPublished:           eventsv1.Event_TYPE_LOG_PUBLISHED,
}

func (t Type) Proto() eventsv1.Event_Type {
	pb, ok := typeProto[t]
	if !ok {
		return eventsv1.Event_TYPE_UNSPECIFIED
	}
	return pb
}

var dataTypeProto = map[DataType]eventsv1.Event_Data_Type{
	DataTypeUnspecified:   eventsv1.Event_Data_TYPE_UNSPECIFIED,
	DataTypeNone:          eventsv1.Event_Data_TYPE_NONE,
	DataTypeTestExecution: eventsv1.Event_Data_TYPE_TEST_EXECUTION,
	DataTypeCaseExecution: eventsv1.Event_Data_TYPE_CASE_EXECUTION,
	DataTypeLog:           eventsv1.Event_Data_TYPE_LOG,
}

func (t DataType) Proto() eventsv1.Event_Data_Type {
	pb, ok := dataTypeProto[t]
	if !ok {
		return eventsv1.Event_Data_TYPE_UNSPECIFIED
	}
	return pb
}

func (e *ExecutionEvent) Proto() *eventsv1.Event {
	data := &eventsv1.Event_Data{
		Type: e.Data.Type.Proto(),
	}

	switch e.Data.Type {
	case DataTypeUnspecified, DataTypeNone:
	case DataTypeTestExecution:
		if e.Data.TestExecution != nil {
			data.Data = &eventsv1.Event_Data_TestExecution{
				TestExecution: e.Data.TestExecution.Proto(),
			}
		}
	case DataTypeCaseExecution:
		if e.Data.CaseExecution != nil {
			data.Data = &eventsv1.Event_Data_CaseExecution{
				CaseExecution: e.Data.CaseExecution.Proto(),
			}
		}
	case DataTypeLog:
		if e.Data.Log != nil {
			data.Data = &eventsv1.Event_Data_Log{
				Log: e.Data.Log.Proto(),
			}
		}
	}

	return &eventsv1.Event{
		EventId:         e.ID.String(),
		TestExecutionId: e.TestExecID.String(),
		Type:            e.Type.Proto(),
		Data:            data,
		CreateTime:      timestamppb.New(e.CreateTime),
	}
}
