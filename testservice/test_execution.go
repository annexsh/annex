package testservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/pagination"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func (s *Service) GetTestExecution(
	ctx context.Context,
	req *connect.Request[testsv1.GetTestExecutionRequest],
) (*connect.Response[testsv1.GetTestExecutionResponse], error) {
	if err := validateGetTestExecutionRequest(req.Msg); err != nil {
		return nil, err
	}

	execID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	exec, err := s.repo.GetTestExecution(ctx, execID)
	if err != nil {
		return nil, err
	}

	res := &testsv1.GetTestExecutionResponse{
		TestExecution: exec.Proto(),
	}

	if exec.HasInput {
		input, err := s.repo.GetTestExecutionInput(ctx, execID)
		if err != nil {
			return nil, err
		}
		res.Input = input.Proto()
	}

	return connect.NewResponse(res), nil
}

func (s *Service) ListTestExecutions(
	ctx context.Context,
	req *connect.Request[testsv1.ListTestExecutionsRequest],
) (*connect.Response[testsv1.ListTestExecutionsResponse], error) {
	if err := validateListTestExecutionsRequest(req.Msg); err != nil {
		return nil, err
	}

	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithTestExecutionID())
	if err != nil {
		return nil, err
	}

	testExecs, err := s.repo.ListTestExecutions(ctx, testID, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, testExecs, func(testExec *test.TestExecution) test.TestExecutionID {
		return testExec.ID
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListTestExecutionsResponse{
		TestExecutions: testExecs.Proto(),
		NextPageToken:  nextPageTkn,
	}), nil
}

func (s *Service) AckTestExecutionStarted(
	ctx context.Context,
	req *connect.Request[testsv1.AckTestExecutionStartedRequest],
) (*connect.Response[testsv1.AckTestExecutionStartedResponse], error) {
	if err := validateAckTestExecutionStartedRequest(req.Msg); err != nil {
		return nil, err
	}

	execID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	started := &test.StartedTestExecution{
		ID:        execID,
		StartTime: req.Msg.StartTime.AsTime(),
	}

	testExec, err := s.repo.UpdateTestExecutionStarted(ctx, started)
	if err != nil {
		return nil, err
	}

	execEvent := event.NewTestExecutionEvent(eventsv1.Event_TYPE_TEST_EXECUTION_STARTED, testExec.Proto())
	if err = s.eventPub.Publish(testExec.ID.String(), execEvent); err != nil {
		return nil, fmt.Errorf("failed to publish test execution event: %w", err)
	}

	return connect.NewResponse(&testsv1.AckTestExecutionStartedResponse{}), nil
}

func (s *Service) AckTestExecutionFinished(
	ctx context.Context,
	req *connect.Request[testsv1.AckTestExecutionFinishedRequest],
) (*connect.Response[testsv1.AckTestExecutionFinishedResponse], error) {
	if err := validateAckTestExecutionFinishedRequest(req.Msg); err != nil {
		return nil, err
	}

	execID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	finished := &test.FinishedTestExecution{
		ID:         execID,
		FinishTime: req.Msg.FinishTime.AsTime(),
		Error:      req.Msg.Error,
	}

	testExec, err := s.repo.UpdateTestExecutionFinished(ctx, finished)
	if err != nil {
		return nil, fmt.Errorf("failed to update test execution: %w", err)
	}

	execEvent := event.NewTestExecutionEvent(eventsv1.Event_TYPE_TEST_EXECUTION_FINISHED, testExec.Proto())
	if err = s.eventPub.Publish(testExec.ID.String(), execEvent); err != nil {
		return nil, fmt.Errorf("failed to publish test execution event: %w", err)
	}

	return connect.NewResponse(&testsv1.AckTestExecutionFinishedResponse{}), nil
}

func (s *Service) RetryTestExecution(
	ctx context.Context,
	req *connect.Request[testsv1.RetryTestExecutionRequest],
) (*connect.Response[testsv1.RetryTestExecutionResponse], error) {
	if err := validateRetryTestExecutionRequest(req.Msg); err != nil {
		return nil, err
	}

	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	testExec, err := s.executor.retry(ctx, testExecID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.RetryTestExecutionResponse{
		TestExecution: testExec.Proto(),
	}), nil
}
