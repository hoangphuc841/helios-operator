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
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Logger provides enhanced logging functionality.
type Logger struct {
	logr.Logger
}

// NewLogger creates a new Logger with context.
func NewLogger(ctx context.Context, name string) *Logger {
	return &Logger{
		Logger: log.FromContext(ctx).WithName(name),
	}
}

// WithHeliosApp creates a logger with HeliosApp context.
func (l *Logger) WithHeliosApp(name, namespace string) *Logger {
	return &Logger{
		Logger: l.Logger.WithValues("heliosapp", name, "namespace", namespace),
	}
}

// WithResource creates a logger with resource context.
func (l *Logger) WithResource(kind, name, namespace string) *Logger {
	return &Logger{
		Logger: l.Logger.WithValues("kind", kind, "name", name, "namespace", namespace),
	}
}

// WithOperation creates a logger with operation context.
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.WithValues("operation", operation),
	}
}

// WithPhase creates a logger with phase context.
func (l *Logger) WithPhase(phase string) *Logger {
	return &Logger{
		Logger: l.Logger.WithValues("phase", phase),
	}
}

// LogReconcileStart logs the start of a reconcile operation.
func (l *Logger) LogReconcileStart(name, namespace string) {
	l.Info("Starting reconcile", "heliosapp", name, "namespace", namespace)
}

// LogReconcileEnd logs the end of a reconcile operation.
func (l *Logger) LogReconcileEnd(name, namespace string, duration time.Duration, err error) {
	if err != nil {
		l.Error(err, "Reconcile failed", "heliosapp", name, "namespace", namespace, "duration", duration)
	} else {
		l.Info("Reconcile completed", "heliosapp", name, "namespace", namespace, "duration", duration)
	}
}

// LogResourceCreated logs resource creation.
func (l *Logger) LogResourceCreated(kind, name, namespace string) {
	l.Info("Resource created", "kind", kind, "name", name, "namespace", namespace)
}

// LogResourceUpdated logs resource update.
func (l *Logger) LogResourceUpdated(kind, name, namespace string) {
	l.Info("Resource updated", "kind", kind, "name", name, "namespace", namespace)
}

// LogResourceDeleted logs resource deletion.
func (l *Logger) LogResourceDeleted(kind, name, namespace string) {
	l.Info("Resource deleted", "kind", kind, "name", name, "namespace", namespace)
}

// LogPhaseTransition logs phase transition.
func (l *Logger) LogPhaseTransition(from, to string) {
	l.Info("Phase transition", "from", from, "to", to)
}

// LogError logs an error with context.
func (l *Logger) LogError(err error, message string, keysAndValues ...interface{}) {
	l.Error(err, message, keysAndValues...)
}

// LogWarning logs a warning with context.
func (l *Logger) LogWarning(message string, keysAndValues ...interface{}) {
	l.Info("WARNING: "+message, keysAndValues...)
}

// LogInfo logs info with context.
func (l *Logger) LogInfo(message string, keysAndValues ...interface{}) {
	l.Info(message, keysAndValues...)
}

// LogDebug logs debug info with context.
func (l *Logger) LogDebug(message string, keysAndValues ...interface{}) {
	l.V(1).Info(message, keysAndValues...)
}

// LogTrace logs trace info with context.
func (l *Logger) LogTrace(message string, keysAndValues ...interface{}) {
	l.V(2).Info(message, keysAndValues...)
}

// LogWebhookValidation logs webhook validation.
func (l *Logger) LogWebhookValidation(operation, result string, duration time.Duration) {
	l.Info("Webhook validation", "operation", operation, "result", result, "duration", duration)
}

// LogMetrics logs metrics.
func (l *Logger) LogMetrics(metric string, value float64, labels map[string]string) {
	keysAndValues := make([]interface{}, 0, len(labels)*2+2)
	keysAndValues = append(keysAndValues, "metric", metric, "value", value)
	for k, v := range labels {
		keysAndValues = append(keysAndValues, k, v)
	}
	l.Info("Metrics", keysAndValues...)
}

// LogHealthCheck logs health check results.
func (l *Logger) LogHealthCheck(check, status string, duration time.Duration) {
	l.Info("Health check", "check", check, "status", status, "duration", duration)
}

// LogCertificateEvent logs certificate events.
func (l *Logger) LogCertificateEvent(event, certName, namespace string) {
	l.Info("Certificate event", "event", event, "certificate", certName, "namespace", namespace)
}

// LogWebhookEvent logs webhook events.
func (l *Logger) LogWebhookEvent(event, webhookName, namespace string) {
	l.Info("Webhook event", "event", event, "webhook", webhookName, "namespace", namespace)
}

// LogControllerEvent logs controller events.
func (l *Logger) LogControllerEvent(event, controllerName, namespace string) {
	l.Info("Controller event", "event", event, "controller", controllerName, "namespace", namespace)
}

// LogReconcilePhase logs reconcile phase.
func (l *Logger) LogReconcilePhase(phase, message string, keysAndValues ...interface{}) {
	allKeysAndValues := append([]interface{}{"phase", phase}, keysAndValues...)
	l.Info("Reconcile phase: "+message, allKeysAndValues...)
}

// LogResourceStatus logs resource status.
func (l *Logger) LogResourceStatus(kind, name, namespace, status string) {
	l.Info("Resource status", "kind", kind, "name", name, "namespace", namespace, "status", status)
}

// LogNamespaceEvent logs namespace events.
func (l *Logger) LogNamespaceEvent(event, namespace string) {
	l.Info("Namespace event", "event", event, "namespace", namespace)
}

// LogFinalizerEvent logs finalizer events.
func (l *Logger) LogFinalizerEvent(event, finalizer, name, namespace string) {
	l.Info("Finalizer event", "event", event, "finalizer", finalizer, "name", name, "namespace", namespace)
}
