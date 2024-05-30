package fake

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

var _ client.WorkflowRun = (*WorkflowRun)(nil)

type WorkflowRun struct {
	mu            sync.RWMutex
	id            string
	runID         string
	startSignalCh chan struct{}
	finished      bool
	result        any
	err           error
	getterChs     []chan any
	stopCh        chan struct{}
	isStopped     bool
}

func StartWorkflowRun(ctx context.Context, timeout time.Duration, id string) *WorkflowRun {
	wr := &WorkflowRun{
		id:            id,
		runID:         uuid.NewString(),
		startSignalCh: make(chan struct{}, 1),
		getterChs:     []chan any{},
		stopCh:        make(chan struct{}),
		isStopped:     false,
	}
	go func() {
		ticker := time.NewTicker(timeout)
		defer ticker.Stop()
		select {
		case <-wr.stopCh:
			wr.mu.Lock()
			wr.err = errors.New("fake workflow run stopped")
			return
		case <-ctx.Done():
			wr.err = fmt.Errorf("fake workflow run timeout: start workflow signal not received: %w", ctx.Err())
		case <-ticker.C:
			wr.err = fmt.Errorf("fake workflow run timeout: start workflow signal not received before timeout %s", timeout)
		case <-wr.startSignalCh:
		}
		wr.mu.Lock()
		wr.finished = true
		for _, getter := range wr.getterChs {
			close(getter)
		}
		wr.getterChs = []chan any{}
		wr.mu.Unlock()
	}()
	return wr
}

func (f *WorkflowRun) GetID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.id
}

func (f *WorkflowRun) GetRunID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.runID
}

func (f *WorkflowRun) Get(ctx context.Context, valuePtr any) error {
	f.mu.RLock()
	if f.err != nil {
		return f.err
	}

	if f.finished {
		valuePtr = f.result
		f.mu.RUnlock()
		return nil
	}

	getCh := make(chan any, 1)
	f.getterChs = append(f.getterChs, getCh)
	f.mu.RUnlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case valuePtr = <-getCh:
	}

	return nil
}

func (f *WorkflowRun) GetWithOptions(ctx context.Context, valuePtr any, _ client.WorkflowRunGetOptions) error {
	return f.Get(ctx, valuePtr)
}

func (f *WorkflowRun) signalStart() {
	close(f.startSignalCh)
}

func (f *WorkflowRun) stop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.isStopped {
		close(f.stopCh)
		f.isStopped = true
	}
}
