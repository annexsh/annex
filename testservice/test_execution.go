package testservice

import (
	"context"
	"fmt"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	"github.com/google/uuid"

	"github.com/annexsh/annex/internal/pagination"
	"github.com/annexsh/annex/test"
)

func (s *Service) GetTestExecution(ctx context.Context, req *testservicev1.GetTestExecutionRequest) (*testservicev1.GetTestExecutionResponse, error) {
	execID, err := test.ParseTestExecutionID(req.Id)
	if err != nil {
		return nil, err
	}

	exec, err := s.repo.GetTestExecution(ctx, execID)
	if err != nil {
		return nil, err
	}

	res := &testservicev1.GetTestExecutionResponse{
		TestExecution: exec.Proto(),
	}

	if exec.HasPayload {
		input, err := s.repo.GetTestExecutionPayload(ctx, execID)
		if err != nil {
			return nil, err
		}
		res.Input = input.Proto()
	}

	return res, nil
}

func (s *Service) ListTestExecutions(ctx context.Context, req *testservicev1.ListTestExecutionsRequest) (*testservicev1.ListTestExecutionsResponse, error) {
	testID, err := uuid.Parse(req.TestId)
	if err != nil {
		return nil, err
	}

	queryPageSize := defaultPageSize + 1

	if req.PageSize > 0 {
		queryPageSize = req.PageSize + 1
	}

	filter := &test.TestExecutionListFilter{
		PageSize: uint32(queryPageSize),
	}

	if req.NextPageToken != "" {
		lastTimestamp, lastID, err := pagination.DecodeNextPageToken(req)
		if err != nil {
			return nil, err
		}
		filter.LastScheduleTime = &lastTimestamp
		filter.LastExecID = &lastID
	}

	var testExecs test.TestExecutionList

	testExecs, err = s.repo.ListTestExecutions(ctx, testID, filter)
	if err != nil {
		return nil, err
	}

	resp := &testservicev1.ListTestExecutionsResponse{}

	hasNextPage := len(testExecs) == int(queryPageSize)
	if hasNextPage {
		testExecs = testExecs[:len(testExecs)-1] // remove page buffer item
		lastExec := testExecs[len(testExecs)-1]
		resp.NextPageToken, err = pagination.EncodeNextPageToken(lastExec.ScheduleTime, lastExec.ID.UUID)
		if err != nil {
			return nil, err
		}
	}

	resp.TestExecutions = testExecs.Proto()
	return resp, nil
}

func (s *Service) AckTestExecutionStarted(ctx context.Context, req *testservicev1.AckTestExecutionStartedRequest) (*testservicev1.AckTestExecutionStartedResponse, error) {
	execID, err := test.ParseTestExecutionID(req.TestExecId)
	if err != nil {
		return nil, err
	}

	started := &test.StartedTestExecution{
		ID:        execID,
		StartTime: req.StartedAt.AsTime(),
	}

	if _, err = s.repo.UpdateStartedTestExecution(ctx, started); err != nil {
		return nil, err
	}

	return &testservicev1.AckTestExecutionStartedResponse{}, nil
}

func (s *Service) AckTestExecutionFinished(ctx context.Context, req *testservicev1.AckTestExecutionFinishedRequest) (*testservicev1.AckTestExecutionFinishedResponse, error) {
	execID, err := test.ParseTestExecutionID(req.TestExecId)
	if err != nil {
		return nil, err
	}

	finished := &test.FinishedTestExecution{
		ID:         execID,
		FinishTime: req.FinishedAt.AsTime(),
		Error:      req.Error,
	}

	if _, err = s.repo.UpdateFinishedTestExecution(ctx, finished); err != nil {
		return nil, fmt.Errorf("failed to update test execution: %w", err)
	}

	return &testservicev1.AckTestExecutionFinishedResponse{}, nil
}

func (s *Service) RetryTestExecution(ctx context.Context, req *testservicev1.RetryTestExecutionRequest) (*testservicev1.RetryTestExecutionResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecId)
	if err != nil {
		return nil, err
	}

	testExec, err := s.executor.retry(ctx, testExecID)
	if err != nil {
		return nil, err
	}

	return &testservicev1.RetryTestExecutionResponse{
		TestExecution: testExec.Proto(),
	}, nil
}
