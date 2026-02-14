package reqx

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/toolkit/cred"
)

const (
	defaultRetryCount        int = 4
	defaultMinBackoff            = 10 * time.Millisecond
	defaultMaxBackoff            = 500 * time.Millisecond
	defaultTimeout               = 10 * time.Second
	defaultAuthRefreshBuffer     = 2 * time.Minute
)

const (
	// JsonApplicationContentType - JSON content type header
	JsonApplicationContentType string = "application/json"
	// FormUrlEncodedApplicationContentType - x-www-form-urlencoded content type header
	FormUrlEncodedApplicationContentType string = "application/x-www-form-urlencoded"
)

var (
	ErrResponseNil     = errors.New("response is nil")
	ErrResponseBodyNil = errors.New("response body is nil")
)

type Impersonation byte

const (
	None Impersonation = iota
	Chrome
	Firefox
	Safari
)

// Options settings to configure the retryable HTTP client
type Options struct {
	RetryCount        int
	MinBackoff        time.Duration
	MaxBackoff        time.Duration
	Timeout           time.Duration
	AuthRefreshBuffer time.Duration
	EnableHttp3       bool
	AutoDecompress    bool
	AutoDecode        bool
	DisableKeepAlives bool
	Impersonation     Impersonation
}

// Config is a function that configures the underlying req.Client (for advanced use cases)
type Config func(*req.Client)

// RefreshTokenFunc is a function that refreshes the token for the given identity
type RefreshTokenFunc func(client *Client, identity cred.Identity) (string, error)

// Client wraps req.Client and provides both authenticated and normal requests
type Client struct {
	client            *req.Client
	authRefreshBuffer time.Duration
	auths             map[string]authStore
	authsMu           sync.RWMutex
	interceptors      map[string]*interceptorStore
	interceptorsMu    sync.RWMutex
}

// NewClient creates a new wrapped client
func NewClient(opts Options, configs ...Config) *Client {
	// Set the default auth refresh buffer if not set
	if opts.AuthRefreshBuffer <= 0 {
		opts.AuthRefreshBuffer = defaultAuthRefreshBuffer
	}

	// Create the retryable client
	client := newRetryClient(opts)

	// InsertIter response error checking callback
	client.OnAfterResponse(func(c *req.Client, resp *req.Response) error {
		return responseErrorCause(resp, resp.Err)
	})

	// Apply any config functions
	for _, applyConfig := range configs {
		applyConfig(client)
	}

	// Return the client
	return &Client{
		client:            client,
		authRefreshBuffer: opts.AuthRefreshBuffer,
		auths:             make(map[string]authStore),
		interceptors:      make(map[string]*interceptorStore),
	}
}

// RegisterAuth registers an authentication provider using a refreshable token with the given label
func (c *Client) RegisterAuth(serviceLabel string, identityReader cred.IdentityReader, refreshFunc func(*Client, cred.Identity) (string, error)) *Client {
	// Lock the auths map
	c.authsMu.Lock()
	defer c.authsMu.Unlock()

	// Create the auth manager
	c.auths[serviceLabel] = newRefreshableAuthStore(c, identityReader, refreshFunc, c.authRefreshBuffer)

	// Return the client
	return c
}

// RegisterToken registers an authentication provider using a fixed token with the given label
func (c *Client) RegisterToken(serviceLabel string, token string) *Client {
	// Lock the auths map
	c.authsMu.Lock()
	defer c.authsMu.Unlock()

	// Create the auth manager
	c.auths[serviceLabel] = newTokenAuthStore(token)

	// Return the client
	return c
}

// UnregisterAuth unregisters an authentication provider with the given label
func (c *Client) UnregisterAuth(serviceLabel string) {
	// Lock the auths map
	c.authsMu.Lock()
	defer c.authsMu.Unlock()

	// Delete the auth manager
	delete(c.auths, serviceLabel)
}

// RegisterInterceptor registers an interceptor with the given label
func (c *Client) RegisterInterceptor(label string, interceptor Interceptor) *Client {
	c.interceptorsMu.Lock()
	defer c.interceptorsMu.Unlock()

	c.interceptors[label] = newInterceptorStore(interceptor)
	return c
}

// UnregisterInterceptor unregisters an interceptor with the given label
func (c *Client) UnregisterInterceptor(label string) {
	c.interceptorsMu.Lock()
	defer c.interceptorsMu.Unlock()

	delete(c.interceptors, label)
}

// R creates a normal request builder
func (c *Client) R() *req.Request {
	return c.client.R()
}

