package ptr

import "unsafe"

// Of returns a pointer to the given value (creates a copy if value is a direct Golang value: int, string, struct, etc.)
func Of[T any](value T) *T {
	return &value
}

// Address returns the address of the given pointer
func Address[T any](ptr *T) uintptr {
	return *(*uintptr)(unsafe.Pointer(ptr))
}
