package log

import (
	"context"
	"log/slog"
	"os"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/lmittmann/tint"
)

type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

func DefaultLogger() Logger {
	return NewLogger()
}
func NewLogger(args ...any) Logger {
	return &logger{slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(args...)}
}

func NewDevLogger(args ...any) Logger {
	return &logger{slog.New(tint.NewHandler(os.Stdout, nil)).With(args...)}
}

type logger struct {
	*slog.Logger
}

func (l *logger) With(args ...any) Logger {
	return &logger{l.Logger.With(args...)}
}

func (l *logger) GRPCLogger() grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func NewNopLogger() Logger {
	return &nopLogger{}
}

type nopLogger struct{}

func (*nopLogger) Log(_ context.Context, _ slog.Level, _ string, _ ...any) {}
func (*nopLogger) Info(_ string, _ ...any)                                 {}
func (*nopLogger) Debug(_ string, _ ...any)                                {}
func (*nopLogger) Warn(_ string, _ ...any)                                 {}
func (*nopLogger) Error(_ string, _ ...any)                                {}
func (*nopLogger) With(...any) Logger                                      { return &nopLogger{} }
func (*nopLogger) GRPCLogger() grpclog.Logger {
	return grpclog.LoggerFunc(func(_ context.Context, _ grpclog.Level, _ string, _ ...any) {})
}
