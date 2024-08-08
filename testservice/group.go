package testservice

import (
	"context"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"go.temporal.io/api/enums/v1"
	"golang.org/x/sync/errgroup"
)

const runnerActiveExpireDuration = time.Minute

func (s *Service) RegisterGroup(
	ctx context.Context,
	req *connect.Request[testsv1.RegisterGroupRequest],
) (*connect.Response[testsv1.RegisterGroupResponse], error) {
	if err := s.repo.CreateGroup(ctx, req.Msg.Context, req.Msg.Name); err != nil {
		return nil, err
	}
	return &connect.Response[testsv1.RegisterGroupResponse]{}, nil
}

func (s *Service) ListGroups(ctx context.Context, req *connect.Request[testsv1.ListGroupsRequest]) (*connect.Response[testsv1.ListGroupsResponse], error) {
	contextID := req.Msg.Context

	groupIDs, err := s.repo.ListGroups(ctx, contextID)
	if err != nil {
		return nil, err
	}

	type processedGroup struct {
		orderID int
		group   *testsv1.Group
	}

	processedCh := make(chan processedGroup, len(groupIDs))
	groups := make([]*testsv1.Group, len(groupIDs))

	const maxConc = 10
	sem := make(chan struct{}, maxConc)
	errg, errCtx := errgroup.WithContext(ctx)

	for i, groupID := range groupIDs {
		sem <- struct{}{}
		errg.Go(func() error {
			runners, isGroupAvail, err := getGroupRunners(errCtx, contextID, groupID, s.workflower)
			if err != nil {
				return err
			}
			processedCh <- processedGroup{
				orderID: i,
				group: &testsv1.Group{
					Context:   contextID,
					Name:      groupID,
					Runners:   runners,
					Available: isGroupAvail,
				},
			}
			<-sem
			return nil
		})
	}

	if err = errg.Wait(); err != nil {
		return nil, err
	}

	close(processedCh)

	for p := range processedCh {
		groups[p.orderID] = p.group
	}
	return connect.NewResponse(&testsv1.ListGroupsResponse{
		Groups: groups,
	}), nil
}

func getGroupRunners(ctx context.Context, contextID string, groupID string, workflower Workflower) ([]*testsv1.Group_Runner, bool, error) {
	taskQueue := getTaskQueue(contextID, groupID)
	taskQueueRes, err := workflower.DescribeTaskQueue(ctx, taskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
	if err != nil {
		return nil, false, err
	}

	isGroupAvail := false
	runners := make([]*testsv1.Group_Runner, len(taskQueueRes.Pollers))

	for j, poller := range taskQueueRes.Pollers {
		runners[j] = &testsv1.Group_Runner{
			Id:             poller.Identity,
			LastAccessTime: poller.LastAccessTime,
		}
		if poller.LastAccessTime.AsTime().Sub(time.Now()) < runnerActiveExpireDuration {
			isGroupAvail = true
		}
	}

	return runners, isGroupAvail, nil
}
