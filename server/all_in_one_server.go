package server

import (
	"context"
	"net/http"
	"time"

	"github.com/annexsh/annex-proto/go/gen/annex/events/v1/eventsv1connect"
	"github.com/annexsh/annex-proto/go/gen/annex/tests/v1/testsv1connect"
	"github.com/jackc/pgx/v5/pgxpool"
	corenats "github.com/nats-io/nats.go"
	workflowservicev1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/nats"
	"github.com/annexsh/annex/postgres"
	"github.com/annexsh/annex/sqlite"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/testservice"
	"github.com/annexsh/annex/workflowservice"
)

func ServeAllInOne(ctx context.Context, cfg AllInOneConfig) error {
	logger := log.NewDevLogger()
	if cfg.StructuredLogging {
		logger = log.NewLogger()
	}

	srv := rpc.NewServer(getHostPort(cfg.Port))

	var pgPool *pgxpool.Pool
	var repo test.Repository
	var err error

	// Repository

	if cfg.SQLite {
		db, err := sqlite.Open(sqlite.WithMigration())
		if err != nil {
			return err
		}
		defer db.Close()
		repo = sqlite.NewTestRepository(sqlite.NewDB(db))
		logger.Info("sqlite db created")
	} else {
		pgCfg := cfg.Postgres
		pgPool, err = postgres.OpenPool(ctx, pgCfg.User, pgCfg.Password, pgCfg.HostPort, postgres.WithMigration())
		if err != nil {
			return err
		}
		defer pgPool.Close()
		repo = postgres.NewTestRepository(postgres.NewDB(pgPool))
		logger.Info("postgres db created")
	}

	if err := repo.CreateContext(ctx, "default"); err != nil {
		return err
	}

	// Pub/Sub

	var nc *corenats.Conn
	if cfg.Nats.Embedded {
		ns, err := runEmbeddedNats(cfg.Nats.HostPort)
		if err != nil {
			return err
		}
		defer ns.Shutdown()
		nc, err = corenats.Connect("", corenats.InProcessServer(ns))
		if err != nil {
			return err
		}
	} else {
		nc, err = corenats.Connect(cfg.Nats.HostPort)
		if err != nil {
			return err
		}
	}
	defer nc.Close()
	pubSub := nats.NewPubSub(nc, nats.WithLogger(logger))

	// Test service

	testSvcLogger := logger.With("service", "test_service")
	workflowProxyClient, err := client.NewLazyClient(client.Options{
		HostPort:  srv.GRPCAddress(),
		Namespace: workflowservice.Namespace,
		Logger:    testSvcLogger.With("component", "temporal_client"),
	})
	if err != nil {
		return err
	}

	testSvc := testservice.New(repo, pubSub, workflowProxyClient, testservice.WithLogger(testSvcLogger))
	testPath, testHandler := testsv1connect.NewTestServiceHandler(testSvc, rpc.WithConnectInterceptors(testSvcLogger))
	srv.RegisterConnect(testPath, testHandler, cfg.CorsOrigins...)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	testClient := testsv1connect.NewTestServiceClient(httpClient, srv.ConnectAddress())

	// Event service

	eventSvcLogger := logger.With("service", "event_service")
	eventSvc := eventservice.New(pubSub, testClient, eventservice.WithLogger(eventSvcLogger))
	eventPath, eventHandler := eventsv1connect.NewEventServiceHandler(eventSvc, rpc.WithConnectInterceptors(eventSvcLogger))
	srv.RegisterConnect(eventPath, eventHandler, cfg.CorsOrigins...)

	// Workflow Proxy service
	wfProxySvcLogger := logger.With("service", "workflow_proxy_service")
	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: workflowservice.Namespace,
		Logger:    wfProxySvcLogger.With("component", "temporal_client"),
	})
	if err != nil {
		return err
	}

	workflowSvc := workflowservice.NewProxyService(testClient, temporalClient.WorkflowService())
	srv.RegisterGRPC(&workflowservicev1.WorkflowService_ServiceDesc, workflowSvc)

	// Misc

	healthSvc := health.NewServer()
	healthSvc.SetServingStatus(workflowservicev1.WorkflowService_ServiceDesc.ServiceName, grpchealthv1.HealthCheckResponse_SERVING)
	srv.RegisterGRPC(&grpchealthv1.Health_ServiceDesc, healthSvc)
	srv.WithGRPCOptions(rpc.WithGRPCInterceptors(logger)...)

	return serve(ctx, srv, logger)
}
