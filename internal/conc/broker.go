package conc

import (
	"context"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	defaultPubBufferSize = 1000
	defaultSubBufferSize = 5
)

type BrokerOption func(opts *brokerOptions)

func WithPublishBufferSize(size int) BrokerOption {
	return func(opts *brokerOptions) {
		opts.pubBufferSize = size
	}
}

func WithSubscribeBufferSize(size int) BrokerOption {
	return func(opts *brokerOptions) {
		opts.subBufferSize = size
	}
}

type Broker[T any] struct {
	options   brokerOptions
	mu        *sync.RWMutex
	topics    map[any]mapset.Set[chan T]
	publishCh chan *brokerMessage[T]
	once      *sync.Once
}

func NewBroker[T any](opts ...BrokerOption) *Broker[T] {
	options := brokerOptions{
		pubBufferSize: defaultPubBufferSize,
		subBufferSize: defaultSubBufferSize,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &Broker[T]{
		options:   options,
		mu:        new(sync.RWMutex),
		topics:    map[any]mapset.Set[chan T]{},
		publishCh: make(chan *brokerMessage[T], options.pubBufferSize),
		once:      new(sync.Once),
	}
}

func (b *Broker[T]) Start(ctx context.Context) {
	go func() {
		stopFn := func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			for _, subs := range b.topics {
				for sub := range subs.Iter() {
					close(sub)
				}
			}
		}

		for {
			select {
			case msg, ok := <-b.publishCh:
				if !ok {
					stopFn()
					return
				}
				b.mu.RLock()
				if topicSubs, ok := b.topics[msg.topic]; ok {
					for sub := range topicSubs.Iter() {
						select {
						case sub <- msg.data:
						default: // subscriber buffer full - drop message
						}
					}
				}
				b.mu.RUnlock()
			case <-ctx.Done():
				stopFn()
				return
			}
		}
	}()
}

func (b *Broker[T]) Publish(topic any, msg T) {
	b.publishCh <- &brokerMessage[T]{
		topic: topic,
		data:  msg,
	}
}

type Unsubscribe func()

func (b *Broker[T]) Subscribe(topic any) (<-chan T, Unsubscribe) {
	b.mu.Lock()
	msgCh := make(chan T, b.options.subBufferSize)
	topicSubs, ok := b.topics[topic]
	if !ok {
		topicSubs = mapset.NewSet[chan T]()
	}
	topicSubs.Add(msgCh)
	b.topics[topic] = topicSubs
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if ts, ok := b.topics[topic]; ok {
			ts.Remove(msgCh)
			if ts.Cardinality() == 0 {
				delete(b.topics, topic)
			}
		}
	}

	return msgCh, unsub
}

func (b *Broker[T]) Stop() {
	b.once.Do(func() {
		close(b.publishCh)
	})
}

type brokerMessage[T any] struct {
	topic any
	data  T
}

type brokerOptions struct {
	pubBufferSize int
	subBufferSize int
}
