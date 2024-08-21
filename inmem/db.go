package inmem

import (
	"sync"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

type DB struct {
	mu               *sync.RWMutex
	contexts         mapset.Set[string]
	groups           mapset.Set[groupKey]
	tests            map[uuid.V7]*test.Test
	defaultInputs    map[uuid.V7]*test.Payload
	testExecs        map[test.TestExecutionID]*test.TestExecution
	testExecPayloads map[test.TestExecutionID][]byte
	caseExecs        map[caseExecKey]*test.CaseExecution
	execLogs         map[uuid.V7]*test.Log
}

func NewDB() *DB {
	return &DB{
		mu:               new(sync.RWMutex),
		contexts:         mapset.NewSet[string](),
		groups:           mapset.NewSet[groupKey](),
		tests:            map[uuid.V7]*test.Test{},
		defaultInputs:    map[uuid.V7]*test.Payload{},
		testExecs:        map[test.TestExecutionID]*test.TestExecution{},
		testExecPayloads: map[test.TestExecutionID][]byte{},
		caseExecs:        map[caseExecKey]*test.CaseExecution{},
		execLogs:         map[uuid.V7]*test.Log{},
	}
}
