package inmem

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/annexhq/annex/event"
	"github.com/annexhq/annex/internal/conc"
	"github.com/annexhq/annex/internal/fake"
	"github.com/annexhq/annex/test"
)

func TestTestExecutionEventSource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	numEvents := 50
	numSubs := 100
	wantTotalReceived := numEvents * numSubs

	// Use buffer size equal to the number of expected messages so that
	// no messages are dropped in the test.
	bufSizeOpt := WithSubscriberBufferSize(numEvents)
	es := NewTestExecutionEventSource(bufSizeOpt)
	defer es.Stop()

	testExecID := test.NewTestExecutionID()
	subStartWG := new(sync.WaitGroup)
	subFinWG := new(sync.WaitGroup)

	wantEvents := make([]*event.ExecutionEvent, numEvents)
	for i := range numEvents {
		wantEvents[i] = fake.GenExecEvent(testExecID)
	}

	totalReceived := atomic.NewInt64(0)

	for i := 0; i < numSubs; i++ {
		subStartWG.Add(1)
		subFinWG.Add(1)
		go func() {
			defer subFinWG.Done()
			sub, unsub := es.Subscribe(ctx, testExecID)
			defer unsub()
			subStartWG.Done()
			j := 0
			for execEvent := range sub {
				require.Equal(t, wantEvents[j], execEvent)
				j++
			}
			totalReceived.Add(int64(j))
		}()
	}

	subStartWG.Wait()

	require.Len(t, es.testExecBrokers, 1)
	require.Contains(t, es.testExecBrokers, testExecID)
	broker, _ := es.testExecBrokers[testExecID]
	waitSubCount(t, broker, numSubs)

	for _, want := range wantEvents {
		es.Publish(want)
	}

	es.Stop()
	subFinWG.Wait()

	waitSubCount(t, broker, 0)
	require.Len(t, es.testExecBrokers, 0)
	require.Equal(t, int64(wantTotalReceived), totalReceived.Load())
}

func waitSubCount[T any](t *testing.T, b *conc.Broker[T], count int) {
	require.Eventually(t, func() bool {
		return b.SubscriberCount() == int64(count)
	}, 300*time.Millisecond, 2*time.Millisecond)
}
