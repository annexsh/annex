package testservice

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	wantBlankContextFieldViolation = &errdetails.BadRequest_FieldViolation{
		Field:       "context",
		Description: "Context can't be blank",
	}
	wantBlankGroupFieldViolation = &errdetails.BadRequest_FieldViolation{
		Field:       "group",
		Description: "Group can't be blank",
	}
	wantBlankTestIDFieldViolation = &errdetails.BadRequest_FieldViolation{
		Field:       "test_id",
		Description: "Test id can't be blank",
	}
	wantTestIDNotUUIDFieldViolation = &errdetails.BadRequest_FieldViolation{
		Field:       "test_id",
		Description: "Test id must be a v7 UUID",
	}
	wantBlankTestExecIDFieldViolation = &errdetails.BadRequest_FieldViolation{
		Field:       "test_execution_id",
		Description: "Test execution id can't be blank",
	}
	wantTestExecIDNotUUIDFieldViolation = &errdetails.BadRequest_FieldViolation{
		Field:       "test_execution_id",
		Description: "Test execution id must be a v7 UUID",
	}
	wantPageSizeFieldViolation = &errdetails.BadRequest_FieldViolation{
		Field:       "page_size",
		Description: fmt.Sprintf(`Page size must be between "0" and "%d"`, maxPageSize),
	}
)

func assertInvalidRequest(t *testing.T, err error, wantFieldViolation *errdetails.BadRequest_FieldViolation) {
	var cErr *connect.Error
	require.ErrorAs(t, err, &cErr)

	assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
	assert.Equal(t, reqValidationBaseErrMsg, cErr.Message())

	require.Len(t, cErr.Details(), 1)
	val, err := cErr.Details()[0].Value()
	require.NoError(t, err)

	badReq, ok := val.(*errdetails.BadRequest)
	require.True(t, ok, "proto message is not BadRequest")

	require.Len(t, badReq.FieldViolations, 1)
	assert.Equal(t, wantFieldViolation, badReq.FieldViolations[0])
}
