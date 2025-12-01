package httpplatform

import "github.com/gin-gonic/gin"

// QueryParamsToMap extracts all query parameters from the request and returns them as a map[string]any.
// If a parameter has multiple values, it returns a []string. Otherwise, it returns a single string value.
//
// Example:
//
//	// Request: GET /users?name=John&age=30&tags=go&tags=web
//	params := httpplatform.QueryParamsToMap(c)
//	// Result: map[string]any{
//	//   "name": "John",
//	//   "age": "30",
//	//   "tags": []string{"go", "web"}
//	// }
func QueryParamsToMap(c *gin.Context) map[string]any {
	queryParams := c.Request.URL.Query()
	result := make(map[string]any, len(queryParams))

	for key, values := range queryParams {
		if len(values) == 1 {
			// Single value: return as string
			result[key] = values[0]
		} else {
			// Multiple values: return as []string
			result[key] = values
		}
	}

	return result
}

// HeadersToMap extracts all request headers and returns them as a map[string]any.
// If a header has multiple values, it returns a []string. Otherwise, it returns a single string value.
// Header names are returned as-is (case-sensitive as received from the client).
//
// Example:
//
//	// Request headers:
//	// Content-Type: application/json
//	// Accept: application/json, text/plain
//	// X-Custom-Header: value1
//	// X-Custom-Header: value2
//	headers := httpplatform.HeadersToMap(c)
//	// Result: map[string]any{
//	//   "Content-Type": "application/json",
//	//   "Accept": []string{"application/json", "text/plain"},
//	//   "X-Custom-Header": []string{"value1", "value2"}
//	// }
func HeadersToMap(c *gin.Context) map[string]any {
	headers := c.Request.Header
	result := make(map[string]any, len(headers))

	for key, values := range headers {
		if len(values) == 1 {
			// Single value: return as string
			result[key] = values[0]
		} else {
			// Multiple values: return as []string
			result[key] = values
		}
	}

	return result
}
