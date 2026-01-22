package middleware

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   Fields
		expected string
	}{
		{
			name:     "empty fields",
			fields:   Fields{},
			expected: "",
		},
		{
			name: "single field",
			fields: Fields{
				"key": "value",
			},
			expected: "key=value",
		},
		{
			name: "multiple fields sorted",
			fields: Fields{
				"status": 200,
				"method": "GET",
				"path":   "/users",
			},
			expected: "method=GET path=/users status=200",
		},
		{
			name: "fields with various types",
			fields: Fields{
				"string": "value",
				"int":    123,
				"bool":   true,
				"float":  45.67,
			},
			expected: "bool=true float=45.67 int=123 string=value",
		},
		{
			name: "fields sorted alphabetically",
			fields: Fields{
				"zebra": "z",
				"apple": "a",
				"mango": "m",
			},
			expected: "apple=a mango=m zebra=z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFields(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := &DefaultLogger{
		logger: log.New(&buf, "[http-platform] ", 0), // No timestamp for easier testing
	}

	ctx := context.Background()

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "test info message", Fields{"key": "value"})

		output := buf.String()
		assert.Contains(t, output, "INFO:")
		assert.Contains(t, output, "test info message")
		assert.Contains(t, output, "key=value")
	})

	t.Run("Error", func(t *testing.T) {
		buf.Reset()
		logger.Error(ctx, "test error message", Fields{"error": "something went wrong"})

		output := buf.String()
		assert.Contains(t, output, "ERROR:")
		assert.Contains(t, output, "test error message")
		assert.Contains(t, output, "error=something went wrong")
	})

	t.Run("Warn", func(t *testing.T) {
		buf.Reset()
		logger.Warn(ctx, "test warning message", Fields{"status": 404})

		output := buf.String()
		assert.Contains(t, output, "WARN:")
		assert.Contains(t, output, "test warning message")
		assert.Contains(t, output, "status=404")
	})

	t.Run("Debug", func(t *testing.T) {
		buf.Reset()
		logger.Debug(ctx, "test debug message", Fields{"trace_id": "abc-123"})

		output := buf.String()
		assert.Contains(t, output, "DEBUG:")
		assert.Contains(t, output, "test debug message")
		assert.Contains(t, output, "trace_id=abc-123")
	})

	t.Run("log without fields", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "message without fields", Fields{})

		output := buf.String()
		assert.Contains(t, output, "INFO:")
		assert.Contains(t, output, "message without fields")
		// Should not contain empty field formatting
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 1)
	})

	t.Run("Close", func(t *testing.T) {
		err := logger.Close()
		assert.NoError(t, err)
	})
}
