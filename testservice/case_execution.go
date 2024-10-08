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
)

func (s *Service) ListCaseExecutions(
	ctx context.Context,
	req *connect.Request[testsv1.ListCaseExecutionsRequest],
) (*connect.Response[testsv1.ListCaseExecutionsResponse], error) {
	if err := validateListCaseExecutionsRequest(req.Msg); err != nil {
		return nil, err
	}

	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithCaseExecutionID())
	if err != nil {
		return nil, err
	}

	execs, err := s.repo.ListCaseExecutions(ctx, testExecID, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, execs, func(exec *test.CaseExecution) test.CaseExecutionID {
		return exec.ID
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListCaseExecutionsResponse{
		CaseExecutions: execs.Proto(),
		NextPageToken:  nextPageTkn,
	}), nil
}

func (s *Service) AckCaseExecutionScheduled(
	ctx context.Context,
	req *connect.Request[testsv1.AckCaseExecutionScheduledRequest],
) (*connect.Response[testsv1.AckCaseExecutionScheduledResponse], error) {
	if err := validateAckCaseExecutionScheduledRequest(req.Msg); err != nil {
		return nil, err
	}

	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	scheduled := &test.ScheduledCaseExecution{
		ID:              test.CaseExecutionID(req.Msg.CaseExecutionId),
		TestExecutionID: testExecID,
		CaseName:        req.Msg.CaseName,
		ScheduleTime:    req.Msg.ScheduleTime.AsTime().UTC(),
	}

	caseExec, err := s.repo.CreateCaseExecutionScheduled(ctx, scheduled)
	if err != nil {
		return nil, fmt.Errorf("failed to create case execution: %w", err)
	}

	execEvent := event.NewCaseExecutionEvent(eventsv1.Event_TYPE_CASE_EXECUTION_SCHEDULED, caseExec.Proto())
	if err = s.eventPub.Publish(caseExec.TestExecutionID.String(), execEvent); err != nil {
		return nil, fmt.Errorf("failed to publish case execution event: %w", err)
	}

	return connect.NewResponse(&testsv1.AckCaseExecutionScheduledResponse{}), nil
}

func (s *Service) AckCaseExecutionStarted(
	ctx context.Context,
	req *connect.Request[testsv1.AckCaseExecutionStartedRequest],
) (*connect.Response[testsv1.AckCaseExecutionStartedResponse], error) {
	if err := validateAckCaseExecutionStartedRequest(req.Msg); err != nil {
		return nil, err
	}

	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	started := &test.StartedCaseExecution{
		ID:              test.CaseExecutionID(req.Msg.CaseExecutionId),
		TestExecutionID: testExecID,
		StartTime:       req.Msg.StartTime.AsTime().UTC(),
	}

	caseExec, err := s.repo.UpdateCaseExecutionStarted(ctx, started)
	if err != nil {
		return nil, fmt.Errorf("failed to create case execution: %w", err)
	}

	execEvent := event.NewCaseExecutionEvent(eventsv1.Event_TYPE_CASE_EXECUTION_STARTED, caseExec.Proto())
	if err = s.eventPub.Publish(caseExec.TestExecutionID.String(), execEvent); err != nil {
		return nil, fmt.Errorf("failed to publish case execution event: %w", err)
	}

	return connect.NewResponse(&testsv1.AckCaseExecutionStartedResponse{}), nil
}

func (s *Service) AckCaseExecutionFinished(
	ctx context.Context,
	req *connect.Request[testsv1.AckCaseExecutionFinishedRequest],
) (*connect.Response[testsv1.AckCaseExecutionFinishedResponse], error) {
	if err := validateAckCaseExecutionFinishedRequest(req.Msg); err != nil {
		return nil, err
	}

	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	finished := &test.FinishedCaseExecution{
		ID:              test.CaseExecutionID(req.Msg.CaseExecutionId),
		TestExecutionID: testExecID,
		FinishTime:      req.Msg.FinishTime.AsTime().UTC(),
		Error:           req.Msg.Error,
	}

	caseExec, err := s.repo.UpdateCaseExecutionFinished(ctx, finished)
	if err != nil {
		return nil, fmt.Errorf("failed to update case execution: %w", err)
	}

	execEvent := event.NewCaseExecutionEvent(eventsv1.Event_TYPE_CASE_EXECUTION_FINISHED, caseExec.Proto())
	if err = s.eventPub.Publish(caseExec.TestExecutionID.String(), execEvent); err != nil {
		return nil, fmt.Errorf("failed to publish case execution event: %w", err)
	}

	return connect.NewResponse(&testsv1.AckCaseExecutionFinishedResponse{}), nil
}
