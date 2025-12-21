package jsonx

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/r3dpixel/toolkit/bytex"
	"github.com/r3dpixel/toolkit/filex"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
)

const (
	DefaultIndent         = "  "
	Extension             = ".json"
	defaultBufferSizeIO   = 32 * bytex.KiB // 32KB buffer for file I/O operations
	defaultBufferSizeJSON = bytex.KiB      // Initial capacity for in-memory JSON encoding
)

// Options for JSON encoding and decoding
type Options struct {
	Pretty bool
	Prefix string
	Indent string
}

// Entity interface of an unknown JSON value
type Entity interface {
	// OnFloat Hook called when a float value is detected
	OnFloat(floatValue float64)
	// OnString Hook called when a string value is detected
	OnString(stringValue string)
	// OnBool Hook called when a boolean value is detected
	OnBool(boolValue bool)
	// OnNull Hook called when a null value is detected
	OnNull()
	// OnArray Hook called when an array value is detected
	OnArray(arrayValue []any)
	// OnObject Hook called when an object value is detected
	OnObject(objectValue map[string]any)
}

// Primitive interface of a JSON primitive value
type Primitive interface {
	// OnValue Hook called when a primitive value is detected
	OnValue(value any)
	// OnNull Hook called when a null value is detected
	OnNull()
	// OnComplex Hook called when a object or array value is detected
	OnComplex(complex any)
}

// HandleEntity parses the given JSON raw bytes using the handlers, according to its detected value type
func HandleEntity(data []byte, entity Entity) error {
	// Parse the JSON value
	var value any
	if err := sonicx.Config.UnmarshalFromString(stringsx.FromBytes(data), &value); err != nil {
		// Return the error
		return err
	}
	// Call the appropriate handler
	HandleEntityValue(value, entity)

	// Return nil (success)
	return nil
}

// HandleEntityValue parses the given JSON marshaled value using the handlers, according to its detected value type
func HandleEntityValue(value any, entity Entity) {
	// Dispatch to the appropriate callback
	switch v := value.(type) {
	case float64:
		entity.OnFloat(v)
	case string:
		entity.OnString(v)
	case bool:
		entity.OnBool(v)
	case nil:
		entity.OnNull()
	case []any:
		entity.OnArray(v)
	case map[string]any:
		entity.OnObject(v)
	}
}

// HandlePrimitive parses the given JSON raw bytes using the handlers, according to its detected value type
func HandlePrimitive(data []byte, primitive Primitive) error {
	// Parse the JSON value
	var value any
	if err := sonicx.Config.UnmarshalFromString(stringsx.FromBytes(data), &value); err != nil {
		// Return the error
		return err
	}
	// Call the appropriate handler
	HandlePrimitiveValue(value, primitive)

	// Return nil (success)
	return nil
}

// HandlePrimitiveValue parses the given JSON marshaled value using the handlers, according to its detected value type
func HandlePrimitiveValue(value any, primitive Primitive) {
	// Dispatch to the appropriate callback
	switch v := value.(type) {
	case string:
		primitive.OnValue(v)
	case float64, bool:
		primitive.OnValue(v)
	case nil:
		primitive.OnNull()
	case []any, map[string]any:
		primitive.OnComplex(v)
	}
}

// String returns the JSON marshaled value as a string
func String(value any) string {
	// Marshal the value to JSON string
	if str, err := sonicx.Config.MarshalToString(value); err == nil {
		return str
	}

	// Return an empty string if an error occurred
	return ""
}

// ExtractJsonFieldNames extracts the JSON field names from the given value
func ExtractJsonFieldNames(v any) []string {
	// Get the type of the value
	t := reflect.TypeOf(v)
	// If the value is a pointer, get the underlying type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Extract the field names from the JSON tags
	fields := make([]string, 0, t.NumField())

	// Iterate over the fields
	for i := 0; i < t.NumField(); i++ {
		// Get the field
		field := t.Field(i)
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		// Get the JSON tag
		jsonTag := field.Tag.Get("json")

		// Skip fields without a JSON tag or with a "-" tag
		if stringsx.IsBlank(jsonTag) || jsonTag == "-" {
			continue
		}

		// Extract the field name
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]

		// InsertIter the field name to the list
		if stringsx.IsNotBlank(fieldName) {
			fields = append(fields, fieldName)
		}
	}

	// Return the list of field names
	return fields
}

// StructToMap efficiently converts a struct to map[string]any using sonic
func StructToMap(v any) (map[string]any, error) {
	// Marshal the value to JSON string
	data, err := sonicx.Config.MarshalToString(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON string to a map[string]any
	var result map[string]any
	err = sonicx.Config.UnmarshalFromString(data, &result)

	// Return the map
	return result, err
}

// ToJSON encodes the item to JSON and writes it to w with optional formatting options
func ToJSON[T any](item T, w io.Writer, opts ...Options) error {
	// Create a new encoder with the specified options
	enc := sonicx.Config.NewEncoder(w)

	// Set the indentation options
	if len(opts) > 0 {
		configureEncoder(enc, opts[0])
	}

	// Encode the item
	return enc.Encode(item)
}

// configureEncoder configures the encoder with the specified options for pretty printing
func configureEncoder(enc sonic.Encoder, opts Options) {
	// Check if pretty printing is enabled
	if !opts.Pretty {
		return
	}

	// Set the indentation options
	indent := opts.Indent
	// Use the default indentation if none was specified
	if len(indent) == 0 {
		indent = DefaultIndent
	}

	// Configure the encoder for pretty printing
	enc.SetIndent(opts.Prefix, indent)
}

// ToFile encodes the item to JSON and writes it to a file at the path with optional formatting options
func ToFile[T any](item T, path string, opts ...Options) error {
	// Create the file
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filex.FilePermission)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the item to JSON and write it to the file buffered
	writer := bufio.NewWriterSize(file, int(defaultBufferSizeIO))
	if err := ToJSON(item, writer, opts...); err != nil {
		return err
	}

	// Flush the buffer
	return writer.Flush()
}

// ToBytes encodes the item to JSON and returns it as a byte slice with optional formatting options
func ToBytes[T any](item T, opts ...Options) ([]byte, error) {
	// Create a buffer for the JSON data
	buf := bytes.NewBuffer(make([]byte, 0, defaultBufferSizeJSON))

	// Encode the item to JSON and write it to the buffer
	if err := ToJSON(item, buf, opts...); err != nil {
		return nil, err
	}

	// Return the buffer bytes
	return buf.Bytes(), nil
}

// FromJSON decodes JSON from the input reader into type [T]
func FromJSON[T any](r io.Reader) (T, error) {
	var item T
	err := sonicx.Config.NewDecoder(r).Decode(&item)
	return item, err
}

// FromFile reads and decodes JSON from a file at the path into type [T]
func FromFile[T any](path string) (T, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		var zero T
		return zero, err
	}
	defer file.Close()

	// Decode the JSON from the file buffered
	reader := bufio.NewReaderSize(file, int(defaultBufferSizeIO))

	// Return the decoded item
	return FromJSON[T](reader)
}

// FromBytes decodes JSON from a byte slice into type [T]
func FromBytes[T any](b []byte) (T, error) {
	return FromJSON[T](bytes.NewReader(b))
}
