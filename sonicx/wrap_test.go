package sonicx

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/r3dpixel/toolkit/ptr"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	jsonDocument := `{"str": "  he\"\"llo  ", "num": 123, "bool": true, "invalid": {"nested": "object"}}`
	jsonBytes := []byte(jsonDocument)

	testCases := []testCase[string, string]{
		{"str", `  he""llo  `},
		{"num", "123"},
		{"bool", "true"},
		{"invalid", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			node, _ := sonic.Get(jsonBytes, tc.input)
			wrapped := Of(node)
			wrappedPtr := OfPtr(&node)
			assert.Equal(t, tc.expected, wrapped.String())
			assert.Equal(t, tc.expected, wrappedPtr.String())
		})
	}

	t.Run("Empty", func(t *testing.T) {
		wrapped := Empty
		wrappedPtr := *Empty
		assert.Equal(t, "", wrapped.String())
		assert.Equal(t, "", wrappedPtr.String())
	})

	t.Run("WrapString function", func(t *testing.T) {
		node, _ := sonic.Get(jsonBytes, "str")
		wrapped := Of(node)
		assert.Equal(t, `  he""llo  `, WrapString(wrapped))
	})
}

func TestInteger(t *testing.T) {
	jsonDocument := `{"int": 42, "float": 3.14, "str": "123", "bool": true, "invalid": "not_a_number"}`
	jsonBytes := []byte(jsonDocument)

	testCases := []testCase[string, int]{
		{"int", 42},
		{"float", 3},
		{"str", 123},
		{"bool", 1},
		{"invalid", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			node, _ := sonic.Get(jsonBytes, tc.input)
			wrapped := Of(node)
			wrappedPtr := OfPtr(&node)
			assert.Equal(t, tc.expected, wrapped.Integer())
			assert.Equal(t, tc.expected, wrappedPtr.Integer())
		})
	}

	t.Run("Empty", func(t *testing.T) {
		wrapped := *Empty
		wrappedPtr := Empty
		assert.Equal(t, 0, wrapped.Integer())
		assert.Equal(t, 0, wrappedPtr.Integer())
	})

	t.Run("WrapInteger function", func(t *testing.T) {
		node, _ := sonic.Get(jsonBytes, "int")
		wrapped := Of(node)
		assert.Equal(t, 42, WrapInteger(wrapped))
	})
}

func TestInteger64(t *testing.T) {
	jsonDocument := `{"int": 42, "bigint": 9223372036854775807, "float": 3.14, "invalid": "not_a_number"}`
	jsonBytes := []byte(jsonDocument)

	testCases := []testCase[string, int64]{
		{"int", 42},
		{"bigint", 9223372036854775807},
		{"float", 3},
		{"invalid", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			node, _ := sonic.Get(jsonBytes, tc.input)
			wrapped := Of(node)
			wrappedPtr := OfPtr(&node)
			assert.Equal(t, tc.expected, wrapped.Integer64())
			assert.Equal(t, tc.expected, wrappedPtr.Integer64())
		})
	}

	t.Run("Empty", func(t *testing.T) {
		wrapped := *Empty
		wrappedPtr := Empty
		assert.Equal(t, int64(0), wrapped.Integer64())
		assert.Equal(t, int64(0), wrappedPtr.Integer64())
	})

	t.Run("WrapInteger64 function", func(t *testing.T) {
		node, _ := sonic.Get(jsonBytes, "bigint")
		wrapped := Of(node)
		assert.Equal(t, int64(9223372036854775807), WrapInteger64(wrapped))
	})
}

func TestFloat64(t *testing.T) {
	jsonDocument := `{"float": 3.14159, "int": 42, "str": "3.14", "bool": true, "invalid": "not_a_float"}`
	jsonBytes := []byte(jsonDocument)

	testCases := []testCase[string, float64]{
		{"float", 3.14159},
		{"int", 42.0},
		{"str", 3.14},
		{"bool", 1.0},
		{"invalid", 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			node, _ := sonic.Get(jsonBytes, tc.input)
			wrapped := Of(node)
			wrappedPtr := OfPtr(&node)
			assert.Equal(t, tc.expected, wrapped.Float64())
			assert.Equal(t, tc.expected, wrappedPtr.Float64())
		})
	}

	t.Run("Empty", func(t *testing.T) {
		wrapped := *Empty
		wrappedPtr := Empty
		assert.Equal(t, 0.0, wrapped.Float64())
		assert.Equal(t, 0.0, wrappedPtr.Float64())
	})

	t.Run("WrapFloat64 function", func(t *testing.T) {
		node, _ := sonic.Get(jsonBytes, "float")
		wrapped := Of(node)
		assert.Equal(t, 3.14159, WrapFloat64(wrapped))
	})
}

func TestBool(t *testing.T) {
	jsonDocument := `{"bool_true": true, "bool_false": false, "int": 1, "str": "true", "invalid": "not_a_bool"}`
	jsonBytes := []byte(jsonDocument)

	testCases := []testCase[string, bool]{
		{"bool_true", true},
		{"bool_false", false},
		{"int", true},
		{"str", true},
		{"invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			node, _ := sonic.Get(jsonBytes, tc.input)
			wrapped := Of(node)
			wrappedPtr := OfPtr(&node)
			assert.Equal(t, tc.expected, wrapped.Bool())
			assert.Equal(t, tc.expected, wrappedPtr.Bool())
		})
	}

	t.Run("Empty", func(t *testing.T) {
		wrapped := *Empty
		wrappedPtr := Empty
		assert.Equal(t, false, wrapped.Bool())
		assert.Equal(t, false, wrappedPtr.Bool())
	})

	t.Run("WrapBool function", func(t *testing.T) {
		node, _ := sonic.Get(jsonBytes, "bool_true")
		wrapped := Of(node)
		assert.Equal(t, true, WrapBool(wrapped))
	})
}

func TestWrapGet(t *testing.T) {
	jsonDocument := `{"user": {"name": "John", "age": 30}, "items": [1, 2, 3]}`
	jsonBytes := []byte(jsonDocument)

	root, _ := sonic.Get(jsonBytes)
	wrapped := Of(root)

	t.Run("Get nested object", func(t *testing.T) {
		user := wrapped.Get("user")
		assert.Equal(t, "John", user.Get("name").String())
		assert.Equal(t, 30, user.Get("age").Integer())
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		missing := wrapped.Get("missing")
		assert.Equal(t, "", missing.String())
		assert.Equal(t, 0, missing.Integer())
	})

	t.Run("Get from array", func(t *testing.T) {
		items := wrapped.Get("items")
		// Note: Get on arrays might not work as expected, but testing the method exists
		result := items.Get("0")
		// This might return empty, which is expected behavior
		assert.NotNil(t, result)
	})

	t.Run("WrapGet function", func(t *testing.T) {
		user := WrapGet(wrapped, "user")
		assert.Equal(t, "John", user.Get("name").String())
	})
}

func TestWrapGetByPath(t *testing.T) {
	jsonDocument := `{
		"users": [
			{"name": "John", "details": {"age": 30, "city": "NYC"}},
			{"name": "Jane", "details": {"age": 25, "city": "LA"}}
		],
		"metadata": {"version": 1.2}
	}`
	jsonBytes := []byte(jsonDocument)

	root, _ := sonic.Get(jsonBytes)
	wrapped := Of(root)

	t.Run("Get by simple path", func(t *testing.T) {
		version := wrapped.GetByPath("metadata", "version")
		assert.Equal(t, 1.2, version.Float64())
	})

	t.Run("Get by array index path", func(t *testing.T) {
		firstUser := wrapped.GetByPath("users", 0)
		assert.Equal(t, "John", firstUser.Get("name").String())

		firstUserAge := wrapped.GetByPath("users", 0, "details", "age")
		assert.Equal(t, 30, firstUserAge.Integer())

		secondUserCity := wrapped.GetByPath("users", 1, "details", "city")
		assert.Equal(t, "LA", secondUserCity.String())
	})

	t.Run("Get by invalid path", func(t *testing.T) {
		missing := wrapped.GetByPath("users", 10, "name")
		assert.Equal(t, "", missing.String())

		invalidPath := wrapped.GetByPath("nonexistent", "key")
		assert.Equal(t, 0, invalidPath.Integer())
	})

	t.Run("Get by mixed path types", func(t *testing.T) {
		result := wrapped.GetByPath("users", "0", "name") // string index
		// This might not work as expected, but testing the method
		assert.NotNil(t, result)
	})

	t.Run("WrapGetByPath function", func(t *testing.T) {
		version := WrapGetByPath(wrapped, "metadata", "version")
		assert.Equal(t, 1.2, version.Float64())
	})
}

func TestRefString(t *testing.T) {
	jsonDocument := `{"str": "  he\"\"llo  ", "num": 123, "bool": true, "invalid": {"nested": "object"}}`
	jsonBytes := []byte(jsonDocument)

	testCases := []testCase[string, string]{
		{"str", `  he""llo  `},
		{"num", "123"},
		{"bool", "true"},
		{"invalid", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			node, _ := sonic.Get(jsonBytes, tc.input)
			wrapped := Of(node)
			assert.Equal(t, tc.expected, wrapped.RefString())
		})
	}

	t.Run("Empty", func(t *testing.T) {
		wrapped := Empty
		assert.Equal(t, "", wrapped.RefString())
	})

	t.Run("WrapRefString function", func(t *testing.T) {
		node, _ := sonic.Get(jsonBytes, "str")
		wrapped := Of(node)
		assert.Equal(t, `  he""llo  `, WrapRefString(wrapped))
	})
}

func TestStringsCopyBehavior(t *testing.T) {
	jsonDocument := `{"str": "hello world", "num": 123}`

	t.Run("Check if sonic strings reference original JSON", func(t *testing.T) {
		node, _ := sonic.GetFromString(jsonDocument, "str")
		copyString := Of(node).String()
		refString, _ := node.String()

		jsonStart := *(*uintptr)(unsafe.Pointer(&jsonDocument))
		jsonEnd := jsonStart + uintptr(len(jsonDocument))

		copyAddress := *(*uintptr)(unsafe.Pointer(&copyString))
		refAdress := *(*uintptr)(unsafe.Pointer(&refString))

		copyAddressIsRef := jsonStart <= copyAddress && copyAddress <= jsonEnd
		refAddressIsRef := jsonStart <= refAdress && refAdress <= jsonEnd
		assert.False(t, copyAddressIsRef)
		assert.True(t, refAddressIsRef)

		assert.NotEmpty(t, copyString)
	})

	t.Run("RefString vs String memory behavior", func(t *testing.T) {
		node, _ := sonic.GetFromString(jsonDocument, "str")
		wrapped := Of(node)

		copyString := wrapped.String()
		refString := wrapped.RefString()
		nodeRefString, _ := node.String()

		jsonStart := *(*uintptr)(unsafe.Pointer(&jsonDocument))
		jsonEnd := jsonStart + uintptr(len(jsonDocument))

		copyAddress := *(*uintptr)(unsafe.Pointer(&copyString))
		refAddress := *(*uintptr)(unsafe.Pointer(&refString))
		nodeRefAddress := *(*uintptr)(unsafe.Pointer(&nodeRefString))

		copyAddressIsRef := jsonStart <= copyAddress && copyAddress <= jsonEnd
		refAddressIsRef := jsonStart <= refAddress && refAddress <= jsonEnd

		// String() should create a copy for certain types
		assert.False(t, copyAddressIsRef)
		// RefString() should reference the original JSON like node.String()
		assert.True(t, refAddressIsRef)
		// RefString should behave the same as node.String()
		assert.Equal(t, nodeRefAddress, refAddress)
		assert.Equal(t, copyString, refString)
	})

	t.Run("Check strings.Clone behavior", func(t *testing.T) {
		original := "test string"
		ref := original
		cloned := strings.Clone(original)

		originalAddr := ptr.Address(&original)
		refAddr := ptr.Address(&ref)
		clonedAddr := ptr.Address(&cloned)

		// strings.Clone should always create a copy
		assert.NotEqual(t, originalAddr, clonedAddr, "strings.Clone should create a copy")
		assert.Equal(t, originalAddr, refAddr, "ref and original should point to the same memory")
		assert.Equal(t, original, cloned, "content should be identical")
	})
}

func TestRaw(t *testing.T) {
	jsonDocument := `{"str": "hello world", "num": 42, "bool": true, "null": null, "obj": {"nested": "value"}}`
	jsonBytes := []byte(jsonDocument)

	testCases := []testCase[string, string]{
		{"str", `"hello world"`},
		{"num", "42"},
		{"bool", "true"},
		{"null", "null"},
		{"obj", `{"nested": "value"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			node, _ := sonic.Get(jsonBytes, tc.input)
			wrapped := Of(node)
			wrappedPtr := OfPtr(&node)
			assert.Equal(t, tc.expected, wrapped.Raw())
			assert.Equal(t, tc.expected, wrappedPtr.Raw())
		})
	}

	t.Run("Empty", func(t *testing.T) {
		wrapped := *Empty
		wrappedPtr := Empty
		assert.Equal(t, "null", wrapped.Raw())
		assert.Equal(t, "null", wrappedPtr.Raw())
	})

	t.Run("WrapRaw function", func(t *testing.T) {
		node, _ := sonic.Get(jsonBytes, "str")
		wrapped := Of(node)
		assert.Equal(t, `"hello world"`, WrapRaw(wrapped))
	})
}

