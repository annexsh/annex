package testservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/test"

	"github.com/annexsh/annex/log"
)

const retryReason = "retry failed test execution"

type executor struct {
	repo     test.Repository
	eventPub event.Publisher
	temporal Workflower
	logger   log.Logger
}

func newExecutor(repo test.Repository, eventPub event.Publisher, workflower Workflower, logger log.Logger) *executor {
	return &executor{
		repo:     repo,
		eventPub: eventPub,
		temporal: workflower,
		logger:   logger,
	}
}

type executeOptions struct {
	payload *testsv1.Payload
}

type executeOption func(opts *executeOptions)

func withInput(input *testsv1.Payload) executeOption {
	return func(opts *executeOptions) {
		opts.payload = input
	}
}

func (e *executor) execute(ctx context.Context, testID uuid.UUID, opts ...executeOption) (*test.TestExecution, error) {
	t, err := e.repo.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	var options executeOptions
	for _, opt := range opts {
		opt(&options)
	}

	execID := test.NewTestExecutionID()
	workflowID := execID.WorkflowID()

	scheduled := &test.ScheduledTestExecution{
		ID:           execID,
		TestID:       t.ID,
		Payload:      nil, // TODO: payload metadata isn't saved: double check if it is needed
		ScheduleTime: time.Now(),
	}
	if options.payload != nil {
		if options.payload.Metadata == nil {
			return nil, errors.New("payload metadata cannot be nil")
		}
		scheduled.Payload = options.payload.Data
	}

	testExec, err := e.repo.CreateScheduledTestExecution(ctx, scheduled)
	if err != nil {
		return nil, err
	}

	execEvent := event.NewTestExecutionEvent(eventsv1.Event_TYPE_TEST_EXECUTION_SCHEDULED, testExec.Proto())
	if err = e.eventPub.Publish(testExec.ID.String(), execEvent); err != nil {
		return nil, fmt.Errorf("failed to publish test execution event: %w", err)
	}

	wfOpts := client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                getTaskQueue(t.ContextID, t.GroupID),
		WorkflowExecutionTimeout: 7 * 24 * time.Hour, // 1 week
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}

	if options.payload == nil {
		if _, err = e.temporal.ExecuteWorkflow(ctx, wfOpts, t.Name); err != nil {
			return nil, err
		}
	} else {
		if _, err = e.temporal.ExecuteWorkflow(ctx, wfOpts, t.Name, options.payload); err != nil {
			return nil, err
		}
	}

	return testExec, nil
}

