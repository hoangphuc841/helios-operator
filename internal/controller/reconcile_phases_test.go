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

// Package controller tests the specialized reconcile functions for different phases
// of the HeliosApp reconciliation process, including error handling and retry logic.
package controller

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"github.com/hoangphuc841/helios-operator/internal/common"
)

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "conflict error",
			err:      apierrors.NewConflict(schema.GroupResource{}, "test", errors.New("conflict")),
			expected: true,
		},
		{
			name:     "server timeout error",
			err:      apierrors.NewServerTimeout(schema.GroupResource{}, "test", 1),
			expected: true,
		},
		{
			name:     "service unavailable error",
			err:      apierrors.NewServiceUnavailable("service unavailable"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      apierrors.NewTimeoutError("timeout", 1),
			expected: true,
		},
		{
			name:     "too many requests error",
			err:      apierrors.NewTooManyRequestsError("too many requests"),
			expected: true,
		},
		{
			name:     "internal server error",
			err:      apierrors.NewInternalError(errors.New("internal error")),
			expected: true,
		},
		{
			name:     "temporary network error",
			err:      &net.DNSError{Err: "temporary failure", IsTemporary: true},
			expected: true,
		},
		{
			name:     "connection refused error",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "rate limit exceeded error",
			err:      errors.New("rate limit exceeded"),
			expected: true,
		},
		{
			name:     "gateway timeout error",
			err:      errors.New("gateway timeout"),
			expected: true,
		},
		{
			name:     "non-transient error",
			err:      errors.New("validation failed"),
			expected: false,
		},
		{
			name:     "permanent network error",
			err:      &net.DNSError{Err: "no such host", IsTemporary: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.IsTransientError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRequeueDelay(t *testing.T) {
	tests := []struct {
		name       string
		retryCount int
		baseDelay  time.Duration
		expected   time.Duration
	}{
		{
			name:       "no retries",
			retryCount: 0,
			baseDelay:  30 * time.Second,
			expected:   30 * time.Second,
		},
		{
			name:       "first retry",
			retryCount: 1,
			baseDelay:  30 * time.Second,
			expected:   30 * time.Second, // 30 * 2^0 = 30
		},
		{
			name:       "second retry",
			retryCount: 2,
			baseDelay:  30 * time.Second,
			expected:   60 * time.Second, // 30 * 2^1 = 60
		},
		{
			name:       "third retry",
			retryCount: 3,
			baseDelay:  30 * time.Second,
			expected:   120 * time.Second, // 30 * 2^2 = 120
		},
		{
			name:       "high retry count (should be capped)",
			retryCount: 10,
			baseDelay:  30 * time.Second,
			expected:   5 * time.Minute, // Should be capped at 5 minutes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.GetRequeueDelay(tt.retryCount, tt.baseDelay)
			// Allow for jitter (±25%)
			minExpected := tt.expected - time.Duration(float64(tt.expected)*0.25)
			maxExpected := tt.expected + time.Duration(float64(tt.expected)*0.25)

			assert.True(t, result >= minExpected, "Result %v should be >= %v", result, minExpected)
			assert.True(t, result <= maxExpected, "Result %v should be <= %v", result, maxExpected)
		})
	}
}

func TestReconcileResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := common.NewSuccessResult("Pipeline", "Success", "Pipeline reconciled successfully")

		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
		assert.Equal(t, time.Duration(0), result.RequeueAfter)
		assert.Equal(t, "Pipeline", result.Status.Phase)
		assert.Equal(t, "Success", result.Status.Status)
		assert.Equal(t, "Pipeline reconciled successfully", result.Status.Message)
		assert.False(t, result.Status.LastUpdated.IsZero())
	})

	t.Run("error result with requeue", func(t *testing.T) {
		err := errors.New("test error")
		requeueAfter := 30 * time.Second
		result := common.NewErrorResult("Pipeline", "Failed", "Pipeline reconciliation failed", err, requeueAfter)

		assert.False(t, result.Success)
		assert.Equal(t, err, result.Error)
		assert.Equal(t, requeueAfter, result.RequeueAfter)
		assert.Equal(t, "Pipeline", result.Status.Phase)
		assert.Equal(t, "Failed", result.Status.Status)
		assert.Equal(t, "Pipeline reconciliation failed", result.Status.Message)
		assert.False(t, result.Status.LastUpdated.IsZero())
	})

	t.Run("error result without requeue", func(t *testing.T) {
		err := errors.New("permanent error")
		result := common.NewErrorResult("Pipeline", "Failed", "Permanent error", err, 0)

		assert.False(t, result.Success)
		assert.Equal(t, err, result.Error)
		assert.Equal(t, time.Duration(0), result.RequeueAfter)
		assert.Equal(t, "Pipeline", result.Status.Phase)
		assert.Equal(t, "Failed", result.Status.Status)
		assert.Equal(t, "Permanent error", result.Status.Message)
	})
}

