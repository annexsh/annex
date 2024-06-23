package pagination

import (
	"encoding/base64"
	"errors"
	"time"

	paginationv1 "github.com/annexsh/annex-proto/gen/go/annex/common/pagination/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaginatedRequest interface {
	GetNextPageToken() string
}

func DecodeNextPageToken(req PaginatedRequest) (time.Time, uuid.UUID, error) {
	if req == nil {
		return time.Time{}, uuid.UUID{}, errors.New("paginated request cannot be nil")
	}

	encoded := req.GetNextPageToken()
	if encoded == "" {
		return time.Time{}, uuid.UUID{}, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return time.Time{}, uuid.UUID{}, err
	}

	var tkn paginationv1.PaginationToken
	if err = proto.Unmarshal(decoded, &tkn); err != nil {
		return time.Time{}, uuid.UUID{}, err
	}

	lastID, err := uuid.Parse(tkn.LastId)
	if err != nil {
		return time.Time{}, uuid.UUID{}, err
	}

	if tkn.LastTimestamp == nil {
		return time.Time{}, uuid.UUID{}, errors.New("next page token does not contain last timestamp")
	}

	return tkn.LastTimestamp.AsTime(), lastID, nil
}

func EncodeNextPageToken(lastTimestamp time.Time, lastID uuid.UUID) (string, error) {
	tkn := &paginationv1.PaginationToken{
		LastTimestamp: timestamppb.New(lastTimestamp),
		LastId:        lastID.String(),
	}

	msgb, err := proto.Marshal(tkn)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(msgb), nil
}
