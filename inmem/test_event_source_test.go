package inmem

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/internal/fake"
	"github.com/annexsh/annex/test"
)

func TestTestExecutionEventSource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	numTestExecs := 4
	numClients := numTestExecs * 2
	numEventsPerTestExec := 50
	wantTotalReceived := numEventsPerTestExec * numClients

	// Use buffer size equal to the number of expected messages so that
	// no messages are dropped in the test.
	bufSizeOpt := conc.WithSubscribeBufferSize(numEventsPerTestExec)
	es := NewTestExecutionEventSource(bufSizeOpt)
	go es.Start(ctx)

	testExecIDs := make([]test.TestExecutionID, numTestExecs)
	for i := range numTestExecs {
		testExecIDs[i] = test.NewTestExecutionID()
	}

	wantTestExecEvents := map[test.TestExecutionID][]*event.ExecutionEvent{}

	totalReceived := atomic.NewInt64(0)
	var wg sync.WaitGroup

	for i := range numClients {
		id := testExecIDs[i%len(testExecIDs)]
		sub, unsub := es.Subscribe(id)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer unsub()
			eventCount := 0
			for execEvent := range sub {
				require.Equal(t, wantTestExecEvents[id][eventCount], execEvent)
				eventCount++
			}
			totalReceived.Add(int64(eventCount))
		}()
	}

	for _, id := range testExecIDs {
		for range numEventsPerTestExec {
			nextEvent := fake.GenExecEvent(id)
			wantEvents, ok := wantTestExecEvents[id]
			if !ok {
				wantTestExecEvents[id] = []*event.ExecutionEvent{nextEvent}
			}
			wantEvents = append(wantEvents, nextEvent)
			wantTestExecEvents[id] = wantEvents
			es.Publish(nextEvent)
		}
	}

	es.Stop()
	wg.Wait()

	require.Equal(t, int64(wantTotalReceived), totalReceived.Load())
}
