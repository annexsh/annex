package testservice

import (
	"context"
	"time"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
	testv1 "github.com/annexsh/annex-proto/gen/go/type/test/v1"
	"go.temporal.io/api/enums/v1"
	"golang.org/x/sync/errgroup"
)

const runnerActiveExpireDuration = time.Minute

func (s *Service) RegisterGroup(ctx context.Context, req *testservicev1.RegisterGroupRequest) (*testservicev1.RegisterGroupResponse, error) {
	if err := s.repo.CreateGroup(ctx, req.Context, req.Name); err != nil {
		return nil, err
	}
	return &testservicev1.RegisterGroupResponse{}, nil
}

func (s *Service) ListGroups(ctx context.Context, req *testservicev1.ListGroupsRequest) (*testservicev1.ListGroupsResponse, error) {
	contextID := req.Context

	groupIDs, err := s.repo.ListGroups(ctx, contextID)
	if err != nil {
		return nil, err
	}

	type processedGroup struct {
		orderID int
		group   *testv1.Group
	}

	processedCh := make(chan processedGroup, len(groupIDs))
	groups := make([]*testv1.Group, len(groupIDs))

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
				group: &testv1.Group{
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

	return &testservicev1.ListGroupsResponse{
		Groups: groups,
	}, nil
}

func getGroupRunners(ctx context.Context, contextID string, groupID string, workflower Workflower) ([]*testv1.Group_Runner, bool, error) {
	taskQueue := getTaskQueue(contextID, groupID)
	taskQueueRes, err := workflower.DescribeTaskQueue(ctx, taskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
	if err != nil {
		return nil, false, err
	}

	isGroupAvail := false
	runners := make([]*testv1.Group_Runner, len(taskQueueRes.Pollers))

	for j, poller := range taskQueueRes.Pollers {
		runners[j] = &testv1.Group_Runner{
			Id:             poller.Identity,
			LastAccessTime: poller.LastAccessTime,
		}
		if poller.LastAccessTime.AsTime().Sub(time.Now()) < runnerActiveExpireDuration {
			isGroupAvail = true
		}
	}

	return runners, isGroupAvail, nil
}
