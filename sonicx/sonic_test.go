package sonicx

import (
	"slices"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/r3dpixel/toolkit/ptr"
	"github.com/stretchr/testify/assert"
)

type testCase[T, V any] struct {
	input    T
	expected V
}

func TestArrayToSlice(t *testing.T) {
	jsonDocument := `{"nums": [10, 25, 30, 45], "strs": ["apple", "banana", "cherry"]}`
	jsonBytes := []byte(jsonDocument)

	numsArray, _ := sonic.Get(jsonBytes, "nums")
	strsArray, _ := sonic.Get(jsonBytes, "strs")

	t.Run("Extract all numbers", func(t *testing.T) {
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		result := ArrayToSlice(Of(numsArray), nil, extractor)
		assert.Equal(t, []int64{10, 25, 30, 45}, result)
	})

	t.Run("Extract numbers with filter", func(t *testing.T) {
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		filter := func(v int64) bool { return v >= 30 }
		result := ArrayToSlice(Of(numsArray), filter, extractor)
		assert.Equal(t, []int64{30, 45}, result)
	})

	t.Run("Extract all strings", func(t *testing.T) {
		extractor := func(node *Wrap) string { return node.String() }
		result := ArrayToSlice(Of(strsArray), nil, extractor)
		assert.Equal(t, []string{"apple", "banana", "cherry"}, result)
	})

	t.Run("Empty array", func(t *testing.T) {
		emptyArray := ast.NewArray([]ast.Node{})
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		result := ArrayToSlice(Of(emptyArray), nil, extractor)
		assert.Nil(t, result)
	})

	t.Run("Nil node", func(t *testing.T) {
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		result := ArrayToSlice(Empty, nil, extractor)
		assert.Nil(t, result)
	})
}

func TestArrayToMap(t *testing.T) {
	jsonDocument := `{"nums": [10, 25, 30, 45], "strs": ["apple", "banana", "cherry"]}`
	jsonBytes := []byte(jsonDocument)

	numsArray, _ := sonic.Get(jsonBytes, "nums")
	strsArray, _ := sonic.Get(jsonBytes, "strs")

	t.Run("Extract all numbers to map", func(t *testing.T) {
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		result := ArrayToMap(Of(numsArray), nil, extractor)

		assert.Equal(t, 4, result.Len())

		expectedKeys := []int64{10, 25, 30, 45}
		for _, key := range expectedKeys {
			_, exists := result.Get(key)
			assert.True(t, exists)
		}
	})

	t.Run("Extract numbers with filter to map", func(t *testing.T) {
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		filter := func(v int64) bool { return v >= 30 }
		result := ArrayToMap(Of(numsArray), filter, extractor)

		assert.Equal(t, 2, result.Len())

		expectedKeys := []int64{30, 45}
		for _, key := range expectedKeys {
			_, exists := result.Get(key)
			assert.True(t, exists)
		}
	})

	t.Run("Extract all strings to map", func(t *testing.T) {
		extractor := func(node *Wrap) string { return node.String() }
		result := ArrayToMap(Of(strsArray), nil, extractor)

		assert.Equal(t, 3, result.Len())

		expectedKeys := []string{"apple", "banana", "cherry"}
		for _, key := range expectedKeys {
			_, exists := result.Get(key)
			assert.True(t, exists)
		}
	})

	t.Run("Extract strings with filter to map", func(t *testing.T) {
		extractor := func(node *Wrap) string { return node.String() }
		filter := func(v string) bool { return len(v) > 5 }
		result := ArrayToMap(Of(strsArray), filter, extractor)

		assert.Equal(t, 2, result.Len())

		expectedKeys := []string{"banana", "cherry"}
		for _, key := range expectedKeys {
			_, exists := result.Get(key)
			assert.True(t, exists)
		}
	})

	t.Run("Empty array", func(t *testing.T) {
		emptyArray := ast.NewArray([]ast.Node{})
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		result := ArrayToMap(Of(emptyArray), nil, extractor)
		assert.Equal(t, 0, result.Len())
	})

	t.Run("Nil node", func(t *testing.T) {
		extractor := func(node *Wrap) int64 { return node.Integer64() }
		result := ArrayToMap(Empty, nil, extractor)
		assert.Equal(t, 0, result.Len())
	})
}

func TestArrayToMapOrder(t *testing.T) {
	jsonDocument := `{"items": ["first", "second", "third"]}`
	jsonBytes := []byte(jsonDocument)

	itemsArray, _ := sonic.Get(jsonBytes, "items")

	t.Run("Maintain insertion order", func(t *testing.T) {
		extractor := func(node *Wrap) string { return node.String() }
		result := ArrayToMap(Of(itemsArray), nil, extractor)

		assert.Equal(t, 3, result.Len())

		keys := make([]string, 0, result.Len())
		for key := range result.Keys() {
			keys = append(keys, key)
		}

		expected := []string{"first", "second", "third"}
		assert.True(t, slices.Equal(expected, keys))
	})
}

