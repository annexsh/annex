package postgres

import "github.com/annexsh/annex/test"

func NewTestRepository(db *DB) test.Repository {
	return &testRepository{
		TestReader:          NewTestReader(db),
		TestWriter:          NewTestWriter(db),
		TestExecutionReader: NewTestExecutionReader(db),
		TestExecutionWriter: NewTestExecutionWriter(db),
		CaseExecutionReader: NewCaseExecutionReader(db),
		CaseExecutionWriter: NewCaseExecutionWriter(db),
		LogReader:           NewLogReader(db),
		LogWriter:           NewLogWriter(db),
	}
}

type testRepository struct {
	*TestReader
	*TestWriter
	*TestExecutionReader
	*TestExecutionWriter
	*CaseExecutionReader
	*CaseExecutionWriter
	*LogReader
	*LogWriter
}
