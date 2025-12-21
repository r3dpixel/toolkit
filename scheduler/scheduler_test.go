package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	var count atomic.Int32

	Exec(FromSlice([]int{1, 2, 3, 4, 5}), Options[int]{
		Handler:     func(ctx context.Context, n int) { count.Add(1) },
		Parallelism: 2,
	})

	if count.Load() != 5 {
		t.Errorf("expected 5, got %d", count.Load())
	}
}

func TestExecWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var count atomic.Int32

	Exec(FromSlice([]int{1, 2, 3, 4, 5}), Options[int]{
		Context:     ctx,
		Handler:     func(ctx context.Context, n int) { count.Add(1) },
		Parallelism: 2,
	})

	// With pre-cancelled context, no tasks should be fed
	if count.Load() != 0 {
		t.Errorf("expected 0, got %d", count.Load())
	}
}

func TestExecNilSource(t *testing.T) {
	Exec[int](nil, Options[int]{
		Handler:     func(ctx context.Context, n int) {},
		Parallelism: 2,
	})
}

func TestExecNilHandler(t *testing.T) {
	Exec(FromSlice([]int{1, 2, 3}), Options[int]{
		Parallelism: 2,
	})
}

func TestExecParallelismLimit(t *testing.T) {
	const parallelism = 4
	const totalTasks = 100

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	Exec(FromSlice(make([]int, totalTasks)), Options[int]{
		Parallelism: parallelism,
		Handler: func(ctx context.Context, n int) {
			cur := concurrent.Add(1)
			for {
				max := maxConcurrent.Load()
				if cur <= max || maxConcurrent.CompareAndSwap(max, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			concurrent.Add(-1)
		},
	})

	if maxConcurrent.Load() > parallelism {
		t.Errorf("max concurrent %d exceeded parallelism %d", maxConcurrent.Load(), parallelism)
	}
	if maxConcurrent.Load() < parallelism {
		t.Errorf("expected to reach parallelism %d, only got %d", parallelism, maxConcurrent.Load())
	}
}

func TestPool(t *testing.T) {
	var count atomic.Int32

	pool := NewPool(Options[int]{
		Handler:     func(ctx context.Context, n int) { count.Add(1) },
		Parallelism: 2,
	})

	for i := range 5 {
		pool.Submit(i)
	}
	pool.Close()

	if count.Load() != 5 {
		t.Errorf("expected 5, got %d", count.Load())
	}
}

func TestPoolWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	proceed := make(chan struct{})

	pool := NewPool(Options[int]{
		Context: ctx,
		Handler: func(ctx context.Context, n int) {
			started <- struct{}{} // Signal handler started
			<-proceed             // Block until released
		},
		Parallelism: 1, // Single worker so it's busy
	})

	pool.Submit(1)
	<-started // Wait for handler to start processing
	cancel()  // Cancel while worker is busy

	// Worker is busy and context is cancelled, Submit should return false
	ok := pool.Submit(2)
	if ok {
		t.Error("expected Submit to return false after context cancellation")
	}

	close(proceed) // Release the handler
	pool.Close()
}

func TestPoolParallelismLimit(t *testing.T) {
	const parallelism = 4
	const totalTasks = 100

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	pool := NewPool(Options[int]{
		Parallelism: parallelism,
		Handler: func(ctx context.Context, n int) {
			cur := concurrent.Add(1)
			for {
				max := maxConcurrent.Load()
				if cur <= max || maxConcurrent.CompareAndSwap(max, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			concurrent.Add(-1)
		},
	})

	for i := range totalTasks {
		pool.Submit(i)
	}
	pool.Close()

	if maxConcurrent.Load() > parallelism {
		t.Errorf("max concurrent %d exceeded parallelism %d", maxConcurrent.Load(), parallelism)
	}
	if maxConcurrent.Load() < parallelism {
		t.Errorf("expected to reach parallelism %d, only got %d", parallelism, maxConcurrent.Load())
	}
}

func TestPoolDefaultParallelism(t *testing.T) {
	var count atomic.Int32

	pool := NewPool(Options[int]{
		Handler: func(ctx context.Context, n int) { count.Add(1) },
	})

	pool.Submit(1)
	pool.Submit(2)
	pool.Close()

	if count.Load() != 2 {
		t.Errorf("expected 2, got %d", count.Load())
	}
}
