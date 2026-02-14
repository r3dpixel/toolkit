package reqx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_DefaultOptions(t *testing.T) {
	client := NewClient(Options{})

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.auths)
}

func TestNewClient_CustomOptions(t *testing.T) {
	opts := Options{
		RetryCount:     2,
		MinBackoff:     5 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Timeout:        5 * time.Second,
		EnableHttp3:    false,
		AutoDecompress: true,
		AutoDecode:     false,
		Impersonation:  Chrome,
	}

	client := NewClient(opts)

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestNewClient_WithConfig(t *testing.T) {
	customHeaderSet := false
	config := func(c *req.Client) {
		c.SetCommonHeader("X-Custom-Header", "test-value")
		customHeaderSet = true
	}

	client := NewClient(Options{}, config)

	assert.NotNil(t, client)
	assert.True(t, customHeaderSet)
}

func TestNewClient_WithMultipleConfigs(t *testing.T) {
	config1Called := false
	config2Called := false

	config1 := func(c *req.Client) {
		config1Called = true
	}

	config2 := func(c *req.Client) {
		config2Called = true
	}

	client := NewClient(Options{}, config1, config2)

	assert.NotNil(t, client)
	assert.True(t, config1Called)
	assert.True(t, config2Called)
}

func TestClient_R(t *testing.T) {
	client := NewClient(Options{})

	assert.NotNil(t, client.R())
}

func TestClient_RegisterAuth(t *testing.T) {
	client := NewClient(Options{})
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}
	reader := &mockIdentityReader{identity: testIdentity}

	refreshCalled := false
	refreshFunc := func(c *Client, identity cred.Identity) (string, error) {
		refreshCalled = true
		return "test-token", nil
	}

	client.RegisterAuth("test-service", reader, refreshFunc)

	assert.Contains(t, client.auths, "test-service")
	assert.NotNil(t, client.auths["test-service"])
	assert.False(t, refreshCalled)

	client.AR("test-service").Get("http://example.com")
	assert.True(t, refreshCalled)
}

func TestClient_AR_Success(t *testing.T) {
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}
	validToken := generateTestJWT(time.Now().Add(time.Hour))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Contains(t, auth, "Bearer ")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Options{})
	reader := &mockIdentityReader{identity: testIdentity}
	client.RegisterAuth("test-service", reader, func(c *Client, identity cred.Identity) (string, error) {
		return validToken, nil
	})

	resp, err := client.AR("test-service").Get(server.URL)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_AR_NonExistentLabel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Options{})

	_, err := client.AR("non-existent").Get(server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth manager for service non-existent does not exist")
}

func TestClient_AR_TokenRefreshError(t *testing.T) {
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}

	client := NewClient(Options{})
	reader := &mockIdentityReader{identity: testIdentity}
	client.RegisterAuth("test-service", reader, func(c *Client, identity cred.Identity) (string, error) {
		return "", assert.AnError
	})

	_, err := client.AR("test-service").Get("http://example.com")
	assert.Error(t, err)
}

func TestClient_AR_WithMethodChaining(t *testing.T) {
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}
	validToken := generateTestJWT(time.Now().Add(time.Hour))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-value", r.Header.Get("X-Custom-Header"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer ")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Options{})
	reader := &mockIdentityReader{identity: testIdentity}
	client.RegisterAuth("test-service", reader, func(c *Client, identity cred.Identity) (string, error) {
		return validToken, nil
	})

	resp, err := client.AR("test-service").
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Custom-Header", "test-value").
		Get(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_ErrorHandling_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(Options{})
	_, err := client.R().Get(server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestClient_ErrorHandling_500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(Options{})
	_, err := client.R().Get(server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestClient_RetryBehavior(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient(Options{
		RetryCount: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	})

	resp, err := client.R().Get(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attemptCount, "Should have retried until success")
}

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Options{
		Timeout:    10 * time.Millisecond,
		RetryCount: 0,
	})

	_, err := client.R().Get(server.URL)

	assert.Error(t, err)
}

func TestClient_ImpersonationOptions(t *testing.T) {
	tests := []struct {
		name          string
		impersonation Impersonation
	}{
		{"None", None},
		{"Chrome", Chrome},
		{"Firefox", Firefox},
		{"Safari", Safari},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(Options{
				Impersonation: tt.impersonation,
			})

			assert.NotNil(t, client)
		})
	}
}

