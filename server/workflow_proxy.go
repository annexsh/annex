package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/annexsh/annex-proto/gen/go/annex/tests/v1/testsv1connect"
	workflowservicev1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/annexsh/annex/config"
	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/workflowservice"
)

func ServeWorkflowProxyService(ctx context.Context, cfg config.WorkflowProxyService) error {
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
	testClient := testsv1connect.NewTestServiceClient(httpClient, srv.ConnectAddress())

	workflowSvc := workflowservice.NewProxyService(testClient, temporalClient.WorkflowService())
	srv.RegisterGRPC(&workflowservicev1.WorkflowService_ServiceDesc, workflowSvc)
	srv.WithGRPCOptions(rpc.WithGRPCInterceptors(logger)...)

	return serve(ctx, srv, logger)
}
