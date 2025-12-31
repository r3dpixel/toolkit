package reqx

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
)

// mockInterceptor is a test interceptor that tracks calls
type mockInterceptor struct {
	shouldInterceptFunc func(resp *req.Response, err error) bool
	recoverFunc         func(client *Client, resp *req.Response) error
	applyFunc           func(r *req.Request) *req.Request
	maxRetries          int

	recoverCount   atomic.Int32
	applyCount     atomic.Int32
	interceptCount atomic.Int32
}

func (m *mockInterceptor) ShouldIntercept(resp *req.Response, err error) bool {
	m.interceptCount.Add(1)
	if m.shouldInterceptFunc != nil {
		return m.shouldInterceptFunc(resp, err)
	}
	return resp != nil && resp.StatusCode == http.StatusForbidden
}

func (m *mockInterceptor) Recover(client *Client, resp *req.Response) error {
	m.recoverCount.Add(1)
	if m.recoverFunc != nil {
		return m.recoverFunc(client, resp)
	}
	return nil
}

func (m *mockInterceptor) Apply(r *req.Request) *req.Request {
	m.applyCount.Add(1)
	if m.applyFunc != nil {
		return m.applyFunc(r)
	}
	return r
}

func (m *mockInterceptor) MaxRetries() int {
	if m.maxRetries > 0 {
		return m.maxRetries
	}
	return 1
}

func TestInterceptor_BasicRecovery(t *testing.T) {
	// Server returns 403 once, then 200
	requestCount := atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	interceptor := &mockInterceptor{}

	client := NewClient(Options{RetryCount: 0})
	client.RegisterInterceptor("test", interceptor)

	resp, err := client.IR("test").Get(server.URL)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), interceptor.recoverCount.Load(), "should recover once")
	assert.Equal(t, int32(2), interceptor.applyCount.Load(), "should apply twice (initial + retry)")
}

func TestInterceptor_NoRecoveryNeeded(t *testing.T) {
	// Server always returns 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	interceptor := &mockInterceptor{}

	client := NewClient(Options{RetryCount: 0})
	client.RegisterInterceptor("test", interceptor)

	resp, err := client.IR("test").Get(server.URL)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(0), interceptor.recoverCount.Load(), "should not recover")
	assert.Equal(t, int32(1), interceptor.applyCount.Load(), "should apply once")
}

func TestInterceptor_MaxRetriesExceeded(t *testing.T) {
	// Server always returns 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	interceptor := &mockInterceptor{maxRetries: 2}

	client := NewClient(Options{RetryCount: 0})
	client.RegisterInterceptor("test", interceptor)

	resp, err := client.IR("test").Get(server.URL)

	// Should fail after max retries
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, int32(2), interceptor.recoverCount.Load(), "should recover max times")
}

func TestInterceptor_RecoveryFails(t *testing.T) {
	// Server returns 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	interceptor := &mockInterceptor{
		recoverFunc: func(client *Client, resp *req.Response) error {
			return assert.AnError
		},
	}

	client := NewClient(Options{RetryCount: 0})
	client.RegisterInterceptor("test", interceptor)

	resp, err := client.IR("test").Get(server.URL)

	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, int32(1), interceptor.recoverCount.Load(), "should attempt recovery once")
}

func TestInterceptor_ThunderingHerd_OnlyOneRecovers(t *testing.T) {
	// Server returns 403 for first N requests, then 200
	var requestCount atomic.Int32
	var allowSuccess atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		if allowSuccess.Load() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer server.Close()

	// Use a barrier to ensure all goroutines enter retry hook around the same time
	var readyCount atomic.Int32
	const numRequests = 5

	interceptor := &mockInterceptor{
		recoverFunc: func(client *Client, resp *req.Response) error {
			// Simulate slow recovery so others pile up waiting
			time.Sleep(100 * time.Millisecond)
			allowSuccess.Store(true)
			return nil
		},
	}

	client := NewClient(Options{RetryCount: 0})
	client.RegisterInterceptor("test", interceptor)

	// Prepare all requests first
	requests := make([]*req.Request, numRequests)
	for i := 0; i < numRequests; i++ {
		requests[i] = client.IR("test")
	}

	// Launch all at once
	var wg sync.WaitGroup
	startCh := make(chan struct{})
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(r *req.Request) {
			defer wg.Done()
			readyCount.Add(1)
			<-startCh // Wait for signal
			_, err := r.Get(server.URL)
			results <- err
		}(requests[i])
	}

	// Wait for all goroutines to be ready, then start
	for readyCount.Load() < numRequests {
		time.Sleep(1 * time.Millisecond)
	}
	close(startCh)

	wg.Wait()
	close(results)

	// All should succeed
	for err := range results {
		assert.NoError(t, err)
	}

	// Only ONE recovery should have happened due to thundering herd prevention
	assert.Equal(t, int32(1), interceptor.recoverCount.Load(), "thundering herd: only one should recover")
}

