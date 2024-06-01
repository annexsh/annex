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
	numMsgsPerTopic := 1000
	wantMsgsRecvd := numClients * numMsgsPerTopic // 100,000
	gotNumMsgsRecvd := new(atomic.Int64)
	topics := []string{"foo", "bar", "baz", "qux"}

	// Use buffer size equal to the number of expected messages so that
	// no messages are dropped in the test.
	b := NewBroker[struct{}](WithSubscribeBufferSize(wantMsgsRecvd))
	go b.Start(ctx)

	var wg sync.WaitGroup

	for i := range numClients {
		topic := topics[i%len(topics)]
		sub, unsub := b.Subscribe(topic)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer unsub()
			for range sub {
				gotNumMsgsRecvd.Add(1)
			}
		}()
	}

	for range numMsgsPerTopic {
		for _, topic := range topics {
			b.Publish(topic, struct{}{})
		}
	}

	b.Stop()
	wg.Wait()

	require.Equal(t, int64(wantMsgsRecvd), gotNumMsgsRecvd.Load())
}
