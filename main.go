package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/annexsh/annex/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return errors.New("server type argument required")
	}

	srvType := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)

	switch srvType {
	case "all":
		cfg, err := server.LoadAllInOneConfig()
		if err != nil {
			return err
		}
		return server.ServeAllInOne(ctx, cfg)
	case "test":
		cfg, err := server.LoadTestServiceConfig()
		if err != nil {
			return err
		}
		return server.ServeTestService(ctx, cfg)
	case "event":
		cfg, err := server.LoadEventServiceConfig()
		if err != nil {
			return err
		}
		return server.ServeEventService(ctx, cfg)
	case "workflow-proxy":
		cfg, err := server.LoadWorkflowProxyServiceConfig()
		if err != nil {
			return err
		}
		return server.ServeWorkflowProxyService(ctx, cfg)
	}

	return errors.New("invalid server type")
}
