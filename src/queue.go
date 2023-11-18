package src

import (
	"container/list"
	"sync"
)

// Queue ...
type Queue[T any] interface {
	Init()
	Size() int64
	Enqueue(T) error
	Dequeue() (T, error)
}

// NewQueue ...
func NewQueue[T any]() Queue[T] {
	return &queue[T]{
		l: list.New().Init(),
	}
}

// queue ...
type queue[T any] struct {
	m sync.Mutex
	l *list.List
}

// Init ...
func (q *queue[T]) Init() {
	q.m.Lock()
	q.l.Init()
	q.m.Unlock()
}

// Size ...
func (q *queue[T]) Size() int64 {
	q.m.Lock()
	size := int64(q.l.Len())
	q.m.Unlock()
	return size
}

// Enqueue ...
func (q *queue[T]) Enqueue(v T) error {
	q.m.Lock()
	_ = q.l.PushBack(v)
	q.m.Unlock()
	return nil
}

// Dequeue ...
func (q *queue[T]) Dequeue() (T, error) {
	q.m.Lock()
	e := q.l.Front()
	q.m.Unlock()
	return e.Value.(T), nil
}
