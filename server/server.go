package server

import (
	"context"
	"fmt"
	"net"

	eventservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/eventservice/v1"
	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	workflowservicev1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/grpcsrv"
	"github.com/annexsh/annex/internal/health"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/testservice"
	"github.com/annexsh/annex/workflowservice"
)

type Option func(opts *serverOptions)

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

	tc, err := client.NewLazyClient(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}

	srv := grpcsrv.New(logger)
	reflection.Register(srv)

	healthSvc, err := health.NewGRPCService(ctx, health.Config{
		ServiceNames: []string{health.ServiceNameTest, health.ServiceNameEvent},
		Dependencies: append(deps.healthChecks, health.WithTemporal(tc, cfg.Temporal.Namespace)),
		Logger:       logger,
	})
	if err != nil {
		return err
	}

	testConn, err := grpc.NewClient(fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	testClient := testservicev1.NewTestServiceClient(testConn)

	grpchealthv1.RegisterHealthServer(srv, healthSvc)
	testservicev1.RegisterTestServiceServer(srv, testservice.New(deps.repo, tc, testservice.WithLogger(logger)))
	eventservicev1.RegisterEventServiceServer(srv, event.NewService(deps.eventSrc, deps.repo))
	workflowservicev1.RegisterWorkflowServiceServer(srv, workflowservice.NewProxyService(testClient, tc.WorkflowService()))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", cfg.Port, err)
	}
	defer lis.Close()

	sErrs := make(chan error, 1)
	go func() {
		if serveErr := srv.Serve(lis); err != nil {
			sErrs <- serveErr
		}
	}()
	logger.Info("serving grpc server", "address", lis.Addr().String())

	defer func() {
		logger.Info("waiting for all active grpc connections to close")
		srv.GracefulStop()
		logger.Info("grpc server stopped")
	}()

	select {
	case <-ctx.Done():
		logger.Info("context cancelled: stopping grpc server")
		return nil
	case err = <-deps.errs:
		return fmt.Errorf("dependency failed: %w", err)
	case err = <-sErrs:
		return fmt.Errorf("grpc server failed: %w", err)
	}
}

type serverOptions struct {
	logger log.Logger
}
