package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	testservicev1 "github.com/annexhq/annex-proto/gen/go/rpc/testservice/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexhq/annex/test"
)

func (s *ProxyService) PollWorkflowTaskQueue(ctx context.Context, req *workflowservice.PollWorkflowTaskQueueRequest) (*workflowservice.PollWorkflowTaskQueueResponse, error) {
	res, err := s.workflow.PollWorkflowTaskQueue(ctx, req)
	if err != nil {
		return nil, err
	}

	if res.WorkflowExecution == nil {
		return res, nil
	}

	tkn, err := common.NewProtoTaskTokenSerializer().Deserialize(res.TaskToken)
	if err != nil {
		return nil, err
	}

	if res.Attempt == 1 && tkn.ScheduledEventId == 2 && tkn.StartedEventId == 3 {
		if res.History != nil && len(res.History.Events) > 3 {
			if res.History.Events[0].GetWorkflowExecutionStartedEventAttributes() != nil &&
				res.History.Events[1].GetWorkflowTaskScheduledEventAttributes() != nil &&
				res.History.Events[2].GetWorkflowExecutionStartedEventAttributes() != nil {
				testExecID, err := test.ParseTestWorkflowID(res.WorkflowExecution.WorkflowId)
				if err == nil {
					if _, err = s.test.AckTestExecutionStarted(ctx, &testservicev1.AckTestExecutionStartedRequest{
						TestExecId: testExecID.String(),
						StartedAt:  res.StartedTime,
					}); err != nil {
						return nil, err
					}
				} else if !errors.Is(err, test.ErrorNotTestExecution) {
					return nil, err
				}

			}
		}
	}

	return res, nil
}

func (s *ProxyService) RespondWorkflowTaskCompleted(ctx context.Context, req *workflowservice.RespondWorkflowTaskCompletedRequest) (*workflowservice.RespondWorkflowTaskCompletedResponse, error) {
	tkn, err := common.NewProtoTaskTokenSerializer().Deserialize(req.TaskToken)
	if err != nil {
		return nil, err
	}

	testExecID, err := test.ParseTestWorkflowID(tkn.WorkflowId)
	if err != nil {
		if errors.Is(err, test.ErrorNotTestExecution) {
			return s.workflow.RespondWorkflowTaskCompleted(ctx, req)
		}
		return nil, err
	}

	for _, cmd := range req.Commands {
		switch cmd.CommandType {
		case enums.COMMAND_TYPE_SCHEDULE_ACTIVITY_TASK:
			attrs := cmd.GetScheduleActivityTaskCommandAttributes()
			if attrs == nil {
				continue
			}
			caseExecID, err := test.ParseCaseActivityID(attrs.ActivityId)
			if err != nil {
				if errors.Is(err, test.ErrorNotCaseExecution) {
					continue
				}
				return nil, err
			}

			if _, err = s.test.AckCaseExecutionScheduled(ctx, &testservicev1.AckCaseExecutionScheduledRequest{
				Id:          caseExecID.Int32(),
				TestExecId:  testExecID.String(),
				CaseName:    attrs.ActivityType.Name,
				ScheduledAt: timestamppb.New(time.Now().UTC()),
			}); err != nil {
				return nil, fmt.Errorf("failed to acknowledge scheduled case execution: %w", err)
			}
		case enums.COMMAND_TYPE_COMPLETE_WORKFLOW_EXECUTION:
			attrs := cmd.GetCompleteWorkflowExecutionCommandAttributes()
			if attrs == nil {
				continue
			}

			if _, err = s.test.AckTestExecutionFinished(ctx, &testservicev1.AckTestExecutionFinishedRequest{
				TestExecId: testExecID.String(),
				FinishedAt: timestamppb.New(time.Now().UTC()),
			}); err != nil {
				return nil, err
			}
		case enums.COMMAND_TYPE_FAIL_WORKFLOW_EXECUTION:
			attrs := cmd.GetFailWorkflowExecutionCommandAttributes()
			if attrs == nil {
				continue
			}

			var testExecError *string
			if attrs.Failure != nil && attrs.Failure.Cause != nil {
				testExecError = &attrs.Failure.Cause.Message
			}

			if _, err = s.test.AckTestExecutionFinished(ctx, &testservicev1.AckTestExecutionFinishedRequest{
				TestExecId: testExecID.String(),
				FinishedAt: timestamppb.New(time.Now().UTC()),
				Error:      testExecError,
			}); err != nil {
				return nil, err
			}
		}
	}

	return s.workflow.RespondWorkflowTaskCompleted(ctx, req)
}

