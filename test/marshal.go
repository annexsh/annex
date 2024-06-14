package test

import (
	testv1 "github.com/annexsh/annex-proto/gen/go/type/test/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/internal/ptr"
)

func (t *Test) Proto() *testv1.Test {
	return &testv1.Test{
		Context:    t.ContextID,
		Group:      t.GroupID,
		Id:         t.ID.String(),
		Name:       t.Name,
		HasInput:   t.HasInput,
		CreateTime: timestamppb.New(t.CreateTime),
	}
}

func (t TestList) Proto() []*testv1.Test {
	testspb := make([]*testv1.Test, len(t))
	for i, test := range t {
		testspb[i] = test.Proto()
	}
	return testspb
}

func (p *Payload) Proto() *testv1.Payload {
	return &testv1.Payload{
		Metadata: p.Metadata,
		Data:     p.Data,
	}
}

func (t *TestExecution) Proto() *testv1.TestExecution {
	exec := &testv1.TestExecution{
		Id:           t.ID.String(),
		TestId:       t.TestID.String(),
		Error:        t.Error,
		ScheduleTime: timestamppb.New(t.ScheduleTime),
		StartTime:    nil,
		FinishTime:   nil,
	}
	if t.StartTime != nil {
		exec.StartTime = timestamppb.New(*t.StartTime)
	}
	if t.FinishTime != nil {
		exec.FinishTime = timestamppb.New(*t.FinishTime)
	}
	return exec
}

func (t TestExecutionList) Proto() []*testv1.TestExecution {
	fe := make([]*testv1.TestExecution, len(t))
	for i, testExec := range t {
		fe[i] = testExec.Proto()
	}
	return fe
}

func (c *CaseExecution) Proto() *testv1.CaseExecution {
	exec := &testv1.CaseExecution{
		Id:              c.ID.Int32(),
		CaseName:        c.CaseName,
		TestExecutionId: c.TestExecutionID.String(),
		ScheduleTime:    timestamppb.New(c.ScheduleTime),
		FinishTime:      nil,
		Error:           c.Error,
	}
	if c.StartTime != nil {
		exec.StartTime = timestamppb.New(*c.StartTime)
	}
	if c.FinishTime != nil {
		exec.FinishTime = timestamppb.New(*c.FinishTime)
	}
	return exec
}

func (c CaseExecutionList) Proto() []*testv1.CaseExecution {
	ce := make([]*testv1.CaseExecution, len(c))
	for i, caseExec := range c {
		ce[i] = caseExec.Proto()
	}
	return ce
}

func (l *Log) Proto() *testv1.Log {
	var caseExecID *int32
	if l.CaseExecutionID != nil {
		caseExecID = ptr.Get(l.CaseExecutionID.Int32())
	}

	return &testv1.Log{
		Id:              l.ID.String(),
		TestExecutionId: l.TestExecutionID.String(),
		CaseExecutionId: caseExecID,
		Level:           l.Level,
		Message:         l.Message,
		CreateTime:      timestamppb.New(l.CreateTime),
	}
}

func (l LogList) Proto() []*testv1.Log {
	execLogs := make([]*testv1.Log, len(l))
	for i, log := range l {
		execLogs[i] = log.Proto()
	}
	return execLogs
}
