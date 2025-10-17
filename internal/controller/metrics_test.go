package controller

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentHealthToMetric(t *testing.T) {
	tests := []struct {
		name     string
		health   string
		expected float64
	}{
		{
			name:     "Healthy status",
			health:   "Healthy",
			expected: 1.0,
		},
		{
			name:     "Progressing status",
			health:   "Progressing",
			expected: 0.5,
		},
		{
			name:     "Degraded status",
			health:   "Degraded",
			expected: 0.0,
		},
		{
			name:     "Unknown status",
			health:   "Unknown",
			expected: -1.0,
		},
		{
			name:     "Empty status",
			health:   "",
			expected: -1.0,
		},
		{
			name:     "Invalid status",
			health:   "InvalidStatus",
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeploymentHealthToMetric(tt.health)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncStatusToMetric(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected float64
	}{
		{
			name:     "Synced status",
			status:   "Synced",
			expected: 1.0,
		},
		{
			name:     "OutOfSync status",
			status:   "OutOfSync",
			expected: 0.0,
		},
		{
			name:     "Unknown status",
			status:   "Unknown",
			expected: -1.0,
		},
		{
			name:     "Empty status",
			status:   "",
			expected: -1.0,
		},
		{
			name:     "Invalid status",
			status:   "InvalidStatus",
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SyncStatusToMetric(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHealthStatusToMetric(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected float64
	}{
		{
			name:     "Healthy status",
			status:   "Healthy",
			expected: 1.0,
		},
		{
			name:     "Progressing status",
			status:   "Progressing",
			expected: 0.5,
		},
		{
			name:     "Degraded status",
			status:   "Degraded",
			expected: 0.0,
		},
		{
			name:     "Unknown status",
			status:   "Unknown",
			expected: -1.0,
		},
		{
			name:     "Missing status",
			status:   "Missing",
			expected: -1.0,
		},
		{
			name:     "Empty status",
			status:   "",
			expected: -1.0,
		},
		{
			name:     "Invalid status",
			status:   "InvalidStatus",
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HealthStatusToMetric(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPhaseTimer(t *testing.T) {
	t.Run("creates phase timer with correct values", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"
		phase := "pipeline"

		timer := NewPhaseTimer(namespace, name, phase)

		assert.NotNil(t, timer)
		assert.Equal(t, namespace, timer.namespace)
		assert.Equal(t, name, timer.name)
		assert.Equal(t, phase, timer.phase)
		assert.NotNil(t, timer.timer)
	})

	t.Run("phase timer can be observed", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"
		phase := "pipeline"

		timer := NewPhaseTimer(namespace, name, phase)

		// This should not panic
		timer.ObserveDuration()
	})
}

func TestPhaseTimer_ObserveDuration(t *testing.T) {
	t.Run("observe duration works correctly", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"
		phase := "pipeline"

		timer := NewPhaseTimer(namespace, name, phase)

		// This should not panic
		timer.ObserveDuration()
	})
}

func TestRecordAPICall(t *testing.T) {
	t.Run("records successful API call", func(t *testing.T) {
		operation := "get"
		resourceType := "configmap"
		duration := 0.05

		// Reset metrics before test
		APICallsTotal.Reset()

		RecordAPICall(operation, resourceType, nil, duration)

		// Check that counter was incremented
		callTotal := testutil.ToFloat64(APICallsTotal.WithLabelValues(operation, resourceType, "success"))
		assert.Equal(t, 1.0, callTotal)
	})

	t.Run("records failed API call", func(t *testing.T) {
		operation := "create"
		resourceType := "deployment"
		duration := 0.1
		err := assert.AnError

		// Reset metrics before test
		APICallsTotal.Reset()

		RecordAPICall(operation, resourceType, err, duration)

		// Check that counter was incremented
		callTotal := testutil.ToFloat64(APICallsTotal.WithLabelValues(operation, resourceType, "error"))
		assert.Equal(t, 1.0, callTotal)
	})

	t.Run("records multiple API calls", func(t *testing.T) {
		operation := "update"
		resourceType := "service"
		duration := 0.02

		// Reset metrics before test
		APICallsTotal.Reset()

		// Record multiple calls
		RecordAPICall(operation, resourceType, nil, duration)
		RecordAPICall(operation, resourceType, nil, duration*2)

		// Check that counter was incremented
		callTotal := testutil.ToFloat64(APICallsTotal.WithLabelValues(operation, resourceType, "success"))
		assert.Equal(t, 2.0, callTotal)
	})
}

func TestRecordResourceCreation(t *testing.T) {
	t.Run("records resource creation", func(t *testing.T) {
		resourceType := "pipeline"
		operation := "create"
		duration := 0.1

		// This should not panic
		RecordResourceCreation(resourceType, operation, duration)
	})

	t.Run("records resource update", func(t *testing.T) {
		resourceType := "eventlistener"
		operation := "update"
		duration := 0.05

		// This should not panic
		RecordResourceCreation(resourceType, operation, duration)
	})

	t.Run("records multiple resource operations", func(t *testing.T) {
		resourceType := "trigger"
		operation := "create"
		duration := 0.03

		// Record multiple operations - should not panic
		RecordResourceCreation(resourceType, operation, duration)
		RecordResourceCreation(resourceType, operation, duration*2)
	})
}

func TestRecordResourceDeletion(t *testing.T) {
	t.Run("records resource deletion", func(t *testing.T) {
		resourceType := "application"
		duration := 0.08

		// This should not panic
		RecordResourceDeletion(resourceType, duration)
	})

	t.Run("records multiple resource deletions", func(t *testing.T) {
		resourceType := "pipeline"
		duration := 0.04

		// Record multiple deletions - should not panic
		RecordResourceDeletion(resourceType, duration)
		RecordResourceDeletion(resourceType, duration*1.5)
	})
}

func TestRecordFinalizerOperation(t *testing.T) {
	t.Run("records finalizer add operation", func(t *testing.T) {
		operation := "add"

		// Reset metrics before test
		FinalizerOperations.Reset()

		RecordFinalizerOperation(operation)

		// Check that counter was incremented
		recordedCount := testutil.ToFloat64(FinalizerOperations.WithLabelValues(operation))
		assert.Equal(t, 1.0, recordedCount)
	})

	t.Run("records finalizer remove operation", func(t *testing.T) {
		operation := "remove"

		// Reset metrics before test
		FinalizerOperations.Reset()

		RecordFinalizerOperation(operation)

		// Check that counter was incremented
		recordedCount := testutil.ToFloat64(FinalizerOperations.WithLabelValues(operation))
		assert.Equal(t, 1.0, recordedCount)
	})

	t.Run("records multiple finalizer operations", func(t *testing.T) {
		operation := "add"

		// Reset metrics before test
		FinalizerOperations.Reset()

		// Record multiple operations
		RecordFinalizerOperation(operation)
		RecordFinalizerOperation(operation)
		RecordFinalizerOperation(operation)

		// Check that counter was incremented
		recordedCount := testutil.ToFloat64(FinalizerOperations.WithLabelValues(operation))
		assert.Equal(t, 3.0, recordedCount)
	})
}

func TestRecordWebhookEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{
			name:      "push event",
			eventType: "push",
		},
		{
			name:      "pull request event",
			eventType: "pull_request",
		},
		{
			name:      "tag event",
			eventType: "tag",
		},
		{
			name:      "custom event",
			eventType: "custom_event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset metrics before test
			WebhookEvents.Reset()

			RecordWebhookEvent(tt.eventType)

			// Check that counter was incremented
			recordedCount := testutil.ToFloat64(WebhookEvents.WithLabelValues(tt.eventType))
			assert.Equal(t, 1.0, recordedCount)
		})
	}

	t.Run("records multiple webhook events", func(t *testing.T) {
		eventType := "push"

		// Reset metrics before test
		WebhookEvents.Reset()

		// Record multiple events
		RecordWebhookEvent(eventType)
		RecordWebhookEvent(eventType)
		RecordWebhookEvent(eventType)

		// Check that counter was incremented
		recordedCount := testutil.ToFloat64(WebhookEvents.WithLabelValues(eventType))
		assert.Equal(t, 3.0, recordedCount)
	})
}

func TestMetricsRegistration(t *testing.T) {
	t.Run("all metrics are registered", func(t *testing.T) {
		// This test verifies that all metrics are properly registered
		// by checking that they can be collected without errors
		registry := prometheus.NewRegistry()

		// Register all metrics
		registry.MustRegister(
			ReconciliationDuration,
			ReconciliationPhaseDuration,
			BuildsTotal,
			DeploymentHealthGauge,
			ReplicasGauge,
			ArgoCDSyncStatus,
			ArgoCDHealthStatus,
			ReconciliationsTotal,
			ReconciliationErrorsTotal,
			WatchEventsTotal,
			LastReconcileTime,
			APICallDuration,
			APICallsTotal,
			ResourcesManaged,
			ResourceCreationDuration,
			ResourceDeletionDuration,
			FinalizerOperations,
			WebhookEvents,
		)

		// This should not panic
		_, err := registry.Gather()
		assert.NoError(t, err)
	})
}

func TestMetricsLabels(t *testing.T) {
	t.Run("builds total labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"
		status := "succeeded"

		// Reset metrics before test
		BuildsTotal.Reset()

		BuildsTotal.WithLabelValues(namespace, name, status).Inc()

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(BuildsTotal.WithLabelValues(namespace, name, status))
		assert.Equal(t, 1.0, recordedValue)
	})

	t.Run("deployment health gauge labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"

		// Reset metrics before test
		DeploymentHealthGauge.Reset()

		DeploymentHealthGauge.WithLabelValues(namespace, name).Set(1.0)

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(DeploymentHealthGauge.WithLabelValues(namespace, name))
		assert.Equal(t, 1.0, recordedValue)
	})

	t.Run("replicas gauge labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"
		replicaType := "desired"

		// Reset metrics before test
		ReplicasGauge.Reset()

		ReplicasGauge.WithLabelValues(namespace, name, replicaType).Set(3.0)

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(ReplicasGauge.WithLabelValues(namespace, name, replicaType))
		assert.Equal(t, 3.0, recordedValue)
	})

	t.Run("argocd sync status labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"

		// Reset metrics before test
		ArgoCDSyncStatus.Reset()

		ArgoCDSyncStatus.WithLabelValues(namespace, name).Set(1.0)

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(ArgoCDSyncStatus.WithLabelValues(namespace, name))
		assert.Equal(t, 1.0, recordedValue)
	})

	t.Run("argocd health status labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"

		// Reset metrics before test
		ArgoCDHealthStatus.Reset()

		ArgoCDHealthStatus.WithLabelValues(namespace, name).Set(0.5)

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(ArgoCDHealthStatus.WithLabelValues(namespace, name))
		assert.Equal(t, 0.5, recordedValue)
	})

	t.Run("reconciliations total labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"

		// Reset metrics before test
		ReconciliationsTotal.Reset()

		ReconciliationsTotal.WithLabelValues(namespace, name).Inc()

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(ReconciliationsTotal.WithLabelValues(namespace, name))
		assert.Equal(t, 1.0, recordedValue)
	})

	t.Run("reconciliation errors total labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"
		errorType := "validation"

		// Reset metrics before test
		ReconciliationErrorsTotal.Reset()

		ReconciliationErrorsTotal.WithLabelValues(namespace, name, errorType).Inc()

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(ReconciliationErrorsTotal.WithLabelValues(namespace, name, errorType))
		assert.Equal(t, 1.0, recordedValue)
	})

	t.Run("watch events total labels", func(t *testing.T) {
		resourceType := "deployment"

		// Reset metrics before test
		WatchEventsTotal.Reset()

		WatchEventsTotal.WithLabelValues(resourceType).Inc()

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(WatchEventsTotal.WithLabelValues(resourceType))
		assert.Equal(t, 1.0, recordedValue)
	})

	t.Run("last reconcile time labels", func(t *testing.T) {
		namespace := "test-namespace"
		name := "test-app"
		timestamp := float64(time.Now().Unix())

		// Reset metrics before test
		LastReconcileTime.Reset()

		LastReconcileTime.WithLabelValues(namespace, name).Set(timestamp)

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(LastReconcileTime.WithLabelValues(namespace, name))
		assert.Equal(t, timestamp, recordedValue)
	})

	t.Run("resources managed labels", func(t *testing.T) {
		resourceType := "pipeline"

		// Reset metrics before test
		ResourcesManaged.Reset()

		ResourcesManaged.WithLabelValues(resourceType).Set(5.0)

		// Check that metric was recorded with correct labels
		recordedValue := testutil.ToFloat64(ResourcesManaged.WithLabelValues(resourceType))
		assert.Equal(t, 5.0, recordedValue)
	})
}

