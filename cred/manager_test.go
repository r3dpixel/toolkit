package cred

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func stringPtr(s string) *string {
	return &s
}

func testManagerLifecycle(t *testing.T, m IdentityManager, notFoundErr error) {
	identity := Identity{
		User:   "test-lifecycle-user",
		Secret: "s3cr3t-l1fecycl3-p@ssw0rd!",
	}

	_, err := m.Get()
	assert.Error(t, err, "Get() on a non-existent identity should return an error")
	assert.True(t, errors.Is(err, notFoundErr), "The error should be the expected not-found error")

	err = m.SetAll(identity)
	assert.NoError(t, err, "SetAll() should not return an error")

	retrieved, err := m.Get()
	assert.NoError(t, err, "Get() after SetAll() should not return an error")
	assert.Equal(t, identity, retrieved, "The retrieved identity should match the one that was set")

	err = m.Delete()
	assert.NoError(t, err, "Delete() should not return an error")

	_, err = m.Get()
	assert.True(t, errors.Is(err, notFoundErr), "Get() after Delete() should fail with the not-found error")

	err = m.Delete()
	assert.NoError(t, err, "Deleting a non-existent identity should not return an error")
}

func testManagerSetPartialPayloads(t *testing.T, m IdentityManager, notFoundErr error) {
	user := "partial-user"
	secret := "p@rtial-s3cr3t"

	err := m.Set(IdentityPayload{User: stringPtr(user)})
	assert.NoError(t, err, "Set() with only user should succeed")

	retrievedUser, err := m.GetUser()
	assert.NoError(t, err)
	assert.Equal(t, user, retrievedUser)
	_, err = m.GetSecret()
	assert.True(t, errors.Is(err, notFoundErr), "Secret should not exist yet")

	err = m.Set(IdentityPayload{Secret: stringPtr(secret)})
	assert.NoError(t, err, "Set() with only secret should succeed")

	retrievedSecret, err := m.GetSecret()
	assert.NoError(t, err)
	assert.Equal(t, secret, retrievedSecret)
	retrievedUser, err = m.GetUser() // re-check user
	assert.NoError(t, err)
	assert.Equal(t, user, retrievedUser)

	fullIdentity, err := m.Get()
	assert.NoError(t, err)
	assert.Equal(t, user, fullIdentity.User)
	assert.Equal(t, secret, fullIdentity.Secret)
}

func testManagerSetUserGetUser(t *testing.T, m IdentityManager, notFoundErr error) {
	user := "standalone-user"

	_, err := m.GetUser()
	assert.True(t, errors.Is(err, notFoundErr))

	err = m.SetUser(user)
	assert.NoError(t, err)

	retrieved, err := m.GetUser()
	assert.NoError(t, err)
	assert.Equal(t, user, retrieved)
}

func testManagerSetSecretGetSecret(t *testing.T, m IdentityManager, notFoundErr error) {
	secret := "st@ndal0ne-s3cr3t"

	_, err := m.GetSecret()
	assert.True(t, errors.Is(err, notFoundErr))

	err = m.SetSecret(secret)
	assert.NoError(t, err)

	retrieved, err := m.GetSecret()
	assert.NoError(t, err)
	assert.Equal(t, secret, retrieved)
}

func testManagerGetPartialFailure(t *testing.T, m IdentityManager, notFoundErr error) {
	t.Run("Fails when only user exists", func(t *testing.T) {
		err := m.SetUser("only-user-exists")
		assert.NoError(t, err)

		_, err = m.Get()
		assert.True(t, errors.Is(err, notFoundErr), "Get() should fail if secret is missing")
	})

	t.Run("Fails when only secret exists", func(t *testing.T) {
		_ = m.Delete()
		err := m.SetSecret("only-secret-exists")
		assert.NoError(t, err)

		_, err = m.Get()
		assert.True(t, errors.Is(err, notFoundErr), "Get() should fail if user is missing")
	})
}

func TestManager(t *testing.T) {
	testModes := []struct {
		name        string
		mode        Mode
		notFoundErr error
	}{
		{
			name:        "KeyRing Mode",
			mode:        KeyRing,
			notFoundErr: keyring.ErrNotFound,
		},
		{
			name:        "Env Mode",
			mode:        Env,
			notFoundErr: ErrEnvVarNotFound,
		},
	}

	testCases := []struct {
		name string
		fn   func(t *testing.T, m IdentityManager, notFoundErr error)
	}{
		{"Lifecycle", testManagerLifecycle},
		{"SetPartialPayloads", testManagerSetPartialPayloads},
		{"SetUserGetUser", testManagerSetUserGetUser},
		{"SetSecretGetSecret", testManagerSetSecretGetSecret},
		{"GetPartialFailure", testManagerGetPartialFailure},
	}

	for _, mode := range testModes {
		t.Run(mode.name, func(t *testing.T) {
			credLabel := fmt.Sprintf("cred-test-%s", t.Name())
			m := NewManager(credLabel, mode.mode)
			t.Cleanup(func() {
				_ = m.Delete()
			})

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					_ = m.Delete()
					tc.fn(t, m, mode.notFoundErr)
				})
			}
		})
	}
}

func TestManager_Label(t *testing.T) {
	label := "test-manager-label"
	m := NewManager(label, Env)
	assert.Equal(t, label, m.CredLabel())
}
