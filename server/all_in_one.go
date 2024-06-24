package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/annexsh/annex-proto/gen/go/annex/events/v1/eventsv1connect"
	"github.com/annexsh/annex-proto/gen/go/annex/tests/v1/testsv1connect"
	"github.com/jackc/pgx/v5/pgxpool"
	workflowservicev1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/inmem"
	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/postgres"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/testservice"
	"github.com/annexsh/annex/workflowservice"
)

func ServeAllInOne(ctx context.Context, cfg AllInOneConfig) error {
	var logger log.Logger
	if cfg.Development.Logger {
		logger = log.NewDevLogger("app", "annex")
	}

	srv := rpc.NewServer(fmt.Sprint(":", cfg.Port))

	var pgPool *pgxpool.Pool
	var repo test.Repository
	var eventSrc event.Source
	var err error

	// Repository

	if cfg.Development.InMemory {
		db := inmem.NewDB()
		eventSrc = db.TestExecutionEventSource()
		repo = inmem.NewTestRepository(db)
	} else if cfg.Repository.Postgres != nil {
		pgPool, err = postgres.OpenPool(ctx, cfg.Repository.Postgres.URL(),
			postgres.WithMigration(cfg.Repository.Postgres.SchemaVersion),
		)
		if err != nil {
			return err
		}
		defer pgPool.Close()

		db := postgres.NewDB(pgPool)
		repo = postgres.NewTestRepository(db)
	} else {
		return errors.New("repository config required")
	}

	if err := repo.CreateContext(ctx, "default"); err != nil {
		return err
	}

	// Event source

	if eventSrc == nil {
		if cfg.EventSource.Postgres != nil {
			if pgPool == nil {
				pgPool, err = postgres.OpenPool(ctx, cfg.Repository.Postgres.URL())
				if err != nil {
					return err
				}
				defer pgPool.Close()
			}
			pgES, err := postgres.NewTestExecutionEventSource(ctx, pgPool)
			if err != nil {
				return err
			}
			go pgES.Start(ctx, pgEventSrcErrCallback(logger))
			defer pgES.Stop()
			eventSrc = pgES
		} else {
			return errors.New("event source config required")
		}
	}

	// Temporal

	if cfg.Development.Temporal {
		temporalDevSrv, temporalAddr, err := setupTemporalDevServer()
		if err != nil {
			return err
		}
		cfg.Temporal.HostPort = temporalAddr
		defer temporalDevSrv.Stop()
	}

	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: workflowservice.Namespace,
	})
	if err != nil {
		return err
	}

	workflowProxyClient, err := client.NewLazyClient(client.Options{
		HostPort:  srv.GRPCAddress(),
		Namespace: workflowservice.Namespace,
	})
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	testClient := testsv1connect.NewTestServiceClient(httpClient, srv.ConnectAddress())

	// Test service
	testSvc := testservice.New(repo, workflowProxyClient, testservice.WithLogger(logger))
	srv.RegisterConnect(testsv1connect.NewTestServiceHandler(testSvc, rpc.WithConnectInterceptors(logger)))
	// Event service
	eventSvc := eventservice.New(eventSrc, repo, eventservice.WithLogger(logger))
	srv.RegisterConnect(eventsv1connect.NewEventServiceHandler(eventSvc, rpc.WithConnectInterceptors(logger)))
	// Workflow Proxy service
	workflowSvc := workflowservice.NewProxyService(testClient, temporalClient.WorkflowService())
	srv.RegisterGRPC(&workflowservicev1.WorkflowService_ServiceDesc, workflowSvc)
	healthSvc := health.NewServer()
	healthSvc.SetServingStatus(workflowservicev1.WorkflowService_ServiceDesc.ServiceName, grpchealthv1.HealthCheckResponse_SERVING)
	srv.RegisterGRPC(&grpchealthv1.Health_ServiceDesc, healthSvc)
	srv.WithGRPCOptions(rpc.WithGRPCInterceptors(logger)...)

	return serve(ctx, srv, logger)
}
