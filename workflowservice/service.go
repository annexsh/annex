package workflowservice

import (
	"github.com/annexsh/annex-proto/go/gen/annex/tests/v1/testsv1connect"
	"go.temporal.io/api/workflowservice/v1"
)

const Namespace = "default"

var _ workflowservice.WorkflowServiceServer = (*ProxyService)(nil)

type ProxyService struct {
	workflowservice.UnimplementedWorkflowServiceServer
	workflow workflowservice.WorkflowServiceClient
	test     testsv1connect.TestServiceClient
}

func NewProxyService(
	testClient testsv1connect.TestServiceClient,
	workflowClient workflowservice.WorkflowServiceClient,
) *ProxyService {
	return &ProxyService{
		test:     testClient,
		workflow: workflowClient,
	}
}
