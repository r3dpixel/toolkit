package cred

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func TestKeyRingLifecycle(t *testing.T) {
	testLabel := fmt.Sprintf("cred-test-%s", t.Name())
	testKey := "test-user"
	testValue := "s3cr3t-p@ssw0rd!"

	t.Cleanup(func() {
		_ = DeleteKeyRing(testLabel, testKey)
	})

	t.Run("Get non-existent key fails", func(t *testing.T) {
		_, err := FromKeyRing(testLabel, testKey)
		assert.Error(t, err)
		assert.ErrorIs(t, err, keyring.ErrNotFound)
	})

	t.Run("Set and Get successfully", func(t *testing.T) {
		err := ToKeyRing(testLabel, testKey, testValue)
		assert.NoError(t, err)

		retrievedKey, err := FromKeyRing(testLabel, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testValue, retrievedKey)
	})

	t.Run("Delete successfully", func(t *testing.T) {
		err := DeleteKeyRing(testLabel, testKey)
		assert.NoError(t, err)

		_, err = FromKeyRing(testLabel, testKey)
		assert.ErrorIs(t, err, keyring.ErrNotFound)
	})

	t.Run("Delete non-existent key again", func(t *testing.T) {
		err := DeleteKeyRing(testLabel, testKey)
		assert.NoError(t, err)
	})
}
