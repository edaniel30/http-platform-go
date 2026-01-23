package logger

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
		fields   map[string]any
		expected string
	}{
		{
			name:     "empty fields",
			fields:   map[string]any{},
			expected: "",
		},
		{
			name: "single field",
			fields: map[string]any{
				"key": "value",
			},
			expected: "key=value",
		},
		{
			name: "multiple fields sorted",
			fields: map[string]any{
				"status": 200,
				"method": "GET",
				"path":   "/users",
			},
			expected: "method=GET path=/users status=200",
		},
		{
			name: "fields with various types",
			fields: map[string]any{
				"string": "value",
				"int":    123,
				"bool":   true,
				"float":  45.67,
			},
			expected: "bool=true float=45.67 int=123 string=value",
		},
		{
			name: "fields sorted alphabetically",
			fields: map[string]any{
				"zebra": "z",
				"apple": "a",
				"mango": "m",
			},
			expected: "apple=a mango=m zebra=z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFields(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	testLogger := NewTestLogger(log.New(&buf, "[http-platform] ", 0)) // No timestamp for easier testing

	ctx := context.Background()

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		testLogger.Info(ctx, "test info message", map[string]any{"key": "value"})

		output := buf.String()
		assert.Contains(t, output, "INFO:")
		assert.Contains(t, output, "test info message")
		assert.Contains(t, output, "key=value")
	})

	t.Run("Error", func(t *testing.T) {
		buf.Reset()
		testLogger.Error(ctx, "test error message", map[string]any{"error": "something went wrong"})

		output := buf.String()
		assert.Contains(t, output, "ERROR:")
		assert.Contains(t, output, "test error message")
		assert.Contains(t, output, "error=something went wrong")
	})

	t.Run("Warn", func(t *testing.T) {
		buf.Reset()
		testLogger.Warn(ctx, "test warning message", map[string]any{"status": 404})

		output := buf.String()
		assert.Contains(t, output, "WARN:")
		assert.Contains(t, output, "test warning message")
		assert.Contains(t, output, "status=404")
	})

	t.Run("Debug", func(t *testing.T) {
		buf.Reset()
		testLogger.Debug(ctx, "test debug message", map[string]any{"trace_id": "abc-123"})

		output := buf.String()
		assert.Contains(t, output, "DEBUG:")
		assert.Contains(t, output, "test debug message")
		assert.Contains(t, output, "trace_id=abc-123")
	})

	t.Run("log without fields", func(t *testing.T) {
		buf.Reset()
		testLogger.Info(ctx, "message without fields", map[string]any{})

		output := buf.String()
		assert.Contains(t, output, "INFO:")
		assert.Contains(t, output, "message without fields")
		// Should not contain empty field formatting
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 1)
	})

	t.Run("Close", func(t *testing.T) {
		err := testLogger.Close()
		assert.NoError(t, err)
	})
}
