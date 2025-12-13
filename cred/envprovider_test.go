package cred

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvProvider_Lifecycle(t *testing.T) {
	credLabel := fmt.Sprintf("cred-test-%s", t.Name())
	key := "test-key"
	value := "s3cr3t-p@ssw0rd!"
	p := NewEnvProvider(credLabel)

	t.Cleanup(func() {
		_ = p.Delete(key)
	})

	t.Run("Get non-existent value fails", func(t *testing.T) {
		_, err := p.Get(key)
		assert.Error(t, err, "Expected an error when getting a non-existent value")
	})

	t.Run("Set and Get successfully", func(t *testing.T) {
		err := p.Set(key, value)
		assert.NoError(t, err)

		retrievedKey, err := p.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedKey)
	})

	t.Run("Delete successfully", func(t *testing.T) {
		err := p.Delete(key)
		assert.NoError(t, err)

		_, err = p.Get(key)
		assert.Error(t, err, "Expected an error after deleting the value")
	})
}

func TestEnvProvider_Concurrency(t *testing.T) {
	label := fmt.Sprintf("gemini-test-envProvider-concurrent-%s", t.Name())
	key := "concurrent-key"
	p := NewEnvProvider(label)

	t.Cleanup(func() {
		_ = p.Delete(key)
	})

	err := p.Set(key, "initial-value")
	assert.NoError(t, err)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = p.Get(key)
			_ = p.Set(key, fmt.Sprintf("key-from-%d", id))
			_, _ = p.Get(key)
			_ = p.Delete(key)
			_, _ = p.Get(key)
			_ = p.Set(key, fmt.Sprintf("final-key-from-%d", id))
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent test finished. Run with 'go test -race' to verify safety.")
}

func TestEnvProvider_Label(t *testing.T) {
	label := "test-label"
	p := NewEnvProvider(label)
	assert.Equal(t, label, p.CredLabel())
}