func (s *ProxyService) PollActivityTaskQueue(ctx context.Context, req *workflowservice.PollActivityTaskQueueRequest) (*workflowservice.PollActivityTaskQueueResponse, error) {
	res, err := s.workflow.PollActivityTaskQueue(ctx, req)
	if err != nil {
		return nil, err
	}

	if res.WorkflowExecution == nil || res.ActivityId == "" {
		return res, nil
	}

	testExecID, err := test.ParseTestWorkflowID(res.WorkflowExecution.WorkflowId)
	if err != nil {
		if errors.Is(err, test.ErrorNotTestExecution) {
			return res, nil
		}
		return nil, err
	}

	caseExecID, err := test.ParseCaseActivityID(res.ActivityId)
	if err != nil {
		if errors.Is(err, test.ErrorNotCaseExecution) {
			return res, nil
		}
		return nil, err
	}

	if _, err = s.test.AckCaseExecutionStarted(ctx, &testservicev1.AckCaseExecutionStartedRequest{
		Id:         caseExecID.Int32(),
		TestExecId: testExecID.String(),
		StartedAt:  timestamppb.Now(),
	}); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *ProxyService) RespondActivityTaskCompleted(ctx context.Context, req *workflowservice.RespondActivityTaskCompletedRequest) (*workflowservice.RespondActivityTaskCompletedResponse, error) {
	tkn, err := common.NewProtoTaskTokenSerializer().Deserialize(req.TaskToken)
	if err != nil {
		return nil, err
	}

	if tkn.WorkflowId == "" || tkn.ActivityId == "" {
		return s.workflow.RespondActivityTaskCompleted(ctx, req)
	}

	testExecID, err := test.ParseTestWorkflowID(tkn.WorkflowId)
	if err != nil {
		if errors.Is(err, test.ErrorNotTestExecution) {
			return s.workflow.RespondActivityTaskCompleted(ctx, req)
		}
		return nil, err
	}

	caseExecID, err := test.ParseCaseActivityID(tkn.ActivityId)
	if err != nil {
		if errors.Is(err, test.ErrorNotCaseExecution) {
			return s.workflow.RespondActivityTaskCompleted(ctx, req)
		}
		return nil, err
	}

	if _, err = s.test.AckCaseExecutionFinished(ctx, &testservicev1.AckCaseExecutionFinishedRequest{
		Id:         caseExecID.Int32(),
		TestExecId: testExecID.String(),
		FinishedAt: timestamppb.Now(),
	}); err != nil {
		return nil, err
	}

	return s.workflow.RespondActivityTaskCompleted(ctx, req)
}

func (s *ProxyService) RespondActivityTaskFailed(ctx context.Context, req *workflowservice.RespondActivityTaskFailedRequest) (*workflowservice.RespondActivityTaskFailedResponse, error) {
	tkn, err := common.NewProtoTaskTokenSerializer().Deserialize(req.TaskToken)
	if err != nil {
		return nil, err
	}

	if tkn.WorkflowId == "" || tkn.ActivityId == "" {
		return s.workflow.RespondActivityTaskFailed(ctx, req)
	}

	testExecID, err := test.ParseTestWorkflowID(tkn.WorkflowId)
	if err != nil {
		if errors.Is(err, test.ErrorNotTestExecution) {
			return s.workflow.RespondActivityTaskFailed(ctx, req)
		}
		return nil, err
	}

	caseExecID, err := test.ParseCaseActivityID(tkn.ActivityId)
	if err != nil {
		if errors.Is(err, test.ErrorNotCaseExecution) {
			return s.workflow.RespondActivityTaskFailed(ctx, req)
		}
		return nil, err
	}

	var execErr *string
	if req.Failure != nil && req.Failure.Message != "" {
		execErr = &req.Failure.Message
	}

	if _, err = s.test.AckCaseExecutionFinished(ctx, &testservicev1.AckCaseExecutionFinishedRequest{
		Id:         caseExecID.Int32(),
		TestExecId: testExecID.String(),
		Error:      execErr,
		FinishedAt: timestamppb.Now(),
	}); err != nil {
		return nil, err
	}

	return s.workflow.RespondActivityTaskFailed(ctx, req)
}
