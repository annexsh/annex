package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/annexsh/annex/config"
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
	case config.TypeAll:
		var cfg config.AllServices
		if err := config.Load(&cfg); err != nil {
			return err
		}
		return server.ServeAllInOne(ctx, cfg)
	case config.TypeTest:
		var cfg config.TestService
		if err := config.Load(&cfg); err != nil {
			return err
		}
		return server.ServeTestService(ctx, cfg)
	case config.TypeWorkflowProxy:
		var cfg config.WorkflowProxyService
		if err := config.Load(&cfg); err != nil {
			return err
		}
		return server.ServeWorkflowProxyService(ctx, cfg)
	}

	return errors.New("invalid server type")
}
