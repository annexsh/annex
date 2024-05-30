package workflow

import (
	"context"

	"go.temporal.io/api/workflowservice/v1"
)

func (s *ProxyService) RegisterNamespace(ctx context.Context, req *workflowservice.RegisterNamespaceRequest) (*workflowservice.RegisterNamespaceResponse, error) {
	return s.workflow.RegisterNamespace(ctx, req)
}

func (s *ProxyService) DescribeNamespace(ctx context.Context, req *workflowservice.DescribeNamespaceRequest) (*workflowservice.DescribeNamespaceResponse, error) {
	return s.workflow.DescribeNamespace(ctx, req)
}

func (s *ProxyService) StartWorkflowExecution(ctx context.Context, req *workflowservice.StartWorkflowExecutionRequest) (*workflowservice.StartWorkflowExecutionResponse, error) {
	// Note: no need to create the test execution record here since it will have
	// already been created via rpc.testservice.v1.TestService.ExecuteTest.
	return s.workflow.StartWorkflowExecution(ctx, req)
}

func (s *ProxyService) RespondWorkflowTaskFailed(ctx context.Context, req *workflowservice.RespondWorkflowTaskFailedRequest) (*workflowservice.RespondWorkflowTaskFailedResponse, error) {
	return s.workflow.RespondWorkflowTaskFailed(ctx, req)
}

func (s *ProxyService) GetWorkflowExecutionHistory(ctx context.Context, req *workflowservice.GetWorkflowExecutionHistoryRequest) (*workflowservice.GetWorkflowExecutionHistoryResponse, error) {
	return s.workflow.GetWorkflowExecutionHistory(ctx, req)
}

func (s *ProxyService) GetWorkflowExecutionHistoryReverse(ctx context.Context, req *workflowservice.GetWorkflowExecutionHistoryReverseRequest) (*workflowservice.GetWorkflowExecutionHistoryReverseResponse, error) {
	return s.workflow.GetWorkflowExecutionHistoryReverse(ctx, req)
}

func (s *ProxyService) RespondActivityTaskCompletedById(ctx context.Context, req *workflowservice.RespondActivityTaskCompletedByIdRequest) (*workflowservice.RespondActivityTaskCompletedByIdResponse, error) {
	return s.workflow.RespondActivityTaskCompletedById(ctx, req)
}

func (s *ProxyService) SignalWorkflowExecution(ctx context.Context, req *workflowservice.SignalWorkflowExecutionRequest) (*workflowservice.SignalWorkflowExecutionResponse, error) {
	return s.workflow.SignalWorkflowExecution(ctx, req)
}

func (s *ProxyService) SignalWithStartWorkflowExecution(ctx context.Context, req *workflowservice.SignalWithStartWorkflowExecutionRequest) (*workflowservice.SignalWithStartWorkflowExecutionResponse, error) {
	return s.workflow.SignalWithStartWorkflowExecution(ctx, req)
}

func (s *ProxyService) ResetWorkflowExecution(ctx context.Context, req *workflowservice.ResetWorkflowExecutionRequest) (*workflowservice.ResetWorkflowExecutionResponse, error) {
	return s.workflow.ResetWorkflowExecution(ctx, req)
}

func (s *ProxyService) TerminateWorkflowExecution(ctx context.Context, req *workflowservice.TerminateWorkflowExecutionRequest) (*workflowservice.TerminateWorkflowExecutionResponse, error) {
	return s.workflow.TerminateWorkflowExecution(ctx, req)
}

func (s *ProxyService) GetSystemInfo(ctx context.Context, req *workflowservice.GetSystemInfoRequest) (*workflowservice.GetSystemInfoResponse, error) {
	return s.workflow.GetSystemInfo(ctx, req)
}

func (s *ProxyService) DescribeWorkflowExecution(ctx context.Context, req *workflowservice.DescribeWorkflowExecutionRequest) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	return s.workflow.DescribeWorkflowExecution(ctx, req)
}

