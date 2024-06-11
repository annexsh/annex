package inmem

import "github.com/annexsh/annex/test"

func NewTestRepository(db *DB) test.Repository {
	return &testRepository{
		ContextReader:       NewContextReader(db),
		ContextWriter:       NewContextWriter(db),
		GroupReader:         NewGroupReader(db),
		GroupWriter:         NewGroupWriter(db),
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
	*ContextReader
	*ContextWriter
	*GroupReader
	*GroupWriter
	*TestReader
	*TestWriter
	*TestExecutionReader
	*TestExecutionWriter
	*CaseExecutionReader
	*CaseExecutionWriter
	*LogReader
	*LogWriter
}
