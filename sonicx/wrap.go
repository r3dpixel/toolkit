package sonicx

import (
	"strings"

	"github.com/bytedance/sonic/ast"
)

// Empty - empty sonic node
var Empty = &Wrap{ast.NewNull()}

type Wrap struct {
	ast.Node
}

// Of wraps a Wrap as a Wrap
func Of(node ast.Node) *Wrap {
	return &Wrap{node}
}

// OfPtr wraps a Wrap pointer as a Wrap, returning empty node if nil
func OfPtr(node *ast.Node) *Wrap {
	if node == nil {
		return Empty
	}
	return &Wrap{*node}
}

// Get returns a Wrap for the specified key
func (w *Wrap) Get(key string) *Wrap {
	return OfPtr(w.Node.Get(key))
}

// GetByPath returns a Wrap for the specified path
func (w *Wrap) GetByPath(path ...any) *Wrap {
	return OfPtr(w.Node.GetByPath(path...))
}

// String extracts a string value from the Wrap (with copy)
func (w *Wrap) String() string {
	t := w.TypeSafe()
	raw, _ := w.Node.String()
	if t == ast.V_ANY || t == ast.V_STRING || t == ast.V_NUMBER {
		return strings.Clone(raw)
	}
	return raw
}

// RefString extracts a string value from the Wrap (without copy)
func (w *Wrap) RefString() string {
	raw, _ := w.Node.String()
	return raw
}

// Integer extracts an integer value from the Wrap
func (w *Wrap) Integer() int {
	raw, _ := w.Int64()
	return int(raw)
}

// Integer64 extracts an int64 value from the Wrap
func (w *Wrap) Integer64() int64 {
	raw, _ := w.Int64()
	return raw
}

// Float64 extracts a float64 value from the Wrap
func (w *Wrap) Float64() float64 {
	raw, _ := w.Node.Float64()
	return raw
}

// Bool extracts a boolean value from the Wrap
func (w *Wrap) Bool() bool {
	raw, _ := w.Node.Bool()
	return raw
}

// Raw extracts the raw string from the Wrap
func (w *Wrap) Raw() string {
	raw, _ := w.Node.Raw()
	return raw
}

// Index returns a Wrap for the specified index
func (w *Wrap) Index(i int) *Wrap {
	return OfPtr(w.Node.Index(i))
}

// WrapGet returns a Wrap for the specified key
func WrapGet(w *Wrap, key string) *Wrap {
	return w.Get(key)
}

// WrapGetByPath returns a Wrap for the specified path
func WrapGetByPath(w *Wrap, path ...any) *Wrap {
	return w.GetByPath(path...)
}

// WrapString extracts a string value from the Wrap (with copy)
func WrapString(w *Wrap) string {
	return w.String()
}

// WrapRefString extracts a string value from the Wrap (without copy)
func WrapRefString(w *Wrap) string {
	return w.RefString()
}

// WrapInteger extracts an integer value from the Wrap
func WrapInteger(w *Wrap) int {
	return w.Integer()
}

// WrapInteger64 extracts an int64 value from the Wrap
func WrapInteger64(w *Wrap) int64 {
	return w.Integer64()
}

// WrapFloat64 extracts a float64 value from the Wrap
func WrapFloat64(w *Wrap) float64 {
	return w.Float64()
}

// WrapBool extracts a boolean value from the Wrap
func WrapBool(w *Wrap) bool {
	return w.Bool()
}

// WrapRaw extracts the raw string from the Wrap
func WrapRaw(w *Wrap) string {
	return w.Raw()
}

// WrapIndex returns a Wrap for the specified index
func WrapIndex(w *Wrap, i int) *Wrap {
	return w.Index(i)
}
