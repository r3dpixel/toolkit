package scheduler

import (
	"context"
	"sync"
)

// Options configures a worker pool or an execution
type Options[T any] struct {
	Context     context.Context
	Handler     Handler[T]
	Parallelism int
}

// TaskSource provides tasks to the scheduler
type TaskSource[T any] interface {
	Next(ctx context.Context) (T, bool)
}

// Handler processes tasks from the scheduler
type Handler[T any] func(context.Context, T)

// Pool is a worker pool that accepts tasks at runtime
type Pool[T any] struct {
	ctx     context.Context
	tasks   chan T
	wg      sync.WaitGroup
	handler Handler[T]
}

// NewPool creates a new worker pool with the given options
func NewPool[T any](opts Options[T]) *Pool[T] {
	// Use a background context if none is provided
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Use a single goroutine if no parallelism is specified
	parallelism := opts.Parallelism
	if parallelism <= 0 {
		parallelism = 1
	}

	// Create the pool
	p := &Pool[T]{
		ctx:     ctx,
		tasks:   make(chan T),
		handler: opts.Handler,
	}

	// Spawn workers
	spawnWorkers(ctx, p.tasks, &p.wg, parallelism, opts.Handler)

	// Return the pool
	return p
}

// Submit adds a task to the pool. Returns false if the context is canceled.
func (p *Pool[T]) Submit(task T) bool {
	select {
	case p.tasks <- task:
		return true
	case <-p.ctx.Done():
		return false
	}
}

// Close stops accepting new tasks and waits for all workers to finish
func (p *Pool[T]) Close() {
	close(p.tasks)
	p.wg.Wait()
}

// Exec executes tasks from the source with limited parallelism.
func Exec[T any](source TaskSource[T], opts Options[T]) {
	// If no source or handler is provided, return immediately
	if source == nil || opts.Handler == nil {
		return
	}
	// Use a background context if none is provided
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Use a single goroutine if no parallelism is specified
	parallelism := opts.Parallelism
	if parallelism <= 0 {
		parallelism = 1
	}

	// Create channels for tasks and workers
	tasks := make(chan T)
	var wg sync.WaitGroup

	// Spawn workers
	spawnWorkers(ctx, tasks, &wg, parallelism, opts.Handler)

	// Feed tasks to workers
	feedTasks(ctx, source, tasks)

	// Wait for workers to finish
	wg.Wait()
}

// spawnWorkers spawns a number of workers to process tasks from the given channel
func spawnWorkers[T any](ctx context.Context, tasks <-chan T, wg *sync.WaitGroup, n int, handler Handler[T]) {
	// For each worker, spawn a goroutine to process tasks
	for range n {
		// Add a worker to the wait group
		wg.Add(1)
		// Spawn a goroutine to process tasks
		go func() {
			// Mark the worker as done when it exits
			defer wg.Done()
			// Process tasks
			for task := range tasks {
				handler(ctx, task)
			}
		}()
	}
}

// feedTasks feeds tasks from the given source to the given channel
func feedTasks[T any](ctx context.Context, source TaskSource[T], tasks chan<- T) {
	// Close the channel when the source is exhausted
	defer close(tasks)

	// Loop until the context is canceled or the source is exhausted
	for {
		// Get the next task from the source
		task, ok := source.Next(ctx)
		if !ok {
			return
		}
		// Send the task to the channel
		select {
		case tasks <- task:
		case <-ctx.Done():
			return
		}
	}
}