func (s *ProxyService) ListOpenWorkflowExecutions(ctx context.Context, req *workflowservice.ListOpenWorkflowExecutionsRequest) (*workflowservice.ListOpenWorkflowExecutionsResponse, error) {
	return s.workflow.ListOpenWorkflowExecutions(ctx, req)
}

func (s *ProxyService) DescribeTaskQueue(ctx context.Context, req *workflowservice.DescribeTaskQueueRequest) (*workflowservice.DescribeTaskQueueResponse, error) {
	return s.workflow.DescribeTaskQueue(ctx, req)
}

func (s *ProxyService) CreateSchedule(ctx context.Context, req *workflowservice.CreateScheduleRequest) (*workflowservice.CreateScheduleResponse, error) {
	return s.workflow.CreateSchedule(ctx, req)
}

func (s *ProxyService) DeleteSchedule(ctx context.Context, req *workflowservice.DeleteScheduleRequest) (*workflowservice.DeleteScheduleResponse, error) {
	return s.workflow.DeleteSchedule(ctx, req)
}

func (s *ProxyService) ListNamespaces(ctx context.Context, req *workflowservice.ListNamespacesRequest) (*workflowservice.ListNamespacesResponse, error) {
	return s.workflow.ListNamespaces(ctx, req)
}

func (s *ProxyService) UpdateNamespace(ctx context.Context, req *workflowservice.UpdateNamespaceRequest) (*workflowservice.UpdateNamespaceResponse, error) {
	return s.workflow.UpdateNamespace(ctx, req)
}

func (s *ProxyService) DeprecateNamespace(ctx context.Context, req *workflowservice.DeprecateNamespaceRequest) (*workflowservice.DeprecateNamespaceResponse, error) {
	return s.workflow.DeprecateNamespace(ctx, req)
}

func (s *ProxyService) RecordActivityTaskHeartbeat(ctx context.Context, req *workflowservice.RecordActivityTaskHeartbeatRequest) (*workflowservice.RecordActivityTaskHeartbeatResponse, error) {
	return s.workflow.RecordActivityTaskHeartbeat(ctx, req)
}

func (s *ProxyService) RecordActivityTaskHeartbeatById(ctx context.Context, req *workflowservice.RecordActivityTaskHeartbeatByIdRequest) (*workflowservice.RecordActivityTaskHeartbeatByIdResponse, error) {
	return s.workflow.RecordActivityTaskHeartbeatById(ctx, req)
}

func (s *ProxyService) RespondActivityTaskFailedById(ctx context.Context, req *workflowservice.RespondActivityTaskFailedByIdRequest) (*workflowservice.RespondActivityTaskFailedByIdResponse, error) {
	return s.workflow.RespondActivityTaskFailedById(ctx, req)
}

func (s *ProxyService) RespondActivityTaskCanceled(ctx context.Context, req *workflowservice.RespondActivityTaskCanceledRequest) (*workflowservice.RespondActivityTaskCanceledResponse, error) {
	return s.workflow.RespondActivityTaskCanceled(ctx, req)
}

func (s *ProxyService) RespondActivityTaskCanceledById(ctx context.Context, req *workflowservice.RespondActivityTaskCanceledByIdRequest) (*workflowservice.RespondActivityTaskCanceledByIdResponse, error) {
	return s.workflow.RespondActivityTaskCanceledById(ctx, req)
}

func (s *ProxyService) RequestCancelWorkflowExecution(ctx context.Context, req *workflowservice.RequestCancelWorkflowExecutionRequest) (*workflowservice.RequestCancelWorkflowExecutionResponse, error) {
	return s.workflow.RequestCancelWorkflowExecution(ctx, req)
}

func (s *ProxyService) DeleteWorkflowExecution(ctx context.Context, req *workflowservice.DeleteWorkflowExecutionRequest) (*workflowservice.DeleteWorkflowExecutionResponse, error) {
	return s.workflow.DeleteWorkflowExecution(ctx, req)
}

