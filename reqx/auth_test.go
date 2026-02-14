package reqx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/stretchr/testify/assert"
)

func generateTestJWT(expiration time.Time) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": jwt.NewNumericDate(expiration),
		"iat": jwt.NewNumericDate(time.Now()),
	})
	signedToken, _ := token.SignedString([]byte("test-secret"))
	return signedToken
}

type mockIdentityReader struct {
	identity cred.Identity
	err      error
	label    string
}

func (m *mockIdentityReader) GetUser() (string, error) {
	return m.identity.User, m.err
}

func (m *mockIdentityReader) GetSecret() (string, error) {
	return m.identity.Secret, m.err
}

func (m *mockIdentityReader) Get() (cred.Identity, error) {
	return m.identity, m.err
}

func (m *mockIdentityReader) CredLabel() string {
	if m.label != "" {
		return m.label
	}
	return "test-auth"
}

func TestExtractTokenExpiration(t *testing.T) {
	t.Run("Valid token", func(t *testing.T) {
		expectedExp := time.Now().Add(time.Hour)
		token := generateTestJWT(expectedExp)
		actualExp := extractTokenExpiration(token)
		assert.WithinDuration(t, expectedExp, actualExp, time.Second, "Expiration should match")
	})

	t.Run("Empty token", func(t *testing.T) {
		exp := extractTokenExpiration("")
		assert.True(t, exp.IsZero(), "Empty token should return zero time")
	})

	t.Run("Malformed token", func(t *testing.T) {
		exp := extractTokenExpiration("not-a-jwt")
		assert.True(t, exp.IsZero(), "Malformed token should return zero time")
	})

	t.Run("Token without expiration", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iat": jwt.NewNumericDate(time.Now()),
		})
		signedToken, _ := token.SignedString([]byte("test-secret"))
		exp := extractTokenExpiration(signedToken)
		assert.True(t, exp.IsZero(), "Token without exp claim should return zero time")
	})
}

