package testservice

import (
	"context"
	"fmt"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"

	"github.com/annexsh/annex/test"
)

func (s *Service) ListTestCaseExecutions(ctx context.Context, req *testservicev1.ListTestCaseExecutionsRequest) (*testservicev1.ListTestCaseExecutionsResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecutionId)
	if err != nil {
		return nil, err
	}

	execs, err := s.repo.ListCaseExecutions(ctx, testExecID)
	if err != nil {
		return nil, err
	}

	return &testservicev1.ListTestCaseExecutionsResponse{
		CaseExecutions: execs.Proto(),
	}, nil
}

func (s *Service) AckCaseExecutionScheduled(ctx context.Context, req *testservicev1.AckCaseExecutionScheduledRequest) (*testservicev1.AckCaseExecutionScheduledResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecutionId)
	if err != nil {
		return nil, err
	}

	scheduled := &test.ScheduledCaseExecution{
		ID:           test.CaseExecutionID(req.CaseExecutionId),
		TestExecID:   testExecID,
		CaseName:     req.CaseName,
		ScheduleTime: req.ScheduleTime.AsTime().UTC(),
	}

	if _, err = s.repo.CreateScheduledCaseExecution(ctx, scheduled); err != nil {
		return nil, fmt.Errorf("failed to create case execution: %w", err)
	}

	return &testservicev1.AckCaseExecutionScheduledResponse{}, nil
}

func (s *Service) AckCaseExecutionStarted(ctx context.Context, req *testservicev1.AckCaseExecutionStartedRequest) (*testservicev1.AckCaseExecutionStartedResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecutionId)
	if err != nil {
		return nil, err
	}

	started := &test.StartedCaseExecution{
		ID:              test.CaseExecutionID(req.CaseExecutionId),
		TestExecutionID: testExecID,
		StartTime:       req.StartTime.AsTime().UTC(),
	}
	_, err = s.repo.UpdateStartedCaseExecution(ctx, started)
	if err != nil {
		return nil, fmt.Errorf("failed to create case execution: %w", err)
	}

	return &testservicev1.AckCaseExecutionStartedResponse{}, nil
}

func (s *Service) AckCaseExecutionFinished(ctx context.Context, req *testservicev1.AckCaseExecutionFinishedRequest) (*testservicev1.AckCaseExecutionFinishedResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecutionId)
	if err != nil {
		return nil, err
	}

	finished := &test.FinishedCaseExecution{
		ID:              test.CaseExecutionID(req.CaseExecutionId),
		TestExecutionID: testExecID,
		FinishTime:      req.FinishTime.AsTime(),
		Error:           req.Error,
	}
	if _, err = s.repo.UpdateFinishedCaseExecution(ctx, finished); err != nil {
		return nil, fmt.Errorf("failed to update case execution: %w", err)
	}

	return &testservicev1.AckCaseExecutionFinishedResponse{}, nil
}
