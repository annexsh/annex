package postgres

import (
	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/test"
)

func marshalTest(t *sqlc.Test) *test.Test {
	return &test.Test{
		ID:         t.ID,
		Project:    t.Project,
		Name:       t.Name,
		HasPayload: t.HasPayload,
		CreateTime: t.CreatedAt.Time,
		Runners: []*test.TestRunner{
			{
				ID:                t.RunnerID,
				LastHeartbeatTime: t.RunnerHeartbeatAt.Time,
			},
		},
	}
}

func marshalTests(tests []*sqlc.Test) []*test.Test {
	testspb := make([]*test.Test, len(tests))
	for i, t := range tests {
		testspb[i] = marshalTest(t)
	}
	return testspb
}

func marshalTestDefaultPayload(payload *sqlc.TestDefaultPayload) *test.Payload {
	return &test.Payload{
		Payload: payload.Payload,
		IsZero:  payload.IsZero,
	}
}

func marshalTestExecPayload(payload *sqlc.TestExecutionPayload) *test.Payload {
	return &test.Payload{
		Payload: payload.Payload,
		IsZero:  false,
	}
}

func marshalTestExec(testExec *sqlc.TestExecution) *test.TestExecution {
	t := &test.TestExecution{
		ID:           testExec.ID,
		TestID:       testExec.TestID,
		HasPayload:   testExec.HasPayload,
		ScheduleTime: testExec.ScheduledAt.Time,
		Error:        testExec.Error,
	}
	if testExec.StartedAt.Valid {
		t.StartTime = &testExec.StartedAt.Time
	}
	if testExec.FinishedAt.Valid {
		t.FinishTime = &testExec.FinishedAt.Time
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
		ID:           caseExec.ID,
		TestExecID:   caseExec.TestExecID,
		CaseName:     caseExec.CaseName,
		ScheduleTime: caseExec.ScheduledAt.Time,
		Error:        caseExec.Error,
	}
	if caseExec.StartedAt.Valid {
		c.StartTime = &caseExec.StartedAt.Time
	}
	if caseExec.FinishedAt.Valid {
		c.FinishTime = &caseExec.FinishedAt.Time
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

func marshalExecLog(log *sqlc.Log) *test.ExecutionLog {
	return &test.ExecutionLog{
		ID:         log.ID,
		TestExecID: log.TestExecID,
		CaseExecID: log.CaseExecID,
		Level:      log.Level,
		Message:    log.Message,
		CreateTime: log.CreatedAt.Time,
	}
}

func marshalExecLogs(logs []*sqlc.Log) []*test.ExecutionLog {
	execLogs := make([]*test.ExecutionLog, len(logs))
	for i, log := range logs {
		execLogs[i] = marshalExecLog(log)
	}
	return execLogs
}
