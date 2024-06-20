package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/annexsh/annex-proto/gen/go/annex/events/v1/eventsv1connect"
	"github.com/annexsh/annex-proto/gen/go/annex/tests/v1/testsv1connect"
	workflowservicev1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/internal/grpcsrv"
	"github.com/annexsh/annex/internal/health"
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

	closer, errs := serve(ctx, cfg, deps, logger)
	defer closer()

	select {
	case <-ctx.Done():
		logger.Info("context cancelled: stopping server")
		return nil
	case err = <-deps.errs:
		return fmt.Errorf("dependency failed: %w", err)
	case err = <-errs:
		return fmt.Errorf("server failed: %w", err)
	}
}

type serverCloser func()

func serve(ctx context.Context, cfg Config, deps *dependencies, logger log.Logger) (serverCloser, <-chan error) {
	hostPort := fmt.Sprint(":", cfg.Port)
	mux := http.NewServeMux()
	errs := make(chan error, 1)

	// Clients

	testClient := testsv1connect.NewTestServiceClient(
		&http.Client{Timeout: 30 * time.Second},
		"http://"+hostPort+"/connect",
	)

	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		errs <- fmt.Errorf("failed to create temporal client: %w", err)
		return nil, errs
	}

	// Connect

	connectOps := []connect.HandlerOption{
		connect.WithInterceptors(grpcsrv.NewConnectLogInterceptor(logger)),
	}

	testSvc := testservice.New(deps.repo, temporalClient, testservice.WithLogger(logger))
	testsPath, testsHandler := testsv1connect.NewTestServiceHandler(testSvc, connectOps...)
	mux.Handle("/connect"+testsPath, http.StripPrefix("/connect", testsHandler))

	eventSvc := eventservice.NewService(deps.eventSrc, deps.repo)
	eventsPath, eventsHandler := eventsv1connect.NewEventServiceHandler(eventSvc, connectOps...)
	mux.Handle("/connect"+eventsPath, http.StripPrefix("/connect", eventsHandler))

	reflector := grpcreflect.NewStaticReflector(
		testsv1connect.TestServiceName,
		eventsv1connect.EventServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// gRPC

	healthSvc, err := health.NewGRPCService(ctx, health.Config{
		ServiceNames: []string{health.ServiceNameTest, health.ServiceNameEvent},
		Dependencies: append(deps.healthChecks, health.WithTemporal(temporalClient, cfg.Temporal.Namespace)),
		Logger:       logger,
	})
	if err != nil {
		errs <- err
		return nil, errs
	}

	grpcSrv := grpc.NewServer()
	grpchealthv1.RegisterHealthServer(grpcSrv, healthSvc)
	workflowSvc := workflowservice.NewProxyService(testClient, temporalClient.WorkflowService())
	workflowservicev1.RegisterWorkflowServiceServer(grpcSrv, workflowSvc)
	reflection.Register(grpcSrv)
	mux.Handle("/", grpcSrv)

	// Start server

	srv := &http.Server{
		Addr:    hostPort,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	go func() {
		logger.Info("serving connect and grpc apis", "address", hostPort)
		if serveErr := srv.ListenAndServe(); err != nil {
			errs <- serveErr
		}
	}()

	closer := func() {
		logger.Info("gracefully stopping grpc server")
		grpcSrv.GracefulStop()
		logger.Info("gracefully stopping http server")
		closeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if closeErr := srv.Shutdown(closeCtx); closeErr != nil {
			logger.Error("failed to gracefully stop http server", "error", closeErr)
		}
	}

	return closer, errs
}
