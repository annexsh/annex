package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/annexsh/annex-proto/gen/go/annex/tests/v1/testsv1connect"
	workflowservicev1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/workflowservice"
)

func ServeWorkflowProxyService(ctx context.Context, cfg WorkflowProxyServiceConfig) error {
	logger := log.NewLogger("app", "annex-workflow-proxy-service")

	srv := rpc.NewServer(fmt.Sprint(":", cfg.Port))

	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: workflowservice.Namespace,
	})
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	testClient := testsv1connect.NewTestServiceClient(httpClient, cfg.TestServiceURL)

	workflowSvc := workflowservice.NewProxyService(testClient, temporalClient.WorkflowService())
	srv.RegisterGRPC(&workflowservicev1.WorkflowService_ServiceDesc, workflowSvc)
	healthSvc := health.NewServer()
	healthSvc.SetServingStatus(workflowservicev1.WorkflowService_ServiceDesc.ServiceName, grpchealthv1.HealthCheckResponse_SERVING)
	srv.RegisterGRPC(&grpchealthv1.Health_ServiceDesc, healthSvc)
	srv.WithGRPCOptions(rpc.WithGRPCInterceptors(logger)...)

	return serve(ctx, srv, logger)
}
