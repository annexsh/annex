package testservice

//go:generate go run github.com/matryer/moq@latest -out workflower_mock_test.go . Workflower
//go:generate go run github.com/matryer/moq@latest -out repository_mock_test.go -pkg testservice ../test Repository
//go:generate go run github.com/matryer/moq@latest -out event_publisher_mock_test.go -pkg testservice ../event Publisher

import (
	"context"

	"github.com/annexsh/annex-proto/go/gen/annex/tests/v1/testsv1connect"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/test"
)

var _ testsv1connect.TestServiceHandler = (*Service)(nil)

type Workflower interface {
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error)
	GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun
	GetWorkflowHistory(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType enums.HistoryEventFilterType) client.HistoryEventIterator
	ResetWorkflowExecution(ctx context.Context, request *workflowservice.ResetWorkflowExecutionRequest) (*workflowservice.ResetWorkflowExecutionResponse, error)
	CancelWorkflow(ctx context.Context, workflowID string, runID string) error
	DescribeTaskQueue(ctx context.Context, taskQueue string, taskQueueType enums.TaskQueueType) (*workflowservice.DescribeTaskQueueResponse, error)
}

type ServiceOption func(s *Service)

func WithLogger(logger log.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

type Service struct {
	repo       test.Repository
	eventPub   event.Publisher
	workflower Workflower
	executor   *executor
	logger     log.Logger
}

func New(repo test.Repository, eventPub event.Publisher, workflower Workflower, opts ...ServiceOption) *Service {
	s := &Service{
		repo:       repo,
		eventPub:   eventPub,
		workflower: workflower,
		logger:     log.NewNopLogger(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.executor = newExecutor(repo, eventPub, workflower, s.logger)
	return s
}
