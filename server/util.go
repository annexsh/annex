package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/lmittmann/tint"
	"github.com/temporalio/cli/temporalcli/devserver"

	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
	"github.com/annexsh/annex/postgres"
	"github.com/annexsh/annex/workflowservice"
)

func serve(ctx context.Context, srv *rpc.Server, logger log.Logger) error {
	logger.Info("starting server", "connect.address", srv.ConnectAddress(), "grpc.address", srv.GRPCAddress())

	srvErrCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(); err != nil {
			srvErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("context cancelled: stopping server")
		return nil
	case err := <-srvErrCh:
		return fmt.Errorf("server failed: %w", err)
	}
}

func setupTemporalDevServer() (*devserver.Server, string, error) {
	ip := "127.0.0.1"
	port := devserver.MustGetFreePort()
	address := fmt.Sprintf("%s:%d", ip, port)
	srv, err := devserver.Start(devserver.StartOptions{
		FrontendIP:             ip,
		FrontendPort:           port,
		UIIP:                   ip,
		UIPort:                 8233,
		Namespaces:             []string{workflowservice.Namespace},
		ClusterID:              uuid.NewString(),
		MasterClusterName:      "active",
		CurrentClusterName:     "active",
		InitialFailoverVersion: 1,
		Logger:                 slog.New(tint.NewHandler(os.Stdout, nil)),
		LogLevel:               slog.LevelWarn,
	})
	return srv, address, err
}

func pgEventSrcErrCallback(logger log.Logger) postgres.ErrorCallback {
	return func(err error) {
		logger.Error("event source message error", "error", err)
	}
}
