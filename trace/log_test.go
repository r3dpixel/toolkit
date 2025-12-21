package trace

import (
	"bytes"
	"errors"
	"testing"

	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogErr(t *testing.T) {
	tests := []struct {
		name           string
		logMessage     string
		setupError     func() error
		expectedOutput []string
	}{
		{
			name:       "chained error logging",
			logMessage: "This is the log message",
			setupError: func() error {
				var testInt int
				root := sonicx.Config.UnmarshalFromString(`}`, &testInt)
				err1 := Error().Msg("Layer 1 trace").Field("layer", 1).Wrap(root)
				return Error().Msg("Layer 2 trace").Field("layer", 2).Wrap(err1)
			},
			expectedOutput: []string{
				"This is the log message",
				"layer=2",
				"\n\tLayer 2 trace",
				"\n\tLayer 1 trace",
				"\n\t\"Syntax error at index 0: invalid chars",
			},
		},
		{
			name:       "standard library error",
			logMessage: "Standard error occurred",
			setupError: func() error {
				return errors.New("something went wrong")
			},
			expectedOutput: []string{
				"Standard error occurred",
				"something went wrong",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, cleanup := setupLogger(t)
			defer cleanup()

			err := tt.setupError()
			require.Error(t, err)

			log.Error().Err(err).Msg(tt.logMessage)

			logOutput := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, logOutput, expected)
			}
		})
	}
}

func setupLogger(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()

	var buf bytes.Buffer
	originalLogger := log.Logger
	originalErrorMarshalFunc := zerolog.ErrorMarshalFunc

	zerolog.ErrorMarshalFunc = ErrorMarshalFunc
	writer := ConsoleTraceWriter()
	writer.Out = &buf
	writer.NoColor = true
	log.Logger = log.Logger.Output(writer)

	cleanup := func() {
		log.Logger = originalLogger
		zerolog.ErrorMarshalFunc = originalErrorMarshalFunc
	}

	return &buf, cleanup
}
