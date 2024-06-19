package inmem

import (
	"context"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/test"
)

type TestExecutionEventSource struct {
	broker *conc.Broker[*eventservice.ExecutionEvent]
}

func NewTestExecutionEventSource(opts ...conc.BrokerOption) *TestExecutionEventSource {
	return &TestExecutionEventSource{
		broker: conc.NewBroker[*eventservice.ExecutionEvent](opts...),
	}
}

func (t *TestExecutionEventSource) Start(ctx context.Context) {
	t.broker.Start(ctx)
}

func (t *TestExecutionEventSource) Publish(event *eventservice.ExecutionEvent) {
	t.broker.Publish(event.TestExecID, event)
}

func (t *TestExecutionEventSource) Subscribe(testExecID test.TestExecutionID) (sub <-chan *eventservice.ExecutionEvent, unsub conc.Unsubscribe) {
	return t.broker.Subscribe(testExecID)
}

func (t *TestExecutionEventSource) Stop() {
	t.broker.Stop()
}
