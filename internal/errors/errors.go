package specerr

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrorCode represents a structured error code for categorization
type ErrorCode string

const (
	// Core error codes
	ErrCodeInternal     ErrorCode = "INTERNAL"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeTimeout      ErrorCode = "TIMEOUT"
	ErrCodeNotSupported ErrorCode = "NOT_SUPPORTED"
	ErrCodeIO           ErrorCode = "IO_ERROR"
	ErrCodeParse        ErrorCode = "PARSE_ERROR"
	ErrCodeNetwork      ErrorCode = "NETWORK_ERROR"
	ErrCodeRateLimit    ErrorCode = "RATE_LIMIT"
	ErrCodeConfig       ErrorCode = "CONFIG_ERROR"
)

// ErrorCategory represents the high-level category of the error
type ErrorCategory string

const (
	CategoryTransient ErrorCategory = "transient" // Temporary errors that may succeed on retry
	CategoryPermanent ErrorCategory = "permanent" // Errors that will never succeed on retry
	CategoryUnknown   ErrorCategory = "unknown"   // Unknown error category
)

// SpecError is a structured error with additional context
type SpecError struct {
	Code       ErrorCode      `json:"code"`
	Category   ErrorCategory  `json:"category"`
	Message    string         `json:"message"`
	Underlying error          `json:"-"`
	Context    map[string]any `json:"context,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// Error implements the error interface
func (e *SpecError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Underlying)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is and errors.As
func (e *SpecError) Unwrap() error {
	return e.Underlying
}

// Is checks if the error matches the target using errors.Is
func (e *SpecError) Is(target error) bool {
	if target == nil {
		return false
	}
	// Check if target is a SpecError
	if specErr, ok := target.(*SpecError); ok {
		return e.Code == specErr.Code
	}
	// Check if target is in the error chain
	return errors.Is(e.Underlying, target)
}

// As checks if the error can be cast to the target type using errors.As
func (e *SpecError) As(target any) bool {
	if target == nil {
		return false
	}
	// Check if target is a pointer to SpecError
	if specErrPtr, ok := target.(**SpecError); ok {
		*specErrPtr = e
		return true
	}
	// Check if target is in the error chain
	return errors.As(e.Underlying, target)
}

// IsTransient checks if the error is likely transient and worth retrying
func (e *SpecError) IsTransient() bool {
	switch e.Category {
	case CategoryTransient:
		return true
	case CategoryPermanent:
		return false
	default:
		// Check common transient error patterns
		return isTransientError(e.Underlying)
	}
}

// WithContext adds context key-value pairs to the error
func (e *SpecError) WithContext(key string, value any) *SpecError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithUnderlying wraps an underlying error
func (e *SpecError) WithUnderlying(err error) *SpecError {
	e.Underlying = err
	return e
}

// New creates a new SpecError with the given code and message
func New(code ErrorCode, message string) *SpecError {
	return &SpecError{
		Code:      code,
		Category:  CategoryUnknown,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewPermanent creates a new permanent (non-retryable) error
func NewPermanent(code ErrorCode, message string) *SpecError {
	return &SpecError{
		Code:      code,
		Category:  CategoryPermanent,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewTransient creates a new transient (retryable) error
func NewTransient(code ErrorCode, message string) *SpecError {
	return &SpecError{
		Code:      code,
		Category:  CategoryTransient,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// Wrap wraps an existing error with a SpecError
func Wrap(err error, code ErrorCode, message string) *SpecError {
	if err == nil {
		return nil
	}
	return &SpecError{
		Code:       code,
		Category:   categorizeError(err),
		Message:    message,
		Underlying: err,
		Timestamp:  time.Now(),
	}
}

// WrapPermanent wraps an error as permanent
func WrapPermanent(err error, code ErrorCode, message string) *SpecError {
	if err == nil {
		return nil
	}
	return &SpecError{
		Code:       code,
		Category:   CategoryPermanent,
		Message:    message,
		Underlying: err,
		Timestamp:  time.Now(),
	}
}

// WrapTransient wraps an error as transient
func WrapTransient(err error, code ErrorCode, message string) *SpecError {
	if err == nil {
		return nil
	}
	return &SpecError{
		Code:       code,
		Category:   CategoryTransient,
		Message:    message,
		Underlying: err,
		Timestamp:  time.Now(),
	}
}

// categorizeError determines the error category based on the underlying error
func categorizeError(err error) ErrorCategory {
	if specErr, ok := err.(*SpecError); ok {
		return specErr.Category
	}

	// Check for common transient error patterns
	if isTransientError(err) {
		return CategoryTransient
	}

	return CategoryPermanent
}

// isTransientError checks if an error is likely transient
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check for common transient error patterns
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"temporary failure",
		"timeout",
		"deadline exceeded",
		"i/o timeout",
		"network is unreachable",
		"no route to host",
		"server misbehaving",
		"too many requests",
		"rate limit",
		"service unavailable",
		"503",
		"429",
		"connection reset by peer",
		"broken pipe",
	}

	errLower := strings.ToLower(errStr)
	for _, pattern := range transientPatterns {
		if strings.Contains(errLower, pattern) {
			return true
		}
	}

	return false
}

// IsTransientError checks if an error string contains transient keywords
func IsTransientError(err error) bool {
	return isTransientError(err)
}
