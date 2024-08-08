package testservice

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/google/uuid"

	"github.com/annexsh/annex/test"
)

func (s *Service) RegisterTests(
	ctx context.Context,
	req *connect.Request[testsv1.RegisterTestsRequest],
) (*connect.Response[testsv1.RegisterTestsResponse], error) {
	if err := s.repo.CreateGroup(ctx, req.Msg.Context, req.Msg.Group); err != nil {
		return nil, err
	}

	var defs []*test.TestDefinition

	for _, defpb := range req.Msg.Definitions {
		def := &test.TestDefinition{
			ContextID:    req.Msg.Context,
			GroupID:      req.Msg.Group,
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

	return connect.NewResponse(&testsv1.RegisterTestsResponse{
		Tests: created.Proto(),
	}), nil
}

func (s *Service) GetTest(
	ctx context.Context,
	req *connect.Request[testsv1.GetTestRequest],
) (*connect.Response[testsv1.GetTestResponse], error) {
	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	t, err := s.repo.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.GetTestResponse{
		Test: t.Proto(),
	}), nil
}

func (s *Service) GetTestDefaultInput(
	ctx context.Context,
	req *connect.Request[testsv1.GetTestDefaultInputRequest],
) (*connect.Response[testsv1.GetTestDefaultInputResponse], error) {
	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	payload, err := s.repo.GetTestDefaultInput(ctx, testID)
	if err != nil {
		if errors.Is(err, test.ErrorTestPayloadNotFound) {
			return connect.NewResponse(&testsv1.GetTestDefaultInputResponse{
				DefaultInput: "",
			}), nil
		}
		return nil, err
	}

	return connect.NewResponse(&testsv1.GetTestDefaultInputResponse{
		DefaultInput: string(payload.Data),
	}), nil
}

func (s *Service) ListTests(
	ctx context.Context,
	req *connect.Request[testsv1.ListTestsRequest],
) (*connect.Response[testsv1.ListTestsResponse], error) {
	tests, err := s.repo.ListTests(ctx, req.Msg.Context, req.Msg.Group)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListTestsResponse{
		Tests: tests.Proto(),
	}), nil
}

func (s *Service) ExecuteTest(
	ctx context.Context,
	req *connect.Request[testsv1.ExecuteTestRequest],
) (*connect.Response[testsv1.ExecuteTestResponse], error) {
	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	var opts []executeOption
	if req.Msg.Input != nil {
		opts = append(opts, withInput(req.Msg.Input))
	}

	testExec, err := s.executor.execute(ctx, testID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test: %w", err)
	}

	return connect.NewResponse(&testsv1.ExecuteTestResponse{
		TestExecution: testExec.Proto(),
	}), nil
}

func getTaskQueue(context string, groupName string) string {
	return fmt.Sprintf("%s-%s", context, groupName)
}
