package reqx

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBytes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		}))
		defer server.Close()

		client := NewClient(Options{})
		bytes, err := Bytes(client.R().Get(server.URL))

		require.NoError(t, err)
		assert.Equal(t, []byte("test response"), bytes)
	})

	t.Run("Err propagation", func(t *testing.T) {
		client := NewClient(Options{})
		testErr := errors.New("test error")

		bytes, err := Bytes(client.R().Get("http://invalid-url-that-does-not-exist-12345.com"))

		assert.Error(t, err)
		assert.Nil(t, bytes)
		assert.NotEqual(t, testErr, err)
	})
}

func TestString(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hello world"))
		}))
		defer server.Close()

		client := NewClient(Options{})
		str, err := String(client.R().Get(server.URL))

		require.NoError(t, err)
		assert.Equal(t, "hello world", str)
	})

	t.Run("Empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(Options{})
		str, err := String(client.R().Get(server.URL))

		require.NoError(t, err)
		assert.Equal(t, "", str)
	})

	t.Run("Err propagation", func(t *testing.T) {
		client := NewClient(Options{})

		str, err := String(client.R().Get("http://invalid-url-that-does-not-exist-12345.com"))

		assert.Error(t, err)
		assert.Equal(t, "", str)
	})
}

func TestStream(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("stream content"))
		}))
		defer server.Close()

		client := NewClient(Options{})
		stream, err := Stream(client.R().Get(server.URL))

		require.NoError(t, err)
		require.NotNil(t, stream)
		defer stream.Close()

		content, err := io.ReadAll(stream)
		require.NoError(t, err)
		assert.Equal(t, []byte("stream content"), content)
	})

	t.Run("Err propagation", func(t *testing.T) {
		client := NewClient(Options{})

		stream, err := Stream(client.R().Get("http://invalid-url-that-does-not-exist-12345.com"))

		assert.Error(t, err)
		assert.Nil(t, stream)
	})
}
