package src

import (
	"fmt"
	"sync"
)

// KVStore ...
type KVStore[Value any] interface {
	Init()
	Size() int64
	GetAndDelete(key string) (Value, error)
	Store(key string, value Value) error
	Delete(key string) error
}

// NewKVStore ...
func NewKVStore[Value any]() KVStore[Value] {
	return &kvStore[Value]{}
}

// kvStore ...
type kvStore[T any] struct {
	sm sync.Map
}

// Init is erase map
func (s *kvStore[T]) Init() {
	s.sm.Range(func(k, _ any) bool {
		s.sm.Delete(k)
		return true
	})
}

// Size is the length of kvStore
func (s *kvStore[T]) Size() int64 {
	var i int64
	s.sm.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}

// Get ...
func (s *kvStore[T]) GetAndDelete(k string) (T, error) {
	v, ok := s.sm.LoadAndDelete(k)
	return v.(T), fmt.Errorf("%v", ok)
}

// Store ...
func (s *kvStore[T]) Store(k string, v T) error {
	s.sm.Store(k, v)
	return nil
}

// Delete ...
func (s *kvStore[T]) Delete(k string) error {
	s.sm.Delete(k)
	return nil
}
