package fake

import (
	"sync"

	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"

	"github.com/annexsh/annex/test"
)

type PubSub struct {
	topics *sync.Map
}

func NewPubSub() *PubSub {
	return &PubSub{
		topics: new(sync.Map),
	}
}

func (p *PubSub) Publish(testExecID string, event *eventsv1.Event) error {
	v, ok := p.topics.Load(testExecID)
	if !ok {
		events := make(chan *eventsv1.Event, 10)
		events <- event
		p.topics.Store(testExecID, events)
		return nil
	}

	events := v.(chan *eventsv1.Event)
	events <- event
	return nil
}

func (p *PubSub) Subscribe(testExecID string) (<-chan *eventsv1.Event, func(), error) {
	v, ok := p.topics.Load(testExecID)
	if !ok {
		return nil, nil, test.ErrorTestExecutionNotFound
	}
	unsubNop := func() {}
	return v.(chan *eventsv1.Event), unsubNop, nil
}
