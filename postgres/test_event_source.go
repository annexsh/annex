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

	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/eventservice"
	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/test"
)

const (
	pgChannelName      = "execution_events"
	testExecsTableName = "test_executions"
	caseExecsTableName = "case_executions"
	logsTableName      = "logs"
)

type TestExecutionEventSource struct {
	broker      *conc.Broker[*eventservice.ExecutionEvent]
	pgConn      *pgx.Conn
	connRelease func()
	ctxCancel   context.CancelFunc
}

func NewTestExecutionEventSource(ctx context.Context, pgPool *pgxpool.Pool, opts ...conc.BrokerOption) (*TestExecutionEventSource, error) {
	conn, err := pgPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	pgConn := conn.Conn()
	if _, err = pgConn.Exec(ctx, "listen "+pgChannelName); err != nil {
		return nil, fmt.Errorf("failed to listen to postgres channel '%s'", pgChannelName)
	}

	return &TestExecutionEventSource{
		broker:      conc.NewBroker[*eventservice.ExecutionEvent](opts...),
		pgConn:      pgConn,
		connRelease: conn.Release,
	}, nil
}

func (t *TestExecutionEventSource) Start(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		ctx, cancel := context.WithCancel(ctx)
		t.ctxCancel = cancel
		t.broker.Start(ctx)

		for {
			if err := t.handleNextEvent(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				errCh <- err
			}
		}
	}()

	return errCh
}

func (t *TestExecutionEventSource) Subscribe(testExecID test.TestExecutionID) (<-chan *eventservice.ExecutionEvent, conc.Unsubscribe) {
	return t.broker.Subscribe(testExecID)
}

func (t *TestExecutionEventSource) Stop() {
	if t.ctxCancel != nil {
		t.ctxCancel()
	}
	t.broker.Stop()
	t.connRelease()
}

func (t *TestExecutionEventSource) handleNextEvent(ctx context.Context) error {
	notif, err := t.pgConn.WaitForNotification(ctx)
	if err != nil {
		return err
	}

	var tableMsg tableMessage

	if err = json.Unmarshal([]byte(notif.Payload), &tableMsg); err != nil {
		return err
	}

	execEvent := &eventservice.ExecutionEvent{
		ID:         uuid.New(),
		CreateTime: time.Now().UTC(),
	}

	switch tableMsg.Table {
	case testExecsTableName:
		if tableMsg.Action != "UPDATE" {
			return nil
		}
		var msg eventMessage[*sqlc.TestExecution]
		if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
			return err
		}
		if !msg.Data.FinishTime.Valid {
			return nil
		}

		execEvent.TestExecID = msg.Data.ID
		execEvent.Data.Type = eventservice.DataTypeTestExecution
		execEvent.Data.TestExecution = marshalTestExec(msg.Data)
		execEvent.Type = eventservice.TypeTestExecutionFinished
	case caseExecsTableName:
		var msg eventMessage[*sqlc.CaseExecution]
		if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
			return err
		}

		execEvent.TestExecID = msg.Data.TestExecutionID
		execEvent.Data.Type = eventservice.DataTypeCaseExecution
		execEvent.Data.CaseExecution = marshalCaseExec(msg.Data)

		switch tableMsg.Action {
		case "INSERT":
			execEvent.Type = eventservice.TypeCaseExecutionScheduled
		case "UPDATE":
			if !msg.Data.StartTime.Valid && !msg.Data.FinishTime.Valid {
				return nil
			}
			execEvent.Type = eventservice.TypeCaseExecutionStarted
			if msg.Data.FinishTime.Valid {
				execEvent.Type = eventservice.TypeCaseExecutionFinished
			}
		}
	case logsTableName:
		var msg eventMessage[*sqlc.Log]
		if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
			return err
		}
		if tableMsg.Action != "INSERT" {
			return nil
		}
		execEvent.TestExecID = msg.Data.TestExecutionID
		execEvent.Data.Type = eventservice.DataTypeLog
		execEvent.Data.Log = marshalExecLog(msg.Data)
		execEvent.Type = eventservice.TypeLogPublished
	default:
		return nil
	}

	t.broker.Publish(execEvent.TestExecID, execEvent)
	return nil
}

type tableMessage struct {
	Table  string `json:"table"`
	Action string `json:"action"`
}

type eventMessage[T any] struct {
	Data T `json:"data"`
}
