package reqx

import (
	"fmt"
	"sync"

	"github.com/imroc/req/v3"
)

const (
	defaultMaxRetries = 1
)

var (
	ErrInterceptorNotFound = fmt.Errorf("interceptor not found")
	ErrMaxRetriesExceeded  = fmt.Errorf("max interceptor retries exceeded")
)

// Interceptor handles response interception and recovery.
//
// Thread safety: Implementations do NOT need to handle synchronization.
// The interceptor system guarantees that Recover and Apply are never called concurrently.
//
// Thundering herd prevention: When multiple concurrent requests trigger recovery,
// only one Recover call executes. Other goroutines waiting will skip recovery
// and use the newly recovered state.
type Interceptor interface {
	// ShouldIntercept returns true if this response/error should trigger recovery.
	// Both resp and err are provided so status codes can be checked (e.g. 429, 401)
	// even when they're wrapped as errors by the retry logic.
	ShouldIntercept(resp *req.Response, err error) bool

	// Recover performs the recovery action (refresh cookies, re-login, etc.)
	// This is called when ShouldIntercept returns true.
	// The implementation should store any state it needs for Apply.
	Recover(client *Client, resp *req.Response) error

	// Apply applies the current state to the request (set cookies, headers, etc.)
	// This is called before every request attempt to apply stored state.
	Apply(r *req.Request) *req.Request

	// MaxRetries returns the maximum number of recovery attempts per request.
	// Return 0 or negative to use the default (1).
	MaxRetries() int
}

// interceptorStore wraps an Interceptor with thread-safe access
type interceptorStore struct {
	interceptor Interceptor
	mu          sync.RWMutex
	generation  uint64 // Incremented on each successful recovery (thundering herd prevention)
}

// newInterceptorStore creates a new thread-safe interceptor store
func newInterceptorStore(interceptor Interceptor) *interceptorStore {
	return &interceptorStore{
		interceptor: interceptor,
	}
}

// shouldIntercept checks if the response should trigger recovery (read lock)
func (s *interceptorStore) shouldIntercept(resp *req.Response, err error) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.interceptor.ShouldIntercept(resp, err)
}

// getGeneration returns the current recovery generation
func (s *interceptorStore) getGeneration() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.generation
}

// recover performs recovery with exclusive access (thundering herd prevention via generation counter)
// genBefore is the generation captured before waiting on the lock.
// If generation changed while waiting, recovery is skipped (someone else already recovered).
func (s *interceptorStore) recover(client *Client, resp *req.Response, genBefore uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If generation changed while waiting, someone else already recovered
	if s.generation != genBefore {
		return nil
	}

	err := s.interceptor.Recover(client, resp)
	if err == nil {
		s.generation++
	}
	return err
}

// apply applies state to the request (read lock)
func (s *interceptorStore) apply(r *req.Request) *req.Request {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.interceptor.Apply(r)
}

// maxRetries returns the max retries, defaulting if needed
func (s *interceptorStore) maxRetries() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	max := s.interceptor.MaxRetries()
	if max <= 0 {
		return defaultMaxRetries
	}
	return max
}
