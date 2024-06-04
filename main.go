package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/annexsh/annex/server"

	"github.com/annexsh/annex/log"
)

const appName = "annex"

func main() {
	cfg, err := server.LoadConfig(server.WithYAML())
	if err != nil {
		log.DefaultLogger().Error("unable to load config", "component", "main", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := log.NewLogger(cfg.Env, "app", appName)
	if cfg.Env == server.EnvLocal {
		logger = log.NewDevLogger()
	}

	if err = server.Start(ctx, cfg, server.WithLogger(logger)); err != nil {
		logger.Error("annex fatal error", "error", err)
		os.Exit(1)
	}
}
