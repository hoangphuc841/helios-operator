/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReconciliationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ReconciliationError
		expected string
	}{
		{
			name: "basic error",
			err: &ReconciliationError{
				Resource: "Pipeline",
				Name:     "test-pipeline",
				Op:       "create",
				Err:      errors.New("resource already exists"),
			},
			expected: "failed to create Pipeline 'test-pipeline': resource already exists",
		},
		{
			name: "get operation",
			err: &ReconciliationError{
				Resource: "EventListener",
				Name:     "test-listener",
				Op:       "get",
				Err:      errors.New("not found"),
			},
			expected: "failed to get EventListener 'test-listener': not found",
		},
		{
			name: "delete operation",
			err: &ReconciliationError{
				Resource: "Application",
				Name:     "test-app",
				Op:       "delete",
				Err:      errors.New("resource is being used"),
			},
			expected: "failed to delete Application 'test-app': resource is being used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestReconciliationError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &ReconciliationError{
		Resource: "Pipeline",
		Name:     "test",
		Op:       "create",
		Err:      originalErr,
	}

	unwrapped := err.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
	assert.True(t, errors.Is(err, originalErr))
}

func TestNewReconciliationError(t *testing.T) {
	originalErr := errors.New("test error")
	err := NewReconciliationError("Pipeline", "test-pipeline", "update", originalErr)

	assert.NotNil(t, err)
	assert.Equal(t, "Pipeline", err.Resource)
	assert.Equal(t, "test-pipeline", err.Name)
	assert.Equal(t, "update", err.Op)
	assert.Equal(t, originalErr, err.Err)
	assert.Contains(t, err.Error(), "failed to update Pipeline 'test-pipeline'")
}

func TestResourceGenerationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ResourceGenerationError
		expected string
	}{
		{
			name: "with reason",
			err: &ResourceGenerationError{
				Resource: "Pipeline",
				Reason:   "invalid spec",
				Err:      errors.New("missing required field"),
			},
			expected: "failed to generate Pipeline: invalid spec: missing required field",
		},
		{
			name: "without reason",
			err: &ResourceGenerationError{
				Resource: "EventListener",
				Reason:   "",
				Err:      errors.New("template error"),
			},
			expected: "failed to generate EventListener: template error",
		},
		{
			name: "application generation",
			err: &ResourceGenerationError{
				Resource: "Application",
				Reason:   "invalid GitOps repo",
				Err:      errors.New("URL parse failed"),
			},
			expected: "failed to generate Application: invalid GitOps repo: URL parse failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestResourceGenerationError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &ResourceGenerationError{
		Resource: "Pipeline",
		Reason:   "test reason",
		Err:      originalErr,
	}

	unwrapped := err.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
	assert.True(t, errors.Is(err, originalErr))
}