func TestEnhancedReconcileLogic(t *testing.T) {
	// Create a test HeliosApp
	heliosApp := &heliosappv1.HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
		Spec: heliosappv1.HeliosAppSpec{
			GitRepo:   "https://github.com/test/repo",
			ImageRepo: "test-registry/test-app",
		},
	}

	// Create a mock reconciler
	reconciler := &HeliosAppReconciler{
		Client: &mockClient{},
		Scheme: runtime.NewScheme(),
	}

	ctx := context.Background()
	logger := log.FromContext(ctx)

	t.Run("pipeline reconciliation success", func(t *testing.T) {
		result := reconciler.ReconcilePipeline(ctx, heliosApp, logger)

		// Since we're using a mock client, this should fail with a specific error
		// but we can test the structure
		assert.NotNil(t, result)
		assert.NotNil(t, result.Status)
		assert.Equal(t, "Pipeline", result.Status.Phase)
	})

	t.Run("triggers reconciliation success", func(t *testing.T) {
		result := reconciler.ReconcileTriggers(ctx, heliosApp, "test-app", "default", "test-pipeline", "default", "github-secret", map[string]interface{}{}, logger)

		assert.NotNil(t, result)
		assert.NotNil(t, result.Status)
		assert.Equal(t, "Triggers", result.Status.Phase)
	})

	t.Run("argocd reconciliation success", func(t *testing.T) {
		result := reconciler.ReconcileArgoCD(ctx, heliosApp, "test-app", "default", logger)

		assert.NotNil(t, result)
		assert.NotNil(t, result.Status)
		assert.Equal(t, "ArgoCD", result.Status.Phase)
	})

	t.Run("status reconciliation success", func(t *testing.T) {
		result := reconciler.ReconcileStatus(ctx, heliosApp, "test-app", "default", logger)

		assert.NotNil(t, result)
		assert.NotNil(t, result.Status)
		assert.Equal(t, "Status", result.Status.Phase)
	})
}

// mockClient is a simple mock client for testing
type mockClient struct {
	client.Client
}

func (m *mockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return apierrors.NewNotFound(schema.GroupResource{}, key.Name)
}

func (m *mockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (m *mockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return errors.New("mock create error")
}

func (m *mockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return errors.New("mock update error")
}

func (m *mockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return errors.New("mock delete error")
}

func (m *mockClient) Status() client.StatusWriter {
	return &mockStatusWriter{}
}

type mockStatusWriter struct{}

func (m *mockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return errors.New("mock status update error")
}

func (m *mockStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	return errors.New("mock status patch error")
}

func (m *mockStatusWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	return errors.New("mock status create error")
}

// Test the enhanced error handling in the main reconcile loop
func TestEnhancedReconcileWithTransientErrors(t *testing.T) {
	// This test would require a more sophisticated mock setup
	// to simulate transient errors and verify the requeue behavior
	// For now, we'll test the error detection logic

	t.Run("transient error detection", func(t *testing.T) {
		transientErrors := []error{
			apierrors.NewConflict(schema.GroupResource{}, "test", errors.New("conflict")),
			apierrors.NewServerTimeout(schema.GroupResource{}, "test", 1),
			apierrors.NewServiceUnavailable("service unavailable"),
			errors.New("connection refused"),
			errors.New("rate limit exceeded"),
		}

		for _, err := range transientErrors {
			assert.True(t, common.IsTransientError(err), "Error %v should be detected as transient", err)
		}
	})

	t.Run("permanent error detection", func(t *testing.T) {
		permanentErrors := []error{
			errors.New("validation failed"),
			errors.New("invalid configuration"),
			errors.New("resource not found"),
		}

		for _, err := range permanentErrors {
			assert.False(t, common.IsTransientError(err), "Error %v should not be detected as transient", err)
		}
	})
}
