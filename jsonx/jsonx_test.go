package jsonx

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testEntity implements the Entity interface for testing
type testEntity struct {
	handlerName *string
	value       *any
}

func (te *testEntity) OnFloat(v float64)         { *te.handlerName, *te.value = "OnFloat", v }
func (te *testEntity) OnString(v string)         { *te.handlerName, *te.value = "OnString", v }
func (te *testEntity) OnBool(v bool)             { *te.handlerName, *te.value = "OnBool", v }
func (te *testEntity) OnNull()                   { *te.handlerName, *te.value = "OnNull", nil }
func (te *testEntity) OnArray(v []any)           { *te.handlerName, *te.value = "OnArray", v }
func (te *testEntity) OnObject(v map[string]any) { *te.handlerName, *te.value = "OnObject", v }

func parseAndGetResult(t *testing.T, input []byte) (handlerName string, value any, err error) {
	t.Helper()
	handlerName = ""
	value = nil

	entity := &testEntity{
		handlerName: &handlerName,
		value:       &value,
	}

	err = HandleEntity(input, entity)
	return
}

func TestParseJSON_HandlesInteger(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte("123"))
	assert.NoError(t, err)
	if handler != "OnFloat" {
		t.Fatalf("Expected handler OnFloat, but got %s", handler)
	}

	if got := value.(float64); got != 123 {
		t.Errorf("Expected float 123, but got %v", got)
	}
}

func TestParseJSON_HandlesFloat(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte("99.9"))
	assert.NoError(t, err)

	if handler != "OnFloat" {
		t.Fatalf("Expected handler OnFloat, but got %s", handler)
	}

	if got := value.(float64); got != 99.9 {
		t.Errorf("Expected float 99.9, but got %v", got)
	}
}

func TestParseJSON_HandlesString(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte(`"hello"`))
	assert.NoError(t, err)

	if handler != "OnString" {
		t.Fatalf("Expected handler OnString, but got %s", handler)
	}

	if got := value.(string); got != "hello" {
		t.Errorf("Expected string 'hello', but got %q", got)
	}
}

func TestParseJSON_HandlesTrue(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte("true"))
	assert.NoError(t, err)

	if handler != "OnBool" {
		t.Fatalf("Expected handler OnBool, but got %s", handler)
	}

	if got := value.(bool); got != true {
		t.Errorf("Expected bool true, but got %v", got)
	}
}

func TestParseJSON_HandlesFalse(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte("false"))
	assert.NoError(t, err)

	if handler != "OnBool" {
		t.Fatalf("Expected handler OnBool, but got %s", handler)
	}

	if got := value.(bool); got != false {
		t.Errorf("Expected bool false, but got %v", got)
	}
}

func TestParseJSON_HandlesNull(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte("null"))
	assert.NoError(t, err)

	if handler != "OnNull" {
		t.Fatalf("Expected handler OnNull, but got %s", handler)
	}

	assert.Nil(t, value, "Expected nil, but got %v", value)
}

func TestParseJSON_HandlesArray(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte(`[1, "two", true]`))
	assert.NoError(t, err)

	if handler != "OnArray" {
		t.Fatalf("Expected handler OnArray, but got %s", handler)
	}

	expected := []any{float64(1), "two", true}
	if got := value.([]any); !reflect.DeepEqual(got, expected) {
		t.Errorf("Mismatched array value")
	}
}

func TestParseJSON_HandlesObject(t *testing.T) {
	handler, value, err := parseAndGetResult(t, []byte(`{"key":"val"}`))
	assert.NoError(t, err)

	if handler != "OnObject" {
		t.Fatalf("Expected handler OnObject, but got %s", handler)
	}

	expected := map[string]any{"key": "val"}
	if got := value.(map[string]any); !reflect.DeepEqual(got, expected) {
		t.Errorf("Mismatched object value")
	}
}

func TestParseJSON_HandlesMalformed(t *testing.T) {
	_, _, err := parseAndGetResult(t, []byte(`{invalid`))
	assert.Error(t, err)
}

