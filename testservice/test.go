package testservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"

	"github.com/annexsh/annex/internal/pagination"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func (s *Service) RegisterTests(
	ctx context.Context,
	stream *connect.ClientStream[testsv1.RegisterTestsRequest],
) (*connect.Response[testsv1.RegisterTestsResponse], error) {
	const maxTests = 30

	if !stream.Receive() {
		if stream.Err() != nil {
			return nil, stream.Err()
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("no tests received in stream"))
	}

	// Only start transaction once first message has been received
	err := s.repo.ExecuteTx(ctx, func(repo test.Repository) error {
		var contextID string
		var testSuiteID uuid.V7
		var version string
		var testsCh <-chan testsResult

		tests := map[string]*test.Test{}
		inputs := map[string]*test.Payload{}
		createTime := time.Now().UTC()

		for i := 0; true; i++ {
			if i > 0 && !stream.Receive() {
				if stream.Err() != nil {
					return stream.Err()
				}
				break
			} else if i > maxTests-1 {
				return connect.NewError(connect.CodeInvalidArgument, errors.New("exceeded maximum of 30 tests per test suite"))
			}

			msg := stream.Msg()

			if err := validateRegisterTestsMessage(i, msg); err != nil {
				return err
			}

			currTestSuiteID, err := uuid.Parse(msg.TestSuiteId)
			if err != nil {
				return err
			}

			if i == 0 {
				testSuiteID = currTestSuiteID
				contextID = msg.Context
				version = msg.Version

				// Get registration version locks the row until tx is complete (postgres only)
				existingVersion, err := repo.GetTestSuiteVersion(ctx, msg.Context, testSuiteID)
				if err != nil && !errors.Is(err, test.ErrorTestSuiteNotFound) {
					return err
				}

				if existingVersion == version {
					return nil // already registered
				}

				testsCh = getAllTestsAsync(ctx, repo, contextID, testSuiteID)
			} else {
				if err = validateRegisterTestsMessageMismatch(i, msg, contextID, testSuiteID, version); err != nil {
					return err
				}
			}

			t := &test.Test{
				ID:          uuid.New(),
				ContextID:   msg.Context,
				TestSuiteID: currTestSuiteID,
				Name:        msg.Definition.Name,
				HasInput:    msg.Definition.DefaultInput != nil,
				CreateTime:  createTime,
			}

			tests[t.Name] = t

			if t.HasInput {
				inputs[t.Name] = &test.Payload{
					Metadata: msg.Definition.DefaultInput.Metadata,
					Data:     msg.Definition.DefaultInput.Data,
				}
			}
		}

		// Receive all existing tests
		var existing test.TestList
		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-testsCh:
			if result.err != nil {
				return result.err
			}
			existing = result.tests
		}

		err := deleteExcludedTests(ctx, repo, existing, tests)
		if err != nil {
			return err
		}

		err = upsertTests(ctx, repo, tests, inputs)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &connect.Response[testsv1.RegisterTestsResponse]{}, nil
}

func (s *Service) GetTest(
	ctx context.Context,
	req *connect.Request[testsv1.GetTestRequest],
) (*connect.Response[testsv1.GetTestResponse], error) {
	if err := validateGetTestRequest(req.Msg); err != nil {
		return nil, err
	}

	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	t, err := s.repo.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.GetTestResponse{
		Test: t.Proto(),
	}), nil
}

func (s *Service) GetTestDefaultInput(
	ctx context.Context,
	req *connect.Request[testsv1.GetTestDefaultInputRequest],
) (*connect.Response[testsv1.GetTestDefaultInputResponse], error) {
	if err := validateGetTestDefaultInputRequest(req.Msg); err != nil {
		return nil, err
	}

	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	payload, err := s.repo.GetTestDefaultInput(ctx, testID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.GetTestDefaultInputResponse{
		DefaultInput: string(payload.Data),
	}), nil
}

func (s *Service) ListTests(
	ctx context.Context,
	req *connect.Request[testsv1.ListTestsRequest],
) (*connect.Response[testsv1.ListTestsResponse], error) {
	if err := validateListTestsRequest(req.Msg); err != nil {
		return nil, err
	}

	testSuiteID, err := uuid.Parse(req.Msg.TestSuiteId)
	if err != nil {
		return nil, err
	}

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithUUID())
	if err != nil {
		return nil, err
	}

	tests, err := s.repo.ListTests(ctx, req.Msg.Context, testSuiteID, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, tests, func(test *test.Test) uuid.V7 {
		return test.ID
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListTestsResponse{
		Tests:         tests.Proto(),
		NextPageToken: nextPageTkn,
	}), nil
}

func (s *Service) ExecuteTest(
	ctx context.Context,
	req *connect.Request[testsv1.ExecuteTestRequest],
) (*connect.Response[testsv1.ExecuteTestResponse], error) {
	if err := validateExecuteTestRequest(req.Msg); err != nil {
		return nil, err
	}

	testID, err := uuid.Parse(req.Msg.TestId)
	if err != nil {
		return nil, err
	}

	t, err := s.repo.GetTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	if err = validateExecuteTestRequestInputRequired(t.HasInput, req.Msg); err != nil {
		return nil, err
	}

	var opts []executeOption
	if req.Msg.Input != nil {
		opts = append(opts, withInput(req.Msg.Input))
	}

	testExec, err := s.executor.execute(ctx, t, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test: %w", err)
	}

	return connect.NewResponse(&testsv1.ExecuteTestResponse{
		TestExecution: testExec.Proto(),
	}), nil
}

type testsResult struct {
	tests test.TestList
	err   error
}

func getAllTestsAsync(ctx context.Context, repo test.Repository, contextID string, testSuiteID uuid.V7) <-chan testsResult {
	out := make(chan testsResult, 1)

	go func() {
		var result testsResult
		var offsetID *uuid.V7
		var items test.TestList
		pageSize := 50

		defer func() {
			out <- result
			close(out)
		}()

		for {
			page, err := repo.ListTests(ctx, contextID, testSuiteID, test.PageFilter[uuid.V7]{
				Size:     pageSize,
				OffsetID: offsetID,
			})
			if err != nil {
				result.err = err
				return
			}

			items = append(items, page...)
			if len(page) < pageSize {
				result.tests = items
				return
			}

			offsetID = &page[len(page)-1].ID
		}
	}()

	return out
}

func deleteExcludedTests(ctx context.Context, repo test.Repository, existing test.TestList, tests map[string]*test.Test) error {
	for _, e := range existing {
		if _, ok := tests[e.Name]; !ok {
			if err := repo.DeleteTest(ctx, e.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func upsertTests(ctx context.Context, repo test.Repository, tests map[string]*test.Test, inputs map[string]*test.Payload) error {
	for _, t := range tests {
		created, err := repo.CreateTest(ctx, t)
		if err != nil {
			return err
		}
		if in, ok := inputs[t.Name]; ok {
			if err = repo.CreateTestDefaultInput(ctx, created.ID, &test.Payload{
				Metadata: in.Metadata,
				Data:     in.Data,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
