package fake

import (
	"math/rand"
	"sync"
	"time"

	"go.temporal.io/sdk/converter"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

type TestOption func(t *test.Test)

func WithContextID(contextID string) TestOption {
	return func(t *test.Test) {
		t.ContextID = contextID
	}
}

func WithGroupID(groupID string) TestOption {
	return func(t *test.Test) {
		t.GroupID = groupID
	}
}

func WithHasInput(hasInput bool) TestOption {
	return func(t *test.Test) {
		t.HasInput = hasInput
	}
}

func GenInput() *test.Payload {
	p, err := converter.NewJSONPayloadConverter().ToPayload(GenData())
	if err != nil {
		panic(err)
	}
	return &test.Payload{
		Metadata: p.Metadata,
		Data:     p.Data,
	}
}

func GenDefaultInput() *test.Payload {
	var data Data
	p, err := converter.NewJSONPayloadConverter().ToPayload(data)
	if err != nil {
		panic(err)
	}
	return &test.Payload{
		Data: p.Data,
	}
}

func GenTest(opts ...TestOption) *test.Test {
	t := &test.Test{
		ContextID:  "default-context",
		GroupID:    "default-group",
		ID:         uuid.New(),
		Name:       uuid.NewString(),
		HasInput:   true,
		CreateTime: time.Now().UTC(),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func GenScheduledTestExec(testID uuid.V7) *test.ScheduledTestExecution {
	return &test.ScheduledTestExecution{
		ID:           test.NewTestExecutionID(),
		TestID:       testID,
		HasInput:     true,
		ScheduleTime: time.Now().UTC(),
	}
}

func GenStartedTestExec(testExecID test.TestExecutionID) *test.StartedTestExecution {
	return &test.StartedTestExecution{
		ID:        testExecID,
		StartTime: time.Now().UTC(),
	}
}

func GenFinishedTestExec(testExecID test.TestExecutionID, err *string) *test.FinishedTestExecution {
	return &test.FinishedTestExecution{
		ID:         testExecID,
		FinishTime: time.Now().UTC(),
		Error:      err,
	}
}

func GenTestExec(testID uuid.V7) *test.TestExecution {
	return &test.TestExecution{
		ID:           test.NewTestExecutionID(),
		TestID:       testID,
		HasInput:     false,
		ScheduleTime: time.Now().UTC().Add(-2 * time.Millisecond),
		StartTime:    ptr.Get(time.Now().UTC().Add(-time.Millisecond)),
		FinishTime:   ptr.Get(time.Now().UTC()),
		Error:        nil,
	}
}

func GenResetTestExec(existing *test.TestExecution) *test.TestExecution {
	reset := *existing
	reset.ScheduleTime = time.Now().UTC()
	reset.StartTime = nil
	reset.FinishTime = nil
	reset.Error = nil
	return &reset
}

func GenCaseID() test.CaseExecutionID {
	mu.Lock()
	currCaseID++
	mu.Unlock()
	return currCaseID
}

func GenScheduledCaseExec(testExecID test.TestExecutionID) *test.ScheduledCaseExecution {
	return &test.ScheduledCaseExecution{
		ID:              GenCaseID(),
		TestExecutionID: testExecID,
		CaseName:        uuid.NewString(),
		ScheduleTime:    time.Now().UTC(),
	}
}

func GenStartedCaseExec(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) *test.StartedCaseExecution {
	return &test.StartedCaseExecution{
		ID:              caseExecID,
		TestExecutionID: testExecID,
		StartTime:       time.Now().UTC(),
	}
}

func GenFinishedCaseExec(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID, err *string) *test.FinishedCaseExecution {
	return &test.FinishedCaseExecution{
		ID:              caseExecID,
		TestExecutionID: testExecID,
		FinishTime:      time.Now().UTC(),
		Error:           err,
	}
}

func GenCaseExec(testExecID test.TestExecutionID) *test.CaseExecution {
	return &test.CaseExecution{
		ID:              GenCaseID(),
		TestExecutionID: testExecID,
		CaseName:        uuid.NewString(),
		ScheduleTime:    time.Now().UTC().Add(-2 * time.Millisecond),
		StartTime:       ptr.Get(time.Now().UTC().Add(-time.Millisecond)),
		FinishTime:      ptr.Get(time.Now().UTC()),
		Error:           nil,
	}
}

func GenTestExecLog(testExecID test.TestExecutionID) *test.Log {
	return genExecLog(testExecID, nil)
}

func GenTestExecLogs(testExecID test.TestExecutionID, count int) test.LogList {
	logs := make([]*test.Log, count)
	for i := range count {
		logs[i] = genExecLog(testExecID, nil)
	}
	return logs
}

func GenCaseExecLog(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) *test.Log {
	return genExecLog(testExecID, &caseExecID)
}

func GenCaseExecLogs(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID, count int) test.LogList {
	logs := make([]*test.Log, count)
	for i := range count {
		logs[i] = GenCaseExecLog(testExecID, caseExecID)
	}
	return logs
}

func genExecLog(testExecID test.TestExecutionID, caseExecID *test.CaseExecutionID) *test.Log {
	return &test.Log{
		ID:              uuid.New(),
		TestExecutionID: testExecID,
		CaseExecutionID: caseExecID,
		Level:           "INFO",
		Message:         uuid.NewString(),
		CreateTime:      time.Now().UTC(),
	}
}

var (
	mu         = new(sync.RWMutex)
	currCaseID = test.CaseExecutionID(0)
)

type Data struct {
	Foo int
	Bar string
}

func GenData() *Data {
	return &Data{
		Foo: rand.Int(),
		Bar: uuid.NewString(),
	}
}
