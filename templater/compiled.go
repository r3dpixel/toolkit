package templater

import "fmt"

// CompiledTemplate dynamically generated template
type CompiledTemplate[T any] struct {
	format     string
	extractors []Extractor[T]
}

// Execute executes the template with the given type
func (ct *CompiledTemplate[T]) Execute(input T) string {
	if len(ct.extractors) == 0 {
		return ct.format
	}

	values := make([]any, len(ct.extractors))
	for i, extractor := range ct.extractors {
		values[i] = extractor(input)
	}

	return fmt.Sprintf(ct.format, values...)
}
