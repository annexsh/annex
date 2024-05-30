package testservice

import (
	"context"

	testservicev1 "github.com/annexhq/annex-proto/gen/go/rpc/testservice/v1"
	"github.com/google/uuid"

	"github.com/annexhq/annex/internal/ptr"
	"github.com/annexhq/annex/test"
)

func (s *Service) PublishTestExecutionLog(ctx context.Context, req *testservicev1.PublishTestExecutionLogRequest) (*testservicev1.PublishTestExecutionLogResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecId)
	if err != nil {
		return nil, err
	}

	var caseExecID *test.CaseExecutionID
	if req.CaseExecId != nil {
		caseExecID = ptr.Get(test.CaseExecutionID(*req.CaseExecId))
	}

	execLog := &test.ExecutionLog{
		ID:         uuid.New(),
		TestExecID: testExecID,
		CaseExecID: caseExecID,
		Level:      req.Level,
		Message:    req.Message,
		CreateTime: req.CreatedAt.AsTime(),
	}

	if err = s.repo.CreateExecutionLog(ctx, execLog); err != nil {
		return nil, err
	}

	return &testservicev1.PublishTestExecutionLogResponse{
		Id: execLog.ID.String(),
	}, nil
}

func (s *Service) ListTestExecutionLogs(ctx context.Context, req *testservicev1.ListTestExecutionLogsRequest) (*testservicev1.ListTestExecutionLogsResponse, error) {
	testExecID, err := test.ParseTestExecutionID(req.TestExecId)
	if err != nil {
		return nil, err
	}

	logs, err := s.repo.ListExecutionLogs(ctx, testExecID)
	if err != nil {
		return nil, err
	}

	return &testservicev1.ListTestExecutionLogsResponse{
		Logs: logs.Proto(),
	}, nil
}
