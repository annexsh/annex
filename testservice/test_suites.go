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
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

const runnerActiveExpireDuration = time.Minute

func (s *Service) RegisterTestSuite(
	ctx context.Context,
	req *connect.Request[testsv1.RegisterTestSuiteRequest],
) (*connect.Response[testsv1.RegisterTestSuiteResponse], error) {
	if err := validateRegisterTestSuiteRequest(req.Msg); err != nil {
		return nil, err
	}

	testSuite := &test.TestSuite{
		ID:          uuid.New(),
		ContextID:   req.Msg.Context,
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
	}

	id, err := s.repo.CreateTestSuite(ctx, testSuite)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.RegisterTestSuiteResponse{
		Id: id.String(),
	}), nil
}

func (s *Service) ListTestSuites(
	ctx context.Context,
	req *connect.Request[testsv1.ListTestSuitesRequest],
) (*connect.Response[testsv1.ListTestSuitesResponse], error) {
	if err := validateListTestSuitesRequest(req.Msg); err != nil {
		return nil, err
	}

	contextID := req.Msg.Context

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithString())
	if err != nil {
		return nil, err
	}

	testSuites, err := s.repo.ListTestSuites(ctx, contextID, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, testSuites, func(ts *test.TestSuite) string {
		return ts.Name
	})
	if err != nil {
		return nil, err
	}

	type runnersResult struct {
		orderID int
		isAvail bool
		runners test.TestSuiteRunnerList
	}

	runnersResultsCh := make(chan runnersResult, len(testSuites))

	const maxConc = 10
	sem := make(chan struct{}, maxConc)
	errg, errCtx := errgroup.WithContext(ctx)

	for i, suite := range testSuites {
		sem <- struct{}{}
		errg.Go(func() error {
			defer func() { <-sem }()

			runners, isTestSuiteAvail, err := getTestSuiteRunners(errCtx, contextID, suite.ID, s.workflower)
			if err != nil {
				return err
			}
			runnersResultsCh <- runnersResult{
				orderID: i,
				isAvail: isTestSuiteAvail,
				runners: runners,
			}

			return nil
		})
	}

	if err = errg.Wait(); err != nil {
		return nil, err
	}

	close(runnersResultsCh)

	listpb := make([]*testsv1.TestSuite, len(testSuites))
	for result := range runnersResultsCh {
		listpb[result.orderID] = testSuites[result.orderID].Proto(result.isAvail, result.runners...)
	}

	return connect.NewResponse(&testsv1.ListTestSuitesResponse{
		TestSuites:    listpb,
		NextPageToken: nextPageTkn,
	}), nil
}

func getTestSuiteRunners(ctx context.Context, contextID string, testSuiteID uuid.V7, workflower Workflower) (test.TestSuiteRunnerList, bool, error) {
	taskQueue := getTaskQueue(contextID, testSuiteID)
	taskQueueRes, err := workflower.DescribeTaskQueue(ctx, taskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
	if err != nil {
		return nil, false, err
	}

	isTestSuiteAvail := false
	runners := make(test.TestSuiteRunnerList, len(taskQueueRes.Pollers))

	for j, poller := range taskQueueRes.Pollers {
		lastAccessTime := poller.LastAccessTime.AsTime().UTC()
		runners[j] = &test.TestSuiteRunner{
			ID:             poller.Identity,
			LastAccessTime: lastAccessTime,
		}
		if time.Now().UTC().Sub(lastAccessTime) < runnerActiveExpireDuration {
			isTestSuiteAvail = true
		}
	}

	return runners, isTestSuiteAvail, nil
}

func getTaskQueue(context string, testSuiteID uuid.V7) string {
	return fmt.Sprintf("%s-%s", context, testSuiteID.String())
}
