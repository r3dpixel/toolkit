package trace

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorMarshalFunc(t *testing.T) {
	t.Run("Single Traced Err", func(t *testing.T) {
		err := Error().Msg("database connection failed").Field("db_host", "localhost")
		result := ErrorMarshalFunc(err)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok, "Result should be a map")

		assert.Equal(t, "localhost", resultMap["db_host"])
		assert.Contains(t, resultMap[ErrorTraceFieldName], "database connection failed")
		assert.Len(t, resultMap[ErrorTraceFieldName], 1)
	})

	t.Run("Chained Traced Errors", func(t *testing.T) {
		err := Error().Msg("service layer error").Field("service", "user_service").
			Wrap(Error().Msg("repository error").Field("repo", "user_repo").Field("user_id", 123))

		result := ErrorMarshalFunc(err)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)

		assert.Equal(t, "user_service", resultMap["service"])
		assert.Equal(t, "user_repo", resultMap["repo"])
		assert.Equal(t, 123, resultMap["user_id"])

		expectedTrace := []string{"service layer error", "repository error"}
		assert.Equal(t, expectedTrace, resultMap[ErrorTraceFieldName])
	})

	t.Run("Mixed Chain with Standard Go Err", func(t *testing.T) {
		stdErr := errors.New("underlying network error")
		err := Error().Msg("request failed").Field("request_id", "xyz-123").Wrap(stdErr)

		result := ErrorMarshalFunc(err)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)

		assert.Equal(t, "xyz-123", resultMap["request_id"])

		expectedTrace := []string{"request failed", "underlying network error"}
		assert.Equal(t, expectedTrace, resultMap[ErrorTraceFieldName])
	})

	t.Run("Standard Go Err Only", func(t *testing.T) {
		stdErr := errors.New("a simple error")
		result := ErrorMarshalFunc(stdErr)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)

		expectedTrace := []string{"a simple error"}
		assert.Equal(t, expectedTrace, resultMap[ErrorTraceFieldName])
	})

	t.Run("Nil Err", func(t *testing.T) {
		result := ErrorMarshalFunc(nil)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)

		assert.Empty(t, resultMap[ErrorTraceFieldName])
		assert.Len(t, resultMap, 1, "Should only contain the empty _trace field")
	})

	t.Run("Field precedence in chain", func(t *testing.T) {
		err := Error().Msg("outer").Field("status", 500).
			Wrap(Error().Msg("inner").Field("status", 404).Field("user", "guest"))

		result := ErrorMarshalFunc(err)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)

		assert.Equal(t, 500, resultMap["status"])
		assert.Equal(t, "guest", resultMap["user"])
	})
}
