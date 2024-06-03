package workflowservice

import (
	testservicev1 "github.com/annexhq/annex-proto/gen/go/rpc/testservice/v1"
	"go.temporal.io/api/workflowservice/v1"
)

var _ workflowservice.WorkflowServiceServer = (*ProxyService)(nil)

type ProxyService struct {
	workflowservice.UnimplementedWorkflowServiceServer
	workflow workflowservice.WorkflowServiceClient
	test     testservicev1.TestServiceClient
}

func NewProxyService(
	testClient testservicev1.TestServiceClient,
	workflowClient workflowservice.WorkflowServiceClient,
) *ProxyService {
	return &ProxyService{
		test:     testClient,
		workflow: workflowClient,
	}
}
