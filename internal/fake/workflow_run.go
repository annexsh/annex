package fake

import (
	"context"

	"go.temporal.io/sdk/client"

	"github.com/annexsh/annex/uuid"
)

var _ client.WorkflowRun = (*WorkflowRun)(nil)

type WorkflowRun struct {
	id     string
	runID  string
	result any
	err    error
}

func newWorkflowRun(id string, result any, err error) *WorkflowRun {
	return &WorkflowRun{
		id:     id,
		runID:  uuid.NewString(),
		result: result,
		err:    err,
	}
}

func (f *WorkflowRun) GetID() string {
	return f.id
}

func (f *WorkflowRun) GetRunID() string {
	return f.runID
}

func (f *WorkflowRun) Get(_ context.Context, valuePtr any) error {
	valuePtr = f.result
	return f.err
}

func (f *WorkflowRun) GetWithOptions(ctx context.Context, valuePtr any, _ client.WorkflowRunGetOptions) error {
	return f.Get(ctx, valuePtr)
}
