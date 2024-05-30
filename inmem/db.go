package inmem

import (
	"sync"

	"github.com/google/uuid"

	"github.com/annexhq/annex/test"
)

type DB struct {
	mu               *sync.RWMutex
	tests            map[uuid.UUID]*test.Test
	defaultPayloads  map[uuid.UUID]*test.Payload
	testExecs        map[test.TestExecutionID]*test.TestExecution
	testExecPayloads map[test.TestExecutionID][]byte
	caseExecs        map[string]*test.CaseExecution
	execLogs         map[uuid.UUID]*test.ExecutionLog
	events           *TestExecutionEventSource
}

func NewDB() *DB {
	return &DB{
		mu:               new(sync.RWMutex),
		tests:            map[uuid.UUID]*test.Test{},
		defaultPayloads:  map[uuid.UUID]*test.Payload{},
		testExecs:        map[test.TestExecutionID]*test.TestExecution{},
		testExecPayloads: map[test.TestExecutionID][]byte{},
		caseExecs:        map[string]*test.CaseExecution{},
		execLogs:         map[uuid.UUID]*test.ExecutionLog{},
		events:           NewTestExecutionEventSource(),
	}
}

func (d *DB) TestExecutionEventSource() *TestExecutionEventSource {
	return d.events
}
