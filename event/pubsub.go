package event

import eventsv1 "github.com/annexsh/annex-proto/gen/go/annex/events/v1"

type Publisher interface {
	Publish(testExecID string, event *eventsv1.Event) error
}

type Subscriber interface {
	Subscribe(testExecID string) (<-chan *eventsv1.Event, func(), error)
}

type PubSub interface {
	Publisher
	Subscriber
}
