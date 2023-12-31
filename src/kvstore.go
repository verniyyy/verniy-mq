package src

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

// KVStore ...
type KVStore[Key, Value any] interface {
	Init()
	Size() int64
	Get(Key) (Value, error)
	GetAll() ([]Key, []Value, error)
	GetAndDelete(Key) (Value, error)
	Store(Key, Value) error
	Delete(Key) error
}

// NewKVStore ...
func NewKVStore[Key, Value any]() KVStore[Key, Value] {
	return &kvStore[Key, Value]{}
}

// kvStore ...
type kvStore[K, V any] struct {
	sm sync.Map
}

// Init is erase map
func (s *kvStore[K, V]) Init() {
	s.sm.Range(func(k, _ any) bool {
		s.sm.Delete(k)
		return true
	})
}

// Size is the length of kvStore
func (s *kvStore[K, V]) Size() int64 {
	var i int64
	s.sm.Range(func(k, v any) bool {
		i++
		return true
	})
	return i
}

// Get ...
func (s *kvStore[K, V]) Get(k K) (V, error) {
	v, ok := s.sm.Load(k)
	if !ok {
		var v V
		return v, ErrNotFound
	}

	return v.(V), nil
}

// GetAll ...
func (s *kvStore[K, V]) GetAll() ([]K, []V, error) {
	keys := make([]K, 0, 100)
	values := make([]V, 0, 100)
	s.sm.Range(func(k, v any) bool {
		keys = append(keys, k.(K))
		values = append(values, v.(V))
		return true
	})
	return keys, values, nil
}

// GetAndDelete ...
func (s *kvStore[K, V]) GetAndDelete(k K) (V, error) {
	v, ok := s.sm.LoadAndDelete(k)
	if !ok {
		return *new(V), ErrNotFound
	}

	return v.(V), nil
}

// Store ...
func (s *kvStore[K, V]) Store(k K, v V) error {
	s.sm.Store(k, v)
	return nil
}

// Delete ...
func (s *kvStore[K, V]) Delete(k K) error {
	s.sm.Delete(k)
	return nil
}
