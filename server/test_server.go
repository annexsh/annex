package server

import (
	"context"
	"errors"
	"time"

	"github.com/annexsh/annex-proto/go/gen/annex/tests/v1/testsv1connect"
	"github.com/nats-io/nats-server/v2/server"
	corenats "github.com/nats-io/nats.go"
	"go.temporal.io/sdk/client"

	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/nats"
	"github.com/annexsh/annex/postgres"
	"github.com/annexsh/annex/testservice"
	"github.com/annexsh/annex/workflowservice"
)

func ServeTestService(ctx context.Context, cfg TestServiceConfig) error {
	logger := log.NewLogger("service", "test_service")
	srv := rpc.NewServer(getHostPort(cfg.Port))

	pgCfg := cfg.Postgres
	pgPool, err := postgres.OpenPool(ctx, pgCfg.User, pgCfg.Password, pgCfg.HostPort, postgres.WithMigration())
	if err != nil {
		return err
	}
	defer pgPool.Close()
	db := postgres.NewDB(pgPool)
	repo := postgres.NewTestRepository(db)

	if err := repo.CreateContext(ctx, "default"); err != nil {
		return err
	}

	nc, err := corenats.Connect(cfg.Nats.HostPort)
	if err != nil {
		return err
	}
	defer nc.Close()
	pubSub := nats.NewPubSub(nc, nats.WithLogger(logger))

	workflowProxyClient, err := client.NewLazyClient(client.Options{
		HostPort:  srv.GRPCAddress(),
		Namespace: workflowservice.Namespace,
		Logger:    logger.With("component", "temporal_client"),
	})
	if err != nil {
		return err
	}

	testSvc := testservice.New(repo, pubSub, workflowProxyClient, testservice.WithLogger(logger))
	path, handler := testsv1connect.NewTestServiceHandler(testSvc, rpc.WithConnectInterceptors(logger))
	srv.RegisterConnect(path, handler, cfg.CorsOrigins...)

	return serve(ctx, srv, logger)
}

func runEmbeddedNats(hostPort string) (*server.Server, error) {
	ns, err := nats.NewEmbeddedNatsServer(hostPort)
	if err != nil {
		return nil, err
	}
	go ns.Start()
	if !ns.ReadyForConnections(10 * time.Second) {
		return nil, errors.New("embedded nats server unhealthy")
	}
	return ns, nil
}
