package slicesx

import (
	"math/bits"
	"slices"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/r3dpixel/toolkit/structx"
)

// Grow assures the given slice has the length to index elements up to the lastIndex (inclusive)
func Grow[T any](container *[]T, lastIndex int) {
	// Compute the new length
	newLen := NextPowerOfTwo(lastIndex + 1)
	// If the new length is sufficient, return
	if newLen <= len(*container) {
		return
	}

	// Grow the slice
	*container = slices.Grow(*container, newLen-len(*container))

	// Truncate the slice to the new length
	*container = (*container)[:newLen]
}

// NextPowerOfTwo returns the next power of two for the given value
func NextPowerOfTwo(value int) int {
	// Return 1 if the value is 0 or 1
	if value <= 1 {
		return 1
	}

	return 1 << bits.Len(uint(value))
}

// PrependValue prepends the given value to the given slice and returns the new slice
func PrependValue[T any](value T, slice []T) []T {
	return append([]T{value}, slice...)
}

// Map functional map of a slice using the op func
func Map[T, V any](s []T, op func(T) V) []V {
	// Return nil if the slice is empty
	if len(s) == 0 {
		return nil
	}

	// Create a new slice of the same length
	result := make([]V, len(s))

	// Map each element in the slice
	for index, t := range s {
		result[index] = op(t)
	}

	// Return the mapped slice
	return result
}

// MapTo maps the elements of the source slice to the destination slice using the op func
func MapTo[T, V any](src []T, dst []V, op func(T) V) {
	// Compute the minimum length of the slices
	limit := min(len(src), len(dst))
	// Map each element in the source slice to the destination slice
	for index := range limit {
		dst[index] = op(src[index])
	}
}

// DeduplicateStable merges slices into a single slice without duplicates (maintaining order)
func DeduplicateStable[T comparable](slices ...[]T) []T {
	// Return nil if no slices are provided
	if len(slices) == 0 {
		return nil
	}

	// Return the first slice if only one is provided
	if len(slices) == 1 {
		return slices[0]
	}

	// Compute total capacity
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	// Merge the slices using an ordered map
	merged := orderedmap.NewOrderedMapWithCapacity[T, struct{}](totalLen)
	for _, s := range slices {
		for _, item := range s {
			merged.Set(item, structx.Empty)
		}
	}

	// Convert the ordered map to a slice
	result := make([]T, merged.Len())

	// Iterate over the keys of the ordered map and copy them to the result slice
	index := 0
	for item := range merged.Keys() {
		result[index] = item
		index++
	}

	// Return the result slice
	return result
}
