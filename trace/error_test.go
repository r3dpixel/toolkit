package trace

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Chaining(t *testing.T) {
	t.Run("All methods should chain and set values correctly", func(t *testing.T) {
		// 1. Setup
		baseError := errors.New("root cause error")
		initialFields := map[string]any{
			"field1": "value1",
			"field2": 123,
		}

		// 2. Execution
		err := Error().
			Msg("this is the error message").
			Field("field3", true).
			Fields(initialFields).
			Wrap(baseError)

		// 3. Assertions
		assert.Equal(t, "this is the error message", err.msg)
		assert.Equal(t, "value1", err.fields["field1"])
		assert.Equal(t, 123, err.fields["field2"])
		assert.Equal(t, true, err.fields["field3"])
		assert.Equal(t, baseError, err.cause)
	})

	t.Run("Fields should overwrite existing keys", func(t *testing.T) {
		err := Error().
			Field("status", "initial").
			Fields(map[string]any{"status": "overwritten"})

		assert.Equal(t, "overwritten", err.fields["status"])
	})
}

func TestError_InterfaceContracts(t *testing.T) {
	t.Run("Implements error interface", func(t *testing.T) {
		var err error = Error().Msg("test message")
		assert.Equal(t, "test message", err.Error())
	})

	t.Run("Implements Wrapper interface for errors.Is and errors.Unwrap", func(t *testing.T) {
		baseError := errors.New("the root cause")
		wrappedError := Error().Wrap(baseError)

		// Test with errors.Unwrap
		unwrapped := errors.Unwrap(wrappedError)
		assert.Equal(t, baseError, unwrapped)

		// Test with errors.Is
		assert.True(t, errors.Is(wrappedError, baseError))
		assert.False(t, errors.Is(wrappedError, errors.New("some other error")))
	})

	t.Run("Unwrap returns nil if no cause is set", func(t *testing.T) {
		err := Error().Msg("no cause")
		assert.Nil(t, errors.Unwrap(err))
	})
}

func TestError_Msgf(t *testing.T) {
	err := Error().Msgf("user %s with id %d not found", "alice", 42)
	assert.Equal(t, "user alice with id 42 not found", err.Error())
}

func TestError_FieldOperations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Err
		field   string
		wantHas bool
		wantVal any
	}{
		{
			name:    "field exists in current error",
			setup:   func() *Err { return Error().Field("key", "value") },
			field:   "key",
			wantHas: true,
			wantVal: "value",
		},
		{
			name:    "field doesn't exist",
			setup:   func() *Err { return Error().Field("other", "val") },
			field:   "key",
			wantHas: false,
			wantVal: nil,
		},
		{
			name: "field exists in wrapped error",
			setup: func() *Err {
				inner := Error().Field("key", "inner_value")
				return Error().Wrap(inner)
			},
			field:   "key",
			wantHas: true,
			wantVal: "inner_value",
		},
		{
			name: "field exists in deeply nested chain",
			setup: func() *Err {
				deepest := Error().Field("deep", "value")
				middle := Error().Wrap(deepest)
				return Error().Wrap(middle)
			},
			field:   "deep",
			wantHas: true,
			wantVal: "value",
		},
		{
			name: "field shadowing - returns closest value",
			setup: func() *Err {
				inner := Error().Field("key", "inner")
				return Error().Field("key", "outer").Wrap(inner)
			},
			field:   "key",
			wantHas: true,
			wantVal: "outer",
		},
		{
			name: "works through non-Err wrapper",
			setup: func() *Err {
				inner := Error().Field("key", "value")
				wrapped := errors.Join(inner, errors.New("other"))
				return Error().Wrap(wrapped)
			},
			field:   "key",
			wantHas: true,
			wantVal: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.wantHas, err.HasField(tt.field))
			assert.Equal(t, tt.wantVal, err.GetField(tt.field))
		})
	}
}

func TestError_LazyAllocation(t *testing.T) {
	t.Run("no allocation when no fields added", func(t *testing.T) {
		err := Error().Msg("message only")
		assert.Nil(t, err.fields)
	})

	t.Run("allocates on first field", func(t *testing.T) {
		err := Error().Field("key", "value")
		assert.NotNil(t, err.fields)
		assert.Len(t, err.fields, 1)
	})

	t.Run("Fields with empty map doesn't allocate", func(t *testing.T) {
		err := Error().Fields(map[string]any{})
		assert.Nil(t, err.fields)
	})

	t.Run("Fields allocates with correct capacity", func(t *testing.T) {
		err := Error().Fields(map[string]any{"a": 1, "b": 2})
		assert.NotNil(t, err.fields)
		assert.Len(t, err.fields, 2)
	})
}

func TestCodedError(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *CodedErr[int]
		wantCode  int
		wantMsg   string
		wantField string
		wantVal   any
	}{
		{
			name: "basic coded error",
			setup: func() *CodedErr[int] {
				return CodedError[int]().Code(404).Msg("not found")
			},
			wantCode: 404,
			wantMsg:  "not found",
		},
		{
			name: "coded error with fields",
			setup: func() *CodedErr[int] {
				return CodedError[int]().Code(500).Field("component", "db").Msg("error")
			},
			wantCode:  500,
			wantMsg:   "error",
			wantField: "component",
			wantVal:   "db",
		},
		{
			name: "coded error with wrapping",
			setup: func() *CodedErr[int] {
				inner := Error().Field("inner_key", "inner_val")
				return CodedError[int]().Code(400).Wrap(inner).Msg("bad request")
			},
			wantCode:  400,
			wantMsg:   "bad request: ",
			wantField: "inner_key",
			wantVal:   "inner_val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.wantCode, err.GetCode())
			assert.Equal(t, tt.wantMsg, err.Error())
			if tt.wantField != "" {
				assert.Equal(t, tt.wantVal, err.GetField(tt.wantField))
			}
		})
	}
}

func TestCodedError_CustomTypes(t *testing.T) {
	type ErrCode string
	const (
		ErrNotFound     ErrCode = "NOT_FOUND"
		ErrUnauthorized ErrCode = "UNAUTHORIZED"
	)

	err := CodedError[ErrCode]().Code(ErrNotFound).Msg("resource not found")
	assert.Equal(t, ErrNotFound, err.GetCode())
	assert.Equal(t, "resource not found", err.Error())
}
