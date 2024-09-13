package testservice

import (
	"context"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"

	"github.com/annexsh/annex/internal/pagination"
)

func (s *Service) RegisterContext(
	ctx context.Context,
	req *connect.Request[testsv1.RegisterContextRequest],
) (*connect.Response[testsv1.RegisterContextResponse], error) {
	if err := validateRegisterContextRequest(req.Msg); err != nil {
		return nil, err
	}

	if err := s.repo.CreateContext(ctx, req.Msg.Context); err != nil {
		return nil, err
	}
	return connect.NewResponse(&testsv1.RegisterContextResponse{}), nil
}

func (s *Service) ListContexts(
	ctx context.Context,
	req *connect.Request[testsv1.ListContextsRequest],
) (*connect.Response[testsv1.ListContextsResponse], error) {
	if err := validateListContextsRequest(req.Msg); err != nil {
		return nil, err
	}

	filter, err := pagination.FilterFromRequest(req.Msg, pagination.WithString())
	if err != nil {
		return nil, err
	}

	contexts, err := s.repo.ListContexts(ctx, filter)
	if err != nil {
		return nil, err
	}

	nextPageTkn, err := pagination.NextPageTokenFromItems(filter.Size, contexts, func(id string) string {
		return id
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&testsv1.ListContextsResponse{
		Contexts:      contexts,
		NextPageToken: nextPageTkn,
	}), nil
}
