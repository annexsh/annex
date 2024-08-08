package testservice

import (
	"context"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
)

func (s *Service) ListContexts(
	ctx context.Context,
	_ *connect.Request[testsv1.ListContextsRequest],
) (*connect.Response[testsv1.ListContextsResponse], error) {
	contexts, err := s.repo.ListContexts(ctx)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&testsv1.ListContextsResponse{
		Contexts: contexts,
	}), nil
}

func (s *Service) RegisterContext(
	ctx context.Context,
	req *connect.Request[testsv1.RegisterContextRequest],
) (*connect.Response[testsv1.RegisterContextResponse], error) {
	if err := s.repo.CreateContext(ctx, req.Msg.Context); err != nil {
		return nil, err
	}
	return connect.NewResponse(&testsv1.RegisterContextResponse{}), nil
}
