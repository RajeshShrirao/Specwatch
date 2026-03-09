package specerr

import (
	"errors"
	"testing"
	"time"
)

func TestSpecError(t *testing.T) {
	// Test New
	err := New(ErrCodeInternal, "test error")
	if err.Code != ErrCodeInternal {
		t.Errorf("expected code %s, got %s", ErrCodeInternal, err.Code)
	}
	if err.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", err.Message)
	}
	if err.Category != CategoryUnknown {
		t.Errorf("expected category %s, got %s", CategoryUnknown, err.Category)
	}
}

func TestSpecErrorWrapping(t *testing.T) {
	original := errors.New("original error")
	err := Wrap(original, ErrCodeNetwork, "network failed")

	if err.Code != ErrCodeNetwork {
		t.Errorf("expected code %s, got %s", ErrCodeNetwork, err.Code)
	}
	if err.Underlying != original {
		t.Errorf("expected underlying error to be original")
	}
}

func TestSpecErrorTransient(t *testing.T) {
	// Test transient error detection
	transientErr := NewTransient(ErrCodeNetwork, "connection refused")
	if !transientErr.IsTransient() {
		t.Error("expected transient error to be transient")
	}

	// Test permanent error detection
	permanentErr := NewPermanent(ErrCodeNotFound, "file not found")
	if permanentErr.IsTransient() {
		t.Error("expected permanent error to not be transient")
	}
}

func TestSpecErrorContext(t *testing.T) {
	err := New(ErrCodeInternal, "test error").
		WithContext("file", "test.go").
		WithContext("line", 42)

	if err.Context == nil {
		t.Error("expected context to be set")
	}
	if err.Context["file"] != "test.go" {
		t.Errorf("expected file in context, got %v", err.Context["file"])
	}
	if err.Context["line"] != 42 {
		t.Errorf("expected line in context, got %v", err.Context["line"])
	}
}

func TestIsRetryable(t *testing.T) {
	// Test transient error
	transientErr := NewTransient(ErrCodeNetwork, "timeout")
	if !IsRetryable(transientErr) {
		t.Error("expected transient error to be retryable")
	}

	// Test permanent error
	permanentErr := NewPermanent(ErrCodeNotFound, "not found")
	if IsRetryable(permanentErr) {
		t.Error("expected permanent error to not be retryable")
	}

	// Test wrapped transient error
	wrappedErr := WrapTransient(errors.New("connection refused"), ErrCodeNetwork, "failed")
	if !IsRetryable(wrappedErr) {
		t.Error("expected wrapped transient error to be retryable")
	}
}

func TestRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("expected max attempts 3, got %d", config.MaxAttempts)
	}
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected initial delay 100ms, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 5*time.Second {
		t.Errorf("expected max delay 5s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("expected multiplier 2.0, got %f", config.Multiplier)
	}
}

func TestRetryOptions(t *testing.T) {
	config := DefaultRetryConfig()

	// Apply options
	WithMaxAttempts(5)(config)
	WithInitialDelay(200 * time.Millisecond)(config)
	WithMaxDelay(10 * time.Second)(config)
	WithMultiplier(1.5)(config)

	if config.MaxAttempts != 5 {
		t.Errorf("expected max attempts 5, got %d", config.MaxAttempts)
	}
	if config.InitialDelay != 200*time.Millisecond {
		t.Errorf("expected initial delay 200ms, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 10*time.Second {
		t.Errorf("expected max delay 10s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 1.5 {
		t.Errorf("expected multiplier 1.5, got %f", config.Multiplier)
	}
}

func TestRetry(t *testing.T) {
	attemptCount := 0

	// Function that succeeds on third attempt
	fn := func() (int, error) {
		attemptCount++
		if attemptCount < 3 {
			return 0, NewTransient(ErrCodeNetwork, "temporary failure")
		}
		return attemptCount, nil
	}

	result, err := Retry(fn, WithMaxAttempts(5), WithInitialDelay(1*time.Millisecond))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != 3 {
		t.Errorf("expected result 3, got %d", result)
	}
}

func TestRetryMaxAttempts(t *testing.T) {
	attemptCount := 0

	fn := func() (int, error) {
		attemptCount++
		return 0, NewPermanent(ErrCodeInternal, "always fails")
	}

	_, err := Retry(fn, WithMaxAttempts(3), WithInitialDelay(1*time.Millisecond))

	if err == nil {
		t.Error("expected error")
	}
	if attemptCount != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
}

func TestRetryResult(t *testing.T) {
	attemptCount := 0

	fn := func() (int, error) {
		attemptCount++
		if attemptCount < 2 {
			return 0, NewTransient(ErrCodeNetwork, "temporary failure")
		}
		return attemptCount, nil
	}

	result := ExecuteRetry(fn, WithMaxAttempts(3), WithInitialDelay(1*time.Millisecond))

	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}
	if result.Result != 2 {
		t.Errorf("expected result 2, got %d", result.Result)
	}
	if result.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", result.Attempts)
	}
	if result.Duration == 0 {
		t.Error("expected duration to be set")
	}
}
