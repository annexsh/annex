package testservice

import (
	"context"
	"fmt"
	"time"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	testv1 "github.com/annexsh/annex-proto/gen/go/type/test/v1"
	"github.com/google/uuid"
	"go.temporal.io/api/enums/v1"

	"github.com/annexsh/annex/test"
)

const runnerActiveExpireDuration = time.Minute

func (s *Service) RegisterTests(ctx context.Context, req *testservicev1.RegisterTestsRequest) (*testservicev1.RegisterTestsResponse, error) {
	var defs []*test.TestDefinition

	for _, defpb := range req.Definitions {
		def := &test.TestDefinition{
			Context:      req.Context,
			Group:        req.Group,
			TestID:       uuid.New(),
			Name:         defpb.Name,
			DefaultInput: nil,
		}

		if defpb.DefaultInput != nil {
			def.DefaultInput = &test.Payload{
				Data: defpb.DefaultInput.Data,
			}
		}

		defs = append(defs, def)
	}

	created, err := s.repo.CreateTests(ctx, defs...)
	if err != nil {
		return nil, err
	}

	return &testservicev1.RegisterTestsResponse{
		Tests: created.Proto(),
	}, nil
}

func (s *Service) GetTestDefaultInput(ctx context.Context, req *testservicev1.GetTestDefaultInputRequest) (*testservicev1.GetTestDefaultInputResponse, error) {
	testID, err := uuid.Parse(req.TestId)
	if err != nil {
		return nil, err
	}

	payload, err := s.repo.GetTestDefaultInput(ctx, testID)
	if err != nil {
		return nil, err
	}
	return &testservicev1.GetTestDefaultInputResponse{
		DefaultInput: string(payload.Data),
	}, nil
}

func (s *Service) ListTests(ctx context.Context, _ *testservicev1.ListTestsRequest) (*testservicev1.ListTestsResponse, error) {
	tests, err := s.repo.ListTests(ctx)
	if err != nil {
		return nil, err
	}

	return &testservicev1.ListTestsResponse{
		Tests: tests.Proto(),
	}, nil
}

func (s *Service) ExecuteTest(ctx context.Context, req *testservicev1.ExecuteTestRequest) (*testservicev1.ExecuteTestResponse, error) {
	testID, err := uuid.Parse(req.TestId)
	if err != nil {
		return nil, err
	}

	var opts []executeOption
	if req.Input != nil {
		opts = append(opts, withInput(req.Input))
	}

	testExec, err := s.executor.execute(ctx, testID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test: %w", err)
	}

	return &testservicev1.ExecuteTestResponse{
		TestExecution: testExec.Proto(),
	}, nil
}

func (s *Service) ListTestRunners(ctx context.Context, req *testservicev1.ListTestRunnersRequest) (*testservicev1.ListTestRunnersResponse, error) {
	taskQueue := getTaskQueue(req.Context, req.Group)
	taskQueueRes, err := s.workflower.DescribeTaskQueue(ctx, taskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
	if err != nil {
		return nil, err
	}

	runners := make([]*testv1.Runner, len(taskQueueRes.Pollers))
	for i, poller := range taskQueueRes.Pollers {
		runners[i] = &testv1.Runner{
			Context:        req.Context,
			Group:          req.Group,
			Id:             poller.Identity,
			LastAccessTime: poller.LastAccessTime,
			Active:         poller.LastAccessTime.AsTime().Sub(time.Now()) < runnerActiveExpireDuration,
		}
	}

	return &testservicev1.ListTestRunnersResponse{
		Runners: runners,
	}, nil
}

func getTaskQueue(context string, groupName string) string {
	return fmt.Sprintf("%s-%s", context, groupName)
}
