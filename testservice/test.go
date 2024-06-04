package testservice

import (
	"context"
	"fmt"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	"github.com/google/uuid"

	"github.com/annexsh/annex/test"
)

func (s *Service) RegisterTests(ctx context.Context, req *testservicev1.RegisterTestsRequest) (*testservicev1.RegisterTestsResponse, error) {
	var defs []*test.TestDefinition

	for _, defpb := range req.Definitions {
		def := &test.TestDefinition{
			TestID:         uuid.New(),
			Project:        defpb.Project,
			Name:           defpb.Name,
			DefaultPayload: nil,
			RunnerID:       req.RunnerId,
		}

		if defpb.DefaultPayload != nil {
			def.DefaultPayload = &test.Payload{
				Payload: defpb.DefaultPayload.Data,
				IsZero:  defpb.DefaultPayload.IsZero,
			}
		}

		defs = append(defs, def)
	}

	created, err := s.repo.CreateTests(ctx, defs...)
	if err != nil {
		return nil, err
	}

	return &testservicev1.RegisterTestsResponse{
		Registered: created.Proto(),
	}, nil
}

func (s *Service) GetTestDefaultPayload(ctx context.Context, req *testservicev1.GetTestDefaultPayloadRequest) (*testservicev1.GetTestDefaultPayloadResponse, error) {
	testID, err := uuid.Parse(req.TestId)
	if err != nil {
		return nil, err
	}

	payload, err := s.repo.GetTestDefaultPayload(ctx, testID)
	if err != nil {
		return nil, err
	}
	return &testservicev1.GetTestDefaultPayloadResponse{
		DefaultPayload: string(payload.Payload),
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
	if req.Payload != nil {
		opts = append(opts, withPayload(req.Payload))
	}

	testExec, err := s.executor.execute(ctx, testID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test: %w", err)
	}

	return &testservicev1.ExecuteTestResponse{
		TestExecution: testExec.Proto(),
	}, nil
}
