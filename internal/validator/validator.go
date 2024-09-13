package validator

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/cohesivestack/valgo"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type Option func(v *Validator)

func WithBaseErrorMessage(msg string) Option {
	return func(v *Validator) {
		v.baseErrMsg = msg
	}
}

type Validator struct {
	*valgo.Validation
	baseErrMsg string
}

func New(opts ...Option) *Validator {
	v := &Validator{
		Validation: valgo.New(),
		baseErrMsg: "validation failed",
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v *Validator) Error() error {
	if v.Validation.Valid() {
		return nil
	}

	var joinedErr error

	for _, err := range v.Validation.Errors() {
		for _, msg := range err.Messages() {
			fieldErr := fmt.Errorf("'%s': %s", err.Name(), msg)
			if joinedErr == nil {
				joinedErr = fieldErr
			} else {
				joinedErr = fmt.Errorf("%w; %w", joinedErr, fieldErr)
			}
		}
	}

	return fmt.Errorf("%s: %w", v.baseErrMsg, joinedErr)
}

func (v *Validator) ConnectError() error {
	if v.Validation.Valid() {
		return nil
	}

	cErr := connect.NewError(connect.CodeInvalidArgument, errors.New(v.baseErrMsg))

	badReq := &errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{},
	}

	for _, err := range v.Validation.Errors() {
		desc := ""

		for i, msg := range err.Messages() {
			if i == 0 {
				desc = msg
			} else {
				desc += ", " + msg
			}
		}

		badReq.FieldViolations = append(badReq.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       err.Name(),
			Description: desc,
		})
	}

	detail, err := connect.NewErrorDetail(badReq)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, v.Error())
	}

	cErr.AddDetail(detail)

	return cErr
}
