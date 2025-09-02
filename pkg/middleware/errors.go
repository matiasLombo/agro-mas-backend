package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	TraceID string                 `json:"trace_id,omitempty"`
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// BusinessLogicError represents domain-specific errors
type BusinessLogicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ErrorHandler provides centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Process any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Generate trace ID for debugging
			traceID := generateTraceID()
			
			// Log the error for monitoring
			logError(c, err, traceID)
			
			// Convert error to appropriate HTTP response
			statusCode, errorResponse := convertErrorToResponse(err, traceID)
			
			c.JSON(statusCode, errorResponse)
			return
		}
	}
}

// RecoveryHandler handles panics and converts them to errors
func RecoveryHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		traceID := generateTraceID()
		
		// Log panic details
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)
		
		logPanic(c, err, string(stack[:length]), traceID)
		
		// Return standardized error response
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred",
			TraceID: traceID,
		})
		
		c.Abort()
	})
}

// ValidationErrorHandler handles validation errors from Gin binding
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check for validation errors
		for _, err := range c.Errors {
			if err.Type == gin.ErrorTypeBind {
				validationErrors := extractValidationErrors(err)
				
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation failed",
					"code":    "VALIDATION_ERROR",
					"message": "The request contains invalid data",
					"details": gin.H{
						"validation_errors": validationErrors,
					},
				})
				return
			}
		}
	}
}

// DatabaseErrorHandler handles database-specific errors
func DatabaseErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		for _, err := range c.Errors {
			if isDatabaseError(err.Err) {
				statusCode, errorResponse := handleDatabaseError(err.Err)
				c.JSON(statusCode, errorResponse)
				return
			}
		}
	}
}

// RateLimitErrorHandler handles rate limiting errors
func RateLimitErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		for _, err := range c.Errors {
			if isRateLimitError(err.Err) {
				c.JSON(http.StatusTooManyRequests, ErrorResponse{
					Error:   "Rate limit exceeded",
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Too many requests. Please try again later.",
					Details: map[string]interface{}{
						"retry_after": "60s",
					},
				})
				return
			}
		}
	}
}

// Helper functions

func generateTraceID() string {
	// In a real implementation, you'd use a proper trace ID generation
	// This is a simplified version
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func logError(c *gin.Context, err *gin.Error, traceID string) {
	// In production, you'd use structured logging (e.g., logrus, zap)
	fmt.Printf("[ERROR] %s | %s %s | %s | TraceID: %s | Error: %v\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		c.Request.Method,
		c.Request.RequestURI,
		c.ClientIP(),
		traceID,
		err.Err,
	)
}

func logPanic(c *gin.Context, err interface{}, stack string, traceID string) {
	fmt.Printf("[PANIC] %s | %s %s | %s | TraceID: %s | Panic: %v\nStack:\n%s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		c.Request.Method,
		c.Request.RequestURI,
		c.ClientIP(),
		traceID,
		err,
		stack,
	)
}

func convertErrorToResponse(err *gin.Error, traceID string) (int, ErrorResponse) {
	switch {
	case isValidationError(err.Err):
		return http.StatusBadRequest, ErrorResponse{
			Error:   "Validation error",
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: traceID,
		}
	case isAuthenticationError(err.Err):
		return http.StatusUnauthorized, ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTHENTICATION_REQUIRED",
			Message: err.Error(),
			TraceID: traceID,
		}
	case isAuthorizationError(err.Err):
		return http.StatusForbidden, ErrorResponse{
			Error:   "Access forbidden",
			Code:    "ACCESS_FORBIDDEN",
			Message: err.Error(),
			TraceID: traceID,
		}
	case isNotFoundError(err.Err):
		return http.StatusNotFound, ErrorResponse{
			Error:   "Resource not found",
			Code:    "RESOURCE_NOT_FOUND",
			Message: err.Error(),
			TraceID: traceID,
		}
	case isConflictError(err.Err):
		return http.StatusConflict, ErrorResponse{
			Error:   "Resource conflict",
			Code:    "RESOURCE_CONFLICT",
			Message: err.Error(),
			TraceID: traceID,
		}
	case isDatabaseError(err.Err):
		_, response := handleDatabaseError(err.Err)
		response.TraceID = traceID
		return http.StatusInternalServerError, response
	default:
		return http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred",
			TraceID: traceID,
		}
	}
}

