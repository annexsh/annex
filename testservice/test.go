package testservice

import (
	"context"
	"fmt"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	"github.com/google/uuid"

	"github.com/annexsh/annex/test"
)

func (s *Service) RegisterTests(ctx context.Context, req *testservicev1.RegisterTestsRequest) (*testservicev1.RegisterTestsResponse, error) {
	if err := s.repo.CreateGroup(ctx, req.Context, req.Group); err != nil {
		return nil, err
	}

	var defs []*test.TestDefinition

	for _, defpb := range req.Definitions {
		def := &test.TestDefinition{
			ContextID:    req.Context,
			GroupID:      req.Group,
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

func (s *Service) ListTests(ctx context.Context, req *testservicev1.ListTestsRequest) (*testservicev1.ListTestsResponse, error) {
	tests, err := s.repo.ListTests(ctx, req.Context, req.Group)
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

func getTaskQueue(context string, groupName string) string {
	return fmt.Sprintf("%s-%s", context, groupName)
}
