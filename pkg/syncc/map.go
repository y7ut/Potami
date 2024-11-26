package syncc

import "sync"

type Map[K comparable, V any] struct {
	sm sync.Map
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	any, loaded := m.sm.LoadOrStore(key, value)
	return any.(V), loaded
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	any, loaded := m.sm.LoadAndDelete(key)
	return any.(V), loaded
}

func (m *Map[K, V]) Delete(key K) {
	m.sm.Delete(key)
}

func (m *Map[K, V]) Swap(key K, value V) (actual V, loaded bool) {
	any, loaded := m.sm.Swap(key, value)
	return any.(V), loaded
}

func (m *Map[K, V]) CompareAndSwap(key, old, new any) bool {
	return m.sm.CompareAndSwap(key, old, new)
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.sm.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (m *Map[K, V]) CompareAndDelete(key, old any) bool {
	return m.sm.CompareAndDelete(key, old)
}

func (m *Map[K, V]) Store(key K, value V) {
	m.sm.Store(key, value)
}

func (m *Map[K, V]) Load(key K) (value V, loaded bool) {
	any, loaded := m.sm.Load(key)
	if any == nil || !loaded {
		return *new(V), false
	}
	return any.(V), loaded
}
