package rpc

import (
	"context"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

func toGRPCLogger(logger Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		logger.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
