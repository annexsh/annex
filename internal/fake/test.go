package fake

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/converter"

	"github.com/annexsh/annex/internal/ptr"
	"github.com/annexsh/annex/test"
)

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

func GenTestDefinition() *test.TestDefinition {
	return &test.TestDefinition{
		TestID:       uuid.New(),
		Group:        uuid.NewString(),
		Name:         uuid.NewString(),
		DefaultInput: GenDefaultInput(),
		RunnerID:     uuid.NewString(),
	}
}

func GenTest() *test.Test {
	return &test.Test{
		ID:       uuid.New(),
		Group:    uuid.NewString(),
		Name:     uuid.NewString(),
		HasInput: true,
		Runners: []*test.Runner{
			{
				ID:                uuid.NewString(),
				LastHeartbeatTime: time.Now(),
				IsActive:          true,
			},
		},
		CreateTime: time.Now(),
	}
}

func GenScheduledTestExec(testID uuid.UUID) *test.ScheduledTestExecution {
	return &test.ScheduledTestExecution{
		ID:           test.NewTestExecutionID(),
		TestID:       testID,
		Payload:      GenInput().Data,
		ScheduleTime: time.Now(),
	}
}

func GenStartedTestExec(testExecID test.TestExecutionID) *test.StartedTestExecution {
	return &test.StartedTestExecution{
		ID:        testExecID,
		StartTime: time.Now(),
	}
}

func GenFinishedTestExec(testExecID test.TestExecutionID, err *string) *test.FinishedTestExecution {
	return &test.FinishedTestExecution{
		ID:         testExecID,
		FinishTime: time.Now(),
		Error:      err,
	}
}

func GenTestExec(testID uuid.UUID) *test.TestExecution {
	return &test.TestExecution{
		ID:           test.NewTestExecutionID(),
		TestID:       testID,
		HasInput:     false,
		ScheduleTime: time.Now().Add(-2 * time.Millisecond),
		StartTime:    ptr.Get(time.Now().Add(-time.Millisecond)),
		FinishTime:   ptr.Get(time.Now()),
		Error:        nil,
	}
}

func GenCaseID() test.CaseExecutionID {
	mu.Lock()
	currCaseID++
	mu.Unlock()
	return currCaseID
}

func GenScheduledCaseExec(testExecID test.TestExecutionID) *test.ScheduledCaseExecution {
	return &test.ScheduledCaseExecution{
		ID:           GenCaseID(),
		TestExecID:   testExecID,
		CaseName:     uuid.NewString(),
		ScheduleTime: time.Now().UTC(),
	}
}

func GenStartedCaseExec(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) *test.StartedCaseExecution {
	return &test.StartedCaseExecution{
		ID:              caseExecID,
		TestExecutionID: testExecID,
		StartTime:       time.Now(),
	}
}

func GenFinishedCaseExec(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID, err *string) *test.FinishedCaseExecution {
	return &test.FinishedCaseExecution{
		ID:              caseExecID,
		TestExecutionID: testExecID,
		FinishTime:      time.Now(),
		Error:           err,
	}
}

func GenCaseExec(testExecID test.TestExecutionID) *test.CaseExecution {
	return &test.CaseExecution{
		ID:              GenCaseID(),
		TestExecutionID: testExecID,
		CaseName:        uuid.NewString(),
		ScheduleTime:    time.Now().Add(-2 * time.Millisecond),
		StartTime:       ptr.Get(time.Now().Add(-time.Millisecond)),
		FinishTime:      ptr.Get(time.Now()),
		Error:           nil,
	}
}

func GenTestExecLog(testExecID test.TestExecutionID) *test.Log {
	return genExecLog(testExecID, nil)
}

func GenTestExecLogs(count int, testExecID test.TestExecutionID) []*test.Log {
	logs := make([]*test.Log, count)
	for i := range count {
		logs[i] = genExecLog(testExecID, nil)
	}
	return logs
}

func GenCaseExecLog(testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) *test.Log {
	return genExecLog(testExecID, nil)
}

func GenCaseExecLogs(count int, testExecID test.TestExecutionID, caseExecID test.CaseExecutionID) []*test.Log {
	logs := make([]*test.Log, count)
	for i := range count {
		logs[i] = genExecLog(testExecID, &caseExecID)
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
		CreateTime:      time.Now(),
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
