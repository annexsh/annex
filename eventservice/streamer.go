package eventservice

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/test"

	"github.com/annexsh/annex/log"
)

const eventStreamBufferSize = 50

type streamer struct {
	eventSrc   event.Source
	execReader ExecutionReader
	logger     log.Logger
}

func newStreamer(eventSrc event.Source, execReader ExecutionReader, logger log.Logger) *streamer {
	return &streamer{
		eventSrc:   eventSrc,
		execReader: execReader,
		logger:     logger,
	}
}

func (s *streamer) streamTestExecutionEvents(ctx context.Context, id test.TestExecutionID) (<-chan *event.ExecutionEvent, <-chan error) {
	out := make(chan *event.ExecutionEvent, eventStreamBufferSize)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		testExec, err := s.execReader.GetTestExecution(ctx, id)
		if err != nil {
			errs <- fmt.Errorf("failed to get test execution: %w", err)
			return
		}

		sub, unsub := s.eventSrc.Subscribe(id)
		defer unsub()

		// Produce events for cases that existed prior to stream
		existingEvents, err := s.getExistingEvents(ctx, testExec)
		if err != nil {
			errs <- err
			return
		}
		seenEventIDs := mapset.NewSet[uuid.UUID]()
		for _, existing := range existingEvents {
			out <- existing
			seenEventIDs.Add(existing.ID)
			if existing.Type == event.TypeTestExecutionFinished {
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-sub:
				if !ok {
					return
				}
				if !seenEventIDs.Contains(e.ID) {
					out <- e
				}
				if e.Type == event.TypeTestExecutionFinished {
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
) ([]*event.ExecutionEvent, error) {
	events := []*event.ExecutionEvent{event.NewTestExecutionEvent(event.TypeTestExecutionScheduled, testExec)}

	if testExec.StartTime == nil {
		return events, nil
	}
	events = append(events, event.NewTestExecutionEvent(event.TypeTestExecutionStarted, testExec))

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

		events = append(events, event.NewCaseExecutionEvent(event.TypeCaseExecutionScheduled, caseExec))

		if caseExec.StartTime != nil {
			events = append(events, event.NewCaseExecutionEvent(event.TypeCaseExecutionStarted, caseExec))
		}

		caseLogEvents, ok := caseLogEventsMap[caseExec.ID]
		if ok {
			for _, logEvent := range caseLogEvents {
				events = append(events, logEvent)
			}
		}

		if caseExec.FinishTime != nil {
			events = append(events, event.NewCaseExecutionEvent(event.TypeCaseExecutionFinished, caseExec))
		}
	}

	// Queue all current test execution log events after the last case execution
	if len(testLogEvents) > 0 {
		for _, testLogEvent := range testLogEvents {
			events = append(events, testLogEvent)
		}
	}

	if testExec.FinishTime != nil {
		events = append(events, event.NewTestExecutionEvent(event.TypeTestExecutionFinished, testExec))
	}

	return events, nil
}

func (s *streamer) getLogEvents(ctx context.Context, testExecID test.TestExecutionID) (
	[]*event.ExecutionEvent,
	map[test.CaseExecutionID][]*event.ExecutionEvent,
	error,
) {
	currLogs, err := s.execReader.ListLogs(ctx, testExecID)
	if err != nil {
		return nil, nil, err
	}

	var testLogEvents []*event.ExecutionEvent
	caseLogEvents := map[test.CaseExecutionID][]*event.ExecutionEvent{}

	for _, log := range currLogs {
		logEvent := event.NewLogEvent(event.TypeLogPublished, log)

		if log.CaseExecutionID == nil {
			testLogEvents = append(testLogEvents, logEvent)
		} else {
			caseLogEvent := logEvent
			if got, ok := caseLogEvents[*log.CaseExecutionID]; ok {
				got = append(got, caseLogEvent)
				caseLogEvents[*log.CaseExecutionID] = got
			} else {
				caseLogEvents[*log.CaseExecutionID] = []*event.ExecutionEvent{caseLogEvent}
			}
		}
	}

	return testLogEvents, caseLogEvents, nil
}
