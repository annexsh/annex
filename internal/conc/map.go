package conc

import (
	"sync"
)

type Map[T any] struct {
	mu    sync.RWMutex
	items map[any]T
}

func NewMap[T any]() *Map[T] {
	return &Map[T]{
		mu:    sync.RWMutex{},
		items: map[any]T{},
	}
}

func (m *Map[T]) Load(key any) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.items[key]
	return val, ok
}

func (m *Map[T]) LoadDefault(key any, def T) (T, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if val, ok := m.items[key]; ok {
		return val, ok
	}
	m.items[key] = def
	return def, false
}

func (m *Map[T]) Set(key any, val T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = val
}

func (m *Map[T]) Delete(key any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *Map[T]) Range(f func(key any, value T) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.items {
		ok := f(k, v)
		if !ok {
			return
		}
	}
}

func (m *Map[T]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}