func (s *ProxyService) ListClosedWorkflowExecutions(ctx context.Context, req *workflowservice.ListClosedWorkflowExecutionsRequest) (*workflowservice.ListClosedWorkflowExecutionsResponse, error) {
	return s.workflow.ListClosedWorkflowExecutions(ctx, req)
}

func (s *ProxyService) ListWorkflowExecutions(ctx context.Context, req *workflowservice.ListWorkflowExecutionsRequest) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	return s.workflow.ListWorkflowExecutions(ctx, req)
}

func (s *ProxyService) ListArchivedWorkflowExecutions(ctx context.Context, req *workflowservice.ListArchivedWorkflowExecutionsRequest) (*workflowservice.ListArchivedWorkflowExecutionsResponse, error) {
	return s.workflow.ListArchivedWorkflowExecutions(ctx, req)
}

func (s *ProxyService) ScanWorkflowExecutions(ctx context.Context, req *workflowservice.ScanWorkflowExecutionsRequest) (*workflowservice.ScanWorkflowExecutionsResponse, error) {
	return s.workflow.ScanWorkflowExecutions(ctx, req)
}

func (s *ProxyService) CountWorkflowExecutions(ctx context.Context, req *workflowservice.CountWorkflowExecutionsRequest) (*workflowservice.CountWorkflowExecutionsResponse, error) {
	return s.workflow.CountWorkflowExecutions(ctx, req)
}

func (s *ProxyService) GetSearchAttributes(ctx context.Context, req *workflowservice.GetSearchAttributesRequest) (*workflowservice.GetSearchAttributesResponse, error) {
	return s.workflow.GetSearchAttributes(ctx, req)
}

func (s *ProxyService) RespondQueryTaskCompleted(ctx context.Context, req *workflowservice.RespondQueryTaskCompletedRequest) (*workflowservice.RespondQueryTaskCompletedResponse, error) {
	return s.workflow.RespondQueryTaskCompleted(ctx, req)
}

func (s *ProxyService) ResetStickyTaskQueue(ctx context.Context, req *workflowservice.ResetStickyTaskQueueRequest) (*workflowservice.ResetStickyTaskQueueResponse, error) {
	return s.workflow.ResetStickyTaskQueue(ctx, req)
}

func (s *ProxyService) QueryWorkflow(ctx context.Context, req *workflowservice.QueryWorkflowRequest) (*workflowservice.QueryWorkflowResponse, error) {
	return s.workflow.QueryWorkflow(ctx, req)
}

func (s *ProxyService) GetClusterInfo(ctx context.Context, req *workflowservice.GetClusterInfoRequest) (*workflowservice.GetClusterInfoResponse, error) {
	return s.workflow.GetClusterInfo(ctx, req)
}

func (s *ProxyService) ListTaskQueuePartitions(ctx context.Context, req *workflowservice.ListTaskQueuePartitionsRequest) (*workflowservice.ListTaskQueuePartitionsResponse, error) {
	return s.workflow.ListTaskQueuePartitions(ctx, req)
}

func (s *ProxyService) DescribeSchedule(ctx context.Context, req *workflowservice.DescribeScheduleRequest) (*workflowservice.DescribeScheduleResponse, error) {
	return s.workflow.DescribeSchedule(ctx, req)
}

func (s *ProxyService) UpdateSchedule(ctx context.Context, req *workflowservice.UpdateScheduleRequest) (*workflowservice.UpdateScheduleResponse, error) {
	return s.workflow.UpdateSchedule(ctx, req)
}

func (s *ProxyService) PatchSchedule(ctx context.Context, req *workflowservice.PatchScheduleRequest) (*workflowservice.PatchScheduleResponse, error) {
	return s.workflow.PatchSchedule(ctx, req)
}

