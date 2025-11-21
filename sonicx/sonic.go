package sonicx

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/elliotchance/orderedmap/v3"
	"github.com/r3dpixel/toolkit/structx"
)

var Default = sonic.Config{
	NoNullSliceOrMap:        true,
	NoValidateJSONMarshaler: true,
	NoValidateJSONSkip:      true,
	CompactMarshaler:        true,
	CopyString:              true,
}.Froze()

var StableSort = sonic.Config{
	NoNullSliceOrMap:        true,
	NoValidateJSONMarshaler: true,
	NoValidateJSONSkip:      true,
	CompactMarshaler:        true,
	CopyString:              true,
	SortMapKeys:             true,
}.Froze()

var Config = Default

// GetFromString returns a Wrap for the node using the specified path
func GetFromString(src string, path ...any) (*Wrap, error) {
	node, err := sonic.GetFromString(src, path...)
	if err != nil {
		return nil, err
	}

	return Of(node), nil
}

// Get returns a Wrap for the node using the specified path
func Get(src []byte, path ...any) (*Wrap, error) {
	node, err := sonic.Get(src, path...)
	if err != nil {
		return nil, err
	}

	return Of(node), nil
}

// GetCopyFromString returns a Wrap for the node using the specified path
func GetCopyFromString(src string, path ...any) (*Wrap, error) {
	node, err := sonic.GetCopyFromString(src, path...)
	if err != nil {
		return nil, err
	}

	return Of(node), nil
}

// GetWithOptions returns a Wrap for the node using the specified path and options
func GetWithOptions(src []byte, opts ast.SearchOptions, path ...any) (*Wrap, error) {
	node, err := sonic.GetWithOptions(src, opts, path...)
	if err != nil {
		return nil, err
	}

	return Of(node), nil
}

// ArrayToMap returns an orderedmap.OrderedMap of the results that pass the filter from the sonic node array, projected using the extractor func
func ArrayToMap[T comparable](node *Wrap, filter func(T) bool, extractor func(*Wrap) T) *orderedmap.OrderedMap[T, struct{}] {
	values := orderedmap.NewOrderedMap[T, struct{}]()

	if node == nil || !node.Valid() {
		return values
	}
	_ = node.Load()
	length, err := node.Len()
	if err != nil || length == 0 {
		return values
	}

	for i := 0; i < length; i++ {
		if item := node.Index(i); item.Valid() {
			value := extractor(item)
			if filter == nil || filter(value) {
				values.Set(value, structx.Empty)
			}
		}
	}

	return values
}

// ArrayToSlice returns a slice of the results that pass the filter from the sonic node array, projected using the extractor func
func ArrayToSlice[T any](node *Wrap, filter func(T) bool, extractor func(*Wrap) T) []T {
	if node == nil || !node.Valid() {
		return nil
	}
	_ = node.Load()
	length, err := node.Len()
	if err != nil || length == 0 {
		return nil
	}
	values := make([]T, 0, length)

	for i := 0; i < length; i++ {
		if item := node.Index(i); item.Valid() {
			value := extractor(item)
			if filter == nil || filter(value) {
				values = append(values, value)
			}
		}
	}

	return values
}
