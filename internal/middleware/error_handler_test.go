package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/edaniel30/http-platform-go/internal/constants"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestErrorHandler(t *testing.T) {
	logger := NewDefaultLogger()

	t.Run("no errors - continues normally", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(logger))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("handles error added to context", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(logger))
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(NewNotFoundError("resource not found"))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
		var apiErr ApiError
		err := json.Unmarshal(w.Body.Bytes(), &apiErr)
		require.NoError(t, err)
		assert.Equal(t, "resource not found", apiErr.Message)
		assert.Equal(t, "NOT_FOUND", apiErr.Code)
	})

	t.Run("handles panic with error type", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(logger))
		router.GET("/test", func(c *gin.Context) {
			panic(errors.New("something went wrong"))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		var apiErr ApiError
		err := json.Unmarshal(w.Body.Bytes(), &apiErr)
		require.NoError(t, err)
		assert.Contains(t, apiErr.Message, "error")
	})

	t.Run("handles panic with non-error type", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(logger))
		router.GET("/test", func(c *gin.Context) {
			panic("string panic")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		assert.Equal(t, constants.ContentTypeJSON, w.Header().Get("Content-Type"))
		var apiErr ApiError
		err := json.Unmarshal(w.Body.Bytes(), &apiErr)
		require.NoError(t, err)
		assert.Equal(t, "Internal server error panic", apiErr.Message)
	})
}

func TestMapErrorToApiError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "NotFoundError",
			err:            NewNotFoundError("user not found"),
			expectedStatus: 404,
			expectedCode:   "NOT_FOUND",
			expectedMsg:    "user not found",
		},
		{
			name:           "UnauthorizedError",
			err:            NewUnauthorizedError("invalid token"),
			expectedStatus: 401,
			expectedCode:   "UNAUTHORIZED",
			expectedMsg:    "invalid token",
		},
		{
			name:           "ForbiddenError",
			err:            NewForbiddenError("access denied"),
			expectedStatus: 403,
			expectedCode:   "FORBIDDEN",
			expectedMsg:    "access denied",
		},
		{
			name:           "BadRequestError",
			err:            NewBadRequestError("invalid input"),
			expectedStatus: 400,
			expectedCode:   "BAD_REQUEST",
			expectedMsg:    "invalid input",
		},
		{
			name:           "ConflictError",
			err:            NewConflictError("already exists"),
			expectedStatus: 409,
			expectedCode:   "CONFLICT",
			expectedMsg:    "already exists",
		},
		{
			name:           "UnprocessableEntityError",
			err:            NewUnprocessableEntityError("validation failed"),
			expectedStatus: 422,
			expectedCode:   "UNPROCESSABLE_ENTITY",
			expectedMsg:    "validation failed",
		},
		{
			name:           "TooManyRequestsError",
			err:            NewTooManyRequestsError("rate limit exceeded"),
			expectedStatus: 429,
			expectedCode:   "TOO_MANY_REQUESTS",
			expectedMsg:    "rate limit exceeded",
		},
		{
			name:           "InternalServerError",
			err:            NewInternalServerError("internal error"),
			expectedStatus: 500,
			expectedCode:   "INTERNAL_SERVER",
			expectedMsg:    "internal error",
		},
		{
			name:           "ServiceUnavailableError",
			err:            NewServiceUnavailableError("service down"),
			expectedStatus: 503,
			expectedCode:   "SERVICE_UNAVAILABLE",
			expectedMsg:    "service down",
		},
		{
			name:           "ExternalServiceError",
			err:            &ExternalServiceError{Message: "external failed", StatusCode: 502},
			expectedStatus: 502,
			expectedCode:   "EXTERNAL_SERVICE",
			expectedMsg:    "external failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr, _ := mapErrorToApiError(tt.err)
			assert.Equal(t, tt.expectedStatus, apiErr.Status)
			assert.Equal(t, tt.expectedCode, apiErr.Code)
			assert.Equal(t, tt.expectedMsg, apiErr.Message)
		})
	}
}