func TestNewResourceGenerationError(t *testing.T) {
	originalErr := errors.New("test error")
	
	tests := []struct {
		name     string
		resource string
		reason   string
		err      error
	}{
		{
			name:     "with reason",
			resource: "Pipeline",
			reason:   "invalid template",
			err:      originalErr,
		},
		{
			name:     "without reason",
			resource: "Application",
			reason:   "",
			err:      originalErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewResourceGenerationError(tt.resource, tt.reason, tt.err)

			assert.NotNil(t, err)
			assert.Equal(t, tt.resource, err.Resource)
			assert.Equal(t, tt.reason, err.Reason)
			assert.Equal(t, originalErr, err.Err)
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "with underlying error",
			err: &ValidationError{
				Field:  "gitRepo",
				Value:  "invalid-url",
				Reason: "must be a valid URL",
				Err:    errors.New("parse error"),
			},
			expected: "validation failed for field 'gitRepo' (value: 'invalid-url'): must be a valid URL: parse error",
		},
		{
			name: "without underlying error",
			err: &ValidationError{
				Field:  "port",
				Value:  "0",
				Reason: "must be between 1 and 65535",
				Err:    nil,
			},
			expected: "validation failed for field 'port' (value: '0'): must be between 1 and 65535",
		},
		{
			name: "replicas validation",
			err: &ValidationError{
				Field:  "replicas",
				Value:  "-1",
				Reason: "must be non-negative",
				Err:    nil,
			},
			expected: "validation failed for field 'replicas' (value: '-1'): must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	tests := []struct {
		name        string
		err         *ValidationError
		shouldUnwrap bool
	}{
		{
			name: "with underlying error",
			err: &ValidationError{
				Field:  "test",
				Value:  "value",
				Reason: "reason",
				Err:    errors.New("underlying"),
			},
			shouldUnwrap: true,
		},
		{
			name: "without underlying error",
			err: &ValidationError{
				Field:  "test",
				Value:  "value",
				Reason: "reason",
				Err:    nil,
			},
			shouldUnwrap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unwrapped := tt.err.Unwrap()
			if tt.shouldUnwrap {
				assert.NotNil(t, unwrapped)
				assert.Equal(t, tt.err.Err, unwrapped)
			} else {
				assert.Nil(t, unwrapped)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	originalErr := errors.New("test error")
	
	tests := []struct {
		name   string
		field  string
		value  string
		reason string
		err    error
	}{
		{
			name:   "with error",
			field:  "gitRepo",
			value:  "bad-url",
			reason: "invalid URL format",
			err:    originalErr,
		},
		{
			name:   "without error",
			field:  "port",
			value:  "70000",
			reason: "port out of range",
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.value, tt.reason, tt.err)

			assert.NotNil(t, err)
			assert.Equal(t, tt.field, err.Field)
			assert.Equal(t, tt.value, err.Value)
			assert.Equal(t, tt.reason, err.Reason)
			assert.Equal(t, tt.err, err.Err)
		})
	}
}

func TestStatusUpdateError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *StatusUpdateError
		expected string
	}{
		{
			name: "pipeline status update",
			err: &StatusUpdateError{
				Resource: "HeliosApp",
				Name:     "test-app",
				Err:      errors.New("conflict"),
			},
			expected: "failed to update status for HeliosApp 'test-app': conflict",
		},
		{
			name: "application status update",
			err: &StatusUpdateError{
				Resource: "Application",
				Name:     "my-app",
				Err:      errors.New("resource not found"),
			},
			expected: "failed to update status for Application 'my-app': resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestStatusUpdateError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &StatusUpdateError{
		Resource: "HeliosApp",
		Name:     "test",
		Err:      originalErr,
	}

	unwrapped := err.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
	assert.True(t, errors.Is(err, originalErr))
}

func TestNewStatusUpdateError(t *testing.T) {
	originalErr := errors.New("test error")
	err := NewStatusUpdateError("HeliosApp", "test-app", originalErr)

	assert.NotNil(t, err)
	assert.Equal(t, "HeliosApp", err.Resource)
	assert.Equal(t, "test-app", err.Name)
	assert.Equal(t, originalErr, err.Err)
	assert.Contains(t, err.Error(), "failed to update status for HeliosApp 'test-app'")
}

// Test error wrapping with errors.Is and errors.As
func TestErrorWrapping(t *testing.T) {
	t.Run("ReconciliationError wrapping", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := NewReconciliationError("Pipeline", "test", "create", baseErr)

		assert.True(t, errors.Is(wrappedErr, baseErr))

		var reconcileErr *ReconciliationError
		assert.True(t, errors.As(wrappedErr, &reconcileErr))
		assert.Equal(t, "Pipeline", reconcileErr.Resource)
	})

	t.Run("ResourceGenerationError wrapping", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := NewResourceGenerationError("Application", "invalid spec", baseErr)

		assert.True(t, errors.Is(wrappedErr, baseErr))

		var genErr *ResourceGenerationError
		assert.True(t, errors.As(wrappedErr, &genErr))
		assert.Equal(t, "Application", genErr.Resource)
	})

	t.Run("ValidationError wrapping", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := NewValidationError("field", "value", "reason", baseErr)

		assert.True(t, errors.Is(wrappedErr, baseErr))

		var valErr *ValidationError
		assert.True(t, errors.As(wrappedErr, &valErr))
		assert.Equal(t, "field", valErr.Field)
	})

	t.Run("StatusUpdateError wrapping", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := NewStatusUpdateError("HeliosApp", "test", baseErr)

		assert.True(t, errors.Is(wrappedErr, baseErr))

		var statusErr *StatusUpdateError
		assert.True(t, errors.As(wrappedErr, &statusErr))
		assert.Equal(t, "HeliosApp", statusErr.Resource)
	})
}
