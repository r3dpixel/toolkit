package util

// GetOrDefault if the value is golang ZERO value it returns the defaultValue, otherwise it returns the value itself
func GetOrDefault[T comparable](value T, defaultValue T) T {
	var zero T
	if value == zero {
		return defaultValue
	}
	return value
}
