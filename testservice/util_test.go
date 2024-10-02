package testservice

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func wantBlankContextFieldViolation(streamIndex ...int) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldViolationName("context", streamIndex...),
		Description: "Context can't be blank",
	}
}

func wantBlankTestSuiteFieldViolation(streamIndex ...int) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldViolationName("test_suite_id", streamIndex...),
		Description: "Test suite id can't be blank",
	}
}

func wantBlankTestIDFieldViolation(streamIndex ...int) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldViolationName("test_id", streamIndex...),
		Description: "Test id can't be blank",
	}
}

func wantTestIDNotUUIDFieldViolation(streamIndex ...int) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldViolationName("test_id", streamIndex...),
		Description: "Test id must be a v7 UUID",
	}
}

func wantBlankTestExecIDFieldViolation(streamIndex ...int) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldViolationName("test_execution_id", streamIndex...),
		Description: "Test execution id can't be blank",
	}
}

func wantTestExecIDNotUUIDFieldViolation(streamIndex ...int) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldViolationName("test_execution_id", streamIndex...),
		Description: "Test execution id must be a v7 UUID",
	}
}

func wantPageSizeFieldViolation(streamIndex ...int) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       fieldViolationName("page_size", streamIndex...),
		Description: fmt.Sprintf(`Page size must be between "0" and "%d"`, maxPageSize),
	}
}

func fieldViolationName(name string, streamIndex ...int) string {
	if len(streamIndex) > 1 {
		panic("only one `streamIndex` permitted")
	} else if len(streamIndex) == 1 {
		return fmt.Sprintf("stream[%d].%s", streamIndex[0], name)
	}
	return name
}

func assertInvalidRequest(t *testing.T, err error, wantFieldViolation *errdetails.BadRequest_FieldViolation, isStream ...bool) {
	baseMsg := reqValidationBaseErrMsg

	if len(isStream) > 1 {
		panic("only one `isStream` permitted")
	} else if len(isStream) == 1 && isStream[0] {
		baseMsg = streamReqValidationBaseErrMsg
	}

	var cErr *connect.Error
	require.ErrorAs(t, err, &cErr)

	assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
	assert.Equal(t, baseMsg, cErr.Message())

	require.Len(t, cErr.Details(), 1)
	val, err := cErr.Details()[0].Value()
	require.NoError(t, err)

	badReq, ok := val.(*errdetails.BadRequest)
	require.True(t, ok, "proto message is not BadRequest")

	require.Len(t, badReq.FieldViolations, 1)
	assert.Equal(t, wantFieldViolation, badReq.FieldViolations[0])
}
