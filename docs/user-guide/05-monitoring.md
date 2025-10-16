# 📊 Monitoring and Observability

This guide covers how to monitor and observe Helios Operator and the applications it manages.

## Table of Contents

1. [Overview](#overview)
2. [Prometheus Metrics](#prometheus-metrics)
3. [Grafana Dashboards](#grafana-dashboards)
4. [Alerting](#alerting)
5. [Logging](#logging)
6. [Best Practices](#best-practices)

---

## Overview

Helios Operator provides comprehensive observability through:

- **Prometheus Metrics**: Detailed metrics about operator performance and application status
- **Structured Logging**: Detailed logs with contextual information
- **Health Checks**: Liveness and readiness probes for operator health
- **Status Conditions**: Kubernetes-native status reporting

---

## Prometheus Metrics

### Operator Metrics

The Helios Operator exposes the following Prometheus metrics on port `8080` at the `/metrics` endpoint:

#### Reconciliation Metrics

```
# Reconciliation duration histogram (seconds)
helios_operator_reconciliation_duration_seconds{namespace,name}

# Reconciliation errors counter
helios_operator_reconciliation_errors_total{namespace,name,phase,error_type}

# Currently reconciling resources gauge
helios_operator_reconciling_resources{namespace}
```

#### Phase-Specific Metrics

```
# Duration of each reconciliation phase (seconds)
helios_operator_reconciliation_phase_duration_seconds{namespace,name,phase}

# Phase breakdown:
# - fetch: Fetching HeliosApp resource
# - pipeline: Creating/updating Tekton Pipeline
# - triggers: Creating/updating Tekton Triggers
# - argocd: Creating/updating ArgoCD Application
# - status: Updating HeliosApp status
```

#### API Call Metrics

```
# Kubernetes API call duration (seconds)
helios_operator_api_call_duration_seconds{operation,resource}

# API call counter
helios_operator_api_calls_total{operation,resource,status}
```

#### Resource Management Metrics

```
# Number of resources managed by the operator
helios_operator_resources_managed{type}

# Resource types: heliosapp, pipeline, eventlistener, argocd_application
```

#### Webhook Metrics

```
# Webhook validation counter
helios_operator_webhook_validations_total{operation,result}

# Webhook validation duration (seconds)
helios_operator_webhook_validation_duration_seconds{operation}

# Operations: create, update, delete
# Results: success, failure
```

### Setting Up Prometheus

#### Option 1: Prometheus Operator

If using Prometheus Operator, a ServiceMonitor is automatically created:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: helios-operator-metrics
  namespace: helios-operator-system
spec:
  endpoints:
  - interval: 30s
    path: /metrics
    port: metrics
    scheme: http
  selector:
    matchLabels:
      control-plane: controller-manager
```

#### Option 2: Prometheus ConfigMap

Add to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'helios-operator'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - helios-operator-system
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_control_plane]
        action: keep
        regex: controller-manager
      - source_labels: [__meta_kubernetes_pod_container_port_name]
        action: keep
        regex: metrics
```

### Querying Metrics

#### Average Reconciliation Duration

```promql
rate(helios_operator_reconciliation_duration_seconds_sum[5m]) 
/ 
rate(helios_operator_reconciliation_duration_seconds_count[5m])
```

#### Reconciliation Error Rate

```promql
rate(helios_operator_reconciliation_errors_total[5m])
```

#### P95 Reconciliation Duration by Phase

```promql
histogram_quantile(0.95,
  rate(helios_operator_reconciliation_phase_duration_seconds_bucket[5m])
)
```

#### API Call Success Rate

```promql
sum(rate(helios_operator_api_calls_total{status="success"}[5m]))
/
sum(rate(helios_operator_api_calls_total[5m]))
* 100
```

---

## Grafana Dashboards

### Dashboard 1: Operator Overview

Create a Grafana dashboard with the following panels:

#### Panel 1: Reconciliation Rate
```promql
rate(helios_operator_reconciliation_duration_seconds_count[5m])
```

#### Panel 2: Average Reconciliation Duration
```promql
rate(helios_operator_reconciliation_duration_seconds_sum[5m]) 
/ 
rate(helios_operator_reconciliation_duration_seconds_count[5m])
```

#### Panel 3: Error Rate by Phase
```promql
sum by (phase) (rate(helios_operator_reconciliation_errors_total[5m]))
```

#### Panel 4: Resources Managed
```promql
helios_operator_resources_managed
```

#### Panel 5: API Call Latency
```promql
histogram_quantile(0.95,
  sum by (le, operation) (
    rate(helios_operator_api_call_duration_seconds_bucket[5m])
  )
)
```

### Dashboard 2: Application Health

Monitor the health of individual HeliosApp resources:

```promql
# Applications by status
count by (status) (
  kube_customresource_heliosapp_status_condition{type="Ready"}
)

# Build success rate
sum(
  rate(helios_operator_reconciliation_phase_duration_seconds_count{phase="pipeline"}[5m])
)

# ArgoCD sync status
count by (sync_status) (
  argocd_app_info
)
```

### Example Dashboard JSON

```json
{
  "dashboard": {
    "title": "Helios Operator Metrics",
    "panels": [
      {
        "title": "Reconciliation Duration (P95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(helios_operator_reconciliation_duration_seconds_bucket[5m]))"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Error Rate by Phase",
        "targets": [
          {
            "expr": "sum by (phase) (rate(helios_operator_reconciliation_errors_total[5m]))",
            "legendFormat": "{{phase}}"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Managed Resources",
        "targets": [
          {
            "expr": "helios_operator_resources_managed",
            "legendFormat": "{{type}}"
          }
        ],
        "type": "stat"
      }
    ]
  }
}
```

---

## Alerting

### Recommended Alerts

#### High Error Rate

```yaml
alert: HeliosOperatorHighErrorRate
expr: |
  rate(helios_operator_reconciliation_errors_total[5m]) > 0.1
for: 10m
labels:
  severity: warning
annotations:
  summary: "High error rate in Helios Operator"
  description: "Helios Operator is experiencing {{ $value }} errors per second"
```

#### Slow Reconciliation

```yaml
alert: HeliosOperatorSlowReconciliation
expr: |
  histogram_quantile(0.95,
    rate(helios_operator_reconciliation_duration_seconds_bucket[5m])
  ) > 30
for: 15m
labels:
  severity: warning
annotations:
  summary: "Slow reconciliation in Helios Operator"
  description: "P95 reconciliation duration is {{ $value }}s"
```

#### Operator Down

```yaml
alert: HeliosOperatorDown
expr: |
  up{job="helios-operator"} == 0
for: 5m
labels:
  severity: critical
annotations:
  summary: "Helios Operator is down"
  description: "Helios Operator has been down for more than 5 minutes"
```

#### Webhook Failures

```yaml
alert: HeliosWebhookFailures
expr: |
  rate(helios_operator_webhook_validations_total{result="failure"}[5m]) > 0.1
for: 10m
labels:
  severity: warning
annotations:
  summary: "High webhook validation failure rate"
  description: "Webhook validations are failing at {{ $value }} per second"
```

#### API Call Failures

```yaml
alert: HeliosAPICallFailures
expr: |
  (
    sum(rate(helios_operator_api_calls_total{status="failure"}[5m]))
    /
    sum(rate(helios_operator_api_calls_total[5m]))
  ) > 0.05
for: 10m
labels:
  severity: warning
annotations:
  summary: "High API call failure rate"
  description: "{{ $value | humanizePercentage }} of API calls are failing"
```

### PrometheusRule Example

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: helios-operator-alerts
  namespace: helios-operator-system
spec:
  groups:
    - name: helios-operator
      interval: 30s
      rules:
        - alert: HeliosOperatorHighErrorRate
          expr: rate(helios_operator_reconciliation_errors_total[5m]) > 0.1
          for: 10m
          labels:
            severity: warning
            component: helios-operator
          annotations:
            summary: "High error rate in Helios Operator"
            description: "Error rate is {{ $value }} errors/sec"
            runbook_url: "https://docs.example.com/runbooks/helios-high-error-rate"

        - alert: HeliosOperatorDown
          expr: up{job="helios-operator"} == 0
          for: 5m
          labels:
            severity: critical
            component: helios-operator
          annotations:
            summary: "Helios Operator is down"
            description: "The operator has been unavailable for 5+ minutes"
            runbook_url: "https://docs.example.com/runbooks/helios-operator-down"

        - alert: HeliosSlowReconciliation
          expr: histogram_quantile(0.95, rate(helios_operator_reconciliation_duration_seconds_bucket[5m])) > 30
          for: 15m
          labels:
            severity: warning
            component: helios-operator
          annotations:
            summary: "Slow reconciliation performance"
            description: "P95 duration is {{ $value }}s (threshold: 30s)"
```

---

## Logging

### Log Levels

Helios Operator uses structured logging with the following levels:

- **Debug**: Detailed debugging information
- **Info**: General informational messages about normal operation
- **Warning**: Warning messages about potential issues
- **Error**: Error messages about failures

### Log Format

All logs are structured JSON:

```json
{
  "level": "info",
  "ts": "2025-10-16T10:30:00.000Z",
  "msg": "Creating Tekton Pipeline",
  "namespace": "default",
  "name": "my-app",
  "pipeline": "my-app-pipeline",
  "phase": "pipeline"
}
```

### Viewing Logs

```bash
# View operator logs
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager -f

# Filter by log level
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager | jq 'select(.level=="error")'

# Filter by application
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager | jq 'select(.name=="my-app")'

# View reconciliation phases
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager | jq 'select(.phase!=null)'
```

### Log Aggregation

#### Loki Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-config
data:
  loki.yaml: |
    scrape_configs:
      - job_name: helios-operator
        kubernetes_sd_configs:
          - role: pod
            namespaces:
              names:
                - helios-operator-system
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_label_control_plane]
            action: keep
            regex: controller-manager
        pipeline_stages:
          - json:
              expressions:
                level: level
                namespace: namespace
                name: name
                phase: phase
          - labels:
              level:
              namespace:
              name:
              phase:
```

#### LogQL Queries

```logql
# All errors
{job="helios-operator"} |= "error"

# Specific application logs
{job="helios-operator",name="my-app"}

# Reconciliation errors
{job="helios-operator"} | json | level="error" | phase!=""

# Pipeline phase logs
{job="helios-operator"} | json | phase="pipeline"
```

---

## Best Practices

### 1. Metric Collection

- **Scrape Interval**: Use 30s scrape interval for most metrics
- **Retention**: Keep metrics for at least 30 days
- **Cardinality**: Be mindful of label cardinality (namespace, name)
- **Aggregation**: Use recording rules for frequently queried metrics

### 2. Dashboarding

- **Overview Dashboard**: High-level operator health
- **Detail Dashboard**: Per-application metrics
- **SLO Dashboard**: Track SLIs and SLOs
- **Use variables**: Namespace and app name as dashboard variables

### 3. Alerting

- **Alert on SLOs**: Focus on user-impacting issues
- **Runbooks**: Include runbook links in alerts
- **Severity Levels**: Use critical, warning, info appropriately
- **Alert Grouping**: Group related alerts
- **Notification Channels**: Route to appropriate teams

### 4. Logging

- **Structured Logging**: Always use structured logs
- **Log Levels**: Use appropriate levels
- **Context**: Include namespace, name, phase in logs
- **Retention**: Keep logs for at least 14 days
- **Cost Management**: Use log sampling for high-volume logs

### 5. Health Checks

The operator exposes health endpoints:

```bash
# Liveness probe
curl http://localhost:8081/healthz

# Readiness probe
curl http://localhost:8081/readyz
```

Configure Kubernetes probes:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8081
  initialDelaySeconds: 15
  periodSeconds: 20

readinessProbe:
  httpGet:
    path: /readyz
    port: 8081
  initialDelaySeconds: 5
  periodSeconds: 10
```

### 6. Resource Monitoring

Monitor operator resource usage:

```promql
# CPU usage
rate(container_cpu_usage_seconds_total{pod=~"helios-operator.*"}[5m])

# Memory usage
container_memory_working_set_bytes{pod=~"helios-operator.*"}

# Resource requests/limits
kube_pod_container_resource_requests{pod=~"helios-operator.*"}
kube_pod_container_resource_limits{pod=~"helios-operator.*"}
```

---

## Example Monitoring Stack Setup

### Complete Monitoring Stack

```bash
# 1. Install Prometheus Operator
kubectl create namespace monitoring
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring

# 2. Create ServiceMonitor for Helios Operator
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: helios-operator
  namespace: helios-operator-system
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
  endpoints:
  - port: metrics
    interval: 30s
EOF

# 3. Create PrometheusRule for alerts
kubectl apply -f config/prometheus/prometheus-rules.yaml

# 4. Access Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

# 5. Import Helios dashboards
# Upload dashboard JSON from docs/grafana/
```

### Verification

```bash
# Check metrics are being scraped
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090

# Navigate to http://localhost:9090/targets
# Verify helios-operator target is "UP"

# Query metrics
# Navigate to http://localhost:9090/graph
# Run: helios_operator_reconciliation_duration_seconds_count
```

---

## Troubleshooting Monitoring

### Metrics Not Appearing

```bash
# Check operator is exposing metrics
kubectl port-forward -n helios-operator-system deployment/helios-operator-controller-manager 8080:8080
curl http://localhost:8080/metrics

# Check ServiceMonitor is created
kubectl get servicemonitor -n helios-operator-system

# Check Prometheus targets
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
# Visit http://localhost:9090/targets
```

### High Cardinality Issues

If you experience high cardinality:

```bash
# Check metric cardinality
curl http://localhost:8080/metrics | grep helios_operator | wc -l

# Identify high-cardinality metrics
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
# Query: topk(10, count by (__name__)({__name__=~"helios_operator.*"}))
```

### Missing Logs

```bash
# Check operator logs
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager

# Check log level
kubectl get deployment -n helios-operator-system helios-operator-controller-manager -o yaml | grep LOG_LEVEL

# Increase log verbosity
kubectl set env deployment/helios-operator-controller-manager -n helios-operator-system LOG_LEVEL=debug
```

---

## Next Steps

- **Alerting Setup**: Configure alert receivers (Slack, PagerDuty, etc.)
- **Dashboard Import**: Import pre-built Grafana dashboards
- **SLO Definition**: Define and track Service Level Objectives
- **Runbook Creation**: Create runbooks for common issues

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Kubernetes Monitoring Guide](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-usage-monitoring/)
- [HeliosApp Status Conditions](02-helios-app-spec.md#status-conditions)
