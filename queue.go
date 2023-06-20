package quill

import "sync"

type queue[T any] struct {
	lock     sync.RWMutex
	elements []T
}

func newQueue[T any]() *queue[T] {
	return &queue[T]{
		elements: make([]T, 0),
	}
}

func (q *queue[T]) Push(x ...T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.elements = append(q.elements, x...)
}

func (q *queue[T]) Top() T {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.elements[0]
}

func (q *queue[T]) Pop() T {
	q.lock.Lock()
	defer q.lock.Unlock()
	top := q.elements[0]
	var x T
	q.elements[0] = x
	q.elements = q.elements[1:]
	return top
}

func (q *queue[T]) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.elements)
}
