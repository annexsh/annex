package nats

import (
	"fmt"

	eventsv1 "github.com/annexsh/annex-proto/go/gen/annex/events/v1"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/annexsh/annex/log"
)

const defaultSubBufferSize = 50

type PubSubOption func(opts *pubSubOptions)

func WithLogger(logger log.Logger) PubSubOption {
	return func(opts *pubSubOptions) {
		opts.logger = logger
	}
}

func WithSubscriptionBufferSize(bufferSize int) PubSubOption {
	return func(opts *pubSubOptions) {
		opts.bufferSize = bufferSize
	}
}

type pubSubOptions struct {
	bufferSize int
	logger     log.Logger
}

type PubSub struct {
	conn *nats.Conn
	opts pubSubOptions
}

func NewPubSub(conn *nats.Conn, opts ...PubSubOption) *PubSub {
	options := pubSubOptions{
		logger:     log.DefaultLogger(),
		bufferSize: defaultSubBufferSize,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return &PubSub{
		conn: conn,
		opts: options,
	}
}

func (p *PubSub) Publish(testExecID string, event *eventsv1.Event) error {
	msgb, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal nats message: %w", err)
	}
	return p.conn.Publish(testExecID, msgb)
}

func (p *PubSub) Subscribe(testExecID string) (<-chan *eventsv1.Event, func(), error) {
	logger := p.opts.logger.With("subject", testExecID)

	ch := make(chan *eventsv1.Event, p.opts.bufferSize)

	sub, err := p.conn.Subscribe(testExecID, func(msg *nats.Msg) {
		out := &eventsv1.Event{}
		if err := proto.Unmarshal(msg.Data, out); err != nil {
			logger.Error("failed to unmarshal nats message", "error", err, "message", string(msg.Data))
			return
		}
		ch <- out
	})
	if err != nil {
		return nil, nil, err
	}

	if err = sub.SetPendingLimits(p.opts.bufferSize, nats.DefaultSubPendingBytesLimit); err != nil {
		return nil, nil, err
	}

	unsub := func() {
		defer close(ch)
		if uErr := sub.Unsubscribe(); err != nil {
			logger.Error("failed to unsubscribe from nats subject", "error", uErr)
		}
	}

	return ch, unsub, nil
}
