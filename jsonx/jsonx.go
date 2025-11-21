package jsonx

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/r3dpixel/toolkit/bytex"
	"github.com/r3dpixel/toolkit/filex"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
)

const (
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
	OnFloat(floatValue float64)
	OnString(stringValue string)
	OnBool(boolValue bool)
	OnNull()
	OnArray(arrayValue []any)
	OnObject(objectValue map[string]any)
}

// Primitive interface of a JSON primitive value
type Primitive interface {
	OnValue(value any)
	OnNull()
	OnComplex(complex any)
}

// HandleEntity parses the given JSON raw bytes using the handlers, according to its detected value type
func HandleEntity(data []byte, entity Entity) error {
	var value any
	if err := sonicx.Config.UnmarshalFromString(stringsx.FromBytes(data), &value); err != nil {
		return err
	}
	HandleEntityValue(value, entity)
	return nil
}

// HandleEntityValue parses the given JSON marshaled value using the handlers, according to its detected value type
func HandleEntityValue(value any, entity Entity) {
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
	var value any
	if err := sonicx.Config.UnmarshalFromString(stringsx.FromBytes(data), &value); err != nil {
		return err
	}
	HandlePrimitiveValue(value, primitive)
	return nil
}

// HandlePrimitiveValue parses the given JSON marshaled value using the handlers, according to its detected value type
func HandlePrimitiveValue(value any, primitive Primitive) {
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
	if str, err := sonicx.Config.MarshalToString(value); err == nil {
		return str
	}

	return stringsx.Empty
}

// ExtractJsonFieldNames extracts the JSON field names from the given value
func ExtractJsonFieldNames(v any) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if stringsx.IsBlank(jsonTag) || jsonTag == "-" {
			continue
		}

		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]

		if stringsx.IsNotBlank(fieldName) {
			fields = append(fields, fieldName)
		}
	}

	return fields
}

// StructToMap efficiently converts a struct to map[string]any using sonic
func StructToMap(v any) (map[string]any, error) {
	data, err := sonicx.Config.MarshalToString(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = sonicx.Config.UnmarshalFromString(data, &result)
	return result, err
}

// ToJSON encodes the item to JSON and writes it to w with optional formatting options
func ToJSON[T any](item T, w io.Writer, opts ...Options) error {
	enc := sonicx.Config.NewEncoder(w)

	if len(opts) > 0 {
		opt := opts[0]
		if opt.Pretty {
			enc.SetIndent(opt.Prefix, opt.Indent)
		}
	}

	return enc.Encode(item)
}

// ToFile encodes the item to JSON and writes it to a file at the path with optional formatting options
func ToFile[T any](item T, path string, opts ...Options) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filex.FilePermission)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, int(defaultBufferSizeIO))
	if err := ToJSON(item, writer, opts...); err != nil {
		return err
	}
	return writer.Flush()
}

// ToBytes encodes the item to JSON and returns it as a byte slice with optional formatting options
func ToBytes[T any](item T, opts ...Options) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, defaultBufferSizeJSON))
	if err := ToJSON(item, buf, opts...); err != nil {
		return nil, err
	}

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
	file, err := os.Open(path)
	if err != nil {
		var zero T
		return zero, err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, int(defaultBufferSizeIO))
	return FromJSON[T](reader)
}

// FromBytes decodes JSON from a byte slice into type [T]
func FromBytes[T any](b []byte) (T, error) {
	return FromJSON[T](bytes.NewReader(b))
}
