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
		var cfg server.AllInOneConfig
		if err := server.LoadConfig(&cfg); err != nil {
			return err
		}
		return server.ServeAllInOne(ctx, cfg)
	case "test":
		var cfg server.TestServiceConfig
		if err := server.LoadConfig(&cfg); err != nil {
			return err
		}
		return server.ServeTestService(ctx, cfg)
	case "workflow-proxy":
		var cfg server.WorkflowProxyServiceConfig
		if err := server.LoadConfig(&cfg); err != nil {
			return err
		}
		return server.ServeWorkflowProxyService(ctx, cfg)
	}

	return errors.New("invalid server type")
}
