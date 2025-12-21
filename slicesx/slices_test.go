package slicesx

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type propertyTestCase[T, V any] struct {
	name     string
	input    T
	expected V
}

type growInput struct {
	slice     []int
	lastIndex int
}
type growResult struct {
	expectedLen    int
	expectedCapGTE int
}

func TestGrow(t *testing.T) {
	testCases := []propertyTestCase[growInput, growResult]{
		{name: "Grow a nil slice", input: growInput{nil, 9}, expected: growResult{16, 16}},
		{name: "Grow an existing slice", input: growInput{make([]int, 5), 20}, expected: growResult{32, 32}},
		{name: "Do not grow if new length is smaller", input: growInput{make([]int, 16), 10}, expected: growResult{16, 16}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.input.slice
			Grow(&s, tc.input.lastIndex)
			assert.Len(t, s, tc.expected.expectedLen)
			assert.GreaterOrEqual(t, cap(s), tc.expected.expectedCapGTE)
		})
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	testCases := []propertyTestCase[int, int]{
		{name: "Input 0", input: 0, expected: 1},
		{name: "Input 3", input: 3, expected: 4},
		{name: "Input 9", input: 9, expected: 16},
		{name: "Input 1024", input: 1024, expected: 2048},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, NextPowerOfTwo(tc.input))
		})
	}
}

