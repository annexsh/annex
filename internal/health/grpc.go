package health

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"
	grpchealth "google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/annexhq/annex/log"
)

const (
	dependencyCheckInterval = 5 * time.Minute
	dependencyCheckTimeout  = 10 * time.Second
	maxConc                 = 5
)

const (
	ServiceNameTest     = "rpc.testservice.v1.TestService"
	ServiceNameEvent    = "rpc.eventservice.v1.EventService"
	ServiceNameWorkflow = "temporal.api.workflowservice.v1.WorkflowService"
	ServiceNamePostgres = "postgres"
)

type ErrorLogger interface {
	Error(msg string, args ...any)
}

type Config struct {
	ServiceNames []string            // service names to register
	Dependencies []DependencyChecker // dependencies to register (optional)
	Logger       ErrorLogger         // logger for health check errors (optional)
}

type GRPCService struct {
	*grpchealth.Server
	deps   []DependencyChecker
	stop   chan struct{}
	logger ErrorLogger
}

// NewGRPCService returns a new health service. All registered dependency checks are
// performed on function invocation and an error is returned if any fail.
// Thereafter, all checks are scheduled at 5 minute intervals as background
// processes, where any failing check sets the overall health service status to
// 'not serving'.
func NewGRPCService(ctx context.Context, config Config) (*GRPCService, error) {
	if len(config.ServiceNames) == 0 {
		return nil, errors.New("at least one service name required")
	}
	if config.Logger == nil {
		config.Logger = log.NewNopLogger()
	}

	s := &GRPCService{
		Server: grpchealth.NewServer(),
		logger: config.Logger,
	}

	errg, depCtx := errgroup.WithContext(ctx)
	errg.SetLimit(maxConc)

	for _, dep := range config.Dependencies {
		currDep := dep
		errg.Go(func() error {
			retryInterval := 5 * time.Second
			bo := backoff.WithMaxRetries(backoff.NewConstantBackOff(retryInterval), 3)

			op := func() error {
				checkCtx, cancel := context.WithTimeout(depCtx, dependencyCheckTimeout)
				defer cancel()
				if err := currDep.Check(checkCtx); err != nil {
					retryMsg := fmt.Sprintf("unhealthy dependency '%s', retrying in %s", currDep.ServiceName(), retryInterval)
					s.logger.Error(retryMsg, "error", err)
					return fmt.Errorf("unhealthy dependency '%s': %w", currDep.ServiceName(), err)
				}
				s.Server.SetServingStatus(currDep.ServiceName(), grpchealthv1.HealthCheckResponse_SERVING)
				return nil
			}

			// Retry checks only on startup
			return backoff.Retry(op, backoff.WithContext(bo, depCtx))
		})
		s.deps = append(s.deps, currDep)
	}

	if err := errg.Wait(); err != nil {
		return nil, err
	}

	for _, name := range config.ServiceNames {
		s.SetServingStatus(name, grpchealthv1.HealthCheckResponse_SERVING)
	}

	go s.scheduleDependencyChecks(ctx)

	return s, nil
}

func (s *GRPCService) Check(ctx context.Context, req *grpchealthv1.HealthCheckRequest) (*grpchealthv1.HealthCheckResponse, error) {
	if req.Service != "" {
		return s.Server.Check(ctx, req)
	}

	for _, dep := range s.deps {
		res, err := s.Check(ctx, &grpchealthv1.HealthCheckRequest{
			Service: dep.ServiceName(),
		})
		if err != nil {
			return nil, err
		}
		if res.Status != grpchealthv1.HealthCheckResponse_SERVING {
			return res, nil
		}
	}

	return &grpchealthv1.HealthCheckResponse{
		Status: grpchealthv1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *GRPCService) scheduleDependencyChecks(ctx context.Context) {
	ticker := time.NewTicker(dependencyCheckInterval)
	sem := make(chan struct{}, maxConc)

	for {
		select {
		case <-ticker.C:
			wg := new(sync.WaitGroup)
			sem <- struct{}{}

			for _, dep := range s.deps {
				wg.Add(1)
				currDep := dep
				go func() {
					checkCtx, cancel := context.WithTimeout(ctx, dependencyCheckTimeout)
					defer cancel()
					defer wg.Done()
					if err := currDep.Check(checkCtx); err != nil {
						s.Server.SetServingStatus(currDep.ServiceName(), grpchealthv1.HealthCheckResponse_NOT_SERVING)
						s.logger.Error("unhealthy dependency", "service.name", currDep.ServiceName(), "error", err)
						return
					}
					s.Server.SetServingStatus(currDep.ServiceName(), grpchealthv1.HealthCheckResponse_SERVING)
				}()

				wg.Wait()
			}
		case <-s.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

type DependencyChecker interface {
	Check(ctx context.Context) error
	ServiceName() string
}

func NewDependencyChecker(serviceName string, check func(ctx context.Context) error) DependencyChecker {
	return &dependency{
		serviceName: serviceName,
		check:       check,
	}
}

type dependency struct {
	serviceName string
	check       func(ctx context.Context) error
}

func (d *dependency) Check(ctx context.Context) error {
	return d.check(ctx)
}

func (d *dependency) ServiceName() string {
	return d.serviceName
}

func WithTemporal(c client.Client, namespace string) DependencyChecker {
	return NewDependencyChecker(ServiceNameWorkflow, func(ctx context.Context) error {
		if _, err := c.CheckHealth(ctx, &client.CheckHealthRequest{}); err != nil {
			return err
		}

		in := &workflowservice.DescribeNamespaceRequest{
			Namespace: namespace,
		}
		wfService := c.WorkflowService()
		if _, err := wfService.DescribeNamespace(ctx, in); err != nil {
			return err
		}

		_, err := wfService.ListOpenWorkflowExecutions(ctx, &workflowservice.ListOpenWorkflowExecutionsRequest{
			Namespace: namespace,
		})
		return err
	})
}

func WithPostgres(pg *pgxpool.Pool) DependencyChecker {
	return NewDependencyChecker(ServiceNamePostgres, pg.Ping)
}
