package testservice

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"go.temporal.io/api/enums/v1"
	"golang.org/x/sync/errgroup"

	"github.com/annexsh/annex/internal/pagination"
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

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithString())
	if err != nil {
		return nil, err
	}

	groupIDs, err := s.repo.ListGroups(ctx, contextID, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, groupIDs, func(id string) string {
		return id
	})
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
			defer func() { <-sem }()

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
		Groups:        groups,
		NextPageToken: nextPageTkn,
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
		if time.Now().UTC().Sub(poller.LastAccessTime.AsTime().UTC()) < runnerActiveExpireDuration {
			isGroupAvail = true
		}
	}

	return runners, isGroupAvail, nil
}

func getTaskQueue(context string, groupName string) string {
	return fmt.Sprintf("%s-%s", context, groupName)
}
