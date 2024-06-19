package eventservice

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	"github.com/annexsh/annex/test"

	"github.com/annexsh/annex/log"
)

const eventStreamBufferSize = 50

type streamer struct {
	eventSource EventSource
	execReader  ExecutionReader
	logger      log.Logger
}

func newStreamer(eventSource EventSource, execReader ExecutionReader, logger log.Logger) *streamer {
	return &streamer{
		eventSource: eventSource,
		execReader:  execReader,
		logger:      logger,
	}
}

func (s *streamer) streamTestExecutionEvents(ctx context.Context, id test.TestExecutionID) (<-chan *ExecutionEvent, <-chan error) {
	out := make(chan *ExecutionEvent, eventStreamBufferSize)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		testExec, err := s.execReader.GetTestExecution(ctx, id)
		if err != nil {
			errs <- fmt.Errorf("failed to get test execution: %w", err)
		}

		sub, unsub := s.eventSource.Subscribe(id)
		defer unsub()

		// Produce events for cases that existed prior to stream
		existingEvents, err := s.getExistingEvents(ctx, testExec)
		if err != nil {
			errs <- err
			return
		}
		seenEventIDs := mapset.NewSet[uuid.UUID]()
		for _, event := range existingEvents {
			out <- event
			seenEventIDs.Add(event.ID)
			if event.Type == TypeTestExecutionFinished {
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-sub:
				if !ok {
					return
				}
				if !seenEventIDs.Contains(event.ID) {
					out <- event
				}
				if event.Type == TypeTestExecutionFinished {
					return
				}
			}
		}
	}()

	return out, errs
}

func (s *streamer) getExistingEvents(
	ctx context.Context,
	testExec *test.TestExecution,
) ([]*ExecutionEvent, error) {
	events := []*ExecutionEvent{NewTestExecutionEvent(TypeTestExecutionScheduled, testExec)}

	if testExec.StartTime == nil {
		return events, nil
	}
	events = append(events, NewTestExecutionEvent(TypeTestExecutionStarted, testExec))

	currCaseExecs, err := s.execReader.ListCaseExecutions(ctx, testExec.ID)
	if err != nil {
		return nil, err
	}

	testLogEvents, caseLogEventsMap, err := s.getLogEvents(ctx, testExec.ID)
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
			if nextTestLogEvent.Data.Log.CreateTime.After(caseExec.ScheduleTime) {
				break
			}
			events = append(events, nextTestLogEvent)
			testLogEvents = testLogEvents[1:]
		}

		events = append(events, NewCaseExecutionEvent(TypeCaseExecutionScheduled, caseExec))

		if caseExec.StartTime != nil {
			events = append(events, NewCaseExecutionEvent(TypeCaseExecutionStarted, caseExec))
		}

		caseLogEvents, ok := caseLogEventsMap[caseExec.ID]
		if ok {
			for _, logEvent := range caseLogEvents {
				events = append(events, logEvent)
			}
		}

		if caseExec.FinishTime != nil {
			events = append(events, NewCaseExecutionEvent(TypeCaseExecutionFinished, caseExec))
		}
	}

	// Queue all current test execution log events after the last case execution
	if len(testLogEvents) > 0 {
		for _, testLogEvent := range testLogEvents {
			events = append(events, testLogEvent)
		}
	}

	if testExec.FinishTime != nil {
		events = append(events, NewTestExecutionEvent(TypeTestExecutionFinished, testExec))
	}

	return events, nil
}

func (s *streamer) getLogEvents(ctx context.Context, testExecID test.TestExecutionID) (
	[]*ExecutionEvent,
	map[test.CaseExecutionID][]*ExecutionEvent,
	error,
) {
	currLogs, err := s.execReader.ListLogs(ctx, testExecID)
	if err != nil {
		return nil, nil, err
	}

	var testLogEvents []*ExecutionEvent
	caseLogEvents := map[test.CaseExecutionID][]*ExecutionEvent{}

	for _, log := range currLogs {
		event := NewLogEvent(TypeLogPublished, log)

		if log.CaseExecutionID == nil {
			testLogEvents = append(testLogEvents, event)
		} else {
			caseLogEvent := event
			if got, ok := caseLogEvents[*log.CaseExecutionID]; ok {
				got = append(got, caseLogEvent)
				caseLogEvents[*log.CaseExecutionID] = got
			} else {
				caseLogEvents[*log.CaseExecutionID] = []*ExecutionEvent{caseLogEvent}
			}
		}
	}

	return testLogEvents, caseLogEvents, nil
}
