/*
Copyright 2025.
*/

package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReconciliationDuration tracks how long reconciliations take
	ReconciliationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "heliosapp_reconciliation_duration_seconds",
			Help:    "Time spent reconciling HeliosApp resources",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~32s
		},
		[]string{"namespace", "name", "result"}, // result: success, error, requeue
	)

	// ReconciliationPhaseDuration tracks duration of each reconciliation phase
	ReconciliationPhaseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "heliosapp_reconciliation_phase_duration_seconds",
			Help:    "Time spent in each reconciliation phase",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"namespace", "name", "phase"}, // phase: fetch, pipeline, triggers, argocd, status
	)

	// BuildsTotal tracks total number of builds per status
	BuildsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_builds_total",
			Help: "Total number of builds by status",
		},
		[]string{"namespace", "name", "status"}, // status: succeeded, failed, running
	)

	// DeploymentHealthGauge tracks current deployment health
	DeploymentHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heliosapp_deployment_health",
			Help: "Deployment health status (1=Healthy, 0.5=Progressing, 0=Degraded, -1=Unknown)",
		},
		[]string{"namespace", "name"},
	)

	// ReplicasGauge tracks replica counts
	ReplicasGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heliosapp_replicas",
			Help: "Number of replicas",
		},
		[]string{"namespace", "name", "type"}, // type: desired, ready
	)

	// ArgoCDSyncStatus tracks ArgoCD sync status
	ArgoCDSyncStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heliosapp_argocd_sync_status",
			Help: "ArgoCD sync status (1=Synced, 0=OutOfSync, -1=Unknown)",
		},
		[]string{"namespace", "name"},
	)

	// ArgoCDHealthStatus tracks ArgoCD health status
	ArgoCDHealthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heliosapp_argocd_health_status",
			Help: "ArgoCD health status (1=Healthy, 0.5=Progressing, 0=Degraded, -1=Unknown)",
		},
		[]string{"namespace", "name"},
	)

	// ReconciliationsTotal tracks total reconciliation attempts
	ReconciliationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_reconciliations_total",
			Help: "Total number of reconciliations",
		},
		[]string{"namespace", "name"},
	)

	// ReconciliationErrorsTotal tracks reconciliation errors by type
	ReconciliationErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_reconciliation_errors_total",
			Help: "Total number of reconciliation errors by error type",
		},
		[]string{"namespace", "name", "error_type"}, // error_type: fetch, pipeline, triggers, argocd, status, validation
	)

	// WatchEventsTotal tracks watch event triggers
	WatchEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_watch_events_total",
			Help: "Total number of watch events that triggered reconciliation",
		},
		[]string{"resource_type"}, // resource_type: argocd_app, pipelinerun, deployment
	)

	// LastReconcileTime tracks timestamp of last reconciliation
	LastReconcileTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heliosapp_last_reconcile_timestamp_seconds",
			Help: "Unix timestamp of last successful reconciliation",
		},
		[]string{"namespace", "name"},
	)

	// APICallDuration tracks duration of Kubernetes API calls
	APICallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "heliosapp_api_call_duration_seconds",
			Help:    "Duration of Kubernetes API calls",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~4s
		},
		[]string{"operation", "resource_type", "result"}, // operation: get, create, update, delete; result: success, error
	)

	// APICallsTotal tracks total API calls
	APICallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_api_calls_total",
			Help: "Total number of Kubernetes API calls",
		},
		[]string{"operation", "resource_type", "result"},
	)

	// ResourcesManaged tracks number of resources managed by the operator
	ResourcesManaged = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heliosapp_resources_managed_total",
			Help: "Total number of resources managed by the operator",
		},
		[]string{"resource_type"}, // resource_type: pipeline, eventlistener, trigger, application
	)

	// ResourceCreationDuration tracks time to create/update individual resources
	ResourceCreationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "heliosapp_resource_creation_duration_seconds",
			Help:    "Time spent creating or updating individual resources",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~4s
		},
		[]string{"resource_type", "operation"}, // resource_type: pipeline, eventlistener, trigger, application; operation: create, update
	)

	// ResourceDeletionDuration tracks time to delete resources during cleanup
	ResourceDeletionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "heliosapp_resource_deletion_duration_seconds",
			Help:    "Time spent deleting resources during cleanup",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~4s
		},
		[]string{"resource_type"}, // resource_type: pipeline, eventlistener, trigger, application
	)

	// FinalizerOperations tracks finalizer add/remove operations
	FinalizerOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_finalizer_operations_total",
			Help: "Total number of finalizer operations",
		},
		[]string{"operation"}, // operation: add, remove
	)

	// WebhookEvents tracks webhook events received
	WebhookEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_webhook_events_total",
			Help: "Total number of webhook events received",
		},
		[]string{"event_type"}, // event_type: push, pull_request, tag
	)
)