// IR creates an intercepted request with automatic recovery
func (c *Client) IR(label string) *req.Request {
	// Get the interceptor store
	c.interceptorsMu.RLock()
	store, exists := c.interceptors[label]
	c.interceptorsMu.RUnlock()

	// Return an error if the interceptor does not exist
	if !exists {
		return c.client.R().OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			return fmt.Errorf("interceptor %s does not exist", label)
		})
	}

	// Create the request and apply current state
	r := store.apply(c.client.R())

	// Set retry count based on interceptor's max retries
	r.SetRetryCount(store.maxRetries())

	// Capture generation at IR() time - "the state I started with"
	initGen := store.getGeneration()

	// Add retry condition - triggers retry when interceptor says so
	r.AddRetryCondition(func(resp *req.Response, err error) bool {
		return store.shouldIntercept(resp, err)
	})

	// Add retry hook - recovers and re-applies state before retry
	r.AddRetryHook(func(resp *req.Response, err error) {
		currentGen := store.getGeneration()

		// If generation changed since I started, someone else recovered
		// Just apply new state and retry, don't recover again
		if currentGen != initGen {
			store.apply(r)
			initGen = currentGen // update for next retry
			return
		}

		// Otherwise, I need to recover
		if recoverErr := store.recover(c, resp, currentGen); recoverErr == nil {
			store.apply(r)
			initGen = store.getGeneration() // update after my recovery
		}
	})

	return r
}

// AR creates an authenticated request with automatic token refresh
func (c *Client) AR(serviceLabel string) *req.Request {
	// Get the auth manager
	c.authsMu.RLock()
	authManager, exists := c.auths[serviceLabel]
	c.authsMu.RUnlock()

	// Return an error if the auth manager does not exist
	if !exists {
		// Set the error on the request with the hook
		return c.client.R().OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			return fmt.Errorf("auth manager for service %s does not exist", serviceLabel)
		})
	}

	// Get the token
	token, err := authManager.getValidToken()
	if err != nil {
		return c.client.R().OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			return err
		})
	}

	// Return the request builder with the token set
	return c.client.R().SetBearerAuthToken(token)
}

// newRetryClient returns an http client with a retry mechanism
func newRetryClient(opts Options) *req.Client {
	// Set default values if needed
	if opts.RetryCount < 0 {
		opts.RetryCount = defaultRetryCount
	}
	if opts.MinBackoff <= 0 {
		opts.MinBackoff = defaultMinBackoff
	}
	if opts.MaxBackoff <= 0 {
		opts.MaxBackoff = defaultMaxBackoff
	}
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}

	// Create the client
	client := req.NewClient().
		// Timeout
		SetTimeout(opts.Timeout).
		// Retry count
		SetCommonRetryCount(opts.RetryCount).
		// Retry back-off exponential
		SetCommonRetryBackoffInterval(opts.MinBackoff, opts.MaxBackoff).
		// Sets the retry condition (will be retried on any error or non 2XX code)
		SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			return responseErrorCause(resp, err) != nil
		})

	// Enable HTTP/3 if requested
	if opts.EnableHttp3 {
		client.EnableHTTP3()
	}

	// Enable auto decompression if requested
	if opts.AutoDecompress {
		client.EnableAutoDecompress()
	}

	// Disable auto decoding if requested
	if !opts.AutoDecode {
		client.DisableAutoDecode()
	}

	// Disable keep-alive connections if requested
	if opts.DisableKeepAlives {
		client.DisableKeepAlives()
	}

	// Impersonate Chrome, Firefox, or Safari if requested
	switch opts.Impersonation {
	case Chrome:
		client.ImpersonateChrome()
	case Firefox:
		client.ImpersonateFirefox()
	case Safari:
		client.ImpersonateSafari()
	case None:
	default:
	}

	// Return the client
	return client
}

// responseErrorCause returns the error cause for the given response (safely wraps the response)
// Only possible cases:
// - response == nil, err != nil
// - response != nil, err != nil
// - response != nil, err == nil
// There is NO case where both response and err are nil
func responseErrorCause(response *req.Response, err error) error {
	switch {
	// If err is present, return it
	case err != nil:
		return err
	// If the response is nil, return an error
	case response == nil:
		return ErrResponseNil
	// If the response has an error nested, return it
	case response.Err != nil:
		return response.Err
	// If the response has an error HTTP status code (defined per the custom logic), return it
	case response.IsErrorState():
		return fmt.Errorf("error request %s %s with status %d", response.Request.Method, rawURL(response.Request), response.StatusCode)
	// If the response has a body and the content type is JSON, return an error if the body is nil
	case response.GetContentType() == JsonApplicationContentType && response.Body == nil:
		return ErrResponseBodyNil
	}

	// Return nil if no error
	return nil
}

// rawURL returns the raw URL of the given request
func rawURL(req *req.Request) string {
	// Return an empty string if the request is nil
	if req == nil {
		return ""
	}
	// Return the raw URL
	return req.RawURL
}
