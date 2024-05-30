package conc

import (
	"context"
	"sync/atomic"
)

const (
	defaultInternalBufferSize = 1000
	defaultSubBuffer          = 5
)

type Broker[T any] struct {
	options   brokerOptions
	publishCh chan T
	subCh     chan chan T
	unsubCh   chan chan T
	subCount  atomic.Int64
	isStarted *Value[bool]
	isStopped *Value[bool]
}

func NewBroker[T any](opts ...BrokerOption) *Broker[T] {
	options := brokerOptions{
		pubBufferSize: defaultInternalBufferSize,
		subBufferSize: defaultSubBuffer,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &Broker[T]{
		options:   options,
		publishCh: make(chan T, options.pubBufferSize),
		subCh:     make(chan chan T, defaultInternalBufferSize),
		unsubCh:   make(chan chan T, defaultInternalBufferSize),
		isStarted: NewValue(false),
		isStopped: NewValue(false),
	}
}

func (b *Broker[T]) Start(ctx context.Context) {
	if b.isStarted.Get() {
		return
	}

	b.isStarted.Set(true)
	subs := map[chan T]struct{}{}

	unsub := func(sub chan T) {
		close(sub)
		delete(subs, sub)
		b.subCount.Add(-1)
	}

	stop := func() {
		b.isStopped.Set(true)
		for msgCh := range subs {
			unsub(msgCh)
		}
	}

	for {
		select {
		case <-ctx.Done():
			stop()
			return
		case msgCh := <-b.subCh:
			subs[msgCh] = struct{}{}
			b.subCount.Add(1)
		case msgCh := <-b.unsubCh:
			unsub(msgCh)
		case msg, ok := <-b.publishCh:
			if !ok {
				stop()
				return
			}
			for msgCh := range subs {
				select {
				case msgCh <- msg:
				default: // subscriber buffer full - drop message
				}
			}
		}
	}
}

func (b *Broker[T]) Stop() {
	if !b.isStopped.Get() {
		close(b.publishCh)
	}
}

func (b *Broker[T]) Subscribe() chan T {
	if b.isStopped.Get() {
		// Safeguard to force no-op on new subscribers after stop
		msgCh := make(chan T)
		close(msgCh)
		return msgCh
	}
	msgCh := make(chan T, b.options.subBufferSize)
	b.subCh <- msgCh
	return msgCh
}

func (b *Broker[T]) Unsubscribe(sub chan T) {
	if !b.isStopped.Get() {
		select {
		case b.unsubCh <- sub:
		default:
		}
	}
}

func (b *Broker[T]) Publish(msg T) {
	if !b.isStopped.Get() {
		b.publishCh <- msg
	}
}

func (b *Broker[T]) SubscriberCount() int64 {
	return b.subCount.Load()
}

type BrokerOption func(opts *brokerOptions)

func WithPublishBufferSize(size int) BrokerOption {
	return func(opts *brokerOptions) {
		opts.pubBufferSize = size
	}
}

func WithSubscriberBufferSize(size int) BrokerOption {
	return func(opts *brokerOptions) {
		opts.subBufferSize = size
	}
}

type brokerOptions struct {
	pubBufferSize int
	subBufferSize int
}
