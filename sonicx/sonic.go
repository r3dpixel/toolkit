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
	// Create an ordered map to store the results
	values := orderedmap.NewOrderedMap[T, struct{}]()

	// Check if the node is valid
	if node == nil || !node.Valid() {
		return values
	}

	// Load the node
	_ = node.Load()

	// Load the number of items in the array
	length, err := node.Len()
	if err != nil || length == 0 {
		return values
	}

	// Iterate over the items in the array and add them to the ordered map
	for i := 0; i < length; i++ {
		// Get the item
		if item := node.Index(i); item.Valid() {
			// Extract the value using the extractor func
			value := extractor(item)
			// InsertIter the value to the ordered map if it passes the filter
			if filter == nil || filter(value) {
				// InsertIter the value to the ordered map
				values.Set(value, structx.Empty)
			}
		}
	}

	// Return the ordered map
	return values
}

// ArrayToSlice returns a slice of the results that pass the filter from the sonic node array, projected using the extractor func
func ArrayToSlice[T any](node *Wrap, filter func(T) bool, extractor func(*Wrap) T) []T {
	// Check if the node is valid and has any items
	if node == nil || !node.Valid() {
		return nil
	}

	// Load the node
	_ = node.Load()

	// Load the number of items in the array
	length, err := node.Len()
	if err != nil || length == 0 {
		return nil
	}

	// Create a slice to store the results
	values := make([]T, 0, length)

	// Iterate over the items in the array and add them to the slice
	for i := 0; i < length; i++ {
		// Get the item
		if item := node.Index(i); item.Valid() {
			// Extract the value using the extractor func
			value := extractor(item)
			// InsertIter the value to the slice if it passes the filter
			if filter == nil || filter(value) {
				values = append(values, value)
			}
		}
	}

	// Return the slice
	return values
}