func TestString_ReturnsJSONString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string value",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "integer value",
			input:    42,
			expected: "42",
		},
		{
			name:     "float value",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "null",
		},
		{
			name:     "array value",
			input:    []int{1, 2, 3},
			expected: "[1,2,3]",
		},
		{
			name:     "object value",
			input:    map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "empty array",
			input:    []int{},
			expected: "[]",
		},
		{
			name:     "empty object",
			input:    map[string]string{},
			expected: "{}",
		},
		{
			name:     "unmarshallable value",
			input:    make(chan int),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := String(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// testPrimitive implements the Primitive interface for testing
type testPrimitive struct {
	handlerName *string
	value       *any
}

func (tp *testPrimitive) OnValue(v any)   { *tp.handlerName, *tp.value = "OnValue", v }
func (tp *testPrimitive) OnNull()         { *tp.handlerName, *tp.value = "OnNull", nil }
func (tp *testPrimitive) OnComplex(v any) { *tp.handlerName, *tp.value = "OnComplex", v }

func parsePrimitiveAndGetResult(t *testing.T, input []byte) (handlerName string, value any, err error) {
	t.Helper()
	handlerName = ""
	value = nil

	primitive := &testPrimitive{
		handlerName: &handlerName,
		value:       &value,
	}

	err = HandlePrimitive(input, primitive)
	return
}

func TestHandleEntityValue_HandlesFloat(t *testing.T) {
	handlerName := ""
	value := any(nil)

	entity := &testEntity{
		handlerName: &handlerName,
		value:       &value,
	}

	HandleEntityValue(42.5, entity)

	if handlerName != "OnFloat" {
		t.Fatalf("Expected handler OnFloat, but got %s", handlerName)
	}

	if got := value.(float64); got != 42.5 {
		t.Errorf("Expected float 42.5, but got %v", got)
	}
}

func TestHandleEntityValue_HandlesString(t *testing.T) {
	handlerName := ""
	value := any(nil)

	entity := &testEntity{
		handlerName: &handlerName,
		value:       &value,
	}

	HandleEntityValue("test", entity)

	if handlerName != "OnString" {
		t.Fatalf("Expected handler OnString, but got %s", handlerName)
	}

	if got := value.(string); got != "test" {
		t.Errorf("Expected string 'test', but got %q", got)
	}
}

func TestHandleEntityValue_HandlesBool(t *testing.T) {
	handlerName := ""
	value := any(nil)

	entity := &testEntity{
		handlerName: &handlerName,
		value:       &value,
	}

	HandleEntityValue(true, entity)

	if handlerName != "OnBool" {
		t.Fatalf("Expected handler OnBool, but got %s", handlerName)
	}

	if got := value.(bool); got != true {
		t.Errorf("Expected bool true, but got %v", got)
	}
}

func TestHandleEntityValue_HandlesNull(t *testing.T) {
	handlerName := ""
	value := any(nil)

	entity := &testEntity{
		handlerName: &handlerName,
		value:       &value,
	}

	HandleEntityValue(nil, entity)

	if handlerName != "OnNull" {
		t.Fatalf("Expected handler OnNull, but got %s", handlerName)
	}

	assert.Nil(t, value, "Expected nil, but got %v", value)
}

func TestHandleEntityValue_HandlesArray(t *testing.T) {
	handlerName := ""
	value := any(nil)

	entity := &testEntity{
		handlerName: &handlerName,
		value:       &value,
	}

	testArray := []any{1.0, "test", true}
	HandleEntityValue(testArray, entity)

	if handlerName != "OnArray" {
		t.Fatalf("Expected handler OnArray, but got %s", handlerName)
	}

	if got := value.([]any); !reflect.DeepEqual(got, testArray) {
		t.Errorf("Mismatched array value")
	}
}

func TestHandleEntityValue_HandlesObject(t *testing.T) {
	handlerName := ""
	value := any(nil)

	entity := &testEntity{
		handlerName: &handlerName,
		value:       &value,
	}

	testObject := map[string]any{"key": "value"}
	HandleEntityValue(testObject, entity)

	if handlerName != "OnObject" {
		t.Fatalf("Expected handler OnObject, but got %s", handlerName)
	}

	if got := value.(map[string]any); !reflect.DeepEqual(got, testObject) {
		t.Errorf("Mismatched object value")
	}
}

func TestHandlePrimitive_HandlesString(t *testing.T) {
	handler, value, err := parsePrimitiveAndGetResult(t, []byte(`"hello"`))
	assert.NoError(t, err)

	if handler != "OnValue" {
		t.Fatalf("Expected handler OnValue, but got %s", handler)
	}

	if got := value.(string); got != "hello" {
		t.Errorf("Expected string 'hello', but got %q", got)
	}
}

func TestHandlePrimitive_HandlesFloat(t *testing.T) {
	handler, value, err := parsePrimitiveAndGetResult(t, []byte("123.45"))
	assert.NoError(t, err)

	if handler != "OnValue" {
		t.Fatalf("Expected handler OnValue, but got %s", handler)
	}

	if got := value.(float64); got != 123.45 {
		t.Errorf("Expected float 123.45, but got %v", got)
	}
}

func TestHandlePrimitive_HandlesBool(t *testing.T) {
	handler, value, err := parsePrimitiveAndGetResult(t, []byte("true"))
	assert.NoError(t, err)

	if handler != "OnValue" {
		t.Fatalf("Expected handler OnValue, but got %s", handler)
	}

	if got := value.(bool); got != true {
		t.Errorf("Expected bool true, but got %v", got)
	}
}

func TestHandlePrimitive_HandlesNull(t *testing.T) {
	handler, value, err := parsePrimitiveAndGetResult(t, []byte("null"))
	assert.NoError(t, err)

	if handler != "OnNull" {
		t.Fatalf("Expected handler OnNull, but got %s", handler)
	}

	assert.Nil(t, value, "Expected nil, but got %v", value)
}

func TestHandlePrimitive_HandlesArray(t *testing.T) {
	handler, value, err := parsePrimitiveAndGetResult(t, []byte(`[1, 2, 3]`))
	assert.NoError(t, err)

	if handler != "OnComplex" {
		t.Fatalf("Expected handler OnComplex, but got %s", handler)
	}

	expected := []any{float64(1), float64(2), float64(3)}
	if got := value.([]any); !reflect.DeepEqual(got, expected) {
		t.Errorf("Mismatched array value")
	}
}

func TestHandlePrimitive_HandlesObject(t *testing.T) {
	handler, value, err := parsePrimitiveAndGetResult(t, []byte(`{"key":"value"}`))
	assert.NoError(t, err)

	if handler != "OnComplex" {
		t.Fatalf("Expected handler OnComplex, but got %s", handler)
	}

	expected := map[string]any{"key": "value"}
	if got := value.(map[string]any); !reflect.DeepEqual(got, expected) {
		t.Errorf("Mismatched object value")
	}
}

func TestHandlePrimitive_HandlesMalformed(t *testing.T) {
	_, _, err := parsePrimitiveAndGetResult(t, []byte(`{invalid`))
	assert.Error(t, err)
}

func TestHandlePrimitiveValue_HandlesString(t *testing.T) {
	handlerName := ""
	value := any(nil)

	primitive := &testPrimitive{
		handlerName: &handlerName,
		value:       &value,
	}

	HandlePrimitiveValue("test", primitive)

	if handlerName != "OnValue" {
		t.Fatalf("Expected handler OnValue, but got %s", handlerName)
	}

	if got := value.(string); got != "test" {
		t.Errorf("Expected string 'test', but got %q", got)
	}
}

func TestHandlePrimitiveValue_HandlesFloat(t *testing.T) {
	handlerName := ""
	value := any(nil)

	primitive := &testPrimitive{
		handlerName: &handlerName,
		value:       &value,
	}

	HandlePrimitiveValue(42.5, primitive)

	if handlerName != "OnValue" {
		t.Fatalf("Expected handler OnValue, but got %s", handlerName)
	}

	if got := value.(float64); got != 42.5 {
		t.Errorf("Expected float 42.5, but got %v", got)
	}
}

func TestHandlePrimitiveValue_HandlesBool(t *testing.T) {
	handlerName := ""
	value := any(nil)

	primitive := &testPrimitive{
		handlerName: &handlerName,
		value:       &value,
	}

	HandlePrimitiveValue(true, primitive)

	if handlerName != "OnValue" {
		t.Fatalf("Expected handler OnValue, but got %s", handlerName)
	}

	if got := value.(bool); got != true {
		t.Errorf("Expected bool true, but got %v", got)
	}
}

func TestHandlePrimitiveValue_HandlesNull(t *testing.T) {
	handlerName := ""
	value := any(nil)

	primitive := &testPrimitive{
		handlerName: &handlerName,
		value:       &value,
	}

	HandlePrimitiveValue(nil, primitive)

	if handlerName != "OnNull" {
		t.Fatalf("Expected handler OnNull, but got %s", handlerName)
	}

	assert.Nil(t, value, "Expected nil, but got %v", value)
}

func TestHandlePrimitiveValue_HandlesArray(t *testing.T) {
	handlerName := ""
	value := any(nil)

	primitive := &testPrimitive{
		handlerName: &handlerName,
		value:       &value,
	}

	testArray := []any{1.0, "test", true}
	HandlePrimitiveValue(testArray, primitive)

	if handlerName != "OnComplex" {
		t.Fatalf("Expected handler OnComplex, but got %s", handlerName)
	}

	if got := value.([]any); !reflect.DeepEqual(got, testArray) {
		t.Errorf("Mismatched array value")
	}
}

func TestHandlePrimitiveValue_HandlesObject(t *testing.T) {
	handlerName := ""
	value := any(nil)

	primitive := &testPrimitive{
		handlerName: &handlerName,
		value:       &value,
	}

	testObject := map[string]any{"key": "value"}
	HandlePrimitiveValue(testObject, primitive)

	if handlerName != "OnComplex" {
		t.Fatalf("Expected handler OnComplex, but got %s", handlerName)
	}

	if got := value.(map[string]any); !reflect.DeepEqual(got, testObject) {
		t.Errorf("Mismatched object value")
	}
}

func TestExtractJsonFieldNames(t *testing.T) {
	type TestStruct struct {
		Name       string `json:"name"`
		Age        int    `json:"age"`
		Email      string `json:"email"`
		Private    string `json:"-"`
		NoTag      string
		unexported string `json:"unexported"`
	}

	expected := []string{"name", "age", "email"}
	result := ExtractJsonFieldNames(TestStruct{})

	assert.Equal(t, expected, result)
}

func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	t.Run("Basic struct conversion", func(t *testing.T) {
		input := TestStruct{
			Name:   "John",
			Age:    30,
			Active: true,
		}

		result, err := StructToMap(input)
		assert.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, float64(30), result["age"])
		assert.Equal(t, true, result["active"])
	})

	t.Run("Empty struct", func(t *testing.T) {
		input := TestStruct{}
		result, err := StructToMap(input)
		assert.NoError(t, err)
		assert.Equal(t, "", result["name"])
		assert.Equal(t, float64(0), result["age"])
		assert.Equal(t, false, result["active"])
	})

	t.Run("Nil input", func(t *testing.T) {
		result, err := StructToMap(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestToJSON(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	t.Run("No options - compact", func(t *testing.T) {
		input := TestStruct{Name: "Alice", Age: 25, Email: "alice@example.com"}
		var buf bytes.Buffer
		err := ToJSON(input, &buf)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"name":"Alice","age":25,"email":"alice@example.com"}`, buf.String())
	})

	t.Run("Pretty with indent", func(t *testing.T) {
		input := TestStruct{Name: "Bob", Age: 30, Email: "bob@example.com"}
		var buf bytes.Buffer
		err := ToJSON(input, &buf, Options{Pretty: true, Indent: "  "})
		assert.NoError(t, err)
		expected := "{\n  \"name\": \"Bob\",\n  \"age\": 30,\n  \"email\": \"bob@example.com\"\n}\n"
		assert.Equal(t, expected, buf.String())
	})

	t.Run("Custom prefix and indent", func(t *testing.T) {
		input := TestStruct{Name: "Charlie", Age: 35, Email: "charlie@example.com"}
		var buf bytes.Buffer
		err := ToJSON(input, &buf, Options{Pretty: true, Prefix: ">>", Indent: "\t"})
		assert.NoError(t, err)
		expected := "{\n>>\t\"name\": \"Charlie\",\n>>\t\"age\": 35,\n>>\t\"email\": \"charlie@example.com\"\n>>}\n"
		assert.Equal(t, expected, buf.String())
	})

	t.Run("Primitive types", func(t *testing.T) {
		var buf bytes.Buffer
		err := ToJSON(42, &buf)
		assert.NoError(t, err)
		assert.Equal(t, "42\n", buf.String())
	})

	t.Run("Array", func(t *testing.T) {
		var buf bytes.Buffer
		err := ToJSON([]int{1, 2, 3}, &buf)
		assert.NoError(t, err)
		assert.JSONEq(t, "[1,2,3]", buf.String())
	})
}

func TestToFile(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	t.Run("Write to file - compact", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test.json"
		input := TestStruct{Name: "Test", ID: 123}

		err := ToFile(input, tmpFile)
		assert.NoError(t, err)

		data, err := os.ReadFile(tmpFile)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"name":"Test","id":123}`, string(data))
	})

	t.Run("Write to file - pretty", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test_pretty.json"
		input := TestStruct{Name: "Pretty", ID: 456}

		err := ToFile(input, tmpFile, Options{Pretty: true, Indent: "  "})
		assert.NoError(t, err)

		data, err := os.ReadFile(tmpFile)
		assert.NoError(t, err)
		expected := "{\n  \"name\": \"Pretty\",\n  \"id\": 456\n}\n"
		assert.Equal(t, expected, string(data))
	})

	t.Run("Overwrite existing file", func(t *testing.T) {
		tmpFile := t.TempDir() + "/overwrite.json"

		// Write first data
		err := ToFile(TestStruct{Name: "First", ID: 1}, tmpFile)
		assert.NoError(t, err)

		// Overwrite with new data
		err = ToFile(TestStruct{Name: "Second", ID: 2}, tmpFile)
		assert.NoError(t, err)

		data, err := os.ReadFile(tmpFile)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"name":"Second","id":2}`, string(data))
	})

	t.Run("Invalid path", func(t *testing.T) {
		input := TestStruct{Name: "Test", ID: 123}
		err := ToFile(input, "/invalid/nonexistent/path/test.json")
		assert.Error(t, err)
	})
}

func TestToBytes(t *testing.T) {
	type TestStruct struct {
		Value string `json:"value"`
		Count int    `json:"count"`
	}

	t.Run("Compact format", func(t *testing.T) {
		input := TestStruct{Value: "test", Count: 5}
		result, err := ToBytes(input)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"value":"test","count":5}`, string(result))
	})

	t.Run("Pretty format", func(t *testing.T) {
		input := TestStruct{Value: "pretty", Count: 10}
		result, err := ToBytes(input, Options{Pretty: true, Indent: "  "})
		assert.NoError(t, err)
		expected := "{\n  \"value\": \"pretty\",\n  \"count\": 10\n}\n"
		assert.Equal(t, expected, string(result))
	})

	t.Run("Nil value", func(t *testing.T) {
		result, err := ToBytes[any](nil)
		assert.NoError(t, err)
		assert.Equal(t, "null\n", string(result))
	})

	t.Run("Empty slice", func(t *testing.T) {
		result, err := ToBytes([]int{})
		assert.NoError(t, err)
		assert.JSONEq(t, "[]", string(result))
	})
}

