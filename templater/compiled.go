package templater

import "fmt"

// CompiledTemplate dynamically generated template
type CompiledTemplate[T any] struct {
	format     string
	extractors []Extractor[T]
}

// Execute executes the template with the given type
func (ct *CompiledTemplate[T]) Execute(input T) string {
	// No extractors, return the format string as-is
	if len(ct.extractors) == 0 {
		return ct.format
	}

	// Extract values from the input
	values := make([]any, len(ct.extractors))
	// Iterate over the extractors
	for i, extractor := range ct.extractors {
		// Extract the value
		values[i] = extractor(input)
	}

	// Format the string using the extracted values
	return fmt.Sprintf(ct.format, values...)
}
