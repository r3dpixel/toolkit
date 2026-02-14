package cred

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// FromKeyRing retrieves a value from the OS keyring by label and key.
func FromKeyRing(credLabel, key string) (string, error) {
	return keyring.Get(credLabel, key)
}

// ToKeyRing stores a key-value pair in the OS keyring under the given label.
func ToKeyRing(credLabel, key, value string) error {
	return keyring.Set(credLabel, key, value)
}

// DeleteKeyRing removes a key from the OS keyring, ignoring not found errors.
func DeleteKeyRing(credLabel, key string) error {
	err := keyring.Delete(credLabel, key)

	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}

	return err
}
