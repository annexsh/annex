package event

import (
	"context"

	eventservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/eventservice/v1"

	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/test"
)

var _ eventservicev1.EventServiceServer = (*Service)(nil)

type EventSource interface {
	Subscribe(testExecID test.TestExecutionID) (sub <-chan *ExecutionEvent, unsub conc.Unsubscribe)
}

type ExecutionReader interface {
	GetTestExecution(ctx context.Context, id test.TestExecutionID) (*test.TestExecution, error)
	ListCaseExecutions(ctx context.Context, id test.TestExecutionID) (test.CaseExecutionList, error)
	ListExecutionLogs(ctx context.Context, id test.TestExecutionID) (test.ExecutionLogList, error)
}

type ServiceOption func(s *Service)

func WithLogger(logger log.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

func With(logger log.Logger) ServiceOption {
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

func (s Service) StreamTestExecutionEvents(req *eventservicev1.StreamTestExecutionEventsRequest, stream eventservicev1.EventService_StreamTestExecutionEventsServer) error {
	ctx := stream.Context()

	testExecID, err := test.ParseTestExecutionID(req.TestExecId)
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
			if err = stream.Send(&eventservicev1.StreamTestExecutionEventsResponse{
				Event: event.Proto(),
			}); err != nil {
				return err
			}
		case err = <-errs:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
