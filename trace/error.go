package trace

import (
	"errors"
	"fmt"
	"maps"

	"github.com/r3dpixel/toolkit/stringsx"
)

// Err containing a chain of causes (linked list of errors)
type Err struct {
	msg    string
	fields map[string]any
	cause  error
}

// Error creates a new Err instance
func Error() *Err {
	return &Err{}
}

// Err returns the error message, implementing the error interface
func (e *Err) Error() string {
	// If there is no cause, return the message directly
	if e.cause == nil {
		return e.msg
	}

	// If the message is blank, return the cause's error message'
	if stringsx.IsBlank(e.msg) {
		return e.cause.Error()
	}

	// Return the error message, and the cause's error message (flows down the chain)
	return e.msg + ": " + e.cause.Error()
}

// Unwrap returns the underlying cause error
func (e *Err) Unwrap() error {
	return e.cause
}

// Field adds a single key-value pair to the error's fields
func (e *Err) Field(key string, value any) *Err {
	if e.fields == nil {
		e.fields = make(map[string]any)
	}
	e.fields[key] = value
	return e
}

// HasField checks if the error has a field with the given key
func (e *Err) HasField(field string) bool {
	if _, hasField := e.fields[field]; hasField {
		return true
	}

	// Check down the error chain
	if e.cause != nil {
		var tracedErr *Err
		if errors.As(e.cause, &tracedErr) {
			return tracedErr.HasField(field)
		}
	}

	return false
}

// GetField returns the value of the field with the given key
func (e *Err) GetField(field string) any {
	if value, hasField := e.fields[field]; hasField {
		return value
	}

	// Check down the error chain
	if e.cause != nil {
		var tracedErr *Err
		if errors.As(e.cause, &tracedErr) {
			return tracedErr.GetField(field)
		}
	}

	return nil
}

// Fields copies all key-value pairs from the provided map to the error's fields
func (e *Err) Fields(f map[string]any) *Err {
	if len(f) == 0 {
		return e
	}
	if e.fields == nil {
		e.fields = make(map[string]any, len(f))
	}
	maps.Copy(e.fields, f)
	return e
}

// Msg sets the error message
func (e *Err) Msg(msg string) *Err {
	e.msg = msg
	return e
}

// Msgf sets the error message
func (e *Err) Msgf(msg string, args ...any) *Err {
	e.msg = fmt.Sprintf(msg, args...)
	return e
}

// Wrap sets the underlying cause error
func (e *Err) Wrap(cause error) *Err {
	e.cause = cause
	return e
}

// CodedErr is a generic error type that embeds Err and includes a typed code field
type CodedErr[T any] struct {
	Err
	code T
}

// CodedError creates a new CodedErr instance with the given code
func CodedError[T any]() *CodedErr[T] {
	return &CodedErr[T]{}
}

// Code sets the error code
func (e *CodedErr[T]) Code(code T) *CodedErr[T] {
	e.code = code
	return e
}

// GetCode returns the error code
func (e *CodedErr[T]) GetCode() T {
	return e.code
}

// Field adds a single key-value pair to the error's fields (overrides Err.Field to return *CodedErr[T])
func (e *CodedErr[T]) Field(key string, value any) *CodedErr[T] {
	e.Err.Field(key, value)
	return e
}

// Fields copies all key-value pairs from the provided map to the error's fields (overrides Err.Fields to return *CodedErr[T])
func (e *CodedErr[T]) Fields(f map[string]any) *CodedErr[T] {
	e.Err.Fields(f)
	return e
}

// Msg sets the error message (overrides Err.Msg to return *CodedErr[T])
func (e *CodedErr[T]) Msg(msg string) *CodedErr[T] {
	e.Err.Msg(msg)
	return e
}

// Msgf sets the error message (overrides Err.Msgf to return *CodedErr[T])
func (e *CodedErr[T]) Msgf(msg string, args ...any) *CodedErr[T] {
	e.Err.Msgf(msg, args...)
	return e
}

// Wrap sets the underlying cause error (overrides Err.Wrap to return *CodedErr[T])
func (e *CodedErr[T]) Wrap(cause error) *CodedErr[T] {
	e.Err.Wrap(cause)
	return e
}