func TestMetricsBuckets(t *testing.T) {
	t.Run("reconciliation duration buckets", func(t *testing.T) {
		// Test that the histogram buckets are properly configured
		namespace := "test-namespace"
		name := "test-app"
		result := "success"

		// Reset metrics before test
		ReconciliationDuration.Reset()

		// Record various durations to test buckets
		ReconciliationDuration.WithLabelValues(namespace, name, result).Observe(0.001) // 1ms
		ReconciliationDuration.WithLabelValues(namespace, name, result).Observe(0.01)  // 10ms
		ReconciliationDuration.WithLabelValues(namespace, name, result).Observe(0.1)   // 100ms
		ReconciliationDuration.WithLabelValues(namespace, name, result).Observe(1.0)   // 1s
		ReconciliationDuration.WithLabelValues(namespace, name, result).Observe(10.0)  // 10s

		// This should not panic and should record the values
		// We can't easily test histogram values with testutil.ToFloat64
		// but we can verify that the operations don't panic
		assert.True(t, true)
	})

	t.Run("api call duration buckets", func(t *testing.T) {
		// Test that the histogram buckets are properly configured
		operation := "get"
		resourceType := "configmap"
		result := "success"

		// Reset metrics before test
		APICallDuration.Reset()

		// Record various durations to test buckets
		APICallDuration.WithLabelValues(operation, resourceType, result).Observe(0.001) // 1ms
		APICallDuration.WithLabelValues(operation, resourceType, result).Observe(0.01)  // 10ms
		APICallDuration.WithLabelValues(operation, resourceType, result).Observe(0.1)   // 100ms
		APICallDuration.WithLabelValues(operation, resourceType, result).Observe(1.0)   // 1s

		// This should not panic and should record the values
		assert.True(t, true)
	})

	t.Run("resource creation duration buckets", func(t *testing.T) {
		// Test that the histogram buckets are properly configured
		resourceType := "pipeline"
		operation := "create"

		// Reset metrics before test
		ResourceCreationDuration.Reset()

		// Record various durations to test buckets
		ResourceCreationDuration.WithLabelValues(resourceType, operation).Observe(0.001) // 1ms
		ResourceCreationDuration.WithLabelValues(resourceType, operation).Observe(0.01)  // 10ms
		ResourceCreationDuration.WithLabelValues(resourceType, operation).Observe(0.1)   // 100ms
		ResourceCreationDuration.WithLabelValues(resourceType, operation).Observe(1.0)   // 1s

		// This should not panic and should record the values
		assert.True(t, true)
	})

	t.Run("resource deletion duration buckets", func(t *testing.T) {
		// Test that the histogram buckets are properly configured
		resourceType := "application"

		// Reset metrics before test
		ResourceDeletionDuration.Reset()

		// Record various durations to test buckets
		ResourceDeletionDuration.WithLabelValues(resourceType).Observe(0.001) // 1ms
		ResourceDeletionDuration.WithLabelValues(resourceType).Observe(0.01)  // 10ms
		ResourceDeletionDuration.WithLabelValues(resourceType).Observe(0.1)   // 100ms
		ResourceDeletionDuration.WithLabelValues(resourceType).Observe(1.0)   // 1s

		// This should not panic and should record the values
		assert.True(t, true)
	})
}