func TestUnmarshallCopyStruct(t *testing.T) {
	wrapper := struct {
		A string
		B string
		C int
	}{}

	jsonDocument := `{"A": "hello world", "B": "this is another string", "C": 123}`
	jsonRef := ptr.Address(&jsonDocument)

	err := sonic.UnmarshalString(jsonDocument, &wrapper)
	assert.NoError(t, err)
	aRef := ptr.Address(&wrapper.A)

	err = Default.UnmarshalFromString(jsonDocument, &wrapper)
	assert.NoError(t, err)
	aCopy := ptr.Address(&wrapper.A)

	assert.Equal(t, jsonRef+uintptr(7), aRef, "jsonDocument and ref should point to the same memory")
	assert.NotEqual(t, aCopy, aRef, "copy and ref should point to the different memory")
}

func TestUnmarshallCopyMap(t *testing.T) {
	var wrapper map[string]interface{}

	jsonDocument := `{"A": "hello world", "B": "this is another string", "C": 123}`
	jsonRef := ptr.Address(&jsonDocument)

	err := sonic.UnmarshalString(jsonDocument, &wrapper)
	assert.NoError(t, err)
	str := wrapper["A"].(string)
	aRef := ptr.Address(&str)

	err = Default.UnmarshalFromString(jsonDocument, &wrapper)
	assert.NoError(t, err)
	str = wrapper["A"].(string)
	aCopy := ptr.Address(&str)

	assert.Equal(t, jsonRef+uintptr(7), aRef, "jsonDocument and ref should point to the same memory")
	assert.NotEqual(t, aCopy, aRef, "copy and ref should point to the different memory")
}

func TestGetFromString(t *testing.T) {
	jsonDocument := `{"user": {"name": "Alice", "age": 30}, "items": [1, 2, 3]}`

	t.Run("Get root", func(t *testing.T) {
		wrapped, err := GetFromString(jsonDocument)
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "Alice", wrapped.Get("user").Get("name").String())
	})

	t.Run("Get nested path", func(t *testing.T) {
		wrapped, err := GetFromString(jsonDocument, "user", "name")
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "Alice", wrapped.String())
	})

	t.Run("Get array element", func(t *testing.T) {
		wrapped, err := GetFromString(jsonDocument, "items", 1)
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, 2, wrapped.Integer())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		wrapped, err := GetFromString(`{invalid}`)
		assert.Error(t, err)
		assert.Nil(t, wrapped)
	})

	t.Run("Non-existent path", func(t *testing.T) {
		wrapped, err := GetFromString(jsonDocument, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, wrapped)
	})
}

func TestGet(t *testing.T) {
	jsonDocument := []byte(`{"user": {"name": "Bob", "age": 25}, "items": [10, 20, 30]}`)

	t.Run("Get root", func(t *testing.T) {
		wrapped, err := Get(jsonDocument)
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "Bob", wrapped.Get("user").Get("name").String())
	})

	t.Run("Get nested path", func(t *testing.T) {
		wrapped, err := Get(jsonDocument, "user", "age")
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, 25, wrapped.Integer())
	})

	t.Run("Get array element", func(t *testing.T) {
		wrapped, err := Get(jsonDocument, "items", 2)
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, 30, wrapped.Integer())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		wrapped, err := Get([]byte(`{invalid}`))
		assert.Error(t, err)
		assert.Nil(t, wrapped)
	})
}

func TestGetCopyFromString(t *testing.T) {
	jsonDocument := `{"data": "test string"}`

	t.Run("Get with copy", func(t *testing.T) {
		wrapped, err := GetCopyFromString(jsonDocument, "data")
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "test string", wrapped.String())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		wrapped, err := GetCopyFromString(`{invalid}`)
		assert.Error(t, err)
		assert.Nil(t, wrapped)
	})
}

func TestGetWithOptions(t *testing.T) {
	jsonDocument := []byte(`{"user": {"name": "Charlie", "age": 35}}`)

	t.Run("Get with options", func(t *testing.T) {
		opts := ast.SearchOptions{}
		wrapped, err := GetWithOptions(jsonDocument, opts, "user", "name")
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "Charlie", wrapped.String())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		opts := ast.SearchOptions{}
		wrapped, err := GetWithOptions([]byte(`{invalid}`), opts)
		assert.NoError(t, err)
		assert.NotNil(t, wrapped)
	})
}
