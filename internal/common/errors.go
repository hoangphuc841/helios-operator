/*
Copyright 2025.
*/

package common

import "fmt"

// ReconciliationError represents an error that occurred during reconciliation
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

// NewReconciliationError creates a new ReconciliationError
func NewReconciliationError(resource, name, op string, err error) *ReconciliationError {
	return &ReconciliationError{
		Resource: resource,
		Name:     name,
		Op:       op,
		Err:      err,
	}
}

// ResourceGenerationError represents an error that occurred while generating a resource
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

// NewResourceGenerationError creates a new ResourceGenerationError
func NewResourceGenerationError(resource, reason string, err error) *ResourceGenerationError {
	return &ResourceGenerationError{
		Resource: resource,
		Reason:   reason,
		Err:      err,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field  string // Field that failed validation
	Value  string // Invalid value
	Reason string // Reason for validation failure
	Err    error  // Underlying error (optional)
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation failed for field '%s' (value: '%s'): %s: %v", e.Field, e.Value, e.Reason, e.Err)
	}
	return fmt.Sprintf("validation failed for field '%s' (value: '%s'): %s", e.Field, e.Value, e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, value, reason string, err error) *ValidationError {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
		Err:    err,
	}
}

// StatusUpdateError represents an error that occurred while updating status
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

// NewStatusUpdateError creates a new StatusUpdateError
func NewStatusUpdateError(resource, name string, err error) *StatusUpdateError {
	return &StatusUpdateError{
		Resource: resource,
		Name:     name,
		Err:      err,
	}
}
