package conc

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBroker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	numClients := 100
	numMsgs := 1000
	wantCount := numClients * numMsgs
	gotCount := new(atomic.Int64)

	// Use buffer size equal to the number of expected messages so that
	// no messages are dropped in the test.
	b := NewBroker[struct{}](WithSubscriberBufferSize(wantCount + 1))
	go b.Start(ctx)

	var wg sync.WaitGroup

	var subs []chan struct{}
	for i := 0; i < numClients; i++ {
		subs = append(subs, b.Subscribe())
	}

	for _, sub := range subs {
		wg.Add(1)
		go func(sub chan struct{}) {
			defer wg.Done()
			for range sub {
				gotCount.Add(1)
			}
		}(sub)
	}

	waitSubCount(t, b, numClients)

	for i := 0; i < numMsgs; i++ {
		b.Publish(struct{}{})
	}

	b.Stop()
	wg.Wait()

	waitSubCount(t, b, 0)

	require.Equal(t, int64(wantCount), gotCount.Load())
}

func waitSubCount[T any](t *testing.T, b *Broker[T], count int) {
	require.Eventually(t, func() bool {
		return b.SubscriberCount() == int64(count)
	}, 200*time.Millisecond, 2*time.Millisecond)
}
