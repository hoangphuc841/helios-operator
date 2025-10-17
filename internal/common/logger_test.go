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
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

// mockLogSink is a mock implementation of logr.LogSink for testing
type mockLogSink struct {
	infoCalled  bool
	errorCalled bool
	lastLevel   int
	lastMessage string
	lastKeys    []interface{}
}

func (m *mockLogSink) Init(info logr.RuntimeInfo) {}

func (m *mockLogSink) Enabled(level int) bool {
	return true
}

func (m *mockLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	m.infoCalled = true
	m.lastLevel = level
	m.lastMessage = msg
	m.lastKeys = keysAndValues
}

func (m *mockLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	m.errorCalled = true
	m.lastMessage = msg
	m.lastKeys = keysAndValues
}

func (m *mockLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	// Return the same sink to maintain state
	return m
}

func (m *mockLogSink) WithName(name string) logr.LogSink {
	// Return the same sink to maintain state
	return m
}

// Ensure mockLogSink implements logr.LogSink interface
var _ logr.LogSink = &mockLogSink{}

// createMockLogger creates a logger with a mock sink for testing
func createMockLogger() (*Logger, *mockLogSink) {
	mock := &mockLogSink{}
	return &Logger{Logger: logr.New(mock)}, mock
}

func TestNewLogger(t *testing.T) {
	t.Run("creates logger with context", func(t *testing.T) {
		ctx := context.Background()
		logger := NewLogger(ctx, "test-logger")

		assert.NotNil(t, logger)
		assert.NotNil(t, logger.Logger)
	})
}

func TestLogger_WithHeliosApp(t *testing.T) {
	t.Run("adds HeliosApp context", func(t *testing.T) {
		logger, _ := createMockLogger()

		result := logger.WithHeliosApp("test-app", "test-namespace")

		assert.NotNil(t, result)
		// The result should be a valid logger (may be same or different instance)
		assert.NotNil(t, result.Logger)
	})
}

func TestLogger_WithResource(t *testing.T) {
	t.Run("adds resource context", func(t *testing.T) {
		logger, _ := createMockLogger()

		result := logger.WithResource("Pipeline", "test-pipeline", "test-namespace")

		assert.NotNil(t, result)
		// The result should be a valid logger (may be same or different instance)
		assert.NotNil(t, result.Logger)
	})
}

func TestLogger_WithOperation(t *testing.T) {
	t.Run("adds operation context", func(t *testing.T) {
		logger, _ := createMockLogger()

		result := logger.WithOperation("create")

		assert.NotNil(t, result)
		// The result should be a valid logger (may be same or different instance)
		assert.NotNil(t, result.Logger)
	})
}

func TestLogger_WithPhase(t *testing.T) {
	t.Run("adds phase context", func(t *testing.T) {
		logger, _ := createMockLogger()

		result := logger.WithPhase("Pipeline")

		assert.NotNil(t, result)
		// The result should be a valid logger (may be same or different instance)
		assert.NotNil(t, result.Logger)
	})
}

