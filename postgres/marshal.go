package postgres

import (
	"go.temporal.io/sdk/converter"

	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/test"
)

func marshalTest(t *sqlc.Test) *test.Test {
	return &test.Test{
		ContextID:  t.ContextID,
		GroupID:    t.GroupID,
		ID:         t.ID,
		Name:       t.Name,
		HasInput:   t.HasInput,
		CreateTime: t.CreateTime,
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
		Data:     input.Data,
		Metadata: map[string][]byte{"encoding": []byte(converter.MetadataEncodingJSON)},
	}
}

func marshalTestExecInput(input *sqlc.TestExecutionInput) *test.Payload {
	return &test.Payload{
		Data:     input.Data,
		Metadata: map[string][]byte{"encoding": []byte(converter.MetadataEncodingJSON)},
	}
}

func marshalTestExec(testExec *sqlc.TestExecution) *test.TestExecution {
	return &test.TestExecution{
		ID:           testExec.ID,
		TestID:       testExec.TestID,
		HasInput:     testExec.HasInput,
		ScheduleTime: testExec.ScheduleTime,
		StartTime:    testExec.StartTime,
		FinishTime:   testExec.FinishTime,
		Error:        testExec.Error,
	}
}

func marshalTestExecs(testExecs []*sqlc.TestExecution) []*test.TestExecution {
	te := make([]*test.TestExecution, len(testExecs))
	for i, testExec := range testExecs {
		te[i] = marshalTestExec(testExec)
	}
	return te
}

func marshalCaseExec(caseExec *sqlc.CaseExecution) *test.CaseExecution {
	return &test.CaseExecution{
		ID:              caseExec.ID,
		TestExecutionID: caseExec.TestExecutionID,
		CaseName:        caseExec.CaseName,
		ScheduleTime:    caseExec.ScheduleTime,
		StartTime:       caseExec.StartTime,
		FinishTime:      caseExec.FinishTime,
		Error:           caseExec.Error,
	}
}

func marshalCaseExecs(caseExecs []*sqlc.CaseExecution) []*test.CaseExecution {
	ce := make([]*test.CaseExecution, len(caseExecs))
	for i, caseExec := range caseExecs {
		ce[i] = marshalCaseExec(caseExec)
	}
	return ce
}

func marshalLog(log *sqlc.Log) *test.Log {
	return &test.Log{
		ID:              log.ID,
		TestExecutionID: log.TestExecutionID,
		CaseExecutionID: log.CaseExecutionID,
		Level:           log.Level,
		Message:         log.Message,
		CreateTime:      log.CreateTime,
	}
}

func marshalExecLogs(logs []*sqlc.Log) []*test.Log {
	execLogs := make([]*test.Log, len(logs))
	for i, log := range logs {
		execLogs[i] = marshalLog(log)
	}
	return execLogs
}
