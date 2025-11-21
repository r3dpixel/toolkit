package slicesx

import (
	"math/bits"
	"slices"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/r3dpixel/toolkit/structx"
)

// Grow assures the given slice has the length to index elements up to the lastIndex (inclusive)
func Grow[T any](container *[]T, lastIndex int) {
	newLen := NextPowerOfTwo(lastIndex + 1)
	if newLen <= len(*container) {
		return
	}

	*container = slices.Grow(*container, newLen-len(*container))

	*container = (*container)[:newLen]
}

// NextPowerOfTwo returns the next power of two for the given value
func NextPowerOfTwo(value int) int {
	if value <= 1 {
		return 1
	}

	return 1 << bits.Len(uint(value-1))
}

// PrependValue prepends the given value to the given slice and returns the new slice
func PrependValue[T any](value T, slice []T) []T {
	return append([]T{value}, slice...)
}

// Map functional map of a slice using the op func
func Map[T, V any](s []T, op func(T) V) []V {
	if len(s) == 0 {
		return nil
	}

	result := make([]V, len(s))

	for index, t := range s {
		result[index] = op(t)
	}

	return result
}

func MapTo[T, V any](src []T, dst []V, op func(T) V) {
	limit := min(len(src), len(dst))
	for index := range limit {
		dst[index] = op(src[index])
	}
}

// Merge merges two slices into a single slice without duplicates
func Merge[T comparable](a []T, b []T) []T {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	merged := make(map[T]struct{}, len(a)+len(b))
	for _, item := range a {
		merged[item] = structx.Empty
	}
	for _, item := range b {
		merged[item] = structx.Empty
	}
	result := make([]T, len(merged))
	index := 0
	for item := range merged {
		result[index] = item
		index++
	}
	return result
}

// MergeStable merges two slices into a single slice without duplicates (maintaining order)
func MergeStable[T comparable](a []T, b []T) []T {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	merged := orderedmap.NewOrderedMapWithCapacity[T, struct{}](len(a) + len(b))
	for _, item := range a {
		merged.Set(item, structx.Empty)
	}
	for _, item := range b {
		merged.Set(item, structx.Empty)
	}
	result := make([]T, merged.Len())
	index := 0
	for item := range merged.Keys() {
		result[index] = item
		index++
	}
	return result
}
