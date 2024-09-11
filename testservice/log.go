package testservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/pagination"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func (s *Service) PublishLog(
	ctx context.Context,
	req *connect.Request[testsv1.PublishLogRequest],
) (*connect.Response[testsv1.PublishLogResponse], error) {
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
		CreateTime:      req.Msg.CreateTime.AsTime().UTC(),
	}

	err = s.repo.ExecuteTx(ctx, func(repo test.Repository) error {
		if err = s.repo.CreateLog(ctx, execLog); err != nil {
			return err
		}

		execEvent := event.NewLogEvent(eventsv1.Event_TYPE_LOG_PUBLISHED, execLog.Proto())
		if err = s.eventPub.Publish(execLog.TestExecutionID.String(), execEvent); err != nil {
			return fmt.Errorf("failed to publish log event: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.PublishLogResponse{
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

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithUUID())
	if err != nil {
		return nil, err
	}

	logs, err := s.repo.ListLogs(ctx, testExecID, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, logs, func(log *test.Log) uuid.V7 {
		return log.ID
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListTestExecutionLogsResponse{
		Logs:          logs.Proto(),
		NextPageToken: nextPageTkn,
	}), nil
}
