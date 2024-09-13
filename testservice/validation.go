package testservice

import (
	"fmt"

	testsv1 "github.com/annexsh/annex-proto/go/gen/annex/tests/v1"
	"github.com/cohesivestack/valgo"
	"go.temporal.io/sdk/converter"

	"github.com/annexsh/annex/internal/validator"
)

const (
	reqValidationBaseErrMsg = "invalid request"
	maxPageSize             = 1000
	maxRegisterTests        = 50
)

func validateRegisterContextRequest(req *testsv1.RegisterContextRequest) error {
	v := newValidator()
	v.Is(validator.Context(req.Context))
	return v.ConnectError()
}

func validateListContextsRequest(req *testsv1.ListContextsRequest) error {
	v := newValidator()
	v.Is(validator.PageSize(req.PageSize, maxPageSize))
	return v.ConnectError()
}

func validateRegisterGroupRequest(req *testsv1.RegisterGroupRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		valgo.String(req.Name, "name").Not().Blank(),
	)
	return v.ConnectError()
}

func validateListGroupsRequest(req *testsv1.ListGroupsRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.PageSize(req.PageSize, maxPageSize),
	)
	return v.ConnectError()
}

func validateRegisterTestsRequest(req *testsv1.RegisterTestsRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.Group(req.Group),
	)

	if len(req.Definitions) == 0 {
		v.AddErrorMessage("definitions", "Definitions must not be empty")
	} else if len(req.Definitions) > maxRegisterTests {
		v.AddErrorMessage("definitions", fmt.Sprintf("Definitions must not exceed a length of %d per request", maxRegisterTests))
	}

	for i, def := range req.Definitions {
		defValidator := valgo.Is(valgo.String(def.Name).Not().Blank())
		if def.DefaultInput != nil {
			validatePayload(defValidator, "default_input", def.DefaultInput)
		}
		v.InRow("definitions", i, defValidator)
	}

	return v.ConnectError()
}

func validateGetTestRequest(req *testsv1.GetTestRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestID(req.TestId),
	)
	return v.ConnectError()
}

func validateGetTestDefaultInputRequest(req *testsv1.GetTestDefaultInputRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestID(req.TestId),
	)
	return v.ConnectError()
}

func validateListTestsRequest(req *testsv1.ListTestsRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.Group(req.Group),
		validator.PageSize(req.PageSize, maxPageSize),
	)
	return v.ConnectError()
}

func validateExecuteTestRequest(req *testsv1.ExecuteTestRequest) error {
	v := newValidator()

	v.Is(
		validator.Context(req.Context),
		validator.TestID(req.TestId),
	)

	if req.Input != nil {
		validatePayload(v.Validation, "input", req.Input)
	}

	return v.ConnectError()
}

func validateExecuteTestRequestInputRequired(requiresInput bool, req *testsv1.ExecuteTestRequest) error {
	v := newValidator()

	if requiresInput {
		v.Is(valgo.Any(req.Input, "input").Not().Nil(fmt.Sprintf("{{title}} required for test '%s'", req.TestId)))
	} else {
		v.Is(valgo.Any(req.Input, "input").Nil(fmt.Sprintf("{{title}} not required for test '%s'", req.TestId)))
	}

	return v.ConnectError()
}

func validateGetTestExecutionRequest(req *testsv1.GetTestExecutionRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
	)
	return v.ConnectError()
}

func validateListTestExecutionsRequest(req *testsv1.ListTestExecutionsRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestID(req.TestId),
		validator.PageSize(req.PageSize, maxPageSize),
	)
	return v.ConnectError()
}

func validateAckTestExecutionStartedRequest(req *testsv1.AckTestExecutionStartedRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		validator.Timestamppb(req.StartTime, "start_time"),
	)
	return v.ConnectError()
}

func validateAckTestExecutionFinishedRequest(req *testsv1.AckTestExecutionFinishedRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		validator.Timestamppb(req.FinishTime, "finish_time"),
	)
	if req.Error != nil {
		v.Is(valgo.StringP(req.Error, "error").Not().Blank())
	}
	return v.ConnectError()
}

func validateRetryTestExecutionRequest(req *testsv1.RetryTestExecutionRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
	)
	return v.ConnectError()
}

func validateListCaseExecutionsRequest(req *testsv1.ListCaseExecutionsRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		validator.PageSize(req.PageSize, maxPageSize),
	)
	return v.ConnectError()
}

func validateAckCaseExecutionScheduledRequest(req *testsv1.AckCaseExecutionScheduledRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		validator.CaseExecID(req.CaseExecutionId),
		valgo.String(req.CaseName, "case_name").Not().Blank(),
		validator.Timestamppb(req.ScheduleTime, "schedule_time"),
	)
	return v.ConnectError()
}

func validateAckCaseExecutionStartedRequest(req *testsv1.AckCaseExecutionStartedRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		validator.CaseExecID(req.CaseExecutionId),
		validator.Timestamppb(req.StartTime, "start_time"),
	)
	return v.ConnectError()
}

func validateAckCaseExecutionFinishedRequest(req *testsv1.AckCaseExecutionFinishedRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		validator.CaseExecID(req.CaseExecutionId),
		validator.Timestamppb(req.FinishTime, "finish_time"),
	)
	if req.Error != nil {
		v.Is(valgo.StringP(req.Error, "error").Not().Blank())
	}
	return v.ConnectError()
}

func validatePublishLogRequest(req *testsv1.PublishLogRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		valgo.String(req.Level, "level").Not().Blank(),
		valgo.String(req.Message, "message").Not().Blank(),
		validator.Timestamppb(req.CreateTime, "create_time"),
	)
	if req.CaseExecutionId != nil {
		v.Is(valgo.Int32P(req.CaseExecutionId, "case_execution_id").GreaterThan(0))
	}
	return v.ConnectError()
}

func validateListTestExecutionLogsRequest(req *testsv1.ListTestExecutionLogsRequest) error {
	v := newValidator()
	v.Is(
		validator.Context(req.Context),
		validator.TestExecID(req.TestExecutionId),
		validator.PageSize(req.PageSize, maxPageSize),
	)
	return v.ConnectError()
}

func validatePayload(v *valgo.Validation, fieldName string, payload *testsv1.Payload) {
	inputValidator := valgo.Is(
		valgo.String(string(payload.Data), "data").Not().Empty(),
	)

	enc, ok := payload.Metadata["encoding"]
	if !ok || string(enc) != converter.MetadataEncodingJSON {
		inputValidator.AddErrorMessage("metadata", "Metadata encoding must be "+converter.MetadataEncodingJSON)
	}

	v.In(fieldName, inputValidator)
}

func newValidator() *validator.Validator {
	return validator.New(validator.WithBaseErrorMessage(reqValidationBaseErrMsg))
}
