package fake

import (
	"time"

	testv1 "github.com/annexhq/annex-proto/gen/go/type/test/v1"
	"github.com/google/uuid"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/sdk/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexhq/annex/test"
)

var _ client.HistoryEventIterator = (*iterator)(nil)

type iterator struct {
	events  chan *history.HistoryEvent
	nextErr error
}

func (i *iterator) HasNext() bool {
	return len(i.events) > 0
}

func (i *iterator) Next() (*history.HistoryEvent, error) {
	if i.nextErr != nil {
		return nil, i.nextErr
	}
	select {
	case event := <-i.events:
		return event, nil
	default:
		return nil, nil
	}
}

func parseTime(timeStr string) *timestamppb.Timestamp {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic("failed to parse time " + timeStr)
	}
	return timestamppb.New(t)
}

// GenCaseFailureHistory generates history inspired by real workflow events for
// a test execution.
//
// The history is based on the following test execution actions:
// - Test execution start
// - Test execution publish log
// - Case 1 execution start (case logs not stored in history)
// - Case 1 execution finish - success
// - Case 2 execution start (case logs not stored in history)
// - Case 2 execution finish - error
// - Test execution finish - error (case 2)
func GenCaseFailureHistory(
	testExecID test.TestExecutionID,
	testExecLogID uuid.UUID, // log published by event 7 local activity
	successCaseExecID test.CaseExecutionID,
	failureCaseExecID test.CaseExecutionID,
) *history.History {
	logResult := struct{ LogID uuid.UUID }{LogID: testExecLogID}

	dc := converter.GetDefaultDataConverter()
	logPayload, err := dc.ToPayload(logResult)
	if err != nil {
		panic("failed to encode test exec log result payload: " + err.Error())
	}

	return &history.History{
		Events: []*history.HistoryEvent{
			{
				EventId:   1,
				EventTime: parseTime("2024-05-28T10:04:10.551328Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
				TaskId:    1048605,
				Attributes: &history.HistoryEvent_WorkflowExecutionStartedEventAttributes{
					WorkflowExecutionStartedEventAttributes: &history.WorkflowExecutionStartedEventAttributes{
						WorkflowType:                    &common.WorkflowType{Name: "FakeTest"},
						TaskQueue:                       &taskqueue.TaskQueue{Name: "default", Kind: enums.TASK_QUEUE_KIND_NORMAL},
						WorkflowExecutionTimeout:        durationpb.New(1000 * time.Second),
						WorkflowRunTimeout:              durationpb.New(1000 * time.Second),
						WorkflowTaskTimeout:             durationpb.New(10 * time.Second),
						OriginalExecutionRunId:          "fe41eee8-bf64-483b-b594-0dcb350502a1",
						Identity:                        "hidden",
						FirstExecutionRunId:             "fe41eee8-bf64-483b-b594-0dcb350502a1",
						Attempt:                         1,
						WorkflowExecutionExpirationTime: parseTime("2024-06-04T10:04:10.551Z"),
						WorkflowId:                      testExecID.WorkflowID(),
					},
				},
			},
			{
				EventId:   2,
				EventTime: parseTime("2024-05-28T10:04:10.551441Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
				TaskId:    1048606,
				Attributes: &history.HistoryEvent_WorkflowTaskScheduledEventAttributes{
					WorkflowTaskScheduledEventAttributes: &history.WorkflowTaskScheduledEventAttributes{
						TaskQueue:           &taskqueue.TaskQueue{Name: "default", Kind: enums.TASK_QUEUE_KIND_NORMAL},
						StartToCloseTimeout: durationpb.New(10 * time.Second),
						Attempt:             1,
					},
				},
			},
			{
				EventId:   3,
				EventTime: parseTime("2024-05-28T10:04:10.569149Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED,
				TaskId:    1048612,
				Attributes: &history.HistoryEvent_WorkflowExecutionSignaledEventAttributes{
					WorkflowExecutionSignaledEventAttributes: &history.WorkflowExecutionSignaledEventAttributes{
						SignalName: testv1.TestSignal_TEST_SIGNAL_START_TEST.String(),
						Input: &common.Payloads{
							Payloads: []*common.Payload{
								{
									Metadata: map[string][]byte{"encoding": []byte("YmluYXJ5L251bGw=")},
								},
							},
						},
						Identity: "hidden",
						Header:   &common.Header{},
					},
				},
			},
			{
				EventId:   4,
				EventTime: parseTime("2024-05-28T10:04:10.579381Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_STARTED,
				TaskId:    1048614,
				Attributes: &history.HistoryEvent_WorkflowTaskStartedEventAttributes{
					WorkflowTaskStartedEventAttributes: &history.WorkflowTaskStartedEventAttributes{
						ScheduledEventId: 2,
						Identity:         "hidden",
						RequestId:        "a27534d0-057c-457c-b586-222094bd4f3b",
						HistorySizeBytes: 414,
					},
				},
			},
			{
				EventId:   5,
				EventTime: parseTime("2024-05-28T10:04:10.596592Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED,
				TaskId:    1048618,
				Attributes: &history.HistoryEvent_WorkflowTaskCompletedEventAttributes{
					WorkflowTaskCompletedEventAttributes: &history.WorkflowTaskCompletedEventAttributes{
						ScheduledEventId: 2,
						StartedEventId:   4,
						Identity:         "hidden",
						WorkerVersion: &common.WorkerVersionStamp{
							BuildId: "81194cb0e5cc14ffb23da13a22ce4152",
						},
						SdkMetadata:      &sdk.WorkflowTaskCompletedMetadata{LangUsedFlags: []uint32{3}, SdkName: "temporal-go", SdkVersion: "1.26.0"},
						MeteringMetadata: &common.MeteringMetadata{},
					},
				},
			},
			{
				EventId:   6,
				EventTime: parseTime("2024-05-28T10:04:10.596617Z"),
				EventType: enums.EVENT_TYPE_TIMER_STARTED,
				TaskId:    1048619,
				Attributes: &history.HistoryEvent_TimerStartedEventAttributes{
					TimerStartedEventAttributes: &history.TimerStartedEventAttributes{
						TimerId:                      "6",
						StartToFireTimeout:           durationpb.New(30 * time.Second),
						WorkflowTaskCompletedEventId: 5,
					},
				},
			},
			{
				EventId:   7,
				EventTime: parseTime("2024-05-28T10:04:10.596623Z"),
				EventType: enums.EVENT_TYPE_MARKER_RECORDED,
				TaskId:    1048620,
				Attributes: &history.HistoryEvent_MarkerRecordedEventAttributes{
					MarkerRecordedEventAttributes: &history.MarkerRecordedEventAttributes{
						MarkerName: "LocalActivity",
						Details: map[string]*common.Payloads{
							"data": {
								Payloads: []*common.Payload{
									{
										Metadata: map[string][]byte{"encoding": []byte("anNvbi9wbGFpbg==")},
										Data:     []byte("eyJBY3Rpdml0eUlEIjoiMSIsIkFjdGl2aXR5VHlwZSI6IlB1Ymxpc2giLCJSZXBsYXlUaW1lIjoiMjAyNC0wNS0yOFQxMDowNDoxMC41ODQ2NjM1WiIsIkF0dGVtcHQiOjEsIkJhY2tvZmYiOjB9"),
									},
								},
							},
							"result": {
								Payloads: []*common.Payload{
									logPayload, // test execution log published by local activity
								},
							},
						},
						WorkflowTaskCompletedEventId: 5,
					},
				},
			},
			{
				EventId:   8,
				EventTime: parseTime("2024-05-28T10:04:10.596632Z"),
				EventType: enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
				TaskId:    1048621,
				Attributes: &history.HistoryEvent_ActivityTaskScheduledEventAttributes{
					ActivityTaskScheduledEventAttributes: &history.ActivityTaskScheduledEventAttributes{
						ActivityId: successCaseExecID.ActivityID(),
						ActivityType: &common.ActivityType{
							Name: "SuccessCase",
						},
						TaskQueue: &taskqueue.TaskQueue{
							Name: "default",
							Kind: enums.TASK_QUEUE_KIND_NORMAL,
						},
						Header:                       &common.Header{},
						ScheduleToCloseTimeout:       durationpb.New(60 * time.Second),
						ScheduleToStartTimeout:       durationpb.New(60 * time.Second),
						StartToCloseTimeout:          durationpb.New(60 * time.Second),
						HeartbeatTimeout:             durationpb.New(0),
						WorkflowTaskCompletedEventId: 5,
						RetryPolicy: &common.RetryPolicy{
							InitialInterval:    durationpb.New(time.Second),
							BackoffCoefficient: 2,
							MaximumInterval:    durationpb.New(100 * time.Second),
							MaximumAttempts:    1,
						},
						UseCompatibleVersion: true,
					},
				},
			},
			{
				EventId:   9,
				EventTime: parseTime("2024-05-28T10:04:10.603343Z"),
				EventType: enums.EVENT_TYPE_ACTIVITY_TASK_STARTED,
				TaskId:    1048629,
				Attributes: &history.HistoryEvent_ActivityTaskStartedEventAttributes{
					ActivityTaskStartedEventAttributes: &history.ActivityTaskStartedEventAttributes{
						ScheduledEventId: 8,
						Identity:         "hidden",
						RequestId:        "a7f20551-57c4-49e5-92f5-9f11c87a836b",
						Attempt:          1,
					},
				},
			},
			{
				EventId:   10,
				EventTime: parseTime("2024-05-28T10:04:11.649503Z"),
				EventType: enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED,
				TaskId:    1048630,
				Attributes: &history.HistoryEvent_ActivityTaskCompletedEventAttributes{
					ActivityTaskCompletedEventAttributes: &history.ActivityTaskCompletedEventAttributes{
						Result: &common.Payloads{
							Payloads: []*common.Payload{
								{
									Metadata: map[string][]byte{"encoding": []byte("anNvbi9wbGFpbg==")},
									Data:     []byte("eyJSZXN1bHQiOm51bGwsIkZpbmlzaGVkIjoiMjAyNC0wNS0yOFQyMDowNDoxMS42NDI2OTgrMTA6MDAiLCJEdXJhdGlvbiI6MTAwMzcyMTg3NX0="),
								},
							},
						},
						ScheduledEventId: 8,
						StartedEventId:   9,
						Identity:         "hidden",
					},
				},
			},
			{
				EventId:   11,
				EventTime: parseTime("2024-05-28T10:04:11.649515Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
				TaskId:    1048631,
				Attributes: &history.HistoryEvent_WorkflowTaskScheduledEventAttributes{
					WorkflowTaskScheduledEventAttributes: &history.WorkflowTaskScheduledEventAttributes{
						TaskQueue: &taskqueue.TaskQueue{
							Name:       "hidden",
							Kind:       enums.TASK_QUEUE_KIND_STICKY,
							NormalName: "default",
						},
						StartToCloseTimeout: durationpb.New(10 * time.Second),
						Attempt:             1,
					},
				},
			},
			{
				EventId:   12,
				EventTime: parseTime("2024-05-28T10:04:11.662908Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_STARTED,
				TaskId:    1048635,
				Attributes: &history.HistoryEvent_WorkflowTaskStartedEventAttributes{
					WorkflowTaskStartedEventAttributes: &history.WorkflowTaskStartedEventAttributes{
						ScheduledEventId: 11,
						Identity:         "hidden",
						RequestId:        "1b374228-7af8-451e-9797-b59b8af00313",
						HistorySizeBytes: 1466,
					},
				},
			},
			{
				EventId:   13,
				EventTime: parseTime("2024-05-28T10:04:11.666495Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED,
				TaskId:    1048639,
				Attributes: &history.HistoryEvent_WorkflowTaskCompletedEventAttributes{
					WorkflowTaskCompletedEventAttributes: &history.WorkflowTaskCompletedEventAttributes{
						ScheduledEventId: 11,
						StartedEventId:   12,
						Identity:         "hidden",
						WorkerVersion: &common.WorkerVersionStamp{
							BuildId: "81194cb0e5cc14ffb23da13a22ce4152",
						},
						SdkMetadata:      &sdk.WorkflowTaskCompletedMetadata{},
						MeteringMetadata: &common.MeteringMetadata{},
					},
				},
			},
			{
				EventId:   14,
				EventTime: parseTime("2024-05-28T10:04:11.666520Z"),
				EventType: enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
				TaskId:    1048640,
				Attributes: &history.HistoryEvent_ActivityTaskScheduledEventAttributes{
					ActivityTaskScheduledEventAttributes: &history.ActivityTaskScheduledEventAttributes{
						ActivityId: failureCaseExecID.ActivityID(),
						ActivityType: &common.ActivityType{
							Name: "FailureCase",
						},
						TaskQueue: &taskqueue.TaskQueue{
							Name: "default",
							Kind: enums.TASK_QUEUE_KIND_NORMAL,
						},
						Header:                       &common.Header{},
						ScheduleToCloseTimeout:       durationpb.New(60 * time.Second),
						ScheduleToStartTimeout:       durationpb.New(60 * time.Second),
						StartToCloseTimeout:          durationpb.New(60 * time.Second),
						HeartbeatTimeout:             durationpb.New(0),
						WorkflowTaskCompletedEventId: 13,
						RetryPolicy: &common.RetryPolicy{
							InitialInterval:    durationpb.New(time.Second),
							BackoffCoefficient: 2,
							MaximumInterval:    durationpb.New(time.Second),
							MaximumAttempts:    1,
						},
						UseCompatibleVersion: true,
					},
				},
			},
			{
				EventId:   15,
				EventTime: parseTime("2024-05-28T10:04:11.670613Z"),
				EventType: enums.EVENT_TYPE_ACTIVITY_TASK_STARTED,
				TaskId:    1048646,
				Attributes: &history.HistoryEvent_ActivityTaskStartedEventAttributes{
					ActivityTaskStartedEventAttributes: &history.ActivityTaskStartedEventAttributes{
						ScheduledEventId: 14,
						Identity:         "hidden",
						RequestId:        "bee10f15-2991-47f5-82c4-2be605848abd",
						Attempt:          1,
					},
				},
			},
			{
				EventId:   16,
				EventTime: parseTime("2024-05-28T10:04:12.679629Z"),
				EventType: enums.EVENT_TYPE_ACTIVITY_TASK_FAILED,
				TaskId:    1048647,
				Attributes: &history.HistoryEvent_ActivityTaskFailedEventAttributes{
					ActivityTaskFailedEventAttributes: &history.ActivityTaskFailedEventAttributes{
						Failure: &failure.Failure{
							Message: "case execution failed: Assertion failed",
							Source:  "GoSDK",
							Cause: &failure.Failure{
								Message: "Assertion failed",
								Source:  "GoSDK",
								FailureInfo: &failure.Failure_ApplicationFailureInfo{
									ApplicationFailureInfo: &failure.ApplicationFailureInfo{},
								},
							},
							FailureInfo: &failure.Failure_ApplicationFailureInfo{
								ApplicationFailureInfo: &failure.ApplicationFailureInfo{
									Type: "wrapError",
								},
							},
						},
						ScheduledEventId: 14,
						StartedEventId:   15,
						Identity:         "hidden",
						RetryState:       enums.RETRY_STATE_MAXIMUM_ATTEMPTS_REACHED,
					},
				},
			},
			{
				EventId:   17,
				EventTime: parseTime("2024-05-28T10:04:12.679638Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
				TaskId:    1048648,
				Attributes: &history.HistoryEvent_WorkflowTaskScheduledEventAttributes{
					WorkflowTaskScheduledEventAttributes: &history.WorkflowTaskScheduledEventAttributes{
						TaskQueue: &taskqueue.TaskQueue{
							Name:       "hidden",
							Kind:       enums.TASK_QUEUE_KIND_STICKY,
							NormalName: "default",
						},
						StartToCloseTimeout: durationpb.New(10 * time.Second),
						Attempt:             1,
					},
				},
			},
			{
				EventId:   18,
				EventTime: parseTime("2024-05-28T10:04:12.692857Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_STARTED,
				TaskId:    1048652,
				Attributes: &history.HistoryEvent_WorkflowTaskStartedEventAttributes{
					WorkflowTaskStartedEventAttributes: &history.WorkflowTaskStartedEventAttributes{
						ScheduledEventId: 17,
						Identity:         "hidden",
						RequestId:        "4e6645eb-08bd-4379-99f3-db2e15fff8d4",
						HistorySizeBytes: 2153,
					},
				},
			},
			{
				EventId:   19,
				EventTime: parseTime("2024-05-28T10:04:12.698410Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED,
				TaskId:    1048656,
				Attributes: &history.HistoryEvent_WorkflowTaskCompletedEventAttributes{
					WorkflowTaskCompletedEventAttributes: &history.WorkflowTaskCompletedEventAttributes{
						ScheduledEventId: 17,
						StartedEventId:   18,
						Identity:         "hidden",
						WorkerVersion: &common.WorkerVersionStamp{
							BuildId: "81194cb0e5cc14ffb23da13a22ce4152",
						},
						SdkMetadata:      &sdk.WorkflowTaskCompletedMetadata{},
						MeteringMetadata: &common.MeteringMetadata{},
					},
				},
			},
			{
				EventId:   20,
				EventTime: parseTime("2024-05-28T10:04:12.698499Z"),
				EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
				TaskId:    1048657,
				Attributes: &history.HistoryEvent_WorkflowExecutionFailedEventAttributes{
					WorkflowExecutionFailedEventAttributes: &history.WorkflowExecutionFailedEventAttributes{
						Failure: &failure.Failure{
							Message: "Assertion failed",
							Source:  "GoSDK",
							FailureInfo: &failure.Failure_ApplicationFailureInfo{
								ApplicationFailureInfo: &failure.ApplicationFailureInfo{},
							},
						},
						RetryState:                   enums.RETRY_STATE_RETRY_POLICY_NOT_SET,
						WorkflowTaskCompletedEventId: 19,
					},
				},
			},
		},
	}
}
