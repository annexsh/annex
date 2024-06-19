package testservice

import (
	"context"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"
	"github.com/google/uuid"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
)

func (s *Service) PublishTestExecutionLog(
	ctx context.Context,
	req *connect.Request[testsv1.PublishTestExecutionLogRequest],
) (*connect.Response[testsv1.PublishTestExecutionLogResponse], error) {
	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	var caseExecID *test.CaseExecutionID
	if req.Msg.CaseExecutionId != nil {
		caseExecID = ptr.Get(test.CaseExecutionID(*req.Msg.CaseExecutionId))
	}

	execLog := &test.Log{
		ID:              uuid.New(),
		TestExecutionID: testExecID,
		CaseExecutionID: caseExecID,
		Level:           req.Msg.Level,
		Message:         req.Msg.Message,
		CreateTime:      req.Msg.CreateTime.AsTime(),
	}

	if err = s.repo.CreateLog(ctx, execLog); err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.PublishTestExecutionLogResponse{
		LogId: execLog.ID.String(),
	}), nil
}

func (s *Service) ListTestExecutionLogs(
	ctx context.Context,
	req *connect.Request[testsv1.ListTestExecutionLogsRequest],
) (*connect.Response[testsv1.ListTestExecutionLogsResponse], error) {
	testExecID, err := test.ParseTestExecutionID(req.Msg.TestExecutionId)
	if err != nil {
		return nil, err
	}

	logs, err := s.repo.ListLogs(ctx, testExecID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListTestExecutionLogsResponse{
		Logs: logs.Proto(),
	}), nil
}
