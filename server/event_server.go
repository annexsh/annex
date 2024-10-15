package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/annexsh/annex-proto/go/gen/annex/events/v1/eventsv1connect"
	"github.com/annexsh/annex-proto/go/gen/annex/tests/v1/testsv1connect"
	corenats "github.com/nats-io/nats.go"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/nats"
)

func ServeEventService(ctx context.Context, cfg EventServiceConfig) error {
	logger := log.NewLogger("app", "annex-event-service")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	testClient := testsv1connect.NewTestServiceClient(httpClient, cfg.TestServiceURL)

	var natsOpts []corenats.Option
	if cfg.Nats.Embedded {
		ns, err := runEmbeddedNats(cfg.Nats.HostPort)
		if err != nil {
			return err
		}
		defer ns.Shutdown()
		natsOpts = append(natsOpts, corenats.InProcessServer(ns))
	}

	var nc *corenats.Conn
	var err error
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

	eventSvc := eventservice.New(pubSub, testClient, eventservice.WithLogger(logger))

	srv := rpc.NewServer(fmt.Sprint(":", cfg.Port))
	path, handler := eventsv1connect.NewEventServiceHandler(eventSvc, rpc.WithConnectInterceptors(logger))
	srv.RegisterConnect(path, handler, cfg.CorsOrigins...)

	return serve(ctx, srv, logger)
}
