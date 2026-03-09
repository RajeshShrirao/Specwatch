package specerr

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RetryConfig contains configuration for retry behavior
type RetryConfig struct {
	MaxAttempts    int           // Maximum number of attempts (default: 3)
	InitialDelay   time.Duration // Initial delay between retries (default: 100ms)
	MaxDelay       time.Duration // Maximum delay between retries (default: 5s)
	Multiplier     float64       // Backoff multiplier (default: 2.0)
	RetryableCodes []ErrorCode   // Specific error codes to retry
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryOption is a functional option for configuring retry
type RetryOption func(*RetryConfig)

// WithMaxAttempts sets the maximum number of retry attempts
func WithMaxAttempts(n int) RetryOption {
	return func(c *RetryConfig) {
		c.MaxAttempts = n
	}
}

// WithInitialDelay sets the initial delay between retries
func WithInitialDelay(d time.Duration) RetryOption {
	return func(c *RetryConfig) {
		c.InitialDelay = d
	}
}

// WithMaxDelay sets the maximum delay between retries
func WithMaxDelay(d time.Duration) RetryOption {
	return func(c *RetryConfig) {
		c.MaxDelay = d
	}
}

// WithMultiplier sets the backoff multiplier
func WithMultiplier(m float64) RetryOption {
	return func(c *RetryConfig) {
		c.Multiplier = m
	}
}

// WithRetryableCodes sets specific error codes that should trigger a retry
func WithRetryableCodes(codes ...ErrorCode) RetryOption {
	return func(c *RetryConfig) {
		c.RetryableCodes = codes
	}
}

// RetryError contains details about a failed retry attempt
type RetryError struct {
	LastError   error         // The last error that occurred
	Attempts    int           // Number of attempts made
	AllErrors   []error       // All errors that occurred during retry attempts
	LastAttempt time.Time     // Time of the last attempt
	Duration    time.Duration // Total time spent retrying
}

// Error implements the error interface
func (e *RetryError) Error() string {
	return fmt.Sprintf("retry failed after %d attempts: %v", e.Attempts, e.LastError)
}

// Unwrap returns the last error
func (e *RetryError) Unwrap() error {
	return e.LastError
}

// IsRetryable checks if the error should trigger a retry
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a SpecError with transient category
	if specErr, ok := err.(*SpecError); ok {
		return specErr.IsTransient()
	}

	// Check for transient error patterns in the error chain
	var lastErr error
	for {
		if lastErr = errors.Unwrap(err); lastErr == nil {
			break
		}
		if isTransientError(lastErr) {
			return true
		}
		err = lastErr
	}

	// Check original error
	return isTransientError(err)
}

// RetryableFunc is a function that can be retried
type RetryableFunc[T any] func() (T, error)

// Retry executes a function with retry logic
func Retry[T any](fn RetryableFunc[T], opts ...RetryOption) (T, error) {
	config := DefaultRetryConfig()
	for _, opt := range opts {
		opt(config)
	}

	var lastErr error
	allErrors := make([]error, 0, config.MaxAttempts)
	delay := config.InitialDelay
	startTime := time.Now()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		allErrors = append(allErrors, err)

		// Check if we should retry (not the last attempt)
		if attempt < config.MaxAttempts {
			// Only delay for transient errors or when RetryableCodes is not specified
			shouldDelay := IsRetryable(err)

			// If specific codes are set, only delay for those
			if len(config.RetryableCodes) > 0 {
				var specErr *SpecError
				if errors.As(err, &specErr) {
					shouldDelay = false
					for _, code := range config.RetryableCodes {
						if specErr.Code == code {
							shouldDelay = true
							break
						}
					}
				}
			}

			if shouldDelay {
				// Wait before next attempt
				time.Sleep(delay)

				// Exponential backoff
				delay = time.Duration(float64(delay) * config.Multiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}
		}
	}

	var zero T
	return zero, &RetryError{
		LastError:   lastErr,
		Attempts:    len(allErrors),
		AllErrors:   allErrors,
		LastAttempt: startTime,
		Duration:    time.Since(startTime),
	}
}

// RetryContext executes a function with retry logic and context support
func RetryContext[T any](ctx context.Context, fn RetryableFunc[T], opts ...RetryOption) (T, error) {
	config := DefaultRetryConfig()
	for _, opt := range opts {
		opt(config)
	}

	var lastErr error
	allErrors := make([]error, 0, config.MaxAttempts)
	delay := config.InitialDelay
	startTime := time.Now()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check context cancellation
		if ctx.Err() != nil {
			var zero T
			return zero, &RetryError{
				LastError:   ctx.Err(),
				Attempts:    attempt,
				AllErrors:   allErrors,
				LastAttempt: startTime,
				Duration:    time.Since(startTime),
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		allErrors = append(allErrors, err)

		// Check if we should retry
		if attempt >= config.MaxAttempts {
			break
		}

		// Check if error is retryable
		if !IsRetryable(err) {
			break
		}

		// Check if specific error codes are retryable
		if len(config.RetryableCodes) > 0 {
			var specErr *SpecError
			if errors.As(err, &specErr) {
				retryable := false
				for _, code := range config.RetryableCodes {
					if specErr.Code == code {
						retryable = true
						break
					}
				}
				if !retryable {
					break
				}
			}
		}

		// Wait before next attempt with context
		select {
		case <-ctx.Done():
			var zero T
			return zero, &RetryError{
				LastError:   ctx.Err(),
				Attempts:    attempt,
				AllErrors:   allErrors,
				LastAttempt: startTime,
				Duration:    time.Since(startTime),
			}
		case <-time.After(delay):
		}

		// Exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	var zero T
	return zero, &RetryError{
		LastError:   lastErr,
		Attempts:    len(allErrors),
		AllErrors:   allErrors,
		LastAttempt: startTime,
		Duration:    time.Since(startTime),
	}
}

// RetryResult allows retry with custom result handling
type RetryResult[T any] struct {
	Result   T
	Err      error
	Attempts int
	Duration time.Duration
}

// ExecuteRetry executes a retryable operation and returns detailed result
func ExecuteRetry[T any](fn RetryableFunc[T], opts ...RetryOption) *RetryResult[T] {
	config := DefaultRetryConfig()
	for _, opt := range opts {
		opt(config)
	}

	var lastErr error
	delay := config.InitialDelay
	startTime := time.Now()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return &RetryResult[T]{
				Result:   result,
				Err:      nil,
				Attempts: attempt,
				Duration: time.Since(startTime),
			}
		}

		lastErr = err

		// Check if we should retry
		if attempt >= config.MaxAttempts {
			break
		}

		// Check if error is retryable
		if !IsRetryable(err) {
			break
		}

		// Wait before next attempt
		time.Sleep(delay)

		// Exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	var zero T
	return &RetryResult[T]{
		Result:   zero,
		Err:      lastErr,
		Attempts: config.MaxAttempts,
		Duration: time.Since(startTime),
	}
}
