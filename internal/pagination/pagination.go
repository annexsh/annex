package pagination

import (
	"encoding/base64"
	"errors"
	"strconv"

	paginationv1 "github.com/annexsh/annex-proto/go/gen/annex/common/pagination/v1"
	"google.golang.org/protobuf/proto"

	"github.com/annexsh/annex/test"
	"github.com/annexsh/annex/uuid"
)

const defaultPageSize = 100

type NextPageToken[T test.Identifier] struct {
	OffsetID T
}

type OffsetOption[T test.Identifier] func(id string) (T, error)

func WithString() OffsetOption[string] {
	return func(id string) (string, error) {
		return id, nil
	}
}

func WithTestExecutionID() OffsetOption[test.TestExecutionID] {
	return test.ParseTestExecutionID
}

func WithCaseExecutionID() OffsetOption[test.CaseExecutionID] {
	return func(id string) (test.CaseExecutionID, error) {
		idInt, err := strconv.Atoi(id)
		return test.CaseExecutionID(idInt), err
	}
}

func WithUUID() OffsetOption[uuid.V7] {
	return uuid.Parse
}

type Request interface {
	GetPageSize() int32
	GetNextPageToken() string
}

func FilterFromRequest[T test.Identifier](req Request, offsetOpt OffsetOption[T]) (test.PageFilter[T], error) {
	filter := test.PageFilter[T]{
		Size: defaultPageSize,
	}

	if req.GetPageSize() > 0 {
		filter.Size = int(req.GetPageSize())
	}

	if req.GetNextPageToken() != "" {
		decodedTkn, err := decodeNextPageToken[T](req.GetNextPageToken(), offsetOpt)
		if err != nil {
			return test.PageFilter[T]{}, err
		}
		filter.OffsetID = &decodedTkn.OffsetID
	}

	return filter, nil
}

type IDGetterFunc[T any, I test.Identifier] func(item T) I

func NextPageTokenFromItems[T any, I test.Identifier](pageSize int, items []T, idGetter IDGetterFunc[T, I]) (string, error) {
	if len(items) == pageSize {
		offsetID := idGetter(items[len(items)-1])
		return encodeNextPageToken(NextPageToken[I]{
			OffsetID: offsetID,
		})
	}
	return "", nil
}

func encodeNextPageToken[T test.Identifier](tkn NextPageToken[T]) (string, error) {
	var offsetID string

	switch id := any(tkn.OffsetID).(type) {
	case string:
		offsetID = id
	case uuid.V7:
		offsetID = id.String()
	case test.TestExecutionID:
		offsetID = id.String()
	case test.CaseExecutionID:
		offsetID = id.String()
	default:
		panic("unsupported offset id type")
	}

	tknpb := &paginationv1.PaginationToken{
		OffsetId: offsetID,
	}

	msgb, err := proto.Marshal(tknpb)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(msgb), nil
}

func decodeNextPageToken[T test.Identifier](tkn string, offsetOpt OffsetOption[T]) (NextPageToken[T], error) {
	if tkn == "" {
		return NextPageToken[T]{}, errors.New("failed to decode next page token: token is empty")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(tkn)
	if err != nil {
		return NextPageToken[T]{}, err
	}

	var tknpb paginationv1.PaginationToken
	if err = proto.Unmarshal(decoded, &tknpb); err != nil {
		return NextPageToken[T]{}, err
	}

	var out NextPageToken[T]

	out.OffsetID, err = offsetOpt(tknpb.OffsetId)
	if err != nil {
		return NextPageToken[T]{}, err
	}

	return out, nil
}
