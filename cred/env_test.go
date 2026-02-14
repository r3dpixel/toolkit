package cred

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvLifecycle(t *testing.T) {
	testLabel := fmt.Sprintf("cred-test-%s", t.Name())
	testKey := "test-user"
	testValue := "s3cr3t-3nv-v@r!"

	t.Cleanup(func() {
		_ = DeleteEnv(testLabel, testKey)
	})

	t.Run("Get non-existent env var fails", func(t *testing.T) {
		_, err := FromEnv(testLabel, testKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not set")
	})

	t.Run("Set and Get successfully", func(t *testing.T) {
		err := ToEnv(testLabel, testKey, testValue)
		assert.NoError(t, err)

		retrievedValue, err := FromEnv(testLabel, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testValue, retrievedValue)
	})

	t.Run("Delete successfully", func(t *testing.T) {
		err := DeleteEnv(testLabel, testKey)
		assert.NoError(t, err)

		_, err = FromEnv(testLabel, testKey)
		assert.Error(t, err, "Expected an error after deleting the env var")
		assert.Contains(t, err.Error(), "not set")
	})

	t.Run("Delete non-existent env var again", func(t *testing.T) {
		err := DeleteEnv(testLabel, testKey)
		assert.NoError(t, err)
	})
}
