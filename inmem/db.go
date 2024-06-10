package inmem

import (
	"sync"

	"github.com/google/uuid"

	"github.com/annexsh/annex/test"
)

type DB struct {
	mu               *sync.RWMutex
	tests            map[uuid.UUID]*test.Test
	defaultInputs    map[uuid.UUID]*test.Payload
	testExecs        map[test.TestExecutionID]*test.TestExecution
	testExecPayloads map[test.TestExecutionID][]byte
	caseExecs        map[string]*test.CaseExecution
	execLogs         map[uuid.UUID]*test.Log
	events           *TestExecutionEventSource
}

func NewDB() *DB {
	return &DB{
		mu:               new(sync.RWMutex),
		tests:            map[uuid.UUID]*test.Test{},
		defaultInputs:    map[uuid.UUID]*test.Payload{},
		testExecs:        map[test.TestExecutionID]*test.TestExecution{},
		testExecPayloads: map[test.TestExecutionID][]byte{},
		caseExecs:        map[string]*test.CaseExecution{},
		execLogs:         map[uuid.UUID]*test.Log{},
		events:           NewTestExecutionEventSource(),
	}
}

func (d *DB) TestExecutionEventSource() *TestExecutionEventSource {
	return d.events
}
