package server

import (
	"context"
	"fmt"

	"github.com/annexsh/annex/internal/rpc"
	"github.com/annexsh/annex/log"
)

func serve(ctx context.Context, srv *rpc.Server, logger log.Logger) error {
	srvErrCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(); err != nil {
			srvErrCh <- err
		}
	}()

	logger.Info("started server", "connect.address", srv.ConnectAddress(), "grpc.address", srv.GRPCAddress())

	select {
	case <-ctx.Done():
		logger.Info("context cancelled: stopping server")
		return srv.Stop()
	case err := <-srvErrCh:
		return fmt.Errorf("server failed: %w", err)
	}
}

func getHostPort(port int) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}
