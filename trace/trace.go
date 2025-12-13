package trace

import (
	"errors"

	"github.com/r3dpixel/toolkit/stringsx"
)

var ErrorTraceFieldName = "_trace"

// ErrorMarshalFunc marshals an error chain into a map with trace information and fields
func ErrorMarshalFunc(err error) interface{} {
	var trace []string
	errMap := make(map[string]interface{}, 4)

	// Iterate through the chain of causes and unwrap next cause
	for cursor := err; cursor != nil; cursor = errors.Unwrap(cursor) {
		var tracedErr *Err
		// If the error is a trace error continue iteration
		if errors.As(cursor, &tracedErr) {
			// Append the current error message
			if !stringsx.IsBlank(tracedErr.msg) {
				trace = append(trace, tracedErr.msg)
			}
			// Extract the fields to the top level
			for key, val := range tracedErr.fields {
				if _, duplicate := errMap[key]; !duplicate {
					errMap[key] = val
				}
			}
		} else {
			// The end of the chain has been reached, append the last error message
			trace = append(trace, cursor.Error())
		}
	}

	// Set the assembled error trace
	errMap[ErrorTraceFieldName] = trace

	// Return the assembled error map
	return errMap
}
