package httpplatform

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestQueryParamsToMap(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected map[string]string
	}{
		{
			name:     "empty query params",
			url:      "/test",
			expected: map[string]string{},
		},
		{
			name: "single value params",
			url:  "/test?name=John&age=30&city=NYC",
			expected: map[string]string{
				"name": "John",
				"age":  "30",
				"city": "NYC",
			},
		},
		{
			name: "multi-value params - first value only",
			url:  "/test?tags=go&tags=web&tags=api&name=test",
			expected: map[string]string{
				"tags": "go", // Only first value
				"name": "test",
			},
		},
		{
			name: "params with special characters",
			url:  "/test?email=test%40example.com&message=hello%20world",
			expected: map[string]string{
				"email":   "test@example.com",
				"message": "hello world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.url, nil)

			result := QueryParamsToMap(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHeadersToMap(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected map[string]string
	}{
		{
			name:     "empty headers",
			headers:  http.Header{},
			expected: map[string]string{},
		},
		{
			name: "single value headers",
			headers: http.Header{
				"Content-Type":   []string{"application/json"},
				"Authorization":  []string{"Bearer token123"},
				"X-Request-Id":   []string{"abc-123"},
			},
			expected: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
				"X-Request-Id":  "abc-123",
			},
		},
		{
			name: "multi-value headers - first value only",
			headers: http.Header{
				"Accept":         []string{"application/json", "text/plain", "*/*"},
				"X-Custom":       []string{"value1", "value2"},
				"Content-Type":   []string{"application/json"},
			},
			expected: map[string]string{
				"Accept":       "application/json", // Only first value
				"X-Custom":     "value1",            // Only first value
				"Content-Type": "application/json",
			},
		},
		{
			name: "headers with various cases",
			headers: http.Header{
				"Content-Type":    []string{"text/html"},
				"content-length":  []string{"1234"},
				"X-CUSTOM-HEADER": []string{"test"},
			},
			expected: map[string]string{
				"Content-Type":    "text/html",
				"content-length":  "1234",
				"X-CUSTOM-HEADER": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			c.Request.Header = tt.headers

			result := HeadersToMap(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTraceID(t *testing.T) {
	t.Run("returns trace ID when present", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// Set trace ID in context
		expectedTraceID := "test-trace-123"
		c.Set("trace_id", expectedTraceID)

		result := GetTraceID(c)
		assert.Equal(t, expectedTraceID, result)
	})

	t.Run("returns empty string when trace ID not present", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		result := GetTraceID(c)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty string when trace ID has wrong type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// Set wrong type (int instead of string)
		c.Set("trace_id", 12345)

		result := GetTraceID(c)
		assert.Equal(t, "", result)
	})
}
