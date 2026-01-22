package httpplatform

import "github.com/gin-gonic/gin"

// QueryParamsToMap extracts all query parameters from the request and returns them as a map[string]string.
// If a parameter has multiple values, only the first value is returned.
// For accessing all values of a multi-value parameter, use c.QueryArray() or c.Request.URL.Query() directly.
//
// Example:
//
//	// Request: GET /users?name=John&age=30&tags=go&tags=web
//	params := httpplatform.QueryParamsToMap(c)
//	// Result: map[string]string{
//	//   "name": "John",
//	//   "age": "30",
//	//   "tags": "go"  // First value only
//	// }
func QueryParamsToMap(c *gin.Context) map[string]string {
	queryParams := c.Request.URL.Query()
	result := make(map[string]string, len(queryParams))

	for key, values := range queryParams {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}

// HeadersToMap extracts all request headers and returns them as a map[string]string.
// If a header has multiple values, only the first value is returned.
// Header names are returned as-is (case-sensitive as received from the client).
// For accessing all values of a multi-value header, use c.Request.Header directly.
//
// Example:
//
//	// Request headers:
//	// Content-Type: application/json
//	// Accept: application/json, text/plain
//	// X-Custom-Header: value1
//	// X-Custom-Header: value2
//	headers := httpplatform.HeadersToMap(c)
//	// Result: map[string]string{
//	//   "Content-Type": "application/json",
//	//   "Accept": "application/json",  // First value only
//	//   "X-Custom-Header": "value1"     // First value only
//	// }
func HeadersToMap(c *gin.Context) map[string]string {
	headers := c.Request.Header
	result := make(map[string]string, len(headers))

	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}
