package server

import (
	"context"
	"fmt"

	"github.com/annexsh/annex-proto/gen/go/annex/events/v1/eventsv1connect"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/postgres"
)

func ServeEventService(ctx context.Context, cfg EventServiceConfig) error {
	logger := log.NewLogger("app", "annex-event-service")

	srv := rpc.NewServer(fmt.Sprint(":", cfg.Port))

	pgPool, err := postgres.OpenPool(ctx, cfg.Postgres.URL())
	if err != nil {
		return err
	}
	defer pgPool.Close()
	db := postgres.NewDB(pgPool)
	eventSrc, err := postgres.NewTestExecutionEventSource(ctx, pgPool)
	if err != nil {
		return err
	}
	go eventSrc.Start(ctx, pgEventSrcErrCallback(logger))
	defer eventSrc.Stop()
	repo := postgres.NewTestRepository(db)

	if err := repo.CreateContext(ctx, "default"); err != nil {
		return err
	}

	eventSvc := eventservice.New(eventSrc, repo, eventservice.WithLogger(logger))
	srv.RegisterConnect(eventsv1connect.NewEventServiceHandler(eventSvc, rpc.WithConnectInterceptors(logger)))

	return serve(ctx, srv, logger)
}
