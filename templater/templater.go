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
	root := &trieNode[T]{children: make(map[rune]*trieNode[T])}

	// Build the trie from tokens
	for _, token := range tokens {
		node := root
		for _, c := range token.GetKey() {
			if node.children[c] == nil {
				node.children[c] = &trieNode[T]{children: make(map[rune]*trieNode[T])}
			}
			node = node.children[c]
		}
		node.isEnd = true
		node.extractor = token.GetExtractor()
	}

	return &Templater[T]{
		trie: root,
	}
}

// Compile compiles a template string into a CompiledTemplate
func (t *Templater[T]) Compile(template string) *CompiledTemplate[T] {
	var format strings.Builder
	var extractors []Extractor[T]

	runes := []rune(template)
	index := 0

	for index < len(runes) {
		// Try to find the longest match starting at position index
		node := t.trie
		matchLen := 0
		var matchedExtractor Extractor[T]

		for searchIndex := index; searchIndex < len(runes); searchIndex++ {
			c := runes[searchIndex]
			if node.children[c] == nil {
				break
			}
			node = node.children[c]
			if node.isEnd {
				// Found a match, remember it (but search for the longest match)
				matchLen = searchIndex - index + 1
				matchedExtractor = node.extractor
			}
		}

		if matchLen > 0 {
			// Found a match starting at index
			format.WriteString("%s")
			extractors = append(extractors, matchedExtractor)
			index += matchLen
		} else {
			// No match, emit the character and move forward
			format.WriteRune(runes[index])
			index++
		}
	}

	return &CompiledTemplate[T]{
		format:     format.String(),
		extractors: extractors,
	}
}

// Execute executes the template with the given type
func (t *Templater[T]) Execute(template string, input T) string {
	return t.Compile(template).Execute(input)
}
