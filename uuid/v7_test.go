package uuid

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	id := New()
	assert.Equal(t, version7, id.Version().String())
}

func TestNewString(t *testing.T) {
	str := NewString()
	id, err := uuid.Parse(str)
	require.NoError(t, err)

	assert.Equal(t, version7, id.Version().String())
}

func TestParse(t *testing.T) {
	v7str := uuid.Must(uuid.NewV7()).String()
	v4str := uuid.NewString()

	v7id, err := Parse(v7str)
	require.NoError(t, err)
	assert.False(t, v7id.Empty())

	_, err = Parse(v4str)
	assert.EqualError(t, err, "uuid is not VERSION_7: found VERSION_4")
}

func TestV7_BeforeAfter(t *testing.T) {
	a := New()
	b := New()
	assert.True(t, a.Before(b))
	assert.False(t, a.After(b))
	assert.False(t, b.Before(a))
	assert.True(t, b.After(a))
}

func TestV7_Empty(t *testing.T) {
	var empty V7
	assert.True(t, empty.Empty())
}