// init registers all metrics with the controller-runtime metrics registry
func init() {
	metrics.Registry.MustRegister(
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
}

// Helper functions to convert status strings to metric values

// DeploymentHealthToMetric converts deployment health string to numeric value
func DeploymentHealthToMetric(health string) float64 {
	switch health {
	case "Healthy":
		return 1.0
	case "Progressing":
		return 0.5
	case "Degraded":
		return 0.0
	default:
		return -1.0 // Unknown
	}
}

// SyncStatusToMetric converts ArgoCD sync status to numeric value
func SyncStatusToMetric(status string) float64 {
	switch status {
	case "Synced":
		return 1.0
	case "OutOfSync":
		return 0.0
	default:
		return -1.0 // Unknown
	}
}

// HealthStatusToMetric converts ArgoCD health status to numeric value
func HealthStatusToMetric(status string) float64 {
	switch status {
	case "Healthy":
		return 1.0
	case "Progressing":
		return 0.5
	case "Degraded":
		return 0.0
	default:
		return -1.0 // Unknown, Missing
	}
}

// PhaseTimer helps track duration of reconciliation phases
type PhaseTimer struct {
	namespace string
	name      string
	phase     string
	timer     *prometheus.Timer
}

// NewPhaseTimer creates a new phase timer and starts tracking
func NewPhaseTimer(namespace, name, phase string) *PhaseTimer {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		ReconciliationPhaseDuration.WithLabelValues(namespace, name, phase).Observe(v)
	}))

	return &PhaseTimer{
		namespace: namespace,
		name:      name,
		phase:     phase,
		timer:     timer,
	}
}

// ObserveDuration stops the timer and records the duration
func (pt *PhaseTimer) ObserveDuration() {
	pt.timer.ObserveDuration()
}

// RecordAPICall is a convenience function to record API call metrics
func RecordAPICall(operation, resourceType string, err error, duration float64) {
	result := "success"
	if err != nil {
		result = "error"
	}

	APICallDuration.WithLabelValues(operation, resourceType, result).Observe(duration)
	APICallsTotal.WithLabelValues(operation, resourceType, result).Inc()
}

// RecordResourceCreation is a convenience function to record resource creation/update metrics
func RecordResourceCreation(resourceType, operation string, duration float64) {
	ResourceCreationDuration.WithLabelValues(resourceType, operation).Observe(duration)
}

// RecordResourceDeletion is a convenience function to record resource deletion metrics
func RecordResourceDeletion(resourceType string, duration float64) {
	ResourceDeletionDuration.WithLabelValues(resourceType).Observe(duration)
}

// RecordFinalizerOperation is a convenience function to record finalizer operations
func RecordFinalizerOperation(operation string) {
	FinalizerOperations.WithLabelValues(operation).Inc()
}

// RecordWebhookEvent is a convenience function to record webhook events
func RecordWebhookEvent(eventType string) {
	WebhookEvents.WithLabelValues(eventType).Inc()
}
