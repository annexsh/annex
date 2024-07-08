package testservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/gen/go/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"
	"github.com/google/uuid"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/pagination"
	"github.com/annexsh/annex/test"
)

func (s *Service) GetTestExecution(
	ctx context.Context,
	req *connect.Request[testsv1.GetTestExecutionRequest],
) (*connect.Response[testsv1.GetTestExecutionResponse], error) {
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
	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	queryPageSize := defaultPageSize + 1

	if req.Msg.PageSize > 0 {
		queryPageSize = req.Msg.PageSize + 1
	}

	filter := &test.TestExecutionListFilter{
		PageSize: uint32(queryPageSize),
	}

	if req.Msg.NextPageToken != "" {
		lastTimestamp, lastID, err := pagination.DecodeNextPageToken(req.Msg)
		if err != nil {
			return nil, err
		}
		filter.LastScheduleTime = &lastTimestamp
		filter.LastTestExecutionID = &lastID
	}

	var testExecs test.TestExecutionList

	testExecs, err = s.repo.ListTestExecutions(ctx, testID, filter)
	if err != nil {
		return nil, err
	}

	res := &testsv1.ListTestExecutionsResponse{}

	hasNextPage := len(testExecs) == int(queryPageSize)
	if hasNextPage {
		testExecs = testExecs[:len(testExecs)-1] // remove page buffer item
		lastExec := testExecs[len(testExecs)-1]
		res.NextPageToken, err = pagination.EncodeNextPageToken(lastExec.ScheduleTime, lastExec.ID.UUID)
		if err != nil {
			return nil, err
		}
	}

	res.TestExecutions = testExecs.Proto()
	return connect.NewResponse(res), nil
}

func (s *Service) AckTestExecutionStarted(
	ctx context.Context,
	req *connect.Request[testsv1.AckTestExecutionStartedRequest],
) (*connect.Response[testsv1.AckTestExecutionStartedResponse], error) {
	execID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	started := &test.StartedTestExecution{
		ID:        execID,
		StartTime: req.Msg.StartTime.AsTime(),
	}

	testExec, err := s.repo.UpdateStartedTestExecution(ctx, started)
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
	execID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	finished := &test.FinishedTestExecution{
		ID:         execID,
		FinishTime: req.Msg.FinishTime.AsTime(),
		Error:      req.Msg.Error,
	}

	testExec, err := s.repo.UpdateFinishedTestExecution(ctx, finished)
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