func TestClient_AutoDecompressOption(t *testing.T) {
	client := NewClient(Options{
		AutoDecompress: true,
	})

	assert.NotNil(t, client)
}

func TestClient_AutoDecodeOption(t *testing.T) {
	t.Run("AutoDecode enabled", func(t *testing.T) {
		client := NewClient(Options{
			AutoDecode: true,
		})

		assert.NotNil(t, client)
	})

	t.Run("AutoDecode disabled", func(t *testing.T) {
		client := NewClient(Options{
			AutoDecode: false,
		})

		assert.NotNil(t, client)
	})
}

func TestClient_ConcurrentAuthRequests(t *testing.T) {
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}
	validToken := generateTestJWT(time.Now().Add(time.Hour))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	refreshCount := 0
	client := NewClient(Options{})
	reader := &mockIdentityReader{identity: testIdentity}
	client.RegisterAuth("test-service", reader, func(c *Client, identity cred.Identity) (string, error) {
		refreshCount++
		time.Sleep(50 * time.Millisecond) // Simulate slow token fetch
		return validToken, nil
	})

	// Make 5 concurrent requests
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := client.AR("test-service").Get(server.URL)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Should only refresh once despite concurrent requests
	assert.Equal(t, 1, refreshCount)
}

func TestClient_MultipleAuthProviders(t *testing.T) {
	identity1 := cred.Identity{User: "user1", Secret: "secret1"}
	identity2 := cred.Identity{User: "user2", Secret: "secret2"}

	token1 := generateTestJWT(time.Now().Add(time.Hour))
	token2 := generateTestJWT(time.Now().Add(time.Hour))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Options{})

	reader1 := &mockIdentityReader{identity: identity1}
	client.RegisterAuth("service-1", reader1, func(c *Client, identity cred.Identity) (string, error) {
		return token1, nil
	})

	reader2 := &mockIdentityReader{identity: identity2}
	client.RegisterAuth("service-2", reader2, func(c *Client, identity cred.Identity) (string, error) {
		return token2, nil
	})

	// Test first auth provider
	resp1, err1 := client.AR("service-1").Get(server.URL)
	assert.NoError(t, err1)
	assert.NotNil(t, resp1)

	// Test second auth provider
	resp2, err2 := client.AR("service-2").Get(server.URL)
	assert.NoError(t, err2)
	assert.NotNil(t, resp2)
}

func TestClient_UnregisterAuth(t *testing.T) {
	client := NewClient(Options{})
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}
	reader := &mockIdentityReader{identity: testIdentity}

	refreshFunc := func(c *Client, identity cred.Identity) (string, error) {
		return "test-token", nil
	}

	client.RegisterAuth("test-service", reader, refreshFunc)
	assert.Contains(t, client.auths, "test-service")

	client.UnregisterAuth("test-service")
	assert.NotContains(t, client.auths, "test-service")
}

func TestClient_RegisterToken(t *testing.T) {
	client := NewClient(Options{})
	fixedToken := "fixed-bearer-token-123"

	client.RegisterToken("test-service", fixedToken)
	assert.Contains(t, client.auths, "test-service")

	authStore := client.auths["test-service"]
	assert.NotNil(t, authStore)

	token, err := authStore.getValidToken()
	assert.NoError(t, err)
	assert.Equal(t, fixedToken, token)
}

func TestClient_RegisterToken_Success(t *testing.T) {
	fixedToken := "fixed-token-xyz"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+fixedToken, auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Options{})
	client.RegisterToken("test-service", fixedToken)

	resp, err := client.AR("test-service").Get(server.URL)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_RegisterToken_MethodChaining(t *testing.T) {
	client := NewClient(Options{}).
		RegisterToken("service-1", "token-1").
		RegisterToken("service-2", "token-2")

	assert.Contains(t, client.auths, "service-1")
	assert.Contains(t, client.auths, "service-2")
}

func TestClient_RegisterAuth_MethodChaining(t *testing.T) {
	testIdentity := cred.Identity{User: "testuser", Secret: "testsecret"}
	reader := &mockIdentityReader{identity: testIdentity}

	client := NewClient(Options{}).
		RegisterAuth("service-1", reader, func(c *Client, identity cred.Identity) (string, error) {
			return "token-1", nil
		}).
		RegisterAuth("service-2", reader, func(c *Client, identity cred.Identity) (string, error) {
			return "token-2", nil
		})

	assert.Contains(t, client.auths, "service-1")
	assert.Contains(t, client.auths, "service-2")
}
