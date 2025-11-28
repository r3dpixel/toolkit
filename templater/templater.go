package templater

import (
	"strings"

	"github.com/r3dpixel/toolkit/stringsx"
)

// Extractor generic function to extract a value from a type
type Extractor[T any] = func(T) string

// Token generic element of a template
type Token[T any] interface {
	GetKey() string
	GetExtractor() Extractor[T]
	GetDescription() string
}

// BasicToken generic definition of a template element
type BasicToken[T any] struct {
	Key       string
	Extractor Extractor[T]
}

// GetKey returns the key of the token
func (t *BasicToken[T]) GetKey() string {
	return t.Key
}

// GetExtractor returns the extractor of the token
func (t *BasicToken[T]) GetExtractor() Extractor[T] {
	return t.Extractor
}

// GetDescription returns the description of the token (NO-OP)
func (t *BasicToken[T]) GetDescription() string {
	return stringsx.Empty
}

// RichToken generic element of a template with a description
type RichToken[T any] struct {
	BasicToken[T]
	Description string
}

// GetDescription returns the description of the token
func (t *RichToken[T]) GetDescription() string {
	return t.Description
}

// trieNode generic trie node
type trieNode[T any] struct {
	children  map[rune]*trieNode[T]
	extractor Extractor[T]
	isEnd     bool
}

// Templater generic template engine
type Templater[T any] struct {
	trie *trieNode[T]
}

// New creates a new Templater instance
func New[T any](tokens ...Token[T]) *Templater[T] {
	// Initialize the trie root
	root := &trieNode[T]{children: make(map[rune]*trieNode[T])}

	// Build the trie from tokens
	for _, token := range tokens {
		// Set the current node to root
		node := root

		// Iterate over the key of the token and add children nodes as needed
		for _, c := range token.GetKey() {
			// Create a new child node if needed
			if node.children[c] == nil {
				// Add the new child node, with the appropriate key
				node.children[c] = &trieNode[T]{children: make(map[rune]*trieNode[T])}
			}
			// Set the current node to the child node
			node = node.children[c]
		}
		// Mark the current node as the end of the token
		node.isEnd = true

		// Set the extractor of the current node
		node.extractor = token.GetExtractor()
	}

	// Return the new Templater instance
	return &Templater[T]{
		trie: root,
	}
}

// Compile compiles a template string into a CompiledTemplate
func (t *Templater[T]) Compile(template string) *CompiledTemplate[T] {
	// Compile the template into a format string and a list of extractors
	var format strings.Builder
	var extractors []Extractor[T]

	// Convert the template input to runes
	runes := []rune(template)
	index := 0

	// Iterate over the runes
	for index < len(runes) {
		// Try to find the longest match starting at the position index
		node := t.trie
		matchLen := 0
		var matchedExtractor Extractor[T]

		// Start searching for a match, from the current index
		for searchIndex := index; searchIndex < len(runes); searchIndex++ {
			// Get the current character
			c := runes[searchIndex]
			// Check if the character is part of the trie chain
			if node.children[c] == nil {
				break
			}
			// Move to the next node in the trie
			node = node.children[c]

			// Check if we reached the end of the trie chain
			if node.isEnd {
				// Found a match, remember it (but search for the longest match)
				matchLen = searchIndex - index + 1
				// Remember the extractor function
				matchedExtractor = node.extractor
			}
		}

		// Check if we found a match
		if matchLen > 0 {
			// Found a match starting at index
			format.WriteString("%s")
			extractors = append(extractors, matchedExtractor)
			// Move the index forward by the match length
			index += matchLen
		} else {
			// No match, emit the character and move forward
			format.WriteRune(runes[index])
			index++
		}
	}

	// Return the compiled template
	return &CompiledTemplate[T]{
		format:     format.String(),
		extractors: extractors,
	}
}

// Execute executes the template with the given type
func (t *Templater[T]) Execute(template string, input T) string {
	return t.Compile(template).Execute(input)
}
