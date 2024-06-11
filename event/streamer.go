package event

import (
	"context"
	"errors"
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

		h, err := s.newEventHandler(ctx, testExec, out)
		if err != nil {
			errs <- err
			return
		}

		if h.isFinished && h.pendingCases.Cardinality() == 0 { // stop early if all cases notified
			out <- h.finishedEvent
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-sub:
				if !ok {
					return
				}

				var handlerFn func(event *ExecutionEvent) error

				switch event.Type {
				case TypeTestExecutionScheduled, TypeTestExecutionStarted, TypeTestExecutionFinished:
					handlerFn = h.handleTestExecution
				case TypeCaseExecutionScheduled, TypeCaseExecutionStarted, TypeCaseExecutionFinished:
					handlerFn = h.handleCase
				case TypeLogPublished:
					handlerFn = h.handleLog
				default:
					s.logger.Error("unknown test execution event received", "event", event)
				}

				if err = handlerFn(event); err != nil {
					errs <- err
					return
				}

				if h.isFinished && h.pendingCases.Cardinality() == 0 {
					out <- h.finishedEvent
					return
				}
			}
		}
	}()

	return out, errs
}

func (s *streamer) newEventHandler(ctx context.Context, testExec *test.TestExecution, out chan *ExecutionEvent) (*eventHandler, error) {
	// Produce events for cases that existed prior to stream
	pendingCases, finishedCases, seenLogs, err := s.sendExistingEvents(ctx, out, testExec.ID)
	if err != nil {
		return nil, err
	}

	handler := &eventHandler{
		out:           out,
		pendingCases:  pendingCases,
		finishedCases: finishedCases,
		seenLogs:      seenLogs,
	}

	if testExec.FinishTime != nil {
		handler.isFinished = true
		handler.finishedEvent = NewTestExecutionEvent(TypeTestExecutionFinished, testExec)
	}

	return handler, nil
}

func (s *streamer) sendExistingEvents(
	ctx context.Context,
	stream chan<- *ExecutionEvent,
	testExecID test.TestExecutionID,
) (mapset.Set[test.CaseExecutionID], mapset.Set[test.CaseExecutionID], mapset.Set[uuid.UUID], error) {
	currCaseExecs, err := s.execReader.ListCaseExecutions(ctx, testExecID)
	if err != nil {
		return nil, nil, nil, err
	}

	testLogEvents, caseLogEventsMap, seenLogs, err := s.getLogEvents(ctx, testExecID)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, event := range testLogEvents {
		stream <- event
	}

	pendingCases := mapset.NewSet[test.CaseExecutionID]()
	finishedCases := mapset.NewSet[test.CaseExecutionID]()

	for _, caseExec := range currCaseExecs {
		events, ok := caseLogEventsMap[caseExec.ID]
		if !ok {
			events = []*ExecutionEvent{}
		}

		if caseExec.FinishTime == nil {
			// Must be added before log events
			events = append([]*ExecutionEvent{NewCaseExecutionEvent(TypeCaseExecutionStarted, caseExec)}, events...)
			pendingCases.Add(caseExec.ID)
		} else {
			// Must be added after log events
			events = append([]*ExecutionEvent{NewCaseExecutionEvent(TypeCaseExecutionFinished, caseExec)}, events...)
			pendingCases.Remove(caseExec.ID)
			finishedCases.Add(caseExec.ID)
		}

		for _, caseEvent := range events {
			stream <- caseEvent
		}
	}

	return pendingCases, finishedCases, seenLogs, nil
}

func (s *streamer) getLogEvents(ctx context.Context, testExecID test.TestExecutionID) (
	[]*ExecutionEvent,
	map[test.CaseExecutionID][]*ExecutionEvent,
	mapset.Set[uuid.UUID],
	error,
) {
	currLogs, err := s.execReader.ListLogs(ctx, testExecID)
	if err != nil {
		return nil, nil, nil, err
	}

	var testLogs []*ExecutionEvent
	caseLogs := map[test.CaseExecutionID][]*ExecutionEvent{}
	seenLogs := mapset.NewSet[uuid.UUID]()

	for _, log := range currLogs {
		seenLogs.Add(log.ID)
		event := NewLogEvent(TypeLogPublished, log)

		if log.CaseExecutionID == nil {
			testLogs = append(testLogs, event)
		} else {
			caseLogEvent := event
			if got, ok := caseLogs[*log.CaseExecutionID]; ok {
				got = append(got, caseLogEvent)
				caseLogs[*log.CaseExecutionID] = got
			} else {
				caseLogs[*log.CaseExecutionID] = []*ExecutionEvent{caseLogEvent}
			}
		}
	}

	return testLogs, caseLogs, seenLogs, nil
}

type eventHandler struct {
	out chan<- *ExecutionEvent
	//stream        eventservicev1.EventService_StreamTestExecutionEventsServer
	pendingCases  mapset.Set[test.CaseExecutionID]
	finishedCases mapset.Set[test.CaseExecutionID]
	seenLogs      mapset.Set[uuid.UUID]
	isFinished    bool
	finishedEvent *ExecutionEvent
}

func (h *eventHandler) handleTestExecution(event *ExecutionEvent) error {
	if _, err := event.Data.GetTestExecution(); err != nil {
		return err // ensure event has valid data
	}
	switch event.Type {
	case TypeTestExecutionScheduled:
		// TODO
	case TypeTestExecutionStarted:
		// TODO
	case TypeTestExecutionFinished:
		h.finishedEvent = event
		h.isFinished = true
	default:
		return errors.New("unknown test execution event type")
	}
	return nil
}

func (h *eventHandler) handleCase(event *ExecutionEvent) error {
	caseExec, err := event.Data.GetCaseExecution()
	if err != nil {
		return err
	}

	switch event.Type {
	case TypeCaseExecutionScheduled:
		// TODO
	case TypeCaseExecutionStarted:
		if h.pendingCases.Contains(caseExec.ID) {
			return nil
		}
		h.pendingCases.Add(caseExec.ID)
	case TypeCaseExecutionFinished:
		if h.finishedCases.Contains(caseExec.ID) {
			return nil
		}
		h.pendingCases.Remove(caseExec.ID)
		h.finishedCases.Add(caseExec.ID)
	default:
		return errors.New("unknown case execution event type")
	}

	h.out <- event
	return nil
}

func (h *eventHandler) handleLog(event *ExecutionEvent) error {
	execLog, err := event.Data.GetLog()
	if err != nil {
		return err
	}
	if h.seenLogs.Contains(execLog.ID) {
		return nil
	}
	h.out <- event
	return nil
}
