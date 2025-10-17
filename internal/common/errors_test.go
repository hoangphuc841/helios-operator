/*
Copyright 2025.
*/

package common

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestReconciliationError(t *testing.T) {
	t.Run("Error message formatting", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewReconciliationError("Pipeline", "test-pipeline", "create", underlyingErr)

		expected := "failed to create Pipeline 'test-pipeline': underlying error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewReconciliationError("Pipeline", "test-pipeline", "create", underlyingErr)

		assert.Equal(t, underlyingErr, err.Unwrap())
	})

	t.Run("Fields are set correctly", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewReconciliationError("EventListener", "test-listener", "update", underlyingErr)

		assert.Equal(t, "EventListener", err.Resource)
		assert.Equal(t, "test-listener", err.Name)
		assert.Equal(t, "update", err.Op)
		assert.Equal(t, underlyingErr, err.Err)
	})
}

func TestResourceGenerationError(t *testing.T) {
	t.Run("Error message with reason", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewResourceGenerationError("Pipeline", "invalid spec", underlyingErr)

		expected := "failed to generate Pipeline: invalid spec: underlying error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error message without reason", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewResourceGenerationError("Pipeline", "", underlyingErr)

		expected := "failed to generate Pipeline: underlying error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewResourceGenerationError("Pipeline", "invalid spec", underlyingErr)

		assert.Equal(t, underlyingErr, err.Unwrap())
	})

	t.Run("Fields are set correctly", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewResourceGenerationError("EventListener", "invalid config", underlyingErr)

		assert.Equal(t, "EventListener", err.Resource)
		assert.Equal(t, "invalid config", err.Reason)
		assert.Equal(t, underlyingErr, err.Err)
	})
}

func TestValidationError(t *testing.T) {
	t.Run("Error message with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewValidationError("name", "invalid-name", "must be alphanumeric", underlyingErr)

		expected := "validation failed for field 'name' (value: 'invalid-name'): must be alphanumeric: underlying error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error message without underlying error", func(t *testing.T) {
		err := NewValidationError("name", "invalid-name", "must be alphanumeric", nil)

		expected := "validation failed for field 'name' (value: 'invalid-name'): must be alphanumeric"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewValidationError("name", "invalid-name", "must be alphanumeric", underlyingErr)

		assert.Equal(t, underlyingErr, err.Unwrap())
	})

	t.Run("Unwrap returns nil when no underlying error", func(t *testing.T) {
		err := NewValidationError("name", "invalid-name", "must be alphanumeric", nil)

		assert.Nil(t, err.Unwrap())
	})

	t.Run("Fields are set correctly", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewValidationError("port", "99999", "must be between 1-65535", underlyingErr)

		assert.Equal(t, "port", err.Field)
		assert.Equal(t, "99999", err.Value)
		assert.Equal(t, "must be between 1-65535", err.Reason)
		assert.Equal(t, underlyingErr, err.Err)
	})
}

func TestStatusUpdateError(t *testing.T) {
	t.Run("Error message formatting", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewStatusUpdateError("HeliosApp", "test-app", underlyingErr)

		expected := "failed to update status for HeliosApp 'test-app': underlying error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewStatusUpdateError("HeliosApp", "test-app", underlyingErr)

		assert.Equal(t, underlyingErr, err.Unwrap())
	})

	t.Run("Fields are set correctly", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := NewStatusUpdateError("HeliosApp", "test-app", underlyingErr)

		assert.Equal(t, "HeliosApp", err.Resource)
		assert.Equal(t, "test-app", err.Name)
		assert.Equal(t, underlyingErr, err.Err)
	})
}

func TestReconcileResult(t *testing.T) {
	t.Run("NewReconcileResult creates correct result", func(t *testing.T) {
		underlyingErr := errors.New("test error")
		requeueAfter := 5 * time.Second

		result := NewReconcileResult(false, underlyingErr, requeueAfter)

		assert.False(t, result.Success)
		assert.Equal(t, underlyingErr, result.Error)
		assert.Equal(t, requeueAfter, result.RequeueAfter)
		assert.Equal(t, PhasePending, result.Status.Phase)
		assert.Equal(t, "Pending", result.Status.Status)
		assert.Empty(t, result.Status.Message)
		assert.Equal(t, 0, result.Status.RetryCount)
		assert.Equal(t, 3, result.Status.MaxRetries)
		assert.True(t, time.Since(result.Status.LastUpdated) < time.Second)
	})

	t.Run("NewSuccessResult creates correct result", func(t *testing.T) {
		phase := "Pipeline"
		status := "Success"
		message := "Pipeline created successfully"

		result := NewSuccessResult(phase, status, message)

		assert.True(t, result.Success)
		assert.Nil(t, result.Error)
		assert.Equal(t, time.Duration(0), result.RequeueAfter)
		assert.Equal(t, phase, result.Status.Phase)
		assert.Equal(t, status, result.Status.Status)
		assert.Equal(t, message, result.Status.Message)
		assert.Equal(t, 0, result.Status.RetryCount)
		assert.Equal(t, 3, result.Status.MaxRetries)
		assert.True(t, time.Since(result.Status.LastUpdated) < time.Second)
	})

	t.Run("NewErrorResult creates correct result", func(t *testing.T) {
		phase := "Pipeline"
		status := "Failed"
		message := "Pipeline creation failed"
		underlyingErr := errors.New("test error")
		requeueAfter := 10 * time.Second

		result := NewErrorResult(phase, status, message, underlyingErr, requeueAfter)

		assert.False(t, result.Success)
		assert.Equal(t, underlyingErr, result.Error)
		assert.Equal(t, requeueAfter, result.RequeueAfter)
		assert.Equal(t, phase, result.Status.Phase)
		assert.Equal(t, status, result.Status.Status)
		assert.Equal(t, message, result.Status.Message)
		assert.Equal(t, 0, result.Status.RetryCount)
		assert.Equal(t, 3, result.Status.MaxRetries)
		assert.True(t, time.Since(result.Status.LastUpdated) < time.Second)
	})
}

