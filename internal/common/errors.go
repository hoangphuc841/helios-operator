/*
Copyright 2025.
*/

package common

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// ReconciliationError represents an error that occurred during reconciliation.
type ReconciliationError struct {
	Resource string // Type of resource being reconciled (e.g., "Pipeline", "EventListener")
	Name     string // Name of the resource
	Op       string // Operation being performed (e.g., "create", "update", "delete")
	Err      error  // Underlying error
}

func (e *ReconciliationError) Error() string {
	return fmt.Sprintf("failed to %s %s '%s': %v", e.Op, e.Resource, e.Name, e.Err)
}

func (e *ReconciliationError) Unwrap() error {
	return e.Err
}

// NewReconciliationError creates a new ReconciliationError.
func NewReconciliationError(resource, name, op string, err error) *ReconciliationError {
	return &ReconciliationError{
		Resource: resource,
		Name:     name,
		Op:       op,
		Err:      err,
	}
}

// ResourceGenerationError represents an error that occurred while generating a resource.
type ResourceGenerationError struct {
	Resource string // Type of resource being generated
	Reason   string // Reason for the error
	Err      error  // Underlying error
}

func (e *ResourceGenerationError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("failed to generate %s: %s: %v", e.Resource, e.Reason, e.Err)
	}
	return fmt.Sprintf("failed to generate %s: %v", e.Resource, e.Err)
}

func (e *ResourceGenerationError) Unwrap() error {
	return e.Err
}

// NewResourceGenerationError creates a new ResourceGenerationError.
func NewResourceGenerationError(resource, reason string, err error) *ResourceGenerationError {
	return &ResourceGenerationError{
		Resource: resource,
		Reason:   reason,
		Err:      err,
	}
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field  string // Field that failed validation
	Value  string // Invalid value
	Reason string // Reason for validation failure
	Err    error  // Underlying error (optional)
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation failed for field '%s' (value: '%s'): %s: %v",
			e.Field, e.Value, e.Reason, e.Err)
	}
	return fmt.Sprintf("validation failed for field '%s' (value: '%s'): %s",
		e.Field, e.Value, e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, value, reason string, err error) *ValidationError {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
		Err:    err,
	}
}

// StatusUpdateError represents an error that occurred while updating status.
type StatusUpdateError struct {
	Resource string // Type of resource
	Name     string // Name of the resource
	Err      error  // Underlying error
}

func (e *StatusUpdateError) Error() string {
	return fmt.Sprintf("failed to update status for %s '%s': %v", e.Resource, e.Name, e.Err)
}

func (e *StatusUpdateError) Unwrap() error {
	return e.Err
}

// NewStatusUpdateError creates a new StatusUpdateError.
func NewStatusUpdateError(resource, name string, err error) *StatusUpdateError {
	return &StatusUpdateError{
		Resource: resource,
		Name:     name,
		Err:      err,
	}
}

// ReconcileResult represents the result of a reconciliation phase.
type ReconcileResult struct {
	Success      bool
	Error        error
	RequeueAfter time.Duration
	Status       ReconcileStatus
}

// ReconcileStatus represents the status of a reconciliation phase.
type ReconcileStatus struct {
	Phase       string
	Status      string
	Message     string
	LastUpdated time.Time
	RetryCount  int
	MaxRetries  int
}

// NewReconcileResult creates a new ReconcileResult.
func NewReconcileResult(success bool, err error, requeueAfter time.Duration) *ReconcileResult {
	return &ReconcileResult{
		Success:      success,
		Error:        err,
		RequeueAfter: requeueAfter,
		Status: ReconcileStatus{
			Phase:       PhasePending,
			Status:      "Pending",
			Message:     "",
			LastUpdated: time.Now(),
			RetryCount:  0,
			MaxRetries:  3,
		},
	}
}

// NewSuccessResult creates a successful ReconcileResult.
func NewSuccessResult(phase, status, message string) *ReconcileResult {
	return &ReconcileResult{
		Success:      true,
		Error:        nil,
		RequeueAfter: 0,
		Status: ReconcileStatus{
			Phase:       phase,
			Status:      status,
			Message:     message,
			LastUpdated: time.Now(),
			RetryCount:  0,
			MaxRetries:  3,
		},
	}
}

// NewErrorResult creates an error ReconcileResult.
func NewErrorResult(phase, status, message string, err error,
	requeueAfter time.Duration,
) *ReconcileResult {
	return &ReconcileResult{
		Success:      false,
		Error:        err,
		RequeueAfter: requeueAfter,
		Status: ReconcileStatus{
			Phase:       phase,
			Status:      status,
			Message:     message,
			LastUpdated: time.Now(),
			RetryCount:  0,
			MaxRetries:  3,
		},
	}
}

// IsTransientError determines if an error is transient and should be retried.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Check Kubernetes API errors that are typically transient.
	if k8serrors.IsConflict(err) ||
		k8serrors.IsServerTimeout(err) ||
		k8serrors.IsServiceUnavailable(err) ||
		k8serrors.IsTimeout(err) ||
		k8serrors.IsTooManyRequests(err) ||
		k8serrors.IsInternalError(err) {
		return true
	}

	// Check network errors that are typically transient.
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Temporary() {
		return true
	}

	// Check for specific error messages that indicate transient issues.
	errStr := err.Error()
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"server is overloaded",
		"rate limit exceeded",
		"too many requests",
		"service unavailable",
		"internal server error",
		"gateway timeout",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// GetRequeueDelay calculates the requeue delay with exponential backoff.
func GetRequeueDelay(retryCount int, baseDelay time.Duration) time.Duration {
	if retryCount <= 0 {
		return baseDelay
	}

	// Exponential backoff with jitter.
	// Use safe bit shifting to avoid overflow
	if retryCount > 30 {
		retryCount = 30 // Cap to prevent overflow
	}
	delay := baseDelay * time.Duration(1<<uint(retryCount-1))

	// Add jitter (±25%).
	jitter := time.Duration(float64(delay) * 0.25)
	delay += time.Duration(float64(jitter) * (2*float64(time.Now().UnixNano()%1000)/1000 - 1))

	// Cap at maximum delay.
	maxDelay := 5 * time.Minute
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}
