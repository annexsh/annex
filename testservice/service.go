package testservice

import (
	"context"

	testservicev1 "github.com/annexhq/annex-proto/gen/go/rpc/testservice/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/annexhq/annex/test"

	"github.com/annexhq/annex/log"
)

const defaultPageSize int32 = 200

var _ testservicev1.TestServiceServer = (*Service)(nil)

type Workflower interface {
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error)
	GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun
	SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error
	GetWorkflowHistory(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType enums.HistoryEventFilterType) client.HistoryEventIterator
	ResetWorkflowExecution(ctx context.Context, request *workflowservice.ResetWorkflowExecutionRequest) (*workflowservice.ResetWorkflowExecutionResponse, error)
	CancelWorkflow(ctx context.Context, workflowID string, runID string) error
}

type ServiceOption func(s *Service)

func WithLogger(logger log.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

type Service struct {
	repo     test.Repository
	executor *executor
	logger   log.Logger
}

func New(repo test.Repository, workflower Workflower, opts ...ServiceOption) *Service {
	s := &Service{
		repo:   repo,
		logger: log.NewNopLogger(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.executor = newExecutor(repo, workflower, s.logger)
	return s
}