func TestLogger_LogReconcileStart(t *testing.T) {
	t.Run("logs reconcile start", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogReconcileStart("test-app", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Starting reconcile", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "heliosapp")
		assert.Contains(t, mock.lastKeys, "test-app")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogReconcileEnd(t *testing.T) {
	t.Run("logs reconcile end with error", func(t *testing.T) {
		logger, mock := createMockLogger()
		err := assert.AnError
		duration := 5 * time.Second

		logger.LogReconcileEnd("test-app", "test-namespace", duration, err)

		assert.True(t, mock.errorCalled)
		assert.Equal(t, "Reconcile failed", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "heliosapp")
		assert.Contains(t, mock.lastKeys, "test-app")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
		assert.Contains(t, mock.lastKeys, "duration")
		assert.Contains(t, mock.lastKeys, duration)
	})

	t.Run("logs reconcile end without error", func(t *testing.T) {
		logger, mock := createMockLogger()
		duration := 5 * time.Second

		logger.LogReconcileEnd("test-app", "test-namespace", duration, nil)

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Reconcile completed", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "heliosapp")
		assert.Contains(t, mock.lastKeys, "test-app")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
		assert.Contains(t, mock.lastKeys, "duration")
		assert.Contains(t, mock.lastKeys, duration)
	})
}

func TestLogger_LogResourceCreated(t *testing.T) {
	t.Run("logs resource creation", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogResourceCreated("Pipeline", "test-pipeline", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Resource created", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "kind")
		assert.Contains(t, mock.lastKeys, "Pipeline")
		assert.Contains(t, mock.lastKeys, "name")
		assert.Contains(t, mock.lastKeys, "test-pipeline")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogResourceUpdated(t *testing.T) {
	t.Run("logs resource update", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogResourceUpdated("Pipeline", "test-pipeline", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Resource updated", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "kind")
		assert.Contains(t, mock.lastKeys, "Pipeline")
		assert.Contains(t, mock.lastKeys, "name")
		assert.Contains(t, mock.lastKeys, "test-pipeline")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogResourceDeleted(t *testing.T) {
	t.Run("logs resource deletion", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogResourceDeleted("Pipeline", "test-pipeline", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Resource deleted", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "kind")
		assert.Contains(t, mock.lastKeys, "Pipeline")
		assert.Contains(t, mock.lastKeys, "name")
		assert.Contains(t, mock.lastKeys, "test-pipeline")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogPhaseTransition(t *testing.T) {
	t.Run("logs phase transition", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogPhaseTransition("Pipeline", "Triggers")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Phase transition", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "from")
		assert.Contains(t, mock.lastKeys, "Pipeline")
		assert.Contains(t, mock.lastKeys, "to")
		assert.Contains(t, mock.lastKeys, "Triggers")
	})
}

func TestLogger_LogError(t *testing.T) {
	t.Run("logs error with context", func(t *testing.T) {
		logger, mock := createMockLogger()
		err := assert.AnError

		logger.LogError(err, "test error", "key1", "value1", "key2", "value2")

		assert.True(t, mock.errorCalled)
		assert.Equal(t, "test error", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "key1")
		assert.Contains(t, mock.lastKeys, "value1")
		assert.Contains(t, mock.lastKeys, "key2")
		assert.Contains(t, mock.lastKeys, "value2")
	})
}

func TestLogger_LogWarning(t *testing.T) {
	t.Run("logs warning with context", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogWarning("test warning", "key1", "value1", "key2", "value2")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "WARNING: test warning", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "key1")
		assert.Contains(t, mock.lastKeys, "value1")
		assert.Contains(t, mock.lastKeys, "key2")
		assert.Contains(t, mock.lastKeys, "value2")
	})
}

func TestLogger_LogInfo(t *testing.T) {
	t.Run("logs info with context", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogInfo("test info", "key1", "value1", "key2", "value2")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "test info", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "key1")
		assert.Contains(t, mock.lastKeys, "value1")
		assert.Contains(t, mock.lastKeys, "key2")
		assert.Contains(t, mock.lastKeys, "value2")
	})
}

func TestLogger_LogDebug(t *testing.T) {
	t.Run("logs debug with context", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogDebug("test debug", "key1", "value1", "key2", "value2")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, 1, mock.lastLevel) // V(1) should set level to 1
		assert.Equal(t, "test debug", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "key1")
		assert.Contains(t, mock.lastKeys, "value1")
		assert.Contains(t, mock.lastKeys, "key2")
		assert.Contains(t, mock.lastKeys, "value2")
	})
}

func TestLogger_LogTrace(t *testing.T) {
	t.Run("logs trace with context", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogTrace("test trace", "key1", "value1", "key2", "value2")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, 2, mock.lastLevel) // V(2) should set level to 2
		assert.Equal(t, "test trace", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "key1")
		assert.Contains(t, mock.lastKeys, "value1")
		assert.Contains(t, mock.lastKeys, "key2")
		assert.Contains(t, mock.lastKeys, "value2")
	})
}

func TestLogger_LogWebhookValidation(t *testing.T) {
	t.Run("logs webhook validation", func(t *testing.T) {
		logger, mock := createMockLogger()
		duration := 100 * time.Millisecond

		logger.LogWebhookValidation("create", "success", duration)

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Webhook validation", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "operation")
		assert.Contains(t, mock.lastKeys, "create")
		assert.Contains(t, mock.lastKeys, "result")
		assert.Contains(t, mock.lastKeys, "success")
		assert.Contains(t, mock.lastKeys, "duration")
		assert.Contains(t, mock.lastKeys, duration)
	})
}

func TestLogger_LogMetrics(t *testing.T) {
	t.Run("logs metrics with labels", func(t *testing.T) {
		logger, mock := createMockLogger()
		labels := map[string]string{
			"label1": "value1",
			"label2": "value2",
		}

		logger.LogMetrics("test_metric", 42.5, labels)

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Metrics", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "metric")
		assert.Contains(t, mock.lastKeys, "test_metric")
		assert.Contains(t, mock.lastKeys, "value")
		assert.Contains(t, mock.lastKeys, 42.5)
		assert.Contains(t, mock.lastKeys, "label1")
		assert.Contains(t, mock.lastKeys, "value1")
		assert.Contains(t, mock.lastKeys, "label2")
		assert.Contains(t, mock.lastKeys, "value2")
	})

	t.Run("logs metrics without labels", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogMetrics("test_metric", 42.5, nil)

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Metrics", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "metric")
		assert.Contains(t, mock.lastKeys, "test_metric")
		assert.Contains(t, mock.lastKeys, "value")
		assert.Contains(t, mock.lastKeys, 42.5)
	})
}

func TestLogger_LogHealthCheck(t *testing.T) {
	t.Run("logs health check", func(t *testing.T) {
		logger, mock := createMockLogger()
		duration := 50 * time.Millisecond

		logger.LogHealthCheck("liveness", "healthy", duration)

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Health check", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "check")
		assert.Contains(t, mock.lastKeys, "liveness")
		assert.Contains(t, mock.lastKeys, "status")
		assert.Contains(t, mock.lastKeys, "healthy")
		assert.Contains(t, mock.lastKeys, "duration")
		assert.Contains(t, mock.lastKeys, duration)
	})
}

func TestLogger_LogCertificateEvent(t *testing.T) {
	t.Run("logs certificate event", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogCertificateEvent("created", "test-cert", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Certificate event", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "event")
		assert.Contains(t, mock.lastKeys, "created")
		assert.Contains(t, mock.lastKeys, "certificate")
		assert.Contains(t, mock.lastKeys, "test-cert")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogWebhookEvent(t *testing.T) {
	t.Run("logs webhook event", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogWebhookEvent("created", "test-webhook", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Webhook event", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "event")
		assert.Contains(t, mock.lastKeys, "created")
		assert.Contains(t, mock.lastKeys, "webhook")
		assert.Contains(t, mock.lastKeys, "test-webhook")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogControllerEvent(t *testing.T) {
	t.Run("logs controller event", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogControllerEvent("started", "test-controller", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Controller event", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "event")
		assert.Contains(t, mock.lastKeys, "started")
		assert.Contains(t, mock.lastKeys, "controller")
		assert.Contains(t, mock.lastKeys, "test-controller")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogReconcilePhase(t *testing.T) {
	t.Run("logs reconcile phase", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogReconcilePhase("Pipeline", "Starting pipeline creation", "key1", "value1")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Reconcile phase: Starting pipeline creation", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "phase")
		assert.Contains(t, mock.lastKeys, "Pipeline")
		assert.Contains(t, mock.lastKeys, "key1")
		assert.Contains(t, mock.lastKeys, "value1")
	})
}

func TestLogger_LogResourceStatus(t *testing.T) {
	t.Run("logs resource status", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogResourceStatus("Pipeline", "test-pipeline", "test-namespace", "Ready")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Resource status", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "kind")
		assert.Contains(t, mock.lastKeys, "Pipeline")
		assert.Contains(t, mock.lastKeys, "name")
		assert.Contains(t, mock.lastKeys, "test-pipeline")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
		assert.Contains(t, mock.lastKeys, "status")
		assert.Contains(t, mock.lastKeys, "Ready")
	})
}

func TestLogger_LogNamespaceEvent(t *testing.T) {
	t.Run("logs namespace event", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogNamespaceEvent("created", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Namespace event", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "event")
		assert.Contains(t, mock.lastKeys, "created")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_LogFinalizerEvent(t *testing.T) {
	t.Run("logs finalizer event", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogFinalizerEvent("added", "test-finalizer", "test-resource", "test-namespace")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "Finalizer event", mock.lastMessage)
		assert.Contains(t, mock.lastKeys, "event")
		assert.Contains(t, mock.lastKeys, "added")
		assert.Contains(t, mock.lastKeys, "finalizer")
		assert.Contains(t, mock.lastKeys, "test-finalizer")
		assert.Contains(t, mock.lastKeys, "name")
		assert.Contains(t, mock.lastKeys, "test-resource")
		assert.Contains(t, mock.lastKeys, "namespace")
		assert.Contains(t, mock.lastKeys, "test-namespace")
	})
}

func TestLogger_Chaining(t *testing.T) {
	t.Run("logger chaining works correctly", func(t *testing.T) {
		logger, mock := createMockLogger()

		// Chain multiple context methods
		chainedLogger := logger.
			WithHeliosApp("test-app", "test-namespace").
			WithResource("Pipeline", "test-pipeline", "test-namespace").
			WithOperation("create").
			WithPhase("Pipeline")

		// Chaining should return a logger (may be same or different instance)
		assert.NotNil(t, chainedLogger)

		// The chained logger should still work
		chainedLogger.LogInfo("test message")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "test message", mock.lastMessage)
	})
}

func TestLogger_ContextPreservation(t *testing.T) {
	t.Run("context is preserved across operations", func(t *testing.T) {
		logger, mock := createMockLogger()

		// Create a logger with context
		contextLogger := logger.WithHeliosApp("test-app", "test-namespace")

		// Use the context logger
		contextLogger.LogInfo("test message", "additional", "value")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "test message", mock.lastMessage)
		// The mock doesn't track the context values, but we can verify the method was called
	})
}

func TestLogger_ErrorHandling(t *testing.T) {
	t.Run("handles nil error gracefully", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogError(nil, "test error")

		assert.True(t, mock.errorCalled)
		assert.Equal(t, "test error", mock.lastMessage)
	})

	t.Run("handles empty message gracefully", func(t *testing.T) {
		logger, mock := createMockLogger()

		logger.LogInfo("")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, "", mock.lastMessage)
	})
}

func TestLogger_LogLevels(t *testing.T) {
	t.Run("V() returns logger with correct level", func(t *testing.T) {
		logger, mock := createMockLogger()

		// Test V(1) for debug
		debugLogger := logger.V(1)
		debugLogger.Info("debug message")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, 1, mock.lastLevel)
		assert.Equal(t, "debug message", mock.lastMessage)

		// Reset mock
		mock.infoCalled = false
		mock.lastLevel = 0
		mock.lastMessage = ""

		// Test V(2) for trace
		traceLogger := logger.V(2)
		traceLogger.Info("trace message")

		assert.True(t, mock.infoCalled)
		assert.Equal(t, 2, mock.lastLevel)
		assert.Equal(t, "trace message", mock.lastMessage)
	})
}
