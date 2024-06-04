package inmem

import "github.com/annexsh/annex/test"

func NewTestRepository(db *DB) test.Repository {
	return &testRepository{
		TestReader:          NewTestReader(db),
		TestWriter:          NewTestWriter(db),
		TestExecutionReader: NewTestExecutionReader(db),
		TestExecutionWriter: NewTestExecutionWriter(db),
		CaseExecutionReader: NewCaseExecutionReader(db),
		CaseExecutionWriter: NewCaseExecutionWriter(db),
		ExecutionLogReader:  NewExecutionLogReader(db),
		ExecutionLogWriter:  NewExecutionLogWriter(db),
	}
}

type testRepository struct {
	*TestReader
	*TestWriter
	*TestExecutionReader
	*TestExecutionWriter
	*CaseExecutionReader
	*CaseExecutionWriter
	*ExecutionLogReader
	*ExecutionLogWriter
}
