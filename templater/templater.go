package templater

import (
	"strings"

	"github.com/r3dpixel/toolkit/iterx"
	"github.com/r3dpixel/toolkit/lexer"
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
	return ""
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

// Templater generic template engine
type Templater[T any] struct {
	lex *lexer.Lexer[rune, Extractor[T]]
}

// New creates a new Templater instance
func New[T any](tokens ...Token[T]) *Templater[T] {
	// Initialize the lexer
	lex := lexer.New[rune, Extractor[T]]()

	// Build the lexer from tokens
	for _, token := range tokens {
		lex.InsertIter(iterx.Runes(token.GetKey()), token.GetExtractor())
	}

	// Return the new Templater instance
	return &Templater[T]{
		lex: lex,
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
		extractor, matchLen, ok := t.lex.LongestMatchSlice(runes[index:])

		// Check if a match was found
		if ok {
			// Found a match starting at index
			format.WriteString("%s")
			extractors = append(extractors, extractor)
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
