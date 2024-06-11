package postgres

import (
	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/test"
)

func marshalTest(t *sqlc.Test) *test.Test {
	return &test.Test{
		Context:    t.Context,
		Group:      t.Group,
		ID:         t.ID,
		Name:       t.Name,
		HasInput:   t.HasInput,
		CreateTime: t.CreateTime.Time,
	}
}

func marshalTests(tests []*sqlc.Test) []*test.Test {
	testspb := make([]*test.Test, len(tests))
	for i, t := range tests {
		testspb[i] = marshalTest(t)
	}
	return testspb
}

func marshalTestDefaultInput(input *sqlc.TestDefaultInput) *test.Payload {
	return &test.Payload{
		Data: input.Data,
	}
}

func marshalTestExecPayload(input *sqlc.TestExecutionInput) *test.Payload {
	return &test.Payload{
		Data: input.Data,
	}
}

func marshalTestExec(testExec *sqlc.TestExecution) *test.TestExecution {
	t := &test.TestExecution{
		ID:           testExec.ID,
		TestID:       testExec.TestID,
		HasInput:     testExec.HasInput,
		ScheduleTime: testExec.ScheduleTime.Time,
		Error:        testExec.Error,
	}
	if testExec.StartTime.Valid {
		t.StartTime = &testExec.StartTime.Time
	}
	if testExec.FinishTime.Valid {
		t.FinishTime = &testExec.FinishTime.Time
	}
	return t
}

func marshalTestExecs(testExecs []*sqlc.TestExecution) []*test.TestExecution {
	te := make([]*test.TestExecution, len(testExecs))
	for i, testExec := range testExecs {
		te[i] = marshalTestExec(testExec)
	}
	return te
}

func marshalCaseExec(caseExec *sqlc.CaseExecution) *test.CaseExecution {
	c := &test.CaseExecution{
		ID:              caseExec.ID,
		TestExecutionID: caseExec.TestExecutionID,
		CaseName:        caseExec.CaseName,
		ScheduleTime:    caseExec.ScheduleTime.Time,
		Error:           caseExec.Error,
	}
	if caseExec.StartTime.Valid {
		c.StartTime = &caseExec.StartTime.Time
	}
	if caseExec.FinishTime.Valid {
		c.FinishTime = &caseExec.FinishTime.Time
	}
	return c
}

func marshalCaseExecs(caseExecs []*sqlc.CaseExecution) []*test.CaseExecution {
	ce := make([]*test.CaseExecution, len(caseExecs))
	for i, caseExec := range caseExecs {
		ce[i] = marshalCaseExec(caseExec)
	}
	return ce
}

func marshalExecLog(log *sqlc.Log) *test.Log {
	return &test.Log{
		ID:              log.ID,
		TestExecutionID: log.TestExecutionID,
		CaseExecutionID: log.CaseExecutionID,
		Level:           log.Level,
		Message:         log.Message,
		CreateTime:      log.CreateTime.Time,
	}
}

func marshalExecLogs(logs []*sqlc.Log) []*test.Log {
	execLogs := make([]*test.Log, len(logs))
	for i, log := range logs {
		execLogs[i] = marshalExecLog(log)
	}
	return execLogs
}