func TestAuthStore(t *testing.T) {
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}

	t.Run("Token expiration caching", func(t *testing.T) {
		bufferTime := 2 * time.Minute
		client := NewClient(Options{AuthRefreshBuffer: bufferTime})
		reader := &mockIdentityReader{identity: testIdentity}
		store := newRefreshableAuthStore(client, reader, nil, bufferTime)

		// Set valid token and verify expiration is cached
		expectedExp := time.Now().Add(time.Hour)
		token := generateTestJWT(expectedExp)
		store.setBearerToken(token)

		cachedExp := store.getTokenExpiration()
		assert.WithinDuration(t, expectedExp, cachedExp, time.Second, "Cached expiration should match token expiration")

		// Set empty token and verify expiration is cleared
		store.setBearerToken("")
		cachedExp = store.getTokenExpiration()
		assert.True(t, cachedExp.IsZero(), "Empty token should clear cached expiration")
	})

	t.Run("getTokenAndCheckExpiry", func(t *testing.T) {
		now := time.Now()
		bufferTime := 2 * time.Minute
		client := NewClient(Options{AuthRefreshBuffer: bufferTime})
		reader := &mockIdentityReader{identity: testIdentity}
		store := newRefreshableAuthStore(client, reader, nil, bufferTime)

		// Empty token should be expired
		token, isExpired := store.getTokenAndCheckExpiryAt(now)
		assert.Empty(t, token)
		assert.True(t, isExpired, "Empty token should be expired")

		// Set and check malformed token
		store.setBearerToken("not-a-jwt")
		token, isExpired = store.getTokenAndCheckExpiryAt(now)
		assert.Equal(t, "not-a-jwt", token)
		assert.True(t, isExpired, "Malformed token should be expired")

		// Set and check future token
		futureToken := generateTestJWT(time.Now().Add(time.Hour))
		store.setBearerToken(futureToken)
		token, isExpired = store.getTokenAndCheckExpiryAt(now)
		assert.Equal(t, futureToken, token)
		assert.False(t, isExpired, "Valid future token should not be expired")

		// Set and check token expiring soon
		soonToken := generateTestJWT(time.Now().Add(bufferTime / 2))
		store.setBearerToken(soonToken)
		token, isExpired = store.getTokenAndCheckExpiryAt(now)
		assert.Equal(t, soonToken, token)
		assert.True(t, isExpired, "Token expiring soon should be considered expired")

		// Set and check past token
		pastToken := generateTestJWT(time.Now().Add(-time.Hour))
		store.setBearerToken(pastToken)
		token, isExpired = store.getTokenAndCheckExpiryAt(now)
		assert.Equal(t, pastToken, token)
		assert.True(t, isExpired, "Past token should be expired")
	})

	t.Run("Initial token fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		var refreshCalls int32
		validToken := generateTestJWT(time.Now().Add(time.Hour))
		mockRefresh := func(c *Client, identity cred.Identity) (string, error) {
			atomic.AddInt32(&refreshCalls, 1)
			return validToken, nil
		}

		client := NewClient(Options{})
		reader := &mockIdentityReader{identity: testIdentity}
		client.RegisterAuth("test-service", reader, mockRefresh)

		resp, err := client.AR("test-service").Get(server.URL)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(1), atomic.LoadInt32(&refreshCalls), "Refresh function should be called once")
	})

	t.Run("Valid token is reused", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		var refreshCalls int32
		mockRefresh := func(c *Client, identity cred.Identity) (string, error) {
			atomic.AddInt32(&refreshCalls, 1)
			return "should-not-be-called", nil
		}

		client := NewClient(Options{})
		reader := &mockIdentityReader{identity: testIdentity}
		client.RegisterAuth("test-service", reader, mockRefresh)

		validToken := generateTestJWT(time.Now().Add(time.Hour))
		client.auths["test-service"].(*refreshableAuthStore).setBearerToken(validToken)

		_, err := client.AR("test-service").Get(server.URL)
		assert.NoError(t, err)
		assert.Equal(t, int32(0), atomic.LoadInt32(&refreshCalls), "Refresh function should not be called for a valid token")
	})

	t.Run("Expired token is refreshed successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		var refreshCalls int32
		newToken := generateTestJWT(time.Now().Add(time.Hour))
		mockRefresh := func(c *Client, identity cred.Identity) (string, error) {
			atomic.AddInt32(&refreshCalls, 1)
			return newToken, nil
		}

		client := NewClient(Options{})
		reader := &mockIdentityReader{identity: testIdentity}
		client.RegisterAuth("test-service", reader, mockRefresh)

		expiredToken := generateTestJWT(time.Now().Add(-time.Hour))
		client.auths["test-service"].(*refreshableAuthStore).setBearerToken(expiredToken)

		_, err := client.AR("test-service").Get(server.URL)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&refreshCalls), "Refresh function should be called for an expired token")
	})

	t.Run("Expired token refresh fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		refreshError := errors.New("failed to refresh")
		mockRefresh := func(c *Client, identity cred.Identity) (string, error) {
			return "", refreshError
		}

		client := NewClient(Options{})
		reader := &mockIdentityReader{identity: testIdentity}
		client.RegisterAuth("test-service", reader, mockRefresh)

		expiredToken := generateTestJWT(time.Now().Add(-time.Hour))
		client.auths["test-service"].(*refreshableAuthStore).setBearerToken(expiredToken)

		_, err := client.AR("test-service").Get(server.URL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to refresh")
		assert.Empty(t, client.auths["test-service"].(*refreshableAuthStore).getBearerToken(), "Internal token should be cleared on failure")
	})

	t.Run("Concurrent requests trigger only one refresh", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		var refreshCalls int32
		newToken := generateTestJWT(time.Now().Add(time.Hour))
		mockRefresh := func(c *Client, identity cred.Identity) (string, error) {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&refreshCalls, 1)
			return newToken, nil
		}

		client := NewClient(Options{})
		reader := &mockIdentityReader{identity: testIdentity}
		client.RegisterAuth("test-service", reader, mockRefresh)

		expiredToken := generateTestJWT(time.Now().Add(-time.Hour))
		client.auths["test-service"].(*refreshableAuthStore).setBearerToken(expiredToken)

		var wg sync.WaitGroup
		numRequests := 10
		wg.Add(numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				defer wg.Done()
				_, err := client.AR("test-service").Get(server.URL)
				assert.NoError(t, err)
			}()
		}

		wg.Wait()

		assert.Equal(t, int32(1), atomic.LoadInt32(&refreshCalls), "Refresh function should only be called once despite concurrent requests")
	})
}
