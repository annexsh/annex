package inmem

import (
	"context"
	"sync"

	"github.com/annexhq/annex/event"
	"github.com/annexhq/annex/internal/conc"
	"github.com/annexhq/annex/test"
)

type TestExecutionEventSourceOption func(es *TestExecutionEventSource)

func WithSubscriberBufferSize(size int) TestExecutionEventSourceOption {
	return func(es *TestExecutionEventSource) {
		es.brokerOpts = append(es.brokerOpts, conc.WithSubscriberBufferSize(size))
	}
}

type TestExecutionEventSource struct {
	mu              *sync.Mutex
	testExecBrokers map[test.TestExecutionID]*conc.Broker[*event.ExecutionEvent]
	brokerOpts      []conc.BrokerOption
}

func NewTestExecutionEventSource(opts ...TestExecutionEventSourceOption) *TestExecutionEventSource {
	es := &TestExecutionEventSource{
		mu:              new(sync.Mutex),
		testExecBrokers: map[test.TestExecutionID]*conc.Broker[*event.ExecutionEvent]{},
	}
	for _, opt := range opts {
		opt(es)
	}
	return es
}

func (t *TestExecutionEventSource) Publish(event *event.ExecutionEvent) {
	key := event.TestExecID
	if b, ok := t.testExecBrokers[key]; ok {
		b.Publish(event)
	}
}

func (t *TestExecutionEventSource) Subscribe(ctx context.Context, testExecID test.TestExecutionID) (sub <-chan *event.ExecutionEvent, unsub func()) {
	t.mu.Lock()

	broker, ok := t.testExecBrokers[testExecID]
	if !ok {
		broker = conc.NewBroker[*event.ExecutionEvent](t.brokerOpts...)
		t.testExecBrokers[testExecID] = broker
		go broker.Start(ctx)
	}
	events := broker.Subscribe()

	t.mu.Unlock()

	finishFn := func() {
		t.mu.Lock()
		broker.Unsubscribe(events)
		if broker.SubscriberCount() == 0 {
			delete(t.testExecBrokers, testExecID)
			broker.Stop() // safe to call multiple times
		}
		t.mu.Unlock()
	}

	return events, finishFn
}

// Stop waits for the active event handler to finish before instructing the
// broker to stop listening for future events and stop all subscriptions.
func (t *TestExecutionEventSource) Stop() {
	for _, broker := range t.testExecBrokers {
		broker.Stop()
	}
}