func TestFromJSON(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	t.Run("Valid JSON", func(t *testing.T) {
		jsonData := `{"name":"Alice","age":25,"email":"alice@example.com"}`
		reader := bytes.NewReader([]byte(jsonData))

		result, err := FromJSON[TestStruct](reader)
		assert.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, 25, result.Age)
		assert.Equal(t, "alice@example.com", result.Email)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		jsonData := `{"name":"Bob","age":invalid}`
		reader := bytes.NewReader([]byte(jsonData))

		_, err := FromJSON[TestStruct](reader)
		assert.Error(t, err)
	})

	t.Run("Primitive type", func(t *testing.T) {
		jsonData := `42`
		reader := bytes.NewReader([]byte(jsonData))

		result, err := FromJSON[int](reader)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("Array", func(t *testing.T) {
		jsonData := `[1,2,3,4,5]`
		reader := bytes.NewReader([]byte(jsonData))

		result, err := FromJSON[[]int](reader)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
	})
}

func TestFromFile(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	t.Run("Read valid file", func(t *testing.T) {
		tmpFile := t.TempDir() + "/read_test.json"
		jsonData := `{"name":"TestName","id":999}`
		err := os.WriteFile(tmpFile, []byte(jsonData), 0644)
		assert.NoError(t, err)

		result, err := FromFile[TestStruct](tmpFile)
		assert.NoError(t, err)
		assert.Equal(t, "TestName", result.Name)
		assert.Equal(t, 999, result.ID)
	})

	t.Run("File not found", func(t *testing.T) {
		_, err := FromFile[TestStruct]("/nonexistent/file.json")
		assert.Error(t, err)
	})

	t.Run("Invalid JSON in file", func(t *testing.T) {
		tmpFile := t.TempDir() + "/invalid.json"
		invalidJSON := `{"name":"Invalid","id":invalid}`
		err := os.WriteFile(tmpFile, []byte(invalidJSON), 0644)
		assert.NoError(t, err)

		_, err = FromFile[TestStruct](tmpFile)
		assert.Error(t, err)
	})

	t.Run("Empty file", func(t *testing.T) {
		tmpFile := t.TempDir() + "/empty.json"
		err := os.WriteFile(tmpFile, []byte(""), 0644)
		assert.NoError(t, err)

		_, err = FromFile[TestStruct](tmpFile)
		assert.Error(t, err)
	})
}

func TestFromBytes(t *testing.T) {
	type TestStruct struct {
		Value string `json:"value"`
		Count int    `json:"count"`
	}

	t.Run("Valid bytes", func(t *testing.T) {
		jsonBytes := []byte(`{"value":"test","count":5}`)
		result, err := FromBytes[TestStruct](jsonBytes)
		assert.NoError(t, err)
		assert.Equal(t, "test", result.Value)
		assert.Equal(t, 5, result.Count)
	})

	t.Run("Invalid bytes", func(t *testing.T) {
		invalidBytes := []byte(`{"value":"test","count":invalid}`)
		_, err := FromBytes[TestStruct](invalidBytes)
		assert.Error(t, err)
	})

	t.Run("Empty bytes", func(t *testing.T) {
		_, err := FromBytes[TestStruct]([]byte{})
		assert.Error(t, err)
	})

	t.Run("Primitive type", func(t *testing.T) {
		jsonBytes := []byte(`"hello"`)
		result, err := FromBytes[string](jsonBytes)
		assert.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("Array type", func(t *testing.T) {
		jsonBytes := []byte(`[1,2,3]`)
		result, err := FromBytes[[]int](jsonBytes)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, result)
	})
}

func TestRoundTrip(t *testing.T) {
	type ComplexStruct struct {
		Name     string            `json:"name"`
		Age      int               `json:"age"`
		Tags     []string          `json:"tags"`
		Metadata map[string]string `json:"metadata"`
	}

	t.Run("ToBytes and FromBytes", func(t *testing.T) {
		original := ComplexStruct{
			Name: "RoundTrip",
			Age:  42,
			Tags: []string{"tag1", "tag2", "tag3"},
			Metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}

		// Encode to bytes
		jsonBytes, err := ToBytes(original)
		assert.NoError(t, err)

		// Decode from bytes
		decoded, err := FromBytes[ComplexStruct](jsonBytes)
		assert.NoError(t, err)

		assert.Equal(t, original.Name, decoded.Name)
		assert.Equal(t, original.Age, decoded.Age)
		assert.Equal(t, original.Tags, decoded.Tags)
		assert.Equal(t, original.Metadata, decoded.Metadata)
	})

	t.Run("ToFile and FromFile", func(t *testing.T) {
		original := ComplexStruct{
			Name: "FileRoundTrip",
			Age:  99,
			Tags: []string{"a", "b"},
			Metadata: map[string]string{
				"x": "y",
			},
		}

		tmpFile := t.TempDir() + "/roundtrip.json"

		// Write to file
		err := ToFile(original, tmpFile, Options{Pretty: true, Indent: "  "})
		assert.NoError(t, err)

		// Read from file
		decoded, err := FromFile[ComplexStruct](tmpFile)
		assert.NoError(t, err)

		assert.Equal(t, original.Name, decoded.Name)
		assert.Equal(t, original.Age, decoded.Age)
		assert.Equal(t, original.Tags, decoded.Tags)
		assert.Equal(t, original.Metadata, decoded.Metadata)
	})
}