func extractValidationErrors(err *gin.Error) []ValidationError {
	// This is a simplified version - in practice you'd use a validation library
	// like go-playground/validator to extract detailed validation errors
	return []ValidationError{
		{
			Field:   "unknown",
			Message: err.Error(),
		},
	}
}

func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "sql") ||
		strings.Contains(errorStr, "database") ||
		strings.Contains(errorStr, "connection") ||
		strings.Contains(errorStr, "postgres")
}

func handleDatabaseError(err error) (int, ErrorResponse) {
	errorStr := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errorStr, "connection"):
		return http.StatusServiceUnavailable, ErrorResponse{
			Error:   "Database connection error",
			Code:    "DATABASE_CONNECTION_ERROR",
			Message: "Database is temporarily unavailable",
		}
	case strings.Contains(errorStr, "duplicate") || strings.Contains(errorStr, "unique"):
		return http.StatusConflict, ErrorResponse{
			Error:   "Duplicate entry",
			Code:    "DUPLICATE_ENTRY",
			Message: "The resource already exists",
		}
	case strings.Contains(errorStr, "foreign key"):
		return http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid reference",
			Code:    "INVALID_REFERENCE",
			Message: "Referenced resource does not exist",
		}
	case strings.Contains(errorStr, "not null"):
		return http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required field",
			Code:    "MISSING_REQUIRED_FIELD",
			Message: "A required field is missing",
		}
	default:
		return http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Message: "An error occurred while processing your request",
		}
	}
}

// Error type checking functions
func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "validation") ||
		strings.Contains(err.Error(), "invalid") ||
		strings.Contains(err.Error(), "required")
}

func isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "unauthorized") ||
		strings.Contains(errorStr, "authentication") ||
		strings.Contains(errorStr, "invalid token") ||
		strings.Contains(errorStr, "token expired")
}

func isAuthorizationError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "forbidden") ||
		strings.Contains(errorStr, "permission") ||
		strings.Contains(errorStr, "access denied") ||
		strings.Contains(errorStr, "insufficient")
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "not found") ||
		strings.Contains(errorStr, "does not exist") ||
		strings.Contains(errorStr, "no rows")
}

func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "already exists") ||
		strings.Contains(errorStr, "conflict") ||
		strings.Contains(errorStr, "duplicate")
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "rate limit") ||
		strings.Contains(errorStr, "too many requests")
}

// Business logic error helpers
func NewValidationError(field, message string, value interface{}) error {
	return fmt.Errorf("validation error on field %s: %s", field, message)
}

func NewNotFoundError(resource string) error {
	return fmt.Errorf("%s not found", resource)
}

func NewConflictError(resource string) error {
	return fmt.Errorf("%s already exists", resource)
}

func NewUnauthorizedError(message string) error {
	return fmt.Errorf("unauthorized: %s", message)
}

func NewForbiddenError(message string) error {
	return fmt.Errorf("forbidden: %s", message)
}

// Error constants for common agricultural marketplace errors
const (
	ErrCodeProductNotFound         = "PRODUCT_NOT_FOUND"
	ErrCodeUserNotFound           = "USER_NOT_FOUND"
	ErrCodeTransactionNotFound    = "TRANSACTION_NOT_FOUND"
	ErrCodeInvalidCUIT           = "INVALID_CUIT"
	ErrCodeEmailAlreadyExists    = "EMAIL_ALREADY_EXISTS"
	ErrCodeCUITAlreadyExists     = "CUIT_ALREADY_EXISTS"
	ErrCodeInsufficientStock     = "INSUFFICIENT_STOCK"
	ErrCodeInvalidPriceRange     = "INVALID_PRICE_RANGE"
	ErrCodeGeolocationRequired   = "GEOLOCATION_REQUIRED"
	ErrCodeImageUploadFailed     = "IMAGE_UPLOAD_FAILED"
	ErrCodeWhatsAppLinkFailed    = "WHATSAPP_LINK_FAILED"
	ErrCodePaymentProcessingFailed = "PAYMENT_PROCESSING_FAILED"
)

// Agricultural marketplace specific error functions
func NewProductNotFoundError() error {
	return fmt.Errorf("product not found")
}

func NewInvalidCUITError(cuit string) error {
	return fmt.Errorf("invalid CUIT format: %s", cuit)
}

func NewInsufficientStockError(requested, available int) error {
	return fmt.Errorf("insufficient stock: requested %d, available %d", requested, available)
}

func NewGeolocationRequiredError() error {
	return fmt.Errorf("geolocation is required for this operation")
}