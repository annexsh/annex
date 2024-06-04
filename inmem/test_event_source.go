package inmem

import (
	"context"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/test"
)

type TestExecutionEventSource struct {
	broker *conc.Broker[*event.ExecutionEvent]
}

func NewTestExecutionEventSource(opts ...conc.BrokerOption) *TestExecutionEventSource {
	return &TestExecutionEventSource{
		broker: conc.NewBroker[*event.ExecutionEvent](opts...),
	}
}

func (t *TestExecutionEventSource) Start(ctx context.Context) {
	t.broker.Start(ctx)
}

func (t *TestExecutionEventSource) Publish(event *event.ExecutionEvent) {
	t.broker.Publish(event.TestExecID, event)
}

func (t *TestExecutionEventSource) Subscribe(testExecID test.TestExecutionID) (sub <-chan *event.ExecutionEvent, unsub conc.Unsubscribe) {
	return t.broker.Subscribe(testExecID)
}

func (t *TestExecutionEventSource) Stop() {
	t.broker.Stop()
}
