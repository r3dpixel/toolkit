package lexer

import (
	"iter"
	"slices"
)

// node represents a node in the trie
type node[T comparable, V any] struct {
	children map[T]*node[T, V]
	value    *V
}

// Lexer is a simple trie-based lexer
type Lexer[T comparable, V any] struct {
	root *node[T, V]
}

// New creates a new Lexer instance
func New[T comparable, V any]() *Lexer[T, V] {
	return &Lexer[T, V]{
		root: &node[T, V]{},
	}
}

// InsertSlice inserts a slice pattern into the trie with the associated value.
func (l *Lexer[T, V]) InsertSlice(pattern []T, value V) {
	l.InsertIter(slices.Values(pattern), value)
}

// InsertIter inserts a pattern into the trie with the associated value.
func (l *Lexer[T, V]) InsertIter(pattern iter.Seq[T], value V) {
	// Set the pointer to the root
	pointer := l.root

	// Iterate over the pattern
	for elem := range pattern {
		// Check if the element exists in the trie
		child, ok := pointer.children[elem]
		// If not, create a new node
		if !ok {
			// Create a new node
			child = &node[T, V]{}
			// Ensure the children map exists
			if pointer.children == nil {
				pointer.children = make(map[T]*node[T, V])
			}
			// Add the child to the trie
			pointer.children[elem] = child
		}
		// Set the pointer to the child
		pointer = child
	}
	// Set the value at the end of the pattern
	pointer.value = &value
}

// MatchSlice returns the value if the entire input exactly matches a pattern.
func (l *Lexer[T, V]) MatchSlice(input []T) (V, bool) {
	return l.Match(slices.Values(input))
}

// FirstMatchSlice returns the first (shortest) pattern that matches a prefix of the input.
func (l *Lexer[T, V]) FirstMatchSlice(input []T) (V, int, bool) {
	return l.FirstMatch(slices.Values(input))
}

// LongestMatchSlice returns the longest pattern that matches a prefix of the input.
func (l *Lexer[T, V]) LongestMatchSlice(input []T) (V, int, bool) {
	return l.LongestMatch(slices.Values(input))
}

// Match returns the value if the entire input exactly matches a pattern.
func (l *Lexer[T, V]) Match(input iter.Seq[T]) (V, bool) {
	var zero V

	//Set the pointer to the root
	pointer := l.root
	// Iterate over the input
	for elem := range input {
		// Check if the element exists in the trie
		child, ok := pointer.children[elem]
		// If not, return false
		if !ok {
			return zero, false
		}
		// Set the pointer to the child
		pointer = child
	}
	// Return the value if it exists
	if pointer.value != nil {
		return *pointer.value, true
	}
	// Return false if the value does not exist
	return zero, false
}

// FirstMatch returns the first (shortest) pattern that matches a prefix of the input.
// Returns the value, number of elements consumed, and whether a match was found.
func (l *Lexer[T, V]) FirstMatch(input iter.Seq[T]) (V, int, bool) {
	var zero V

	// Set the pointer to the root
	pointer := l.root
	consumed := 0
	// Iterate over the input
	for elem := range input {
		// Check if the element exists in the trie
		child, ok := pointer.children[elem]
		// If not, no match
		if !ok {
			return zero, 0, false
		}
		// Set the pointer to the child
		pointer = child
		consumed++
		// Return immediately if a value is found
		if pointer.value != nil {
			return *pointer.value, consumed, true
		}
	}
	// No match found
	return zero, 0, false
}

// LongestMatch returns the longest pattern that matches a prefix of the input.
// Returns the value, number of elements consumed, and whether a match was found.
func (l *Lexer[T, V]) LongestMatch(input iter.Seq[T]) (V, int, bool) {
	var zero V

	// Remember the last value found
	var lastValue *V
	var lastLen int

	// Set the pointer to the root
	pointer := l.root
	consumed := 0
	// Iterate over the input
	for elem := range input {
		// Check if the element exists in the trie
		child, ok := pointer.children[elem]
		// If not, stop traversing
		if !ok {
			break
		}
		// Set the pointer to the child
		pointer = child
		consumed++
		// Remember the last value found
		if pointer.value != nil {
			lastValue = pointer.value
			lastLen = consumed
		}
	}
	// Return the longest match if found
	if lastValue != nil {
		return *lastValue, lastLen, true
	}

	// No match found
	return zero, 0, false
}
