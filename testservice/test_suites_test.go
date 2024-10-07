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
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

func TestService_RegisterTestSuite(t *testing.T) {
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
			req := &testsv1.RegisterTestSuiteRequest{
				Context:     "foo",
				Name:        "bar",
				Description: ptr.Get("lorem ipsum"),
			}

			var wantID string

			r := &RepositoryMock{
				CreateTestSuiteFunc: func(ctx context.Context, testSuite *test.TestSuite) (uuid.V7, error) {
					assert.False(t, testSuite.ID.Empty())
					wantID = testSuite.ID.String()
					assert.Equal(t, req.Context, testSuite.ContextID)
					assert.Equal(t, req.Name, testSuite.Name)
					assert.Equal(t, req.Description, testSuite.Description)
					return testSuite.ID, tt.wantErr
				},
			}

			s := Service{repo: r}

			res, err := s.RegisterTestSuite(context.Background(), connect.NewRequest(req))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantID, res.Msg.Id)
		})
	}
}

func TestService_RegisterTestSuite_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.RegisterTestSuiteRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.RegisterTestSuiteRequest{
				Context: "",
				Name:    "foo",
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "blank name",
			req: &testsv1.RegisterTestSuiteRequest{
				Context: "foo",
				Name:    "",
			},
			wantFieldViolation: &errdetails.BadRequest_FieldViolation{
				Field:       "name",
				Description: "Name can't be blank",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.RegisterTestSuite(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}

func TestService_ListTestSuites(t *testing.T) {
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

			wantPage1 := test.TestSuiteList{fake.GenTestSuite(contextID), fake.GenTestSuite(contextID)}
			wantPage2 := test.TestSuiteList{fake.GenTestSuite(contextID)}

			r := new(RepositoryMock)
			r.ListTestSuitesFunc = func(ctx context.Context, contextID string, filter test.PageFilter[string]) (test.TestSuiteList, error) {
				assert.Equal(t, contextID, contextID)
				assert.Equal(t, pageSize, filter.Size)

				switch len(r.ListTestSuitesCalls()) {
				case 1:
					assert.Nil(t, filter.OffsetID)
					return wantPage1, nil
				case 2:
					assert.Equal(t, wantPage1[pageSize-1].Name, *filter.OffsetID)
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

			req := &testsv1.ListTestSuitesRequest{
				Context:  contextID,
				PageSize: int32(pageSize),
			}
			res, err := s.ListTestSuites(context.Background(), connect.NewRequest(req))
			require.NoError(t, err)
			require.NotEmpty(t, res.Msg.NextPageToken)
			require.Len(t, res.Msg.TestSuites, len(wantPage1))

			wantRunners := []*testsv1.TestSuite_Runner{{Id: runnerID, LastAccessTime: runnerLastAccessTime}}

			for i, suite := range wantPage1 {
				wantTestSuite := &testsv1.TestSuite{
					Id:          suite.ID.String(),
					Context:     contextID,
					Name:        suite.Name,
					Description: suite.Description,
					Runners:     wantRunners,
					Available:   tt.wantAvailable,
				}
				assert.Equal(t, wantTestSuite, res.Msg.TestSuites[i])
			}

			req.NextPageToken = res.Msg.NextPageToken
			res, err = s.ListTestSuites(context.Background(), connect.NewRequest(req))
			require.NoError(t, err)
			assert.Empty(t, res.Msg.NextPageToken)
			require.Len(t, res.Msg.TestSuites, len(wantPage2))

			for i, suite := range wantPage2 {
				wantTestSuite := &testsv1.TestSuite{
					Id:          suite.ID.String(),
					Context:     contextID,
					Name:        suite.Name,
					Description: suite.Description,
					Runners:     wantRunners,
					Available:   tt.wantAvailable,
				}
				assert.Equal(t, wantTestSuite, res.Msg.TestSuites[i])
			}
		})
	}
}

func TestService_ListTestSuites_validation(t *testing.T) {
	tests := []struct {
		name               string
		req                *testsv1.ListTestSuitesRequest
		wantFieldViolation *errdetails.BadRequest_FieldViolation
	}{
		{
			name: "blank context",
			req: &testsv1.ListTestSuitesRequest{
				Context:  "",
				PageSize: 1,
			},
			wantFieldViolation: wantBlankContextFieldViolation(),
		},
		{
			name: "page size less than 0",
			req: &testsv1.ListTestSuitesRequest{
				Context:  "foo",
				PageSize: int32(-1),
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
		{
			name: "page size greater than max",
			req: &testsv1.ListTestSuitesRequest{
				Context:  "foo",
				PageSize: maxPageSize + 1,
			},
			wantFieldViolation: wantPageSizeFieldViolation(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{}
			res, err := s.ListTestSuites(context.Background(), connect.NewRequest(tt.req))
			require.Nil(t, res)
			assertInvalidRequest(t, err, tt.wantFieldViolation)
		})
	}
}
