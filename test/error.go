package test

const (
	ErrorContextAlreadyExists         = testErr("context already exists")
	ErrorTestNotFound                 = testErr("test not found")
	ErrorTestPayloadNotFound          = testErr("test payload not found")
	ErrorTestExecutionNotFound        = testErr("test execution not found")
	ErrorTestExecutionPayloadNotFound = testErr("test execution payload not found")
	ErrorCaseExecutionNotFound        = testErr("case execution not found")
	ErrorLogNotFound                  = testErr("execution log not found")
	ErrorNotTestExecution             = testErr("workflow is not a test execution")
	ErrorNotCaseExecution             = testErr("activity is not a test execution")
)

type testErr string

func (e testErr) Error() string {
	return string(e)
}