// TODO: add safeguard to ensure reset point is after the start test execution signal
func (e *executor) retry(ctx context.Context, execID test.TestExecutionID) (*test.TestExecution, error) {
	testExec, err := e.repo.GetTestExecution(ctx, execID)
	if err != nil {
		return nil, err
	}

	origCaseExecs, err := e.repo.ListCaseExecutions(ctx, testExec.ID)
	if err != nil {
		return nil, err
	}

	origLogs, err := e.repo.ListLogs(ctx, testExec.ID)
	if err != nil {
		return nil, err
	}

	testLogsToDelete := mapset.NewSet[uuid.UUID]()
	caseLogsToDelete := map[test.CaseExecutionID][]uuid.UUID{}
	for _, l := range origLogs {
		if l.CaseExecutionID == nil {
			testLogsToDelete.Add(l.ID)
		} else {
			if caseLogs, ok := caseLogsToDelete[*l.CaseExecutionID]; ok {
				caseLogs = append(caseLogs, l.ID)
				caseLogsToDelete[*l.CaseExecutionID] = caseLogs
			} else {
				caseLogsToDelete[*l.CaseExecutionID] = []uuid.UUID{l.ID}
			}
		}
	}

	caseExecsToDelete := map[test.CaseExecutionID]*test.CaseExecution{}
	for _, c := range origCaseExecs {
		caseExecsToDelete[c.ID] = c
	}

	eventIDsToCaseIDs := map[int64]test.CaseExecutionID{}

	it := e.temporal.GetWorkflowHistory(ctx, testExec.ID.WorkflowID(), "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

	var resetCaseExec *test.CaseExecution
	var resetID int64

	for it.HasNext() {
		event, err := it.Next()
		if err != nil {
			return nil, err
		}

		if isFailedEvent(event.EventType) {
			if activityAttrs := event.GetActivityTaskFailedEventAttributes(); activityAttrs != nil {
				if caseID, ok := eventIDsToCaseIDs[activityAttrs.ScheduledEventId]; ok {
					if caseExec, ok := caseExecsToDelete[caseID]; ok {
						resetCaseExec = caseExec
						delete(caseExecsToDelete, caseID)
					}
				}
			} else if taskAttrs := event.GetWorkflowTaskFailedEventAttributes(); taskAttrs != nil {
				if taskAttrs.Failure.Message == retryReason {
					continue
				}
			}
			break
		}

		if isResettableEvent(event.EventType) {
			resetID = event.EventId
		}

		switch event.EventType {
		case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			attrs := event.GetActivityTaskScheduledEventAttributes()
			activityID, err := test.ParseCaseActivityID(attrs.ActivityId)
			if err != nil {
				return nil, err
			}
			eventIDsToCaseIDs[event.EventId] = activityID
		case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
			attrs := event.GetActivityTaskCompletedEventAttributes()
			// Completed without failure indicates history can be preserved
			if caseID, ok := eventIDsToCaseIDs[attrs.ScheduledEventId]; ok {
				delete(caseExecsToDelete, caseID)
				delete(caseLogsToDelete, caseID)
			}
		case enums.EVENT_TYPE_MARKER_RECORDED:
			attrs := event.GetMarkerRecordedEventAttributes()
			if attrs.MarkerName == "LocalActivity" && attrs.Details != nil {
				if data, ok := attrs.Details["result"]; ok {
					// Local activity should just have a single payload
					if len(data.Payloads) == 1 {
						var logResult struct{ LogID uuid.UUID }
						dc := converter.GetDefaultDataConverter()
						if err = dc.FromPayload(data.Payloads[0], &logResult); err != nil {
							return nil, err
						}
						testLogsToDelete.Remove(logResult.LogID)
					}
				}
			}
		}
	}

	if resetCaseExec != nil {
		caseExecsToDelete[resetCaseExec.ID] = resetCaseExec
	}

	logsToDelete := testLogsToDelete.ToSlice()
	for _, caseExecLogs := range caseLogsToDelete {
		logsToDelete = append(logsToDelete, caseExecLogs...)
	}

	reset, rollback, err := e.repo.ResetTestExecution(ctx, &test.ResetTestExecution{
		ID:                  testExec.ID,
		ResetTime:           time.Now().UTC().Add(25 * time.Hour),
		StaleCaseExecutions: keys(caseExecsToDelete),
		StaleLogs:           logsToDelete,
	})
	if err != nil {
		return nil, err
	}

	_, err = e.temporal.ResetWorkflowExecution(ctx, &workflowservice.ResetWorkflowExecutionRequest{
		Namespace: "default",
		WorkflowExecution: &common.WorkflowExecution{
			WorkflowId: testExec.ID.WorkflowID(),
		},
		Reason:                    retryReason,
		WorkflowTaskFinishEventId: resetID,
	})
	if err != nil {
		if rbErr := rollback(ctx); rbErr != nil {
			e.logger.Error("failed to rollback retry test execution after workflow failed to reset", "error", rbErr)
		}
		return nil, err
	}

	return reset, nil
}

func isResettableEvent(eventType enums.EventType) bool {
	switch eventType {
	case enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED,
		enums.EVENT_TYPE_WORKFLOW_TASK_TIMED_OUT,
		enums.EVENT_TYPE_WORKFLOW_TASK_FAILED,
		enums.EVENT_TYPE_WORKFLOW_TASK_STARTED:
		return true

	}
	return false
}

func isFailedEvent(eventType enums.EventType) bool {
	switch eventType {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_WORKFLOW_TASK_FAILED,
		enums.EVENT_TYPE_ACTIVITY_TASK_FAILED,
		enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED:
		return true

	}
	return false
}

func keys[T comparable, V any](in map[T]V) []T {
	out := make([]T, len(in))
	i := 0
	for k := range in {
		out[i] = k
		i++
	}
	return out
}
