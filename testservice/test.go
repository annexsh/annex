package testservice

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"

	"github.com/annexsh/annex/internal/pagination"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func (s *Service) RegisterTests(
	ctx context.Context,
	req *connect.Request[testsv1.RegisterTestsRequest],
) (*connect.Response[testsv1.RegisterTestsResponse], error) {
	if err := validateRegisterTestsRequest(req.Msg); err != nil {
		return nil, err
	}

	createTime := time.Now().UTC()
	var tests test.TestList

	err := s.repo.ExecuteTx(ctx, func(repo test.Repository) error {
		if err := repo.CreateGroup(ctx, req.Msg.Context, req.Msg.Group); err != nil {
			return err
		}

		for _, defpb := range req.Msg.Definitions {
			t, err := repo.CreateTest(ctx, &test.Test{
				ID:         uuid.New(),
				ContextID:  req.Msg.Context,
				GroupID:    req.Msg.Group,
				Name:       defpb.Name,
				HasInput:   defpb.DefaultInput != nil,
				CreateTime: createTime,
			})
			if err != nil {
				return err
			}

			if t.HasInput {
				if err = repo.CreateTestDefaultInput(ctx, t.ID, &test.Payload{
					Metadata: defpb.DefaultInput.Metadata,
					Data:     defpb.DefaultInput.Data,
				}); err != nil {
					return err
				}
			}

			tests = append(tests, t)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.RegisterTestsResponse{
		Tests: tests.Proto(),
	}), nil
}

func (s *Service) GetTest(
	ctx context.Context,
	req *connect.Request[testsv1.GetTestRequest],
) (*connect.Response[testsv1.GetTestResponse], error) {
	if err := validateGetTestRequest(req.Msg); err != nil {
		return nil, err
	}

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
	if err := validateGetTestDefaultInputRequest(req.Msg); err != nil {
		return nil, err
	}

	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	payload, err := s.repo.GetTestDefaultInput(ctx, testID)
	if err != nil {
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
	if err := validateListTestsRequest(req.Msg); err != nil {
		return nil, err
	}

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithUUID())
	if err != nil {
		return nil, err
	}

	tests, err := s.repo.ListTests(ctx, req.Msg.Context, req.Msg.Group, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, tests, func(test *test.Test) uuid.V7 {
		return test.ID
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListTestsResponse{
		Tests:         tests.Proto(),
		NextPageToken: nextPageTkn,
	}), nil
}

func (s *Service) ExecuteTest(
	ctx context.Context,
	req *connect.Request[testsv1.ExecuteTestRequest],
) (*connect.Response[testsv1.ExecuteTestResponse], error) {
	if err := validateExecuteTestRequest(req.Msg); err != nil {
		return nil, err
	}

	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	t, err := s.repo.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	if err = validateExecuteTestRequestInputRequired(t.HasInput, req.Msg); err != nil {
		return nil, err
	}

	var opts []executeOption
	if req.Msg.Input != nil {
		opts = append(opts, withInput(req.Msg.Input))
	}

	testExec, err := s.executor.execute(ctx, t, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test: %w", err)
	}

	return connect.NewResponse(&testsv1.ExecuteTestResponse{
		TestExecution: testExec.Proto(),
	}), nil
}
