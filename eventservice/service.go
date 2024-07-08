package eventservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/gen/go/annex/events/v1"
	"github.com/annexsh/annex-proto/gen/go/annex/events/v1/eventsv1connect"
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"
	mapset "github.com/deckarep/golang-set/v2"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/log"
)

var _ eventsv1connect.EventServiceHandler = (*Service)(nil)

type EventSubscriber interface {
	Subscribe(testExecID string) (<-chan *eventsv1.Event, func(), error)
}

type ExecutionFetcher interface {
	GetTestExecution(ctx context.Context, req *connect.Request[testsv1.GetTestExecutionRequest]) (*connect.Response[testsv1.GetTestExecutionResponse], error)
	ListCaseExecutions(ctx context.Context, req *connect.Request[testsv1.ListCaseExecutionsRequest]) (*connect.Response[testsv1.ListCaseExecutionsResponse], error)
	ListTestExecutionLogs(ctx context.Context, req *connect.Request[testsv1.ListTestExecutionLogsRequest]) (*connect.Response[testsv1.ListTestExecutionLogsResponse], error)
}

type ServiceOption func(s *Service)

func WithLogger(logger log.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

type Service struct {
	subscriber  EventSubscriber
	execFetcher ExecutionFetcher
	logger      log.Logger
}

func New(subscriber EventSubscriber, execFetcher ExecutionFetcher, opts ...ServiceOption) *Service {
	s := &Service{
		subscriber:  subscriber,
		execFetcher: execFetcher,
		logger:      log.DefaultLogger(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) StreamTestExecutionEvents(
	ctx context.Context,
	req *connect.Request[eventsv1.StreamTestExecutionEventsRequest],
	stream *connect.ServerStream[eventsv1.StreamTestExecutionEventsResponse],
) error {
	testExecID := req.Msg.TestExecutionId
	contextID := req.Msg.Context

	testExecRes, err := s.execFetcher.GetTestExecution(ctx, connect.NewRequest(&testsv1.GetTestExecutionRequest{
		Context:         contextID,
		TestExecutionId: testExecID,
	}))
	if err != nil {
		return fmt.Errorf("failed to get test execution: %w", err)
	}
	testExec := testExecRes.Msg.TestExecution

	sub, unsub, err := s.subscriber.Subscribe(testExecID)
	if err != nil {
		return fmt.Errorf("failed to subscribe to test execution events: %w", err)
	}
	defer unsub()

	// Produce events for cases that existed prior to stream
	existingEvents, err := s.getExistingEvents(ctx, contextID, testExec)
	if err != nil {
		return err
	}

	seenEventIDs := mapset.NewSet[string]()
	for _, existing := range existingEvents {
		if err = stream.Send(&eventsv1.StreamTestExecutionEventsResponse{
			Event: existing,
		}); err != nil {
			return err
		}
		seenEventIDs.Add(existing.EventId)
		if existing.Type == eventsv1.Event_TYPE_TEST_EXECUTION_FINISHED {
			return nil
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-sub:
			if !ok {
				return nil
			}
			if !seenEventIDs.Contains(e.EventId) {
				if err = stream.Send(&eventsv1.StreamTestExecutionEventsResponse{
					Event: e,
				}); err != nil {
					return err
				}
			}
			if e.Type == eventsv1.Event_TYPE_TEST_EXECUTION_FINISHED {
				return nil
			}
		}
	}
}

func (s *Service) getExistingEvents(ctx context.Context, contextID string, testExec *testsv1.TestExecution) ([]*eventsv1.Event, error) {
	events := []*eventsv1.Event{event.NewTestExecutionEvent(eventsv1.Event_TYPE_TEST_EXECUTION_SCHEDULED, testExec)}

	if testExec.StartTime == nil {
		return events, nil
	}
	events = append(events, event.NewTestExecutionEvent(eventsv1.Event_TYPE_TEST_EXECUTION_STARTED, testExec))

	caseExecsRes, err := s.execFetcher.ListCaseExecutions(ctx, connect.NewRequest(&testsv1.ListCaseExecutionsRequest{
		Context:         contextID,
		TestExecutionId: testExec.Id,
	}))
	if err != nil {
		return nil, err
	}
	currCaseExecs := caseExecsRes.Msg.CaseExecutions

	testLogEvents, caseLogEventsMap, err := s.getLogEvents(ctx, contextID, testExec.Id)
	if err != nil {
		return nil, err
	}

	// Queue all current test execution log events if no case executions exist
	if len(currCaseExecs) == 0 {
		for _, testLogEvent := range testLogEvents {
			events = append(events, testLogEvent)
		}
	}

	for _, caseExec := range currCaseExecs {
		// Queue all test execution logs events before the current case
		for {
			if len(testLogEvents) == 0 {
				break
			}
			nextTestLogEvent := testLogEvents[0]
			logData := nextTestLogEvent.Data.Data.(*eventsv1.Event_Data_Log)
			if logData.Log.CreateTime.AsTime().After(caseExec.ScheduleTime.AsTime()) {
				break
			}
			events = append(events, nextTestLogEvent)
			testLogEvents = testLogEvents[1:]
		}

		events = append(events, event.NewCaseExecutionEvent(eventsv1.Event_TYPE_CASE_EXECUTION_SCHEDULED, caseExec))

		if caseExec.StartTime != nil {
			events = append(events, event.NewCaseExecutionEvent(eventsv1.Event_TYPE_CASE_EXECUTION_STARTED, caseExec))
		}

		caseLogEvents, ok := caseLogEventsMap[caseExec.Id]
		if ok {
			for _, logEvent := range caseLogEvents {
				events = append(events, logEvent)
			}
		}

		if caseExec.FinishTime != nil {
			events = append(events, event.NewCaseExecutionEvent(eventsv1.Event_TYPE_CASE_EXECUTION_FINISHED, caseExec))
		}
	}

	// Queue all current test execution log events after the last case execution
	if len(testLogEvents) > 0 {
		for _, testLogEvent := range testLogEvents {
			events = append(events, testLogEvent)
		}
	}

	if testExec.FinishTime != nil {
		events = append(events, event.NewTestExecutionEvent(eventsv1.Event_TYPE_TEST_EXECUTION_FINISHED, testExec))
	}

	return events, nil
}

func (s *Service) getLogEvents(ctx context.Context, contextID string, testExecID string) ([]*eventsv1.Event, map[int32][]*eventsv1.Event, error) {
	logsRes, err := s.execFetcher.ListTestExecutionLogs(ctx, connect.NewRequest(&testsv1.ListTestExecutionLogsRequest{
		Context:         contextID,
		TestExecutionId: testExecID,
	}))
	if err != nil {
		return nil, nil, err
	}
	currLogs := logsRes.Msg.Logs

	var testLogEvents []*eventsv1.Event
	caseLogEvents := map[int32][]*eventsv1.Event{}

	for _, log := range currLogs {
		logEvent := event.NewLogEvent(eventsv1.Event_TYPE_LOG_PUBLISHED, log)

		if log.CaseExecutionId == nil {
			testLogEvents = append(testLogEvents, logEvent)
		} else {
			caseLogEvent := logEvent
			if got, ok := caseLogEvents[*log.CaseExecutionId]; ok {
				got = append(got, caseLogEvent)
				caseLogEvents[*log.CaseExecutionId] = got
			} else {
				caseLogEvents[*log.CaseExecutionId] = []*eventsv1.Event{caseLogEvent}
			}
		}
	}

	return testLogEvents, caseLogEvents, nil
}
