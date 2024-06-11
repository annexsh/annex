package testservice

import (
	"context"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	"github.com/google/uuid"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
)

func (s *Service) PublishTestExecutionLog(ctx context.Context, req *testservicev1.PublishTestExecutionLogRequest) (*testservicev1.PublishTestExecutionLogResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecutionId)
	if err != nil {
		return nil, err
	}

	var caseExecID *test.CaseExecutionID
	if req.CaseExecutionId != nil {
		caseExecID = ptr.Get(test.CaseExecutionID(*req.CaseExecutionId))
	}

	execLog := &test.Log{
		ID:              uuid.New(),
		TestExecutionID: testExecID,
		CaseExecutionID: caseExecID,
		Level:           req.Level,
		Message:         req.Message,
		CreateTime:      req.CreateTime.AsTime(),
	}

	if err = s.repo.CreateLog(ctx, execLog); err != nil {
		return nil, err
	}

	return &testservicev1.PublishTestExecutionLogResponse{
		LogId: execLog.ID.String(),
	}, nil
}

func (s *Service) ListTestExecutionLogs(ctx context.Context, req *testservicev1.ListTestExecutionLogsRequest) (*testservicev1.ListTestExecutionLogsResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecutionId)
	if err != nil {
		return nil, err
	}

	logs, err := s.repo.ListLogs(ctx, testExecID)
	if err != nil {
		return nil, err
	}

	return &testservicev1.ListTestExecutionLogsResponse{
		Logs: logs.Proto(),
	}, nil
}
