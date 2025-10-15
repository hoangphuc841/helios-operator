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
)

// init registers all metrics with the controller-runtime metrics registry
func init() {
	metrics.Registry.MustRegister(
		ReconciliationDuration,
		BuildsTotal,
		DeploymentHealthGauge,
		ReplicasGauge,
		ArgoCDSyncStatus,
		ArgoCDHealthStatus,
		ReconciliationsTotal,
		WatchEventsTotal,
		LastReconcileTime,
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
