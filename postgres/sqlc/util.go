package sqlc

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/annexhq/annex/internal/ptr"
)

type Timestamp struct {
	pgtype.Timestamp
}

func NewTimestamp(time time.Time) Timestamp {
	return Timestamp{
		Timestamp: pgtype.Timestamp{
			Time:  time,
			Valid: true,
		},
	}
}

func NewNullableTimestamp(time *time.Time) Timestamp {
	ts := pgtype.Timestamp{
		Valid: false,
	}
	if time != nil {
		ts.Valid = true
		ts.Time = *time
	}
	return Timestamp{
		Timestamp: ts,
	}
}

func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	var s *string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	if s == nil {
		*ts = Timestamp{}
		return nil
	}

	format := "2006-01-02T15:04:05"
	tim, err := time.Parse(format, *s)
	if err != nil {
		return err
	}
	*ts = Timestamp{
		Timestamp: pgtype.Timestamp{Time: tim, Valid: true},
	}

	return nil
}

type Time struct {
	time.Time
}

func NewTime(time time.Time) Time {
	return Time{
		time,
	}
}

func NewNullableTime(time *time.Time) *Time {
	if time == nil {
		return nil
	}
	return ptr.Get(NewTime(*time))
}

func (t *Time) UnmarshalJSON(b []byte) error {
	str := string(bytes.Trim(b, `"`))
	if str == "null" || str == `""` {
		return nil
	}
	format := "2006-01-02T15:04:05"
	parsed, err := time.Parse(format, str)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func (t *Time) Value() time.Time {
	if t != nil {
		return t.Time
	}
	return time.Time{}
}

func (t *Time) Pointer() *time.Time {
	if t != nil {
		return &t.Time
	}
	return nil
}

func (t *Time) Proto() *timestamppb.Timestamp {
	if t != nil {
		return timestamppb.New(t.Time)
	}
	return &timestamppb.Timestamp{}
}
