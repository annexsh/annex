package uuid

import (
	"fmt"

	"github.com/google/uuid"
)

const version7 = "VERSION_7"

type V7 struct {
	uuid.UUID
}

func (u V7) Empty() bool {
	return u.UUID == uuid.Nil
}

func (u V7) Before(id V7) bool {
	return u.String() < id.String()
}

func (u V7) After(id V7) bool {
	return u.String() > id.String()
}

func New() V7 {
	return V7{uuid.Must(uuid.NewV7())}
}

func NewString() string {
	return New().String()
}

func Parse(s string) (V7, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return V7{}, err
	}
	if u.Version().String() != version7 {
		return V7{}, fmt.Errorf("uuid is not %s: found %s", version7, u.Version().String())
	}
	return V7{u}, err
}
