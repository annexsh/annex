package event

import (
	testv1 "github.com/annexhq/annex-proto/gen/go/type/test/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var typeProto = map[Type]testv1.ExecutionEvent_Type{
	TypeUnspecified:            testv1.ExecutionEvent_TYPE_UNSPECIFIED,
	TypeTestExecutionScheduled: testv1.ExecutionEvent_TYPE_TEST_EXECUTION_SCHEDULED,
	TypeTestExecutionStarted:   testv1.ExecutionEvent_TYPE_TEST_EXECUTION_STARTED,
	TypeTestExecutionFinished:  testv1.ExecutionEvent_TYPE_TEST_EXECUTION_FINISHED,
	TypeCaseExecutionScheduled: testv1.ExecutionEvent_TYPE_CASE_EXECUTION_SCHEDULED,
	TypeCaseExecutionStarted:   testv1.ExecutionEvent_TYPE_CASE_EXECUTION_STARTED,
	TypeCaseExecutionFinished:  testv1.ExecutionEvent_TYPE_CASE_EXECUTION_FINISHED,
	TypeExecutionLogPublished:  testv1.ExecutionEvent_TYPE_EXECUTION_LOG_PUBLISHED,
}

func (t Type) Proto() testv1.ExecutionEvent_Type {
	pb, ok := typeProto[t]
	if !ok {
		return testv1.ExecutionEvent_TYPE_UNSPECIFIED
	}
	return pb
}

var dataTypeProto = map[DataType]testv1.ExecutionEvent_Data_Type{
	DataTypeUnspecified:   testv1.ExecutionEvent_Data_TYPE_UNSPECIFIED,
	DataTypeNone:          testv1.ExecutionEvent_Data_TYPE_NONE,
	DataTypeTestExecution: testv1.ExecutionEvent_Data_TYPE_TEST_EXECUTION,
	DataTypeCaseExecution: testv1.ExecutionEvent_Data_TYPE_CASE_EXECUTION,
	DataTypeExecutionLog:  testv1.ExecutionEvent_Data_TYPE_EXECUTION_LOG,
}

func (t DataType) Proto() testv1.ExecutionEvent_Data_Type {
	pb, ok := dataTypeProto[t]
	if !ok {
		return testv1.ExecutionEvent_Data_TYPE_UNSPECIFIED
	}
	return pb
}

func (e *ExecutionEvent) Proto() *testv1.ExecutionEvent {
	data := &testv1.ExecutionEvent_Data{
		Type: e.Data.Type.Proto(),
	}

	switch e.Data.Type {
	case DataTypeUnspecified, DataTypeNone:
	case DataTypeTestExecution:
		if e.Data.TestExecution != nil {
			data.Data = &testv1.ExecutionEvent_Data_TestExecution{
				TestExecution: e.Data.TestExecution.Proto(),
			}
		}
	case DataTypeCaseExecution:
		if e.Data.CaseExecution != nil {
			data.Data = &testv1.ExecutionEvent_Data_CaseExecution{
				CaseExecution: e.Data.CaseExecution.Proto(),
			}
		}
	case DataTypeExecutionLog:
		if e.Data.ExecutionLog != nil {
			data.Data = &testv1.ExecutionEvent_Data_ExecutionLog{
				ExecutionLog: e.Data.ExecutionLog.Proto(),
			}
		}
	}

	return &testv1.ExecutionEvent{
		EventId:    uuid.NewString(),
		TestExecId: e.TestExecID.String(),
		Type:       e.Type.Proto(),
		Data:       data,
		CreateTime: timestamppb.New(e.CreateTime),
	}
}
