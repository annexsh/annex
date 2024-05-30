package test

import (
	"time"

	testv1 "github.com/annexhq/annex-proto/gen/go/type/test/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexhq/annex/internal/ptr"
)

const runnerTTL = time.Minute + 30*time.Second

func (t *Test) Proto() *testv1.Test {
	testpb := &testv1.Test{
		Id:         t.ID.String(),
		Project:    t.Project,
		Name:       t.Name,
		HasPayload: t.HasPayload,
		CreatedAt:  timestamppb.New(t.CreateTime),
	}

	if len(t.Runners) > 0 {
		runner := t.Runners[0]
		testpb.LastAvailable = &testv1.TestRunner{
			Id:            runner.ID,
			LastHeartbeat: timestamppb.New(runner.LastHeartbeatTime),
			IsActive:      time.Since(runner.LastHeartbeatTime) < runnerTTL,
		}
	}

	return testpb
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
		Data:     p.Payload,
		IsZero:   p.IsZero,
	}
}

func (t *TestExecution) Proto() *testv1.TestExecution {
	exec := &testv1.TestExecution{
		Id:          t.ID.String(),
		TestId:      t.TestID.String(),
		Error:       t.Error,
		ScheduledAt: timestamppb.New(t.ScheduleTime),
		StartedAt:   nil,
		FinishedAt:  nil,
	}
	if t.StartTime != nil {
		exec.StartedAt = timestamppb.New(*t.StartTime)
	}
	if t.FinishTime != nil {
		exec.FinishedAt = timestamppb.New(*t.FinishTime)
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
		Id:          c.ID.Int32(),
		CaseName:    c.CaseName,
		TestExecId:  c.TestExecID.String(),
		ScheduledAt: timestamppb.New(c.ScheduleTime),
		FinishedAt:  nil,
		Error:       c.Error,
	}
	if c.StartTime != nil {
		exec.StartedAt = timestamppb.New(*c.StartTime)
	}
	if c.FinishTime != nil {
		exec.FinishedAt = timestamppb.New(*c.FinishTime)
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

func (l *ExecutionLog) Proto() *testv1.ExecutionLog {
	var caseExecID *int32
	if l.CaseExecID != nil {
		caseExecID = ptr.Get(l.CaseExecID.Int32())
	}

	return &testv1.ExecutionLog{
		Id:         l.ID.String(),
		TestExecId: l.TestExecID.String(),
		CaseExecId: caseExecID,
		Level:      l.Level,
		Message:    l.Message,
		CreatedAt:  timestamppb.New(l.CreateTime),
	}
}

func (l ExecutionLogList) Proto() []*testv1.ExecutionLog {
	execLogs := make([]*testv1.ExecutionLog, len(l))
	for i, log := range l {
		execLogs[i] = log.Proto()
	}
	return execLogs
}
