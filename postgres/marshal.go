package postgres

import (
	"go.temporal.io/sdk/converter"

	"github.com/annexsh/annex/postgres/sqlc"

	"github.com/annexsh/annex/test"
)

func marshalTestSuite(t *sqlc.TestSuite) *test.TestSuite {
	return &test.TestSuite{
		ID:          t.ID,
		ContextID:   t.ContextID,
		Name:        t.Name,
		Description: t.Description,
	}
}

func marshalTestSuites(suites []*sqlc.TestSuite) []*test.TestSuite {
	out := make([]*test.TestSuite, len(suites))
	for i, t := range suites {
		out[i] = marshalTestSuite(t)
	}
	return out
}

func marshalTest(t *sqlc.Test) *test.Test {
	return &test.Test{
		ContextID:   t.ContextID,
		TestSuiteID: t.TestSuiteID,
		ID:          t.ID,
		Name:        t.Name,
		HasInput:    t.HasInput,
		CreateTime:  t.CreateTime,
	}
}

func marshalTests(tests []*sqlc.Test) []*test.Test {
	out := make([]*test.Test, len(tests))
	for i, t := range tests {
		out[i] = marshalTest(t)
	}
	return out
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
	out := make([]*test.TestExecution, len(testExecs))
	for i, testExec := range testExecs {
		out[i] = marshalTestExec(testExec)
	}
	return out
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
	out := make([]*test.CaseExecution, len(caseExecs))
	for i, caseExec := range caseExecs {
		out[i] = marshalCaseExec(caseExec)
	}
	return out
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
	out := make([]*test.Log, len(logs))
	for i, log := range logs {
		out[i] = marshalLog(log)
	}
	return out
}
