package eventservice

import (
	"context"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/gen/go/annex/events/v1"
	"github.com/annexsh/annex-proto/gen/go/annex/events/v1/eventsv1connect"

	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/test"
)

var _ eventsv1connect.EventServiceHandler = (*Service)(nil)

type EventSource interface {
	Subscribe(testExecID test.TestExecutionID) (sub <-chan *ExecutionEvent, unsub conc.Unsubscribe)
}

type ExecutionReader interface {
	GetTestExecution(ctx context.Context, id test.TestExecutionID) (*test.TestExecution, error)
	ListCaseExecutions(ctx context.Context, id test.TestExecutionID) (test.CaseExecutionList, error)
	ListLogs(ctx context.Context, id test.TestExecutionID) (test.LogList, error)
}

type ServiceOption func(s *Service)

func WithLogger(logger log.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

type Service struct {
	streamer *streamer
	logger   log.Logger
}

func NewService(eventSource EventSource, execReader ExecutionReader, opts ...ServiceOption) *Service {
	s := &Service{logger: log.DefaultLogger()}
	s.streamer = newStreamer(eventSource, execReader, s.logger)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s Service) StreamTestExecutionEvents(
	ctx context.Context,
	req *connect.Request[eventsv1.StreamTestExecutionEventsRequest],
	stream *connect.ServerStream[eventsv1.StreamTestExecutionEventsResponse],
) error {
	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return err
	}

	events, errs := s.streamer.streamTestExecutionEvents(ctx, testExecID)
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if err = stream.Send(&eventsv1.StreamTestExecutionEventsResponse{
				Event: event.Proto(),
			}); err != nil {
				return err
			}
		case err, ok := <-errs:
			if ok {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
