package testservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"

	"github.com/annexsh/annex/test"
)

func (s *Service) ListTestCaseExecutions(
	ctx context.Context,
	req *connect.Request[testsv1.ListTestCaseExecutionsRequest],
) (*connect.Response[testsv1.ListTestCaseExecutionsResponse], error) {
	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	execs, err := s.repo.ListCaseExecutions(ctx, testExecID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListTestCaseExecutionsResponse{
		CaseExecutions: execs.Proto(),
	}), nil
}

func (s *Service) AckCaseExecutionScheduled(
	ctx context.Context,
	req *connect.Request[testsv1.AckCaseExecutionScheduledRequest],
) (*connect.Response[testsv1.AckCaseExecutionScheduledResponse], error) {
	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	scheduled := &test.ScheduledCaseExecution{
		ID:           test.CaseExecutionID(req.Msg.CaseExecutionId),
		TestExecID:   testExecID,
		CaseName:     req.Msg.CaseName,
		ScheduleTime: req.Msg.ScheduleTime.AsTime().UTC(),
	}

	if _, err = s.repo.CreateScheduledCaseExecution(ctx, scheduled); err != nil {
		return nil, fmt.Errorf("failed to create case execution: %w", err)
	}

	return connect.NewResponse(&testsv1.AckCaseExecutionScheduledResponse{}), nil
}

func (s *Service) AckCaseExecutionStarted(
	ctx context.Context,
	req *connect.Request[testsv1.AckCaseExecutionStartedRequest],
) (*connect.Response[testsv1.AckCaseExecutionStartedResponse], error) {
	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	started := &test.StartedCaseExecution{
		ID:              test.CaseExecutionID(req.Msg.CaseExecutionId),
		TestExecutionID: testExecID,
		StartTime:       req.Msg.StartTime.AsTime().UTC(),
	}
	_, err = s.repo.UpdateStartedCaseExecution(ctx, started)
	if err != nil {
		return nil, fmt.Errorf("failed to create case execution: %w", err)
	}

	return connect.NewResponse(&testsv1.AckCaseExecutionStartedResponse{}), nil
}

func (s *Service) AckCaseExecutionFinished(
	ctx context.Context,
	req *connect.Request[testsv1.AckCaseExecutionFinishedRequest],
) (*connect.Response[testsv1.AckCaseExecutionFinishedResponse], error) {
	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	finished := &test.FinishedCaseExecution{
		ID:              test.CaseExecutionID(req.Msg.CaseExecutionId),
		TestExecutionID: testExecID,
		FinishTime:      req.Msg.FinishTime.AsTime(),
		Error:           req.Msg.Error,
	}
	if _, err = s.repo.UpdateFinishedCaseExecution(ctx, finished); err != nil {
		return nil, fmt.Errorf("failed to update case execution: %w", err)
	}

	return connect.NewResponse(&testsv1.AckCaseExecutionFinishedResponse{}), nil
}
