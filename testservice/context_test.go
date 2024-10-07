package testservice

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"github.com/annexsh/annex/test"
)

func TestService_RegisterContext(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "success",
			wantErr: nil,
		},
		{
			name:    "error",
			wantErr: errors.New("bang"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextID := "foo"

			r := &RepositoryMock{
				CreateContextFunc: func(ctx context.Context, id string) error {
					assert.Equal(t, contextID, id)
					return tt.wantErr
				},
			}

			s := Service{repo: r}

			req := &testsv1.RegisterContextRequest{
				Context: contextID,
			}
			res, err := s.RegisterContext(context.Background(), connect.NewRequest(req))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}

func TestService_RegisterContext_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.RegisterContextRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.RegisterContextRequest{
				Context: "",
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.RegisterContext(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_ListContexts(t *testing.T) {
	pageSize := 2

	wantPage1 := []string{"a", "b"}
	wantPage2 := []string{"c"}

	r := new(RepositoryMock)
	r.ListContextsFunc = func(ctx context.Context, filter test.PageFilter[string]) ([]string, error) {
		assert.Equal(t, pageSize, filter.Size)

		switch len(r.ListContextsCalls()) {
		case 1:
			assert.Nil(t, filter.OffsetID)
			return wantPage1, nil
		case 2:
			assert.Equal(t, wantPage1[pageSize-1], *filter.OffsetID)
			return wantPage2, nil
		default:
			panic("unexpected list invocation")
		}
	}

	s := Service{repo: r}

	req := &testsv1.ListContextsRequest{
		PageSize: int32(pageSize),
	}
	res, err := s.ListContexts(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage1, res.Msg.Contexts)
	assert.NotEmpty(t, res.Msg.NextPageToken)

	req.NextPageToken = res.Msg.NextPageToken
	res, err = s.ListContexts(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
	assert.Equal(t, wantPage2, res.Msg.Contexts)
	assert.Empty(t, res.Msg.NextPageToken)
}

func TestService_ListContexts_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.ListContextsRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "page size less than 0",
			req: &testsv1.ListContextsRequest{
				PageSize: int32(-1),
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
		{
			name: "page size greater than max",
			req: &testsv1.ListContextsRequest{
				PageSize: maxPageSize + 1,
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.ListContexts(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}