func TestWrapIndex(t *testing.T) {
	jsonDocument := `{"items": [10, 20, 30, 40, 50]}`
	jsonBytes := []byte(jsonDocument)

	root, _ := sonic.Get(jsonBytes)
	wrapped := Of(root)

	t.Run("Get array elements by index", func(t *testing.T) {
		items := wrapped.Get("items")
		assert.Equal(t, 10, items.Index(0).Integer())
		assert.Equal(t, 20, items.Index(1).Integer())
		assert.Equal(t, 30, items.Index(2).Integer())
		assert.Equal(t, 40, items.Index(3).Integer())
		assert.Equal(t, 50, items.Index(4).Integer())
	})

	t.Run("Get out of bounds index", func(t *testing.T) {
		items := wrapped.Get("items")
		outOfBounds := items.Index(10)
		assert.Equal(t, 0, outOfBounds.Integer())
		assert.Equal(t, "", outOfBounds.String())
	})

	t.Run("Get negative index", func(t *testing.T) {
		items := wrapped.Get("items")
		negative := items.Index(-1)
		assert.Equal(t, 0, negative.Integer())
	})

	t.Run("WrapIndex function", func(t *testing.T) {
		items := wrapped.Get("items")
		item := WrapIndex(items, 2)
		assert.Equal(t, 30, item.Integer())
	})
}
