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
}

// NewClient creates a new wrapped client
func NewClient(opts Options, configs ...Config) *Client {
	if opts.AuthRefreshBuffer <= 0 {
		opts.AuthRefreshBuffer = defaultAuthRefreshBuffer
	}

	// Create the retryable client
	client := newRetryClient(opts)

	// Add response error checking callback
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
	}
}

// RegisterAuth registers an authentication provider using a refreshable token with the given label
func (c *Client) RegisterAuth(serviceLabel string, identityReader cred.IdentityReader, refreshFunc func(*Client, cred.Identity) (string, error)) *Client {
	c.authsMu.Lock()
	defer c.authsMu.Unlock()
	c.auths[serviceLabel] = newRefreshableAuthStore(c, identityReader, refreshFunc, c.authRefreshBuffer)

	return c
}

// RegisterToken registers an authentication provider using a fixed token with the given label
func (c *Client) RegisterToken(serviceLabel string, token string) *Client {
	c.authsMu.Lock()
	defer c.authsMu.Unlock()
	c.auths[serviceLabel] = newTokenAuthStore(token)

	return c
}

// UnregisterAuth unregisters an authentication provider with the given label
func (c *Client) UnregisterAuth(serviceLabel string) {
	c.authsMu.Lock()
	defer c.authsMu.Unlock()
	delete(c.auths, serviceLabel)
}

// R creates a normal request builder
func (c *Client) R() *req.Request {
	return c.client.R()
}

// AR creates an authenticated request with automatic token refresh
func (c *Client) AR(serviceLabel string) *req.Request {
	c.authsMu.RLock()
	authManager, exists := c.auths[serviceLabel]
	c.authsMu.RUnlock()

	if !exists {
		return c.client.R().OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			return fmt.Errorf("auth manager for service %s does not exist", serviceLabel)
		})
	}

	token, err := authManager.getValidToken()
	if err != nil {
		return c.client.R().OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			return err
		})
	}

	return c.client.R().SetBearerAuthToken(token)
}

// newRetryClient returns an http client with a retry mechanism
func newRetryClient(opts Options) *req.Client {
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

func responseErrorCause(response *req.Response, err error) error {
	switch {
	case err != nil:
		return err
	case response == nil:
		return ErrResponseNil
	case response.Err != nil:
		return response.Err
	case response.IsErrorState():
		return fmt.Errorf("error request %s %s with status %d", response.Request.Method, rawURL(response.Request), response.StatusCode)
	case response.GetContentType() == JsonApplicationContentType && response.Body == nil:
		return ErrResponseBodyNil
	}

	return nil
}

func rawURL(req *req.Request) string {
	if req == nil {
		return "[]"
	}
	return req.RawURL
}