func TestIsTransientError(t *testing.T) {
	t.Run("nil error returns false", func(t *testing.T) {
		assert.False(t, IsTransientError(nil))
	})

	t.Run("Kubernetes API errors are transient", func(t *testing.T) {
		testCases := []struct {
			name string
			err  error
		}{
			{"Conflict", k8serrors.NewConflict(schema.GroupResource{}, "test", errors.New("conflict"))},
			{"ServerTimeout", k8serrors.NewServerTimeout(schema.GroupResource{}, "test", 1)},
			{"ServiceUnavailable", k8serrors.NewServiceUnavailable("service unavailable")},
			{"Timeout", k8serrors.NewTimeoutError("timeout", 1)},
			{"TooManyRequests", k8serrors.NewTooManyRequestsError("too many requests")},
			{"InternalError", k8serrors.NewInternalError(errors.New("internal error"))},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.True(t, IsTransientError(tc.err), "Expected %s to be transient", tc.name)
			})
		}
	})

	t.Run("Network errors are transient", func(t *testing.T) {
		// Create a temporary network error
		netErr := &net.OpError{
			Op:  "read",
			Net: "tcp",
			Err: &net.DNSError{
				Err:         "temporary failure",
				Name:        "example.com",
				Server:      "8.8.8.8",
				IsTemporary: true,
			},
		}

		assert.True(t, IsTransientError(netErr))
	})

	t.Run("Error messages with transient patterns are transient", func(t *testing.T) {
		transientMessages := []string{
			"connection refused",
			"connection reset",
			"timeout occurred",
			"temporary failure in name resolution",
			"server is overloaded",
			"rate limit exceeded",
			"too many requests",
			"service unavailable",
			"internal server error",
			"gateway timeout",
		}

		for _, msg := range transientMessages {
			t.Run(msg, func(t *testing.T) {
				err := errors.New(msg)
				assert.True(t, IsTransientError(err), "Expected error with message '%s' to be transient", msg)
			})
		}
	})

	t.Run("Case insensitive pattern matching", func(t *testing.T) {
		testCases := []string{
			"CONNECTION REFUSED",
			"Connection Reset",
			"TIMEOUT",
			"Temporary Failure",
			"Server Is Overloaded",
			"Rate Limit Exceeded",
			"Too Many Requests",
			"Service Unavailable",
			"Internal Server Error",
			"Gateway Timeout",
		}

		for _, msg := range testCases {
			t.Run(msg, func(t *testing.T) {
				err := errors.New(msg)
				assert.True(t, IsTransientError(err), "Expected error with message '%s' to be transient", msg)
			})
		}
	})

	t.Run("Non-transient errors return false", func(t *testing.T) {
		nonTransientErrors := []error{
			errors.New("validation failed"),
			errors.New("invalid input"),
			errors.New("not found"),
			errors.New("permission denied"),
			errors.New("resource already exists"),
		}

		for _, err := range nonTransientErrors {
			t.Run(err.Error(), func(t *testing.T) {
				assert.False(t, IsTransientError(err), "Expected error '%s' to not be transient", err.Error())
			})
		}
	})
}

