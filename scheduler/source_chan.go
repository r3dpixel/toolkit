package scheduler

import "context"

// ChanSource wraps a channel as a TaskSource.
type ChanSource[T any] struct {
	ch <-chan T
}

// FromChan creates a TaskSource from a channel.
func FromChan[T any](ch <-chan T) *ChanSource[T] {
	return &ChanSource[T]{ch: ch}
}

// Next returns the next task from the channel, or (zero, false) if the channel is closed.
func (s *ChanSource[T]) Next(ctx context.Context) (T, bool) {
	select {
	case task, ok := <-s.ch:
		return task, ok
	case <-ctx.Done():
		var zero T
		return zero, false
	}
}
