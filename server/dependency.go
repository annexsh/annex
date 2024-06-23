package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/lmittmann/tint"
	"github.com/temporalio/cli/temporalcli/devserver"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/inmem"
	"github.com/annexsh/annex/test"

	"github.com/annexsh/annex/internal/health"
	"github.com/annexsh/annex/postgres"
)

type dependencies struct {
	repo         test.Repository
	eventSrc     eventservice.EventSource
	healthChecks []health.DependencyChecker
	errs         <-chan error
	close        func()
}

func setupPostgresDeps(ctx context.Context, url string, schemaVersion uint) (*dependencies, error) {
	pgPool, err := postgres.OpenPool(ctx, url, schemaVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection for %s: %w", url, err)
	}

	deps := &dependencies{
		healthChecks: []health.DependencyChecker{health.WithPostgres(pgPool)},
	}

	db := postgres.NewDB(pgPool)
	deps.repo = postgres.NewTestRepository(db)

	eventSrc, err := postgres.NewTestExecutionEventSource(ctx, pgPool)
	if err != nil {
		return nil, err
	}

	deps.errs = eventSrc.Start(ctx)
	deps.eventSrc = eventSrc
	deps.close = func() {
		eventSrc.Stop()
		pgPool.Close()
	}

	return deps, nil
}

// Temporary option during initial development phase
func setupInMemoryDeps(ctx context.Context) *dependencies {
	deps := &dependencies{
		errs: make(chan error), // will never be published to
	}

	db := inmem.NewDB()
	deps.repo = inmem.NewTestRepository(db)
	eventSrc := db.TestExecutionEventSource()
	eventSrc.Start(ctx)
	deps.eventSrc = eventSrc
	deps.close = eventSrc.Stop

	return deps
}

// Temporary option during initial development phase
func setupTemporalDevServer(namespace string) (*devserver.Server, string, error) {
	ip := "127.0.0.1"
	port := devserver.MustGetFreePort()
	address := fmt.Sprintf("%s:%d", ip, port)
	srv, err := devserver.Start(devserver.StartOptions{
		FrontendIP:             ip,
		FrontendPort:           port,
		UIIP:                   ip,
		UIPort:                 8233,
		Namespaces:             []string{namespace},
		ClusterID:              uuid.NewString(),
		MasterClusterName:      "active",
		CurrentClusterName:     "active",
		InitialFailoverVersion: 1,
		Logger:                 slog.New(tint.NewHandler(os.Stdout, nil)),
		LogLevel:               slog.LevelWarn,
	})
	return srv, address, err
}