func TestGetRequeueDelay(t *testing.T) {
	baseDelay := 1 * time.Second

	t.Run("retryCount <= 0 returns baseDelay", func(t *testing.T) {
		assert.Equal(t, baseDelay, GetRequeueDelay(0, baseDelay))
		assert.Equal(t, baseDelay, GetRequeueDelay(-1, baseDelay))
	})

	t.Run("exponential backoff calculation", func(t *testing.T) {
		// Test first few retries
		expectedDelays := []time.Duration{
			1 * time.Second, // retryCount = 1
			2 * time.Second, // retryCount = 2
			4 * time.Second, // retryCount = 3
			8 * time.Second, // retryCount = 4
		}

		for i, expected := range expectedDelays {
			retryCount := i + 1
			delay := GetRequeueDelay(retryCount, baseDelay)

			// Allow for jitter (±25%)
			minDelay := time.Duration(float64(expected) * 0.75)
			maxDelay := time.Duration(float64(expected) * 1.25)

			assert.True(t, delay >= minDelay, "Delay %v should be >= %v for retryCount %d", delay, minDelay, retryCount)
			assert.True(t, delay <= maxDelay, "Delay %v should be <= %v for retryCount %d", delay, maxDelay, retryCount)
		}
	})

	t.Run("retryCount is capped at 30", func(t *testing.T) {
		// Test that very high retry counts don't cause overflow
		delay := GetRequeueDelay(100, baseDelay)

		// Should be capped at 5 minutes
		maxDelay := 5 * time.Minute
		assert.True(t, delay <= maxDelay, "Delay %v should be <= %v", delay, maxDelay)
	})

	t.Run("delay is capped at 5 minutes", func(t *testing.T) {
		// Use a very large base delay to test the cap
		largeBaseDelay := 10 * time.Minute
		delay := GetRequeueDelay(5, largeBaseDelay)

		maxDelay := 5 * time.Minute
		assert.True(t, delay <= maxDelay, "Delay %v should be <= %v", delay, maxDelay)
	})

	t.Run("jitter is applied", func(t *testing.T) {
		// Run multiple times to ensure jitter is applied
		delays := make([]time.Duration, 10)
		for i := 0; i < 10; i++ {
			delays[i] = GetRequeueDelay(3, baseDelay)
		}

		// All delays should be different due to jitter
		uniqueDelays := make(map[time.Duration]bool)
		for _, delay := range delays {
			uniqueDelays[delay] = true
		}

		// With jitter, we should have some variation
		assert.True(t, len(uniqueDelays) > 1, "Expected jitter to create variation in delays")
	})

	t.Run("different base delays work correctly", func(t *testing.T) {
		testCases := []struct {
			baseDelay  time.Duration
			retryCount int
		}{
			{100 * time.Millisecond, 1},
			{500 * time.Millisecond, 2},
			{2 * time.Second, 3},
		}

		for _, tc := range testCases {
			t.Run(tc.baseDelay.String(), func(t *testing.T) {
				delay := GetRequeueDelay(tc.retryCount, tc.baseDelay)

				// Delay should be reasonable (not too small or too large)
				// Due to jitter, it might be smaller than base delay for very small base delays
				minReasonableDelay := tc.baseDelay / 2
				maxReasonableDelay := 10 * time.Minute
				assert.True(t, delay >= minReasonableDelay, "Delay %v should be >= %v", delay, minReasonableDelay)
				assert.True(t, delay <= maxReasonableDelay, "Delay %v should be <= %v", delay, maxReasonableDelay)
			})
		}
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("ReconciliationError chains properly", func(t *testing.T) {
		originalErr := errors.New("original error")
		reconcileErr := NewReconciliationError("Pipeline", "test", "create", originalErr)

		// Test error chaining
		var unwrappedErr error
		unwrappedErr = reconcileErr
		for unwrappedErr != nil {
			unwrappedErr = errors.Unwrap(unwrappedErr)
		}

		// Should be able to unwrap to the original error
		assert.True(t, errors.Is(reconcileErr, originalErr))
	})

	t.Run("ResourceGenerationError chains properly", func(t *testing.T) {
		originalErr := errors.New("original error")
		genErr := NewResourceGenerationError("Pipeline", "invalid spec", originalErr)

		// Test error chaining
		assert.True(t, errors.Is(genErr, originalErr))
	})

	t.Run("ValidationError chains properly", func(t *testing.T) {
		originalErr := errors.New("original error")
		validationErr := NewValidationError("name", "invalid", "must be valid", originalErr)

		// Test error chaining
		assert.True(t, errors.Is(validationErr, originalErr))
	})

	t.Run("StatusUpdateError chains properly", func(t *testing.T) {
		originalErr := errors.New("original error")
		statusErr := NewStatusUpdateError("HeliosApp", "test", originalErr)

		// Test error chaining
		assert.True(t, errors.Is(statusErr, originalErr))
	})
}

func TestReconcileStatus(t *testing.T) {
	t.Run("ReconcileStatus fields", func(t *testing.T) {
		status := ReconcileStatus{
			Phase:       "Pipeline",
			Status:      "Success",
			Message:     "Pipeline created successfully",
			LastUpdated: time.Now(),
			RetryCount:  2,
			MaxRetries:  5,
		}

		assert.Equal(t, "Pipeline", status.Phase)
		assert.Equal(t, "Success", status.Status)
		assert.Equal(t, "Pipeline created successfully", status.Message)
		assert.True(t, time.Since(status.LastUpdated) < time.Second)
		assert.Equal(t, 2, status.RetryCount)
		assert.Equal(t, 5, status.MaxRetries)
	})
}
