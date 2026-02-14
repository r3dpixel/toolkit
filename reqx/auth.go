package reqx

import (
	"sync"
	"time"

	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/ptr"
	"github.com/r3dpixel/toolkit/stringsx"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// jwtParser is thread safe and immutable (this instance will be reused for all operations)
	jwtParser = jwt.NewParser()
)

// authStore manages http requests that need bearer authentication
type authStore interface {
	// getValidToken returns a valid token, refreshing it if expired
	getValidToken() (string, error)
}

// tokenAuthStore manages http requests that need bearer authentication with a fixed token
type tokenAuthStore string

// getValidToken returns the token stored in the store
func (t *tokenAuthStore) getValidToken() (string, error) {
	return string(*t), nil
}

// newTokenAuthStore creates a new tokenAuthStore with the provided token
func newTokenAuthStore(token string) *tokenAuthStore {
	return ptr.Of(tokenAuthStore(token))
}

// refreshableAuthStore manages http requests that need bearer authentication refreshing the token optimally
// (completely thread-safe for high-throughput concurrent requests)
type refreshableAuthStore struct {
	client            *Client
	identityReader    cred.IdentityReader
	refreshTokenFunc  RefreshTokenFunc
	authRefreshBuffer time.Duration

	token           string
	tokenExpiration time.Time
	tokenMu         sync.RWMutex
	refreshMu       sync.Mutex
}

// newRefreshableAuthStore creates a new refreshableAuthStore with the provided token refresh function
func newRefreshableAuthStore(client *Client, identityReader cred.IdentityReader, refreshTokenFunc RefreshTokenFunc, authRefreshBuffer time.Duration) *refreshableAuthStore {
	return &refreshableAuthStore{
		client:            client,
		identityReader:    identityReader,
		refreshTokenFunc:  refreshTokenFunc,
		authRefreshBuffer: authRefreshBuffer,
	}
}

// getValidToken returns a valid token, refreshing it if expired
func (as *refreshableAuthStore) getValidToken() (string, error) {
	// Get current time
	now := time.Now()

	// Atomically get a token and check if expired
	token, isExpired := as.getTokenAndCheckExpiryAt(now)
	if !isExpired {
		return token, nil
	}

	// Lock the refresh mutex and refresh the token
	as.refreshMu.Lock()
	defer as.refreshMu.Unlock()

	// Check again after acquiring the lock
	token, isExpired = as.getTokenAndCheckExpiryAt(now)
	if !isExpired {
		return token, nil
	}

	// Get the identity
	identity, err := as.identityReader.Get()
	if err != nil {
		as.setBearerToken("")
		return "", err
	}

	// Refresh the token
	newToken, err := as.refreshTokenFunc(as.client, identity)
	if err != nil {
		as.setBearerToken("")
		return "", err
	}
	as.setBearerToken(newToken)

	// Return the new token
	return newToken, nil
}

// getTokenAndCheckExpiry atomically retrieves the token and checks if it's expired
func (as *refreshableAuthStore) getTokenAndCheckExpiryAt(t time.Time) (token string, isExpired bool) {
	as.tokenMu.RLock()
	defer as.tokenMu.RUnlock()
	token = as.token
	isExpired = as.tokenExpiration.Before(t.Add(as.authRefreshBuffer))
	return
}

// getBearerToken safely retrieves the current bearer token
func (as *refreshableAuthStore) getBearerToken() string {
	as.tokenMu.RLock()
	defer as.tokenMu.RUnlock()
	return as.token
}

// getTokenExpiration safely retrieves the cached token expiration time
func (as *refreshableAuthStore) getTokenExpiration() time.Time {
	as.tokenMu.RLock()
	defer as.tokenMu.RUnlock()
	return as.tokenExpiration
}

// setBearerToken safely sets the bearer token and caches its expiration time
func (as *refreshableAuthStore) setBearerToken(token string) {
	as.tokenMu.Lock()
	defer as.tokenMu.Unlock()
	as.token = token
	as.tokenExpiration = extractTokenExpiration(token)
}

// extractTokenExpiration parses a JWT token and returns its expiration time
func extractTokenExpiration(tokenString string) time.Time {
	// Check if the token is blank
	if stringsx.IsBlank(tokenString) {
		return time.Time{}
	}

	// Parse the token
	parsedToken, _, err := jwtParser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}
	}

	// Extract the expiration time
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}
	}

	// Check if the token has an expiration time
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return time.Time{}
	}

	// Return the expiration time
	return exp.Time
}
