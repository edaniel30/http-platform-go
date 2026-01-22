package httpplatform

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		expected string
	}{
		{
			name:     "with field",
			field:    "Port",
			message:  "Port must be between 1 and 65535",
			expected: "httpplatform: config error [Port]: Port must be between 1 and 65535",
		},
		{
			name:     "without field",
			field:    "",
			message:  "Invalid configuration",
			expected: "httpplatform: config error: Invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ConfigError{
				Field:   tt.field,
				Message: tt.message,
			}

			assert.Equal(t, tt.expected, err.Error())

			// Test errors.As compatibility
			var configErr *ConfigError
			assert.True(t, errors.As(err, &configErr))
			assert.Equal(t, tt.field, configErr.Field)
			assert.Equal(t, tt.message, configErr.Message)
		})
	}
}

func TestRuntimeError(t *testing.T) {
	causeErr := fmt.Errorf("database connection failed")

	tests := []struct {
		name     string
		message  string
		cause    error
		expected string
	}{
		{
			name:     "with cause",
			message:  "server failed to start",
			cause:    causeErr,
			expected: "httpplatform: runtime error: server failed to start: database connection failed",
		},
		{
			name:     "without cause",
			message:  "invalid state",
			cause:    nil,
			expected: "httpplatform: runtime error: invalid state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &RuntimeError{
				Message: tt.message,
				Cause:   tt.cause,
			}

			assert.Equal(t, tt.expected, err.Error())

			// Test errors.As compatibility
			var runtimeErr *RuntimeError
			assert.True(t, errors.As(err, &runtimeErr))
			assert.Equal(t, tt.message, runtimeErr.Message)
			assert.Equal(t, tt.cause, runtimeErr.Cause)

			// Test Unwrap
			assert.Equal(t, tt.cause, err.Unwrap())

			// Test errors.Is with cause
			if tt.cause != nil {
				assert.True(t, errors.Is(err, causeErr))
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("newConfigFieldError", func(t *testing.T) {
		err := newConfigFieldError("Logger", "Logger is required")

		var configErr *ConfigError
		assert.True(t, errors.As(err, &configErr))
		assert.Equal(t, "Logger", configErr.Field)
		assert.Equal(t, "Logger is required", configErr.Message)
		assert.Equal(t, "httpplatform: config error [Logger]: Logger is required", err.Error())
	})

	t.Run("newRuntimeError", func(t *testing.T) {
		causeErr := fmt.Errorf("connection refused")
		err := newRuntimeError("telemetry init failed", causeErr)

		var runtimeErr *RuntimeError
		assert.True(t, errors.As(err, &runtimeErr))
		assert.Equal(t, "telemetry init failed", runtimeErr.Message)
		assert.Equal(t, causeErr, runtimeErr.Cause)
		assert.True(t, errors.Is(err, causeErr))
	})
}
