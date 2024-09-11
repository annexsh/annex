package testservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/test"
)

func TestService_RegisterGroup(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "success",
			wantErr: nil,
		}, {
			name:    "error",
			wantErr: errors.New("bang"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextID := "foo"
			groupID := "bar"

			r := &RepositoryMock{
				CreateGroupFunc: func(ctx context.Context, contextID string, groupID string) error {
					assert.Equal(t, contextID, contextID)
					assert.Equal(t, groupID, groupID)
					return tt.wantErr
				},
			}

			s := Service{repo: r}

			req := &testsv1.RegisterGroupRequest{
				Context: contextID,
				Name:    groupID,
			}
			res, err := s.RegisterGroup(context.Background(), connect.NewRequest(req))
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

func TestService_ListGroups(t *testing.T) {
	tests := []struct {
		name            string
		wantAvailable   bool
		lastAccessedDur time.Duration
	}{
		{
			name:            "list with available runner",
			wantAvailable:   true,
			lastAccessedDur: runnerActiveExpireDuration - time.Second,
		},
		{
			name:            "list with unavailable runner",
			wantAvailable:   false,
			lastAccessedDur: runnerActiveExpireDuration + time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextID := "foo"
			pageSize := 2

			wantPage1 := []string{"a", "b"}
			wantPage2 := []string{"c"}

			r := new(RepositoryMock)
			r.ListGroupsFunc = func(ctx context.Context, contextID string, filter test.PageFilter[string]) ([]string, error) {
				assert.Equal(t, contextID, contextID)
				assert.Equal(t, pageSize, filter.Size)

				switch len(r.ListGroupsCalls()) {
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

			runnerID := "mock-runner"
			runnerLastAccessTime := timestamppb.New(time.Now().UTC().Add(-tt.lastAccessedDur))

			w := &WorkflowerMock{
				DescribeTaskQueueFunc: func(ctx context.Context, taskQueue string, taskQueueType enums.TaskQueueType) (*workflowservice.DescribeTaskQueueResponse, error) {
					assert.NotEmpty(t, taskQueue)
					assert.Equal(t, enums.TASK_QUEUE_TYPE_WORKFLOW, taskQueueType)
					return &workflowservice.DescribeTaskQueueResponse{
						Pollers: []*taskqueue.PollerInfo{
							{
								LastAccessTime: runnerLastAccessTime,
								Identity:       runnerID,
							},
						},
					}, nil
				},
			}

			s := Service{repo: r, workflower: w}

			req := &testsv1.ListGroupsRequest{
				Context:  contextID,
				PageSize: int32(pageSize),
			}
			res, err := s.ListGroups(context.Background(), connect.NewRequest(req))
			require.NoError(t, err)
			require.NotEmpty(t, res.Msg.NextPageToken)
			require.Len(t, res.Msg.Groups, len(wantPage1))

			wantRunners := []*testsv1.Group_Runner{{Id: runnerID, LastAccessTime: runnerLastAccessTime}}

			for i, groupID := range wantPage1 {
				wantGroup := &testsv1.Group{
					Context:   contextID,
					Name:      groupID,
					Runners:   wantRunners,
					Available: tt.wantAvailable,
				}
				assert.Equal(t, wantGroup, res.Msg.Groups[i])
			}

			req.NextPageToken = res.Msg.NextPageToken
			res, err = s.ListGroups(context.Background(), connect.NewRequest(req))
			require.NoError(t, err)
			assert.Empty(t, res.Msg.NextPageToken)
			require.Len(t, res.Msg.Groups, len(wantPage2))

			for i, groupID := range wantPage2 {
				wantGroup := &testsv1.Group{
					Context:   contextID,
					Name:      groupID,
					Runners:   wantRunners,
					Available: tt.wantAvailable,
				}
				assert.Equal(t, wantGroup, res.Msg.Groups[i])
			}
		})
	}
}