func (s *ProxyService) ListScheduleMatchingTimes(ctx context.Context, req *workflowservice.ListScheduleMatchingTimesRequest) (*workflowservice.ListScheduleMatchingTimesResponse, error) {
	return s.workflow.ListScheduleMatchingTimes(ctx, req)
}

func (s *ProxyService) ListSchedules(ctx context.Context, req *workflowservice.ListSchedulesRequest) (*workflowservice.ListSchedulesResponse, error) {
	return s.workflow.ListSchedules(ctx, req)
}

func (s *ProxyService) UpdateWorkerBuildIdCompatibility(ctx context.Context, req *workflowservice.UpdateWorkerBuildIdCompatibilityRequest) (*workflowservice.UpdateWorkerBuildIdCompatibilityResponse, error) {
	return s.workflow.UpdateWorkerBuildIdCompatibility(ctx, req)
}

func (s *ProxyService) GetWorkerBuildIdCompatibility(ctx context.Context, req *workflowservice.GetWorkerBuildIdCompatibilityRequest) (*workflowservice.GetWorkerBuildIdCompatibilityResponse, error) {
	return s.workflow.GetWorkerBuildIdCompatibility(ctx, req)
}

func (s *ProxyService) GetWorkerTaskReachability(ctx context.Context, req *workflowservice.GetWorkerTaskReachabilityRequest) (*workflowservice.GetWorkerTaskReachabilityResponse, error) {
	return s.workflow.GetWorkerTaskReachability(ctx, req)
}

func (s *ProxyService) UpdateWorkflowExecution(ctx context.Context, req *workflowservice.UpdateWorkflowExecutionRequest) (*workflowservice.UpdateWorkflowExecutionResponse, error) {
	return s.workflow.UpdateWorkflowExecution(ctx, req)
}

func (s *ProxyService) PollWorkflowExecutionUpdate(ctx context.Context, req *workflowservice.PollWorkflowExecutionUpdateRequest) (*workflowservice.PollWorkflowExecutionUpdateResponse, error) {
	return s.workflow.PollWorkflowExecutionUpdate(ctx, req)
}

func (s *ProxyService) StartBatchOperation(ctx context.Context, req *workflowservice.StartBatchOperationRequest) (*workflowservice.StartBatchOperationResponse, error) {
	return s.workflow.StartBatchOperation(ctx, req)
}

func (s *ProxyService) StopBatchOperation(ctx context.Context, req *workflowservice.StopBatchOperationRequest) (*workflowservice.StopBatchOperationResponse, error) {
	return s.workflow.StopBatchOperation(ctx, req)
}

func (s *ProxyService) DescribeBatchOperation(ctx context.Context, req *workflowservice.DescribeBatchOperationRequest) (*workflowservice.DescribeBatchOperationResponse, error) {
	return s.workflow.DescribeBatchOperation(ctx, req)
}

func (s *ProxyService) ListBatchOperations(ctx context.Context, req *workflowservice.ListBatchOperationsRequest) (*workflowservice.ListBatchOperationsResponse, error) {
	return s.workflow.ListBatchOperations(ctx, req)
}

func (s *ProxyService) PollNexusTaskQueue(ctx context.Context, req *workflowservice.PollNexusTaskQueueRequest) (*workflowservice.PollNexusTaskQueueResponse, error) {
	return s.workflow.PollNexusTaskQueue(ctx, req)
}

func (s *ProxyService) RespondNexusTaskCompleted(ctx context.Context, req *workflowservice.RespondNexusTaskCompletedRequest) (*workflowservice.RespondNexusTaskCompletedResponse, error) {
	return s.workflow.RespondNexusTaskCompleted(ctx, req)
}

func (s *ProxyService) RespondNexusTaskFailed(ctx context.Context, req *workflowservice.RespondNexusTaskFailedRequest) (*workflowservice.RespondNexusTaskFailedResponse, error) {
	return s.workflow.RespondNexusTaskFailed(ctx, req)
}
