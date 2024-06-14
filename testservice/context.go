package testservice

import (
	"context"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
)

func (s *Service) ListContexts(ctx context.Context, _ *testservicev1.ListContextsRequest) (*testservicev1.ListContextsResponse, error) {
	contexts, err := s.repo.ListContexts(ctx)
	if err != nil {
		return nil, err
	}
	return &testservicev1.ListContextsResponse{
		Contexts: contexts,
	}, nil
}

func (s *Service) RegisterContext(ctx context.Context, req *testservicev1.RegisterContextRequest) (*testservicev1.RegisterContextResponse, error) {
	if err := s.repo.CreateContext(ctx, req.Context); err != nil {
		return nil, err
	}
	return &testservicev1.RegisterContextResponse{}, nil
}
