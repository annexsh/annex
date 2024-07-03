package event

import (
	"github.com/annexsh/annex/internal/conc"
	"github.com/annexsh/annex/test"
)

type Source interface {
	Subscribe(testExecID test.TestExecutionID) (sub <-chan *ExecutionEvent, unsub conc.Unsubscribe)
}
