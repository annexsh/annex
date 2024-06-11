package test

const (
	ErrorTestNotFound          = testErr("test not found")
	ErrorTestExecutionNotFound = testErr("test execution not found")
	ErrorCaseExecutionNotFound = testErr("case execution not found")
	ErrorLogNotFound           = testErr("execution log not found")
	ErrorNotTestExecution      = testErr("workflow is not a test execution")
	ErrorNotCaseExecution      = testErr("activity is not a test execution")
)

type testErr string

func (e testErr) Error() string {
	return string(e)
}
