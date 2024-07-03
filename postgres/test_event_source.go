package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/annexsh/annex/event"
	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/test"
)

const (
	pgChannelName      = "execution_events"
	testExecsTableName = "test_executions"
	caseExecsTableName = "case_executions"
	logsTableName      = "logs"
	pgInsert           = "INSERT"
	pgUpdate           = "UPDATE"
)

type ErrorCallback func(err error)

type TestExecutionEventSource struct {
	broker      *conc.Broker[*event.ExecutionEvent]
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
		broker:      conc.NewBroker[*event.ExecutionEvent](opts...),
		pgConn:      pgConn,
		connRelease: conn.Release,
	}, nil
}

func (t *TestExecutionEventSource) Start(ctx context.Context, errCallback ErrorCallback) {
	ctx, cancel := context.WithCancel(ctx)
	t.ctxCancel = cancel
	t.broker.Start(ctx)

	for {
		if err := t.handleNextEvent(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			errCallback(err)
		}
	}
}

func (t *TestExecutionEventSource) Subscribe(testExecID test.TestExecutionID) (<-chan *event.ExecutionEvent, conc.Unsubscribe) {
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

	var execEvent *event.ExecutionEvent

	switch tableMsg.Table {
	case testExecsTableName:
		var msg eventMessage[*sqlc.TestExecution]
		if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
			return err
		}
		testExec := marshalTestExec(msg.Data)

		switch tableMsg.Action {
		case pgInsert:
			execEvent = event.NewTestExecutionEvent(event.TypeTestExecutionScheduled, testExec)
		case pgUpdate:
			if msg.Data.StartTime.Valid && !msg.Data.FinishTime.Valid {
				execEvent = event.NewTestExecutionEvent(event.TypeTestExecutionStarted, testExec)
			} else if msg.Data.StartTime.Valid && msg.Data.FinishTime.Valid {
				execEvent = event.NewTestExecutionEvent(event.TypeTestExecutionFinished, testExec)
			} else {
				// TODO: log unexpected state error
				return nil
			}
		}
	case caseExecsTableName:
		var msg eventMessage[*sqlc.CaseExecution]
		if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
			return err
		}

		caseExec := marshalCaseExec(msg.Data)

		switch tableMsg.Action {
		case pgInsert:
			execEvent = event.NewCaseExecutionEvent(event.TypeCaseExecutionScheduled, caseExec)
		case pgUpdate:
			if msg.Data.StartTime.Valid && !msg.Data.FinishTime.Valid {
				execEvent = event.NewCaseExecutionEvent(event.TypeCaseExecutionStarted, caseExec)
			} else if msg.Data.StartTime.Valid && msg.Data.FinishTime.Valid {
				execEvent = event.NewCaseExecutionEvent(event.TypeCaseExecutionFinished, caseExec)
			} else {
				// TODO: log unexpected state error
				return nil
			}
		}
	case logsTableName:
		if tableMsg.Action != pgInsert {
			return nil
		}
		var msg eventMessage[*sqlc.Log]
		if err = json.Unmarshal([]byte(notif.Payload), &msg); err != nil {
			return err
		}
		execLog := marshalLog(msg.Data)
		execEvent = event.NewLogEvent(event.TypeLogPublished, execLog)
	}

	if execEvent == nil {
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
