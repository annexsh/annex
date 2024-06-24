package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/annexsh/annex-proto/gen/go/annex/tests/v1/testsv1connect"
	"go.temporal.io/sdk/client"

	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/postgres"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/testservice"
	"github.com/annexsh/annex/workflowservice"
)

func ServeTestService(ctx context.Context, cfg TestServiceConfig) error {
	logger := log.NewLogger("app", "annex-test-service")

	srv := rpc.NewServer(fmt.Sprint(":", cfg.Port))

	var repo test.Repository
	if cfg.Repository.Postgres != nil {
		pgPool, err := postgres.OpenPool(ctx, cfg.Repository.Postgres.URL(),
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

	workflowProxyClient, err := client.NewLazyClient(client.Options{
		HostPort:  srv.GRPCAddress(),
		Namespace: workflowservice.Namespace,
	})
	if err != nil {
		return err
	}

	testSvc := testservice.New(repo, workflowProxyClient, testservice.WithLogger(logger))
	srv.RegisterConnect(testsv1connect.NewTestServiceHandler(testSvc, rpc.WithConnectInterceptors(logger)))

	return serve(ctx, srv, logger)
}
