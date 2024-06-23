package test

import (
	testsv1 "github.com/annexsh/annex-proto/gen/go/annex/tests/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/internal/ptr"
)

func (t *Test) Proto() *testsv1.Test {
	return &testsv1.Test{
		Context:    t.ContextID,
		Group:      t.GroupID,
		Id:         t.ID.String(),
		Name:       t.Name,
		HasInput:   t.HasInput,
		CreateTime: timestamppb.New(t.CreateTime),
	}
}

func (t TestList) Proto() []*testsv1.Test {
	testspb := make([]*testsv1.Test, len(t))
	for i, test := range t {
		testspb[i] = test.Proto()
	}
	return testspb
}

func (p *Payload) Proto() *testsv1.Payload {
	return &testsv1.Payload{
		Metadata: p.Metadata,
		Data:     p.Data,
	}
}

func (t *TestExecution) Proto() *testsv1.TestExecution {
	exec := &testsv1.TestExecution{
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

func (t TestExecutionList) Proto() []*testsv1.TestExecution {
	fe := make([]*testsv1.TestExecution, len(t))
	for i, testExec := range t {
		fe[i] = testExec.Proto()
	}
	return fe
}

func (c *CaseExecution) Proto() *testsv1.CaseExecution {
	exec := &testsv1.CaseExecution{
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

func (c CaseExecutionList) Proto() []*testsv1.CaseExecution {
	ce := make([]*testsv1.CaseExecution, len(c))
	for i, caseExec := range c {
		ce[i] = caseExec.Proto()
	}
	return ce
}

func (l *Log) Proto() *testsv1.Log {
	var caseExecID *int32
	if l.CaseExecutionID != nil {
		caseExecID = ptr.Get(l.CaseExecutionID.Int32())
	}

	return &testsv1.Log{
		Id:              l.ID.String(),
		TestExecutionId: l.TestExecutionID.String(),
		CaseExecutionId: caseExecID,
		Level:           l.Level,
		Message:         l.Message,
		CreateTime:      timestamppb.New(l.CreateTime),
	}
}

func (l LogList) Proto() []*testsv1.Log {
	execLogs := make([]*testsv1.Log, len(l))
	for i, log := range l {
		execLogs[i] = log.Proto()
	}
	return execLogs
}
