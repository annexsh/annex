package rpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"connectrpc.com/connect"
)

type ConnectLogInterceptor struct {
	logger Logger
}

func NewConnectLogInterceptor(logger Logger) *ConnectLogInterceptor {
	return &ConnectLogInterceptor{
		logger: logger,
	}
}

func (c *ConnectLogInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		level := slog.LevelInfo
		keyVals := initConnectLogKeyVals(req.Spec())
		start := time.Now()
		keyVals = append(keyVals, "connect.start_time", start.Format(time.RFC3339))

		res, err := next(ctx, req)
		if err != nil {
			level = slog.LevelError
			keyVals = append(keyVals, "connect.error", err)
		}

		keyVals = append(keyVals, "connect.duration", time.Since(start).String())

		c.logger.Log(ctx, level, "unary rpc called", keyVals...)
		return res, err
	}
}

func (c *ConnectLogInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		return conn
	}
}

func (c *ConnectLogInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		level := slog.LevelInfo
		keyVals := initConnectLogKeyVals(conn.Spec())
		keyVals = append(keyVals, "connect.stream_type", conn.Spec().StreamType.String())

		c.logger.Log(ctx, level, "streaming rpc started", keyVals...)

		err := next(ctx, &logStreamingHandlerConn{
			ctx:                  ctx,
			StreamingHandlerConn: conn,
			logger:               c.logger,
			defaultKeyVals:       initConnectLogKeyVals(conn.Spec()),
		})
		if err != nil {
			level = slog.LevelError
			keyVals = append(keyVals, "connect.error", err)
		}

		c.logger.Log(ctx, level, "streaming rpc finished", keyVals...)
		return err
	}
}

type logStreamingHandlerConn struct {
	connect.StreamingHandlerConn
	ctx            context.Context
	logger         Logger
	defaultKeyVals []any
}

func (l *logStreamingHandlerConn) Send(msg any) error {
	level := slog.LevelInfo
	start := time.Now()
	keyVals := l.defaultKeyVals

	err := l.StreamingHandlerConn.Send(msg)
	err = connect.NewError(connect.CodeUnknown, errors.New("bang"))

	if err != nil {
		level = slog.LevelError
		keyVals = append(keyVals, "connect.error", err)
	}

	keyVals = append(keyVals, "connect.duration", time.Since(start).String())
	l.logger.Log(l.ctx, level, "send stream message", keyVals...)
	return err
}

func initConnectLogKeyVals(spec connect.Spec) []any {
	keyVals := []any{
		"protocol", "connect",
	}

	procParts := strings.Split(spec.Procedure, "/")
	if len(procParts) == 3 {
		keyVals = append(keyVals,
			"connect.service", strings.TrimSuffix(procParts[1], "/"),
			"connect.method", procParts[2],
		)
	} else {
		keyVals = append(keyVals, "connect.procedure", spec.Procedure)
	}

	return keyVals
}
