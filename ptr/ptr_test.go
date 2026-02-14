package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertOfAny[T comparable](t *testing.T, value T) {
	t.Helper()
	ptr := Of(value)
	assert.NotNil(t, ptr, "Of(%v) should not return nil", value)
	assert.Equal(t, value, *ptr, "Of(%v) should return pointer to same value", value)
}

func TestOf(t *testing.T) {
	testCases := []struct {
		name   string
		values []any
	}{
		{
			name:   "bool",
			values: []any{true, false},
		},
		{
			name:   "int",
			values: []any{42, -42, 0},
		},
		{
			name:   "int8",
			values: []any{int8(127), int8(-128), int8(0)},
		},
		{
			name:   "int16",
			values: []any{int16(32767), int16(-32768), int16(0)},
		},
		{
			name:   "int32",
			values: []any{int32(2147483647), int32(-2147483648), int32(0)},
		},
		{
			name:   "int64",
			values: []any{int64(9223372036854775807), int64(-9223372036854775808), int64(0)},
		},
		{
			name:   "uint",
			values: []any{uint(42), uint(0)},
		},
		{
			name:   "uint8",
			values: []any{uint8(255), uint8(0)},
		},
		{
			name:   "uint16",
			values: []any{uint16(65535), uint16(0)},
		},
		{
			name:   "uint32",
			values: []any{uint32(4294967295), uint32(0)},
		},
		{
			name:   "uint64",
			values: []any{uint64(18446744073709551615), uint64(0)},
		},
		{
			name:   "uintptr",
			values: []any{uintptr(0x1234567890abcdef), uintptr(0)},
		},
		{
			name:   "byte",
			values: []any{byte(255), byte(0)},
		},
		{
			name:   "rune",
			values: []any{'A', '😀', rune(0)},
		},
		{
			name:   "float32",
			values: []any{float32(3.14159), float32(-3.14159), float32(0.0)},
		},
		{
			name:   "float64",
			values: []any{3.141592653589793, -3.141592653589793, 0.0},
		},
		{
			name:   "complex64",
			values: []any{complex64(complex(1.5, 2.5)), complex64(complex(0, 0))},
		},
		{
			name:   "complex128",
			values: []any{complex(1.5, 2.5), complex(0, 0)},
		},
		{
			name:   "string",
			values: []any{"hello world", "", "unicode: 你好 🌍"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, value := range tc.values {
				assertOfAny(t, value)
			}
		})
	}
}

func TestAddress(t *testing.T) {
	t.Run("same_value_different_addresses", func(t *testing.T) {
		value1 := 42
		value2 := 12
		addr1 := Address(&value1)
		addr2 := Address(&value2)

		assert.NotEqual(t, addr1, addr2, "expected different addresses for different variables with same value")
	})

	t.Run("same_pointer_same_address", func(t *testing.T) {
		value := 42

		addr1 := Address(&value)
		addr2 := Address(&value)

		assert.Equal(t, addr1, addr2, "expected same address for same pointer")
	})

	t.Run("slice_addresses", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		ref1 := slice[0:]
		ref2 := slice[1:]
		ref3 := slice[2:]
		addr1 := Address(&ref1)
		addr2 := Address(&ref2)
		addr3 := Address(&ref3)

		assert.NotEqual(t, addr1, addr2, "expected different addresses for slice[0] and slice[1]")
		assert.NotEqual(t, addr1, addr3, "expected different addresses for slice[0] and slice[2]")
		assert.NotEqual(t, addr2, addr3, "expected different addresses for slice[1] and slice[2]")

		assert.Equal(t, uintptr(8), addr2-addr1, "expected 8-byte difference between consecutive int addresses")
		assert.Equal(t, uintptr(8), addr3-addr2, "expected 8-byte difference between consecutive int addresses")
	})
}