func TestInterceptor_ThunderingHerd_SecondRecoveryIfFirstBad(t *testing.T) {
	// Scenario: A, B, C all get 403
	// A recovers but produces bad state
	// A, B, C retry - A and B succeed, C gets 403 again
	// C should be able to recover again

	var requestCount atomic.Int32
	var recoveryCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		// First 3 requests fail (initial A, B, C)
		// 4th and 5th succeed (A and B retry)
		// 6th fails (C retry with bad luck)
		// 7th succeeds (C second retry)
		if count <= 3 || count == 6 {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	interceptor := &mockInterceptor{
		maxRetries: 3,
		recoverFunc: func(client *Client, resp *req.Response) error {
			recoveryCount.Add(1)
			time.Sleep(10 * time.Millisecond) // Small delay to simulate work
			return nil
		},
	}

	client := NewClient(Options{RetryCount: 0})
	client.RegisterInterceptor("test", interceptor)

	// Launch 3 concurrent requests
	var wg sync.WaitGroup
	results := make(chan error, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.IR("test").Get(server.URL)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	// All should eventually succeed
	for err := range results {
		assert.NoError(t, err)
	}

	// At least one recovery, possibly two if C needed another
	assert.GreaterOrEqual(t, recoveryCount.Load(), int32(1), "at least one recovery")
}

func TestInterceptor_ApplyStateOnRetry(t *testing.T) {
	// Verify that Apply is called before each retry with updated state
	var cookie string
	var mu sync.RWMutex

	requestCount := atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		gotCookie := r.Header.Get("X-Session")

		if count == 1 {
			// First request has no cookie
			assert.Empty(t, gotCookie)
			w.WriteHeader(http.StatusForbidden)
		} else {
			// Retry should have cookie
			assert.Equal(t, "recovered-session", gotCookie)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	interceptor := &mockInterceptor{
		recoverFunc: func(client *Client, resp *req.Response) error {
			mu.Lock()
			cookie = "recovered-session"
			mu.Unlock()
			return nil
		},
		applyFunc: func(r *req.Request) *req.Request {
			mu.RLock()
			defer mu.RUnlock()
			if cookie != "" {
				r.SetHeader("X-Session", cookie)
			}
			return r
		},
	}

	client := NewClient(Options{RetryCount: 0})
	client.RegisterInterceptor("test", interceptor)

	resp, err := client.IR("test").Get(server.URL)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInterceptor_NonExistentLabel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Options{})

	_, err := client.IR("non-existent").Get(server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interceptor non-existent does not exist")
}

func TestInterceptor_RegisterUnregister(t *testing.T) {
	client := NewClient(Options{})
	interceptor := &mockInterceptor{}

	// Register
	client.RegisterInterceptor("test", interceptor)
	assert.Contains(t, client.interceptors, "test")

	// Unregister
	client.UnregisterInterceptor("test")
	assert.NotContains(t, client.interceptors, "test")
}

func TestInterceptor_MultipleInterceptors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	interceptor1 := &mockInterceptor{}
	interceptor2 := &mockInterceptor{}

	client := NewClient(Options{})
	client.RegisterInterceptor("service1", interceptor1)
	client.RegisterInterceptor("service2", interceptor2)

	// Use service1
	_, err := client.IR("service1").Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), interceptor1.applyCount.Load())
	assert.Equal(t, int32(0), interceptor2.applyCount.Load())

	// Use service2
	_, err = client.IR("service2").Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), interceptor1.applyCount.Load())
	assert.Equal(t, int32(1), interceptor2.applyCount.Load())
}

func TestInterceptor_Check401And429(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"401 Unauthorized", http.StatusUnauthorized},
		{"403 Forbidden", http.StatusForbidden},
		{"429 Too Many Requests", http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := atomic.Int32{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := requestCount.Add(1)
				if count == 1 {
					w.WriteHeader(tt.statusCode)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer server.Close()

			interceptor := &mockInterceptor{
				shouldInterceptFunc: func(resp *req.Response, err error) bool {
					if resp == nil {
						return false
					}
					return resp.StatusCode == tt.statusCode
				},
			}

			client := NewClient(Options{RetryCount: 0})
			client.RegisterInterceptor("test", interceptor)

			resp, err := client.IR("test").Get(server.URL)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, int32(1), interceptor.recoverCount.Load())
		})
	}
}

func TestInterceptorStore_Generation(t *testing.T) {
	interceptor := &mockInterceptor{}
	store := newInterceptorStore(interceptor)

	// Initial generation is 0
	assert.Equal(t, uint64(0), store.getGeneration())

	// After recovery, generation increments
	err := store.recover(nil, nil, 0)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), store.getGeneration())

	// Recovery with wrong generation skips
	err = store.recover(nil, nil, 0) // genBefore=0 but current is 1
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), store.getGeneration()) // unchanged
	assert.Equal(t, int32(1), interceptor.recoverCount.Load())

	// Recovery with correct generation works
	err = store.recover(nil, nil, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), store.getGeneration())
	assert.Equal(t, int32(2), interceptor.recoverCount.Load())
}

func TestInterceptorStore_GenerationSkipsOnRecoveryError(t *testing.T) {
	interceptor := &mockInterceptor{
		recoverFunc: func(client *Client, resp *req.Response) error {
			return assert.AnError
		},
	}
	store := newInterceptorStore(interceptor)

	err := store.recover(nil, nil, 0)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), store.getGeneration()) // not incremented on error
}
