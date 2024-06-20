package grpcsrv

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpcselector "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/annexsh/annex/log"
)

type ConnectLogInterceptor struct {
	logger log.Logger
}

func NewConnectLogInterceptor(logger log.Logger) *ConnectLogInterceptor {
	return &ConnectLogInterceptor{
		logger: logger,
	}
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

func (c *ConnectLogInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		writeLog := c.logger.Info
		keyVals := initConnectLogKeyVals(req.Spec())
		start := time.Now()
		keyVals = append(keyVals, "connect.start_time", start.Format(time.RFC3339))

		res, err := next(ctx, req)
		if err != nil {
			writeLog = c.logger.Error
			keyVals = append(keyVals, "connect.error", err)
		}

		keyVals = append(keyVals, "connect.duration", time.Since(start).String())

		writeLog("unary rpc called", keyVals...)
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
		writeLog := c.logger.Info
		keyVals := initConnectLogKeyVals(conn.Spec())
		keyVals = append(keyVals, "connect.stream_type", conn.Spec().StreamType.String())

		writeLog("streaming rpc started", keyVals...)

		err := next(ctx, &logStreamingHandlerConn{
			StreamingHandlerConn: conn,
			logger:               c.logger,
			defaultKeyVals:       initConnectLogKeyVals(conn.Spec()),
		})
		if err != nil {
			writeLog = c.logger.Error
			keyVals = append(keyVals, "connect.error", err)
		}

		writeLog("streaming rpc finished", keyVals...)
		return err
	}
}

type logStreamingHandlerConn struct {
	connect.StreamingHandlerConn
	logger         log.Logger
	defaultKeyVals []any
}

func (l *logStreamingHandlerConn) Send(msg any) error {
	writeLog := l.logger.Info
	start := time.Now()
	keyVals := l.defaultKeyVals

	err := l.StreamingHandlerConn.Send(msg)
	err = connect.NewError(connect.CodeUnknown, errors.New("bang"))

	if err != nil {
		writeLog = l.logger.Error
		keyVals = append(keyVals, "connect.error", err)
	}

	keyVals = append(keyVals, "connect.duration", time.Since(start).String())
	writeLog("send stream message", keyVals...)
	return err
}

func New(logger log.Logger) *grpc.Server {
	grpcLogger := logger.GRPCLogger()

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.StartCall, grpclog.FinishCall),
		grpclog.WithDurationField(func(duration time.Duration) grpclog.Fields {
			return grpclog.Fields{"duration", duration.String()}
		}),
	}

	recoveryOpts := []grpcrecovery.Option{
		grpcrecovery.WithRecoveryHandler(recoveryHandler()),
	}

	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcselector.UnaryServerInterceptor(
				grpclog.UnaryServerInterceptor(grpcLogger, logOpts...),
				grpcselector.MatchFunc(func(ctx context.Context, callMeta interceptors.CallMeta) bool {
					if callMeta.Service == grpchealthv1.Health_ServiceDesc.ServiceName {
						return false // ignore health probe logs
					}
					return true
				}),
			),
			grpcrecovery.UnaryServerInterceptor(recoveryOpts...),
		),
		grpc.ChainStreamInterceptor(
			grpclog.StreamServerInterceptor(grpcLogger, logOpts...),
			grpcrecovery.StreamServerInterceptor(recoveryOpts...),
		),
	)
}

func recoveryHandler() grpcrecovery.RecoveryHandlerFunc {
	return func(p any) error {
		msg := "recovered from grpc server panic"
		if p != nil {
			return fmt.Errorf("%s: %+v", msg, p)
		}
		return errors.New(msg)
	}
}
