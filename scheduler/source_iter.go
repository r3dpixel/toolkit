package scheduler

import (
	"context"
	"iter"
)

// IterSource wraps an iter.Seq as a TaskSource.
type IterSource[T any] struct {
	next func() (T, bool)
	stop func()
}

// FromIter creates a TaskSource from an iterator.
func FromIter[T any](seq iter.Seq[T]) *IterSource[T] {
	next, stop := iter.Pull(seq)
	return &IterSource[T]{next: next, stop: stop}
}

// Next returns the next task from the iterator.
func (s *IterSource[T]) Next(ctx context.Context) (T, bool) {
	return s.next()
}

// Stop releases resources associated with the iterator.
func (s *IterSource[T]) Stop() {
	s.stop()
}