func TestMapStandardError(t *testing.T) {
	t.Run("io.EOF", func(t *testing.T) {
		apiErr, _ := mapStandardError(io.EOF, Fields{})
		assert.Equal(t, 400, apiErr.Status)
		assert.Equal(t, "EMPTY_BODY", apiErr.Code)
		assert.Equal(t, "Request body is empty", apiErr.Message)
	})

	t.Run("io.ErrUnexpectedEOF", func(t *testing.T) {
		apiErr, _ := mapStandardError(io.ErrUnexpectedEOF, Fields{})
		assert.Equal(t, 400, apiErr.Status)
		assert.Equal(t, "INCOMPLETE_BODY", apiErr.Code)
		assert.Equal(t, "Request body is incomplete", apiErr.Message)
	})

	t.Run("context.Canceled", func(t *testing.T) {
		apiErr, fields := mapStandardError(context.Canceled, Fields{})
		assert.Equal(t, constants.StatusClientClosedRequest, apiErr.Status)
		assert.Equal(t, "REQUEST_CANCELED", apiErr.Code)
		assert.Equal(t, "Request was cancelled by client", apiErr.Message)
		assert.Equal(t, "context_canceled", fields["reason"])
	})

	t.Run("context.DeadlineExceeded", func(t *testing.T) {
		apiErr, fields := mapStandardError(context.DeadlineExceeded, Fields{})
		assert.Equal(t, 408, apiErr.Status)
		assert.Equal(t, "REQUEST_TIMEOUT", apiErr.Code)
		assert.Equal(t, "Request timeout exceeded", apiErr.Message)
		assert.Equal(t, "deadline_exceeded", fields["reason"])
	})

	t.Run("wrapped context.Canceled", func(t *testing.T) {
		wrappedErr := errors.New("wrapped: " + context.Canceled.Error())
		wrappedErr = errors.Join(wrappedErr, context.Canceled)
		apiErr, _ := mapStandardError(wrappedErr, Fields{})
		assert.Equal(t, constants.StatusClientClosedRequest, apiErr.Status)
	})

	t.Run("unknown error", func(t *testing.T) {
		unknownErr := errors.New("some random error")
		apiErr, fields := mapStandardError(unknownErr, Fields{})
		assert.Equal(t, 500, apiErr.Status)
		assert.Equal(t, "UNKNOWN", apiErr.Code)
		assert.Equal(t, "An error occurred", apiErr.Message)
		assert.NotEmpty(t, fields["full_error"])
	})
}

func TestJSONErrors(t *testing.T) {
	t.Run("json.UnmarshalTypeError", func(t *testing.T) {
		typeErr := &json.UnmarshalTypeError{
			Value: "string",
			Type:  nil,
			Field: "age",
		}
		// Set Type to int
		var dummy int
		typeErr.Type = reflect.TypeOf(dummy)

		apiErr, fields := mapErrorToApiError(typeErr)
		assert.Equal(t, 400, apiErr.Status)
		assert.Equal(t, "UNMARSHAL_TYPE", apiErr.Code)
		assert.Contains(t, apiErr.Message, "age")
		assert.Equal(t, "age", fields["field"])
	})

	t.Run("json.SyntaxError", func(t *testing.T) {
		// Create a real syntax error by parsing invalid JSON
		var result map[string]any
		invalidJSON := []byte(`{"key": invalid}`)
		err := json.Unmarshal(invalidJSON, &result)
		require.Error(t, err)

		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			apiErr, fields := mapErrorToApiError(syntaxErr)
			assert.Equal(t, 400, apiErr.Status)
			assert.Equal(t, "JSONSYNTAX", apiErr.Code) // toScreamingSnakeCase converts JSONSyntaxError to JSONSYNTAX
			assert.Contains(t, apiErr.Message, "Invalid JSON syntax")
			assert.NotZero(t, fields["offset"])
		}
	})
}

func TestValidationErrors(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=0,lte=120,min=18"`
	}

	validate := validator.New()

	t.Run("validation errors are properly mapped", func(t *testing.T) {
		test := TestStruct{
			Email: "invalid-email",
			Age:   -1,
		}

		err := validate.Struct(test)
		require.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		require.True(t, ok)

		apiErr, fields := mapErrorToApiError(validationErrs)
		assert.Equal(t, 400, apiErr.Status)
		assert.Equal(t, "VALIDATION", apiErr.Code)
		assert.Equal(t, "Validation error", apiErr.Message)
		assert.NotEmpty(t, fields["validation_errors"])
		assert.NotEmpty(t, apiErr.Cause)
	})

	t.Run("descriptive validation errors have field and reason", func(t *testing.T) {
		test := TestStruct{} // Empty struct will fail validations

		err := validate.Struct(test)
		require.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		require.True(t, ok)

		errs := descriptiveValidationErrors(validationErrs)
		assert.NotEmpty(t, errs)
		assert.Greater(t, len(errs), 0)

		// Check that errors have field and reason
		for _, e := range errs {
			assert.NotEmpty(t, e.Field)
			assert.NotEmpty(t, e.Reason)
		}
	})
}

func TestBuildLogFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?foo=bar", nil)
	c.Request.Header.Set("X-Trace-ID", "test-trace-id")

	// Set trace ID in context
	c.Set(TraceIDKey, "test-trace-id")

	fields := buildLogFields(c)
	assert.NotNil(t, fields)
	// Fields may contain trace_id if it was set
}
