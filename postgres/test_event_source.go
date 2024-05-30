package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/annexhq/annex/postgres/sqlc"

	"github.com/annexhq/annex/event"
	"github.com/annexhq/annex/internal/conc"
	"github.com/annexhq/annex/test"
)

const (
	pgChannelName      = "execution_events"
	testExecsTableName = "test_executions"
	caseExecsTableName = "case_executions"
	logsTableName      = "logs"
)

type TestExecutionEventSource struct {
	children *conc.Map[*conc.Broker[*event.ExecutionEvent]]
	pgConn   *pgx.Conn
	stopCh   chan struct{}
	pgCloser func()
}

func NewTestExecutionEventSource(ctx context.Context, pgPool *pgxpool.Pool) (*TestExecutionEventSource, error) {
	conn, err := pgPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	pgConn := conn.Conn()
	if _, err = pgConn.Exec(ctx, "listen "+pgChannelName); err != nil {
		return nil, fmt.Errorf("failed to listen to postgres channel '%s'", pgChannelName)
	}

	return &TestExecutionEventSource{
		children: conc.NewMap[*conc.Broker[*event.ExecutionEvent]](),
		pgConn:   pgConn,
		stopCh:   make(chan struct{}),
		pgCloser: conn.Release,
	}, nil
}

func (t *TestExecutionEventSource) Start(ctx context.Context) <-chan error {
	fn := func() error {
		for {
			select {
			case <-t.stopCh:
				return nil
			default:
			}

			notif, err := t.pgConn.WaitForNotification(ctx)
			if errors.Is(err, context.Canceled) {
				return nil
			} else if err != nil {
				return err
			}

			var tableMsg tableMessage

			if err = json.Unmarshal([]byte(notif.Payload), &tableMsg); err != nil {
				return err
			}

			execEvent := &event.ExecutionEvent{
				ID:         uuid.New(),
				CreateTime: time.Now().UTC(),
			}

			switch tableMsg.Table {
			case testExecsTableName:
				if tableMsg.Action != "UPDATE" {
					continue
				}
				var msg eventMessage[*sqlc.TestExecution]
				if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
					return err
				}
				if !msg.Data.FinishedAt.Valid {
					continue
				}

				execEvent.TestExecID = msg.Data.ID
				execEvent.Data.Type = event.DataTypeTestExecution
				execEvent.Data.TestExecution = marshalTestExec(msg.Data)
				execEvent.Type = event.TypeTestExecutionFinished
			case caseExecsTableName:
				var msg eventMessage[*sqlc.CaseExecution]
				if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
					return err
				}

				execEvent.TestExecID = msg.Data.TestExecID
				execEvent.Data.Type = event.DataTypeCaseExecution
				execEvent.Data.CaseExecution = marshalCaseExec(msg.Data)

				switch tableMsg.Action {
				case "INSERT":
					execEvent.Type = event.TypeCaseExecutionScheduled
				case "UPDATE":
					if !msg.Data.StartedAt.Valid && !msg.Data.FinishedAt.Valid {
						continue
					}
					execEvent.Type = event.TypeCaseExecutionStarted
					if msg.Data.FinishedAt.Valid {
						execEvent.Type = event.TypeCaseExecutionFinished
					}
				}
			case logsTableName:
				var msg eventMessage[*sqlc.Log]
				if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
					return err
				}
				if tableMsg.Action != "INSERT" {
					continue
				}
				execEvent.TestExecID = msg.Data.TestExecID
				execEvent.Data.Type = event.DataTypeExecutionLog
				execEvent.Data.ExecutionLog = marshalExecLog(msg.Data)
				execEvent.Type = event.TypeExecutionLogPublished
			default:
				continue
			}

			key := execEvent.TestExecID
			if b, ok := t.children.Load(key); ok {
				b.Publish(execEvent)
			}
		}
	}

	errCh := make(chan error, 1)
	go func() {
		if err := fn(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	return errCh
}

func (t *TestExecutionEventSource) Subscribe(ctx context.Context, testExecID test.TestExecutionID) (sub <-chan *event.ExecutionEvent, unsub func()) {
	key := testExecID

	childBroker, ok := t.children.Load(key)
	if !ok {
		childBroker = conc.NewBroker[*event.ExecutionEvent]()
		go childBroker.Start(ctx)
		t.children.Set(key, childBroker)
	}

	events := childBroker.Subscribe()

	finishFn := func() {
		childBroker.Unsubscribe(events)
		if childBroker.SubscriberCount() == 0 {
			t.children.Delete(key)
			childBroker.Stop()
		}
	}

	return events, finishFn
}

// Stop waits for the active event handler to finish before instructing the
// broker to stop listening for future events and stop all subscriptions.
func (t *TestExecutionEventSource) Stop() {
	close(t.stopCh)
	t.children.Range(func(_ any, broker *conc.Broker[*event.ExecutionEvent]) bool {
		broker.Stop()
		return true
	})
	t.pgCloser()
}

type tableMessage struct {
	Table  string `json:"table"`
	Action string `json:"action"`
}

type eventMessage[T any] struct {
	Data T `json:"data"`
}
