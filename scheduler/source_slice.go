package scheduler

import (
	"context"
	"slices"
	"sync"
)

// SliceSource iterates over a slice as a TaskSource.
type SliceSource[T any] struct {
	items []T
	idx   int
	mu    sync.Mutex
}

// FromSlice creates a TaskSource from a slice.
func FromSlice[T any](items []T) *SliceSource[T] {
	return &SliceSource[T]{items: items}
}

// FromSliceClone creates a TaskSource from a slice, cloning the slice.
func FromSliceClone[T any](items []T) *SliceSource[T] {
	return &SliceSource[T]{items: slices.Clone(items)}
}

// Next returns the next task from the slice, or (zero, false) if the slice is exhausted.
func (s *SliceSource[T]) Next(ctx context.Context) (T, bool) {
	// Lock the slice
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we've reached the end of the slice'
	if s.idx >= len(s.items) {
		var zero T
		return zero, false
	}

	// Extract the next task
	task := s.items[s.idx]
	// Increment the index
	s.idx++
	// Return the task
	return task, true
}
