package scheduler

import "context"

// FuncSource generates tasks from a function.
type FuncSource[T any] struct {
	fn func(ctx context.Context) (T, bool)
}

// FromFunc creates a TaskSource from a generator function.
// The function should return (task, true) for each task, or (zero, false) when done.
func FromFunc[T any](fn func(ctx context.Context) (T, bool)) *FuncSource[T] {
	return &FuncSource[T]{fn: fn}
}

// Next calls the generator function.
func (s *FuncSource[T]) Next(ctx context.Context) (T, bool) {
	return s.fn(ctx)
}
