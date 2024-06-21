package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/annexsh/annex-proto/gen/go/annex/events/v1/eventsv1connect"
	"github.com/annexsh/annex-proto/gen/go/annex/tests/v1/testsv1connect"
	workflowservicev1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/internal/health"
	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/testservice"
	"github.com/annexsh/annex/workflowservice"
)

type Option func(opts *serverOptions)

type serverOptions struct {
	logger log.Logger
}

func WithLogger(logger log.Logger) Option {
	return func(opts *serverOptions) {
		opts.logger = logger
	}
}

func Start(ctx context.Context, cfg Config, opts ...Option) error {
	var options serverOptions
	for _, opt := range opts {
		opt(&options)
	}

	logger := log.NewLogger()
	if options.logger != nil {
		logger = options.logger
	}

	var deps *dependencies
	var err error

	if cfg.InMemory {
		deps = setupInMemoryDeps(ctx)
	} else {
		if deps, err = setupPostgresDeps(ctx, cfg.Postgres.URL(), cfg.Postgres.SchemaVersion); err != nil {
			return err
		}
	}

	defer deps.close()

	if err = deps.repo.CreateContext(ctx, "default"); err != nil {
		return err
	}

	if cfg.Temporal.LocalDev {
		temporalSrv, hostPort, err := setupTemporalDevServer(cfg.Temporal.Namespace)
		if err != nil {
			return fmt.Errorf("failed to start temporal dev server")
		}
		defer temporalSrv.Stop()
		cfg.Temporal.HostPort = hostPort
	}

	srv, err := newServer(ctx, cfg, deps, logger)
	if err != nil {
		return err
	}

	srvErrs := make(chan error, 1)
	go func() {
		logger.Info("starting server",
			"connect.address", srv.ConnectAddress(),
			"grpc.address", srv.GRPCAddress(),
		)
		if serveErr := srv.Serve(); err != nil {
			srvErrs <- serveErr
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("context cancelled: stopping server")
		return nil
	case err = <-deps.errs:
		return fmt.Errorf("dependency failed: %w", err)
	case err = <-srvErrs:
		return fmt.Errorf("server failed: %w", err)
	}
}

func newServer(ctx context.Context, cfg Config, deps *dependencies, logger log.Logger) (*rpc.Server, error) {
	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}

	hostPort := fmt.Sprint(":", cfg.Port)
	srv := rpc.NewServer(hostPort)

	// Connect

	testSvc := testservice.New(deps.repo, temporalClient, testservice.WithLogger(logger))
	eventSvc := eventservice.NewService(deps.eventSrc, deps.repo)

	connectOps := []connect.HandlerOption{rpc.WithConnectInterceptors(logger)}
	srv.RegisterConnect(testsv1connect.NewTestServiceHandler(testSvc, connectOps...))
	srv.RegisterConnect(eventsv1connect.NewEventServiceHandler(eventSvc, connectOps...))

	// gRPC

	testClient := testsv1connect.NewTestServiceClient(
		&http.Client{Timeout: 30 * time.Second},
		srv.ConnectAddress(),
	)
	healthSvc, err := health.NewGRPCService(ctx, health.Config{
		ServiceNames: []string{health.ServiceNameTest, health.ServiceNameEvent},
		Dependencies: append(deps.healthChecks, health.WithTemporal(temporalClient, cfg.Temporal.Namespace)),
		Logger:       logger,
	})
	if err != nil {
		return nil, err
	}

	workflowSvc := workflowservice.NewProxyService(testClient, temporalClient.WorkflowService())
	srv.RegisterGRPC(&workflowservicev1.WorkflowService_ServiceDesc, workflowSvc)
	srv.RegisterGRPC(&grpchealthv1.Health_ServiceDesc, healthSvc)
	srv.WithGRPCOptions(rpc.WithGRPCInterceptors(logger)...)

	return srv, nil
}
