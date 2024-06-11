package inmem

import (
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	"github.com/annexsh/annex/test"
)

type DB struct {
	mu               *sync.RWMutex
	contexts         mapset.Set[string]
	groups           mapset.Set[groupKey]
	tests            map[uuid.UUID]*test.Test
	defaultInputs    map[uuid.UUID]*test.Payload
	testExecs        map[test.TestExecutionID]*test.TestExecution
	testExecPayloads map[test.TestExecutionID][]byte
	caseExecs        map[caseExecKey]*test.CaseExecution
	execLogs         map[uuid.UUID]*test.Log
	events           *TestExecutionEventSource
}

func NewDB() *DB {
	return &DB{
		mu:               new(sync.RWMutex),
		contexts:         mapset.NewSet[string](),
		groups:           mapset.NewSet[groupKey](),
		tests:            map[uuid.UUID]*test.Test{},
		defaultInputs:    map[uuid.UUID]*test.Payload{},
		testExecs:        map[test.TestExecutionID]*test.TestExecution{},
		testExecPayloads: map[test.TestExecutionID][]byte{},
		caseExecs:        map[caseExecKey]*test.CaseExecution{},
		execLogs:         map[uuid.UUID]*test.Log{},
		events:           NewTestExecutionEventSource(),
	}
}

func (d *DB) TestExecutionEventSource() *TestExecutionEventSource {
	return d.events
}
