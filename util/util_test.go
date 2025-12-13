package util

import (
	"testing"
)

func TestGetOrDefault(t *testing.T) {
	defaultPtr := new(int)
	*defaultPtr = 100
	valPtr := new(int)
	*valPtr = 50

	testCases := []struct {
		name   string
		testFn func(t *testing.T)
	}{
		{
			name: "int with zero value",
			testFn: func(t *testing.T) {
				result := GetOrDefault(0, 42)
				if result != 42 {
					t.Errorf("Expected 42, got %d", result)
				}
			},
		},
		{
			name: "int with non-zero value",
			testFn: func(t *testing.T) {
				result := GetOrDefault(10, 42)
				if result != 10 {
					t.Errorf("Expected 10, got %d", result)
				}
			},
		},
		{
			name: "string with zero value",
			testFn: func(t *testing.T) {
				result := GetOrDefault("", "default")
				if result != "default" {
					t.Errorf("Expected 'default', got '%s'", result)
				}
			},
		},
		{
			name: "string with non-zero value",
			testFn: func(t *testing.T) {
				result := GetOrDefault("hello", "default")
				if result != "hello" {
					t.Errorf("Expected 'hello', got '%s'", result)
				}
			},
		},
		{
			name: "pointer with nil value",
			testFn: func(t *testing.T) {
				result := GetOrDefault(nil, defaultPtr)
				if result != defaultPtr {
					t.Errorf("Expected pointer %v, got %v", defaultPtr, result)
				}
			},
		},
		{
			name: "pointer with non-nil value",
			testFn: func(t *testing.T) {
				result := GetOrDefault(valPtr, defaultPtr)
				if result != valPtr {
					t.Errorf("Expected pointer %v, got %v", valPtr, result)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFn)
	}
}
