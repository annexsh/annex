package validator

import (
	"net"

	"github.com/cohesivestack/valgo"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexsh/annex/uuid"
)

func Context(context string) valgo.Validator {
	return valgo.String(context, "context").Not().Blank()
}

func TestSuiteID(testSuiteID string) valgo.Validator {
	return UUIDv7(testSuiteID, "test_suite_id")
}

func TestID(testID string) valgo.Validator {
	return UUIDv7(testID, "test_id")
}

func TestExecID(testExecID string) valgo.Validator {
	return UUIDv7(testExecID, "test_execution_id")
}

func CaseExecID(caseExecID int32) valgo.Validator {
	return valgo.Int32(caseExecID, "case_execution_id").GreaterThan(0)
}

func UUIDv7(id string, nameAndTitle ...string) valgo.Validator {
	return valgo.String(id, nameAndTitle...).Not().Blank().Passing(func(str string) bool {
		parsed, err := uuid.Parse(str)
		return err == nil && !parsed.Empty()
	}, "{{title}} must be a v7 UUID")
}

func PageSize(pageSize int32, max int32) valgo.Validator {
	return valgo.Int32(pageSize, "page_size").Between(0, max)
}

func HostPort(hostPort string, nameAndTitle ...string) valgo.Validator {
	nt := []string{"host_port", ""}
	if len(nameAndTitle) > 0 {
		nt = nameAndTitle
	}
	return valgo.String(hostPort, nt...).Passing(func(hp string) bool {
		_, _, err := net.SplitHostPort(hp)
		return err == nil
	}, "{{title}} must be a network address of the form 'host:port'")
}

func Timestamppb(ts *timestamppb.Timestamp, nameAndTitle ...string) valgo.Validator {
	return valgo.Any(ts, nameAndTitle...).Not().Nil().Passing(func(val any) bool {
		tspb := val.(*timestamppb.Timestamp)
		if tspb.GetSeconds() == 0 && tspb.GetNanos() == 0 {
			return false
		}
		return tspb.IsValid()
	}, "{{title}} must be a valid timestamp")
}
