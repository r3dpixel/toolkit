package trace

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

const (
	MODULE   string = "module"
	SERVICE  string = "service"
	ACTIVITY string = "activity"
	PACKAGE  string = "package"
	URL      string = "url"
	PATH     string = "path"
	AGENT    string = "agent"
	ID       string = "id"
	SOURCE   string = "source"
	ORIGIN   string = "origin"
	SIZE     string = "size"
	NAME     string = "name"
)

// ConsoleTraceWriter creates a zerolog console writer configured for trace output
func ConsoleTraceWriter() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:           os.Stderr,
		TimeFormat:    time.RFC3339,
		FieldsExclude: []string{zerolog.ErrorFieldName},
		FormatPrepare: formatPrepare,
		FormatExtra:   formatExtra,
	}
}

// formatPrepare prepares log fields by extracting trace fields into the main message
func formatPrepare(m map[string]interface{}) error {
	// Extract trace messages
	trace, ok := m[zerolog.ErrorFieldName].(map[string]any)
	if !ok {
		return nil
	}
	// Copy all non-trace fields to the main message
	for k, v := range trace {
		// Skip the trace field
		if k == ErrorTraceFieldName {
			continue
		}
		// InsertIter the field to the main message without overwriting existing fields
		// This will provide the first field found from TOP to BOTTOM precedence
		if _, duplicate := m[k]; !duplicate {
			m[k] = v
		}
	}

	// Return nil to indicate no error
	return nil
}

// formatExtra formats additional trace messages and appends them to the buffer
func formatExtra(m map[string]interface{}, buffer *bytes.Buffer) error {
	// Extract trace messages
	trace, ok := m[zerolog.ErrorFieldName].(map[string]any)
	if !ok {
		return nil
	}

	// Check if the type is a slice
	messages, ok := trace[ErrorTraceFieldName].([]any)
	if !ok {
		return nil
	}

	// Append each message to the buffer
	for _, rawMsg := range messages {
		if msg, ok := rawMsg.(string); ok {
			buffer.WriteString(fmt.Sprintf("\n\t%s", msg))
		}
	}

	// Return nil to indicate no error
	return nil
}
