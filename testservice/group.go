package testservice

import (
	"context"

	testservicev1 "github.com/annexsh/annex-proto/gen/go/rpc/testservice/v1"
)

func (s *Service) RegisterGroup(ctx context.Context, req *testservicev1.RegisterGroupRequest) (*testservicev1.RegisterGroupResponse, error) {
	if err := s.repo.CreateGroup(ctx, req.Context, req.Name); err != nil {
		return nil, err
	}
	return &testservicev1.RegisterGroupResponse{}, nil
}
