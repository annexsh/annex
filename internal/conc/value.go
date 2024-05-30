package conc

import (
	"sync"
)

type Value[T any] struct {
	mu  *sync.RWMutex
	val T
}

func NewValue[T any](val T) *Value[T] {
	return &Value[T]{
		mu:  new(sync.RWMutex),
		val: val,
	}
}

func (v *Value[T]) Set(val T) {
	v.mu.Lock()
	v.val = val
	v.mu.Unlock()
}

func (v *Value[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.val
}