func TestPrependValue(t *testing.T) {
	testCases := []propertyTestCase[any, any]{
		{name: "Prepend int", input: struct {
			v int
			s []int
		}{1, []int{2, 3}}, expected: []int{1, 2, 3}},
		{name: "Prepend string", input: struct {
			v string
			s []string
		}{"a", []string{"b", "c"}}, expected: []string{"a", "b", "c"}},
		{name: "Prepend to nil slice", input: struct {
			v int
			s []int
		}{1, nil}, expected: []int{1}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch v := tc.input.(type) {
			case struct {
				v int
				s []int
			}:
				result := PrependValue(v.v, v.s)
				assert.Equal(t, tc.expected, result)
			case struct {
				v string
				s []string
			}:
				result := PrependValue(v.v, v.s)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestMap(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}

	testCases := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "should map integers to strings",
			test: func(t *testing.T) {
				input := []int{1, 2, 3, 4, 5}
				op := func(i int) string {
					return strconv.Itoa(i)
				}
				expected := []string{"1", "2", "3", "4", "5"}

				result := Map(input, op)

				if !reflect.DeepEqual(result, expected) {
					t.Errorf("Map() = %v, want %v", result, expected)
				}
			},
		},
		{
			name: "should map strings to their lengths (integers)",
			test: func(t *testing.T) {
				input := []string{"a", "bb", "ccc"}
				op := func(s string) int {
					return len(s)
				}
				expected := []int{1, 2, 3}

				result := Map(input, op)

				if !reflect.DeepEqual(result, expected) {
					t.Errorf("Map() = %v, want %v", result, expected)
				}
			},
		},
		{
			name: "should handle an empty slice",
			test: func(t *testing.T) {
				var input []int
				op := func(i int) int {
					// This function will never be called
					return i * 2
				}
				result := Map(input, op)

				// Check if the result is a non-nil empty slice
				if len(result) != 0 {
					t.Errorf("Map() on empty slice = %v, want non-nil empty slice []", result)
				}
			},
		},
		{
			name: "should handle a nil slice",
			test: func(t *testing.T) {
				var input []string = nil
				op := func(s string) bool {
					return s != ""
				}

				result := Map(input, op)

				if result != nil {
					t.Errorf("Map() on nil slice = %v, want nil", result)
				}
			},
		},
		{
			name: "should map a slice of structs to a slice of their fields",
			test: func(t *testing.T) {
				input := []user{
					{ID: 1, Name: "Alice"},
					{ID: 2, Name: "Bob"},
					{ID: 3, Name: "Charlie"},
				}
				op := func(u user) string {
					return u.Name
				}
				expected := []string{"Alice", "Bob", "Charlie"}

				result := Map(input, op)

				if !reflect.DeepEqual(result, expected) {
					t.Errorf("Map() = %v, want %v", result, expected)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func TestMapTo(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}

	testCases := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "should map integers to strings in existing slice",
			test: func(t *testing.T) {
				src := []int{1, 2, 3, 4, 5}
				dst := make([]string, len(src))
				op := func(i int) string {
					return strconv.Itoa(i)
				}
				expected := []string{"1", "2", "3", "4", "5"}

				MapTo(src, dst, op)

				if !reflect.DeepEqual(dst, expected) {
					t.Errorf("MapTo() result = %v, want %v", dst, expected)
				}
			},
		},
		{
			name: "should handle mismatched slice lengths - src longer",
			test: func(t *testing.T) {
				src := []int{1, 2, 3}
				dst := make([]string, 2) // Shorter length
				op := func(i int) string {
					return strconv.Itoa(i)
				}

				MapTo(src, dst, op)

				// Only first 2 elements should be mapped
				expected := []string{"1", "2"}
				if !reflect.DeepEqual(dst, expected) {
					t.Errorf("MapTo() with src longer = %v, want %v", dst, expected)
				}
			},
		},
		{
			name: "should handle mismatched slice lengths - dst longer",
			test: func(t *testing.T) {
				src := []int{1, 2}
				dst := make([]string, 4) // Longer length
				op := func(i int) string {
					return strconv.Itoa(i)
				}

				MapTo(src, dst, op)

				// Only first 2 elements should be mapped, rest remain zero values
				expected := []string{"1", "2", "", ""}
				if !reflect.DeepEqual(dst, expected) {
					t.Errorf("MapTo() with dst longer = %v, want %v", dst, expected)
				}
			},
		},
		{
			name: "should handle empty slices",
			test: func(t *testing.T) {
				var src []int
				var dst []string
				op := func(i int) string {
					return strconv.Itoa(i)
				}

				MapTo(src, dst, op)

				// Both should remain empty
				if len(dst) != 0 || len(src) != 0 {
					t.Errorf("MapTo() with empty slices failed")
				}
			},
		},
		{
			name: "should map structs to field values",
			test: func(t *testing.T) {
				src := []user{
					{ID: 1, Name: "Alice"},
					{ID: 2, Name: "Bob"},
				}
				dst := make([]string, len(src))
				op := func(u user) string {
					return u.Name
				}
				expected := []string{"Alice", "Bob"}

				MapTo(src, dst, op)

				if !reflect.DeepEqual(dst, expected) {
					t.Errorf("MapTo() with structs = %v, want %v", dst, expected)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func TestDeduplicateStableStrings(t *testing.T) {
	testCases := []struct {
		name     string
		strs1    []string
		strs2    []string
		expected []string
	}{
		{
			name:     "Empty slices",
			strs1:    []string{},
			strs2:    []string{},
			expected: []string{},
		},
		{
			name:     "First empty, second has values",
			strs1:    []string{},
			strs2:    []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "First has values, second empty",
			strs1:    []string{"x", "y"},
			strs2:    []string{},
			expected: []string{"x", "y"},
		},
		{
			name:     "No duplicates",
			strs1:    []string{"a", "b"},
			strs2:    []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "With duplicates - order preserved",
			strs1:    []string{"a", "b", "c"},
			strs2:    []string{"b", "d", "a"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "All duplicates",
			strs1:    []string{"x", "y"},
			strs2:    []string{"x", "y"},
			expected: []string{"x", "y"},
		},
		{
			name:     "Empty strings included",
			strs1:    []string{"", "a"},
			strs2:    []string{"b", ""},
			expected: []string{"", "a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DeduplicateStable(tc.strs1, tc.strs2)
			assert.Equal(t, tc.expected, result)
			assert.Len(t, result, len(tc.expected))
		})
	}
}

func TestDeduplicateStableFloats(t *testing.T) {
	testCases := []struct {
		name     string
		floats1  []float64
		floats2  []float64
		expected []float64
	}{
		{
			name:     "Empty slices",
			floats1:  []float64{},
			floats2:  []float64{},
			expected: []float64{},
		},
		{
			name:     "First empty, second has values",
			floats1:  []float64{},
			floats2:  []float64{1.5, 2.7},
			expected: []float64{1.5, 2.7},
		},
		{
			name:     "First has values, second empty",
			floats1:  []float64{3.14, 2.71},
			floats2:  []float64{},
			expected: []float64{3.14, 2.71},
		},
		{
			name:     "No duplicates",
			floats1:  []float64{1.1, 2.2},
			floats2:  []float64{3.3, 4.4},
			expected: []float64{1.1, 2.2, 3.3, 4.4},
		},
		{
			name:     "With duplicates - order preserved",
			floats1:  []float64{1.5, 2.5, 3.5},
			floats2:  []float64{2.5, 4.5, 1.5},
			expected: []float64{1.5, 2.5, 3.5, 4.5},
		},
		{
			name:     "All duplicates",
			floats1:  []float64{1.23, 4.56},
			floats2:  []float64{1.23, 4.56},
			expected: []float64{1.23, 4.56},
		},
		{
			name:     "Zero values included",
			floats1:  []float64{0.0, 1.1},
			floats2:  []float64{2.2, 0.0},
			expected: []float64{0.0, 1.1, 2.2},
		},
		{
			name:     "Negative values - order preserved",
			floats1:  []float64{-1.5, 2.5},
			floats2:  []float64{-2.5, 1.5},
			expected: []float64{-1.5, 2.5, -2.5, 1.5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DeduplicateStable(tc.floats1, tc.floats2)
			assert.Equal(t, tc.expected, result)
			assert.Len(t, result, len(tc.expected))
		})
	}
}
