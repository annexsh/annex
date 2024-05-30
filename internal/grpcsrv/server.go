package grpcsrv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpcselector "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/annexhq/annex/log"
)

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
