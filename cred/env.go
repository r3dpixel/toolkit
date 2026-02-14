package cred

import (
	"errors"
	"os"
	"strings"

	"github.com/r3dpixel/toolkit/symbols"
)

var ErrEnvVarNotFound = errors.New("environment variable not set")

// toEnvVarName converts a label and key into an uppercase environment variable name.
func toEnvVarName(credLabel, key string) string {
	var b strings.Builder
	b.Grow(len(credLabel) + 1 + len(key))
	b.WriteString(credLabel)
	b.WriteByte(symbols.UnderscoreByte)
	b.WriteString(key)
	return strings.ToUpper(b.String())
}

// FromEnv retrieves a value from an environment variable based on label and key.
func FromEnv(credLabel, key string) (string, error) {
	value, ok := os.LookupEnv(toEnvVarName(credLabel, key))
	if !ok {
		return "", ErrEnvVarNotFound
	}
	return value, nil
}

// ToEnv sets an environment variable value based on label and key.
func ToEnv(credLabel, key, value string) error {
	return os.Setenv(toEnvVarName(credLabel, key), value)
}

// DeleteEnv removes an environment variable based on label and key.
func DeleteEnv(credLabel, key string) error {
	return os.Unsetenv(toEnvVarName(credLabel, key))
}
