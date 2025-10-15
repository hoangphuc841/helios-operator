# 📊 Prometheus Metrics Reference

Complete reference for all Prometheus metrics exposed by the Helios Operator.

## 🎯 **Overview**

The Helios Operator exposes comprehensive metrics for monitoring, alerting, and observability. All metrics are available at the `/metrics` endpoint on port 8443.

## 📈 **Available Metrics**

### 1. **Reconciliation Metrics**

#### `heliosapp_reconciliation_total`

**Type**: Counter  
**Description**: Total number of reconciliation attempts  
**Labels**:

- `controller`: Controller name (always "heliosapp")
- `result`: Result of reconciliation ("success", "error", "requeue")

**Example**:

```
heliosapp_reconciliation_total{controller="heliosapp",result="success"} 150
heliosapp_reconciliation_total{controller="heliosapp",result="error"} 5
```

#### `heliosapp_reconciliation_duration_seconds`

**Type**: Histogram  
**Description**: Time spent on reconciliation operations  
**Labels**:

- `name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Buckets**: `[0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60]`

**Example**:

```
heliosapp_reconciliation_duration_seconds_bucket{name="my-app",namespace="default",le="0.5"} 45
heliosapp_reconciliation_duration_seconds_bucket{name="my-app",namespace="default",le="1.0"} 78
heliosapp_reconciliation_duration_seconds_count{name="my-app",namespace="default"} 100
heliosapp_reconciliation_duration_seconds_sum{name="my-app",namespace="default"} 125.5
```

### 2. **Pipeline Metrics**

#### `heliosapp_pipeline_status_total`

**Type**: Counter  
**Description**: Total number of pipeline runs by status  
**Labels**:

- `status`: Pipeline status ("succeeded", "failed", "running", "unknown")
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_pipeline_status_total{status="succeeded",app_name="my-app",namespace="default"} 25
heliosapp_pipeline_status_total{status="failed",app_name="my-app",namespace="default"} 2
heliosapp_pipeline_status_total{status="running",app_name="my-app",namespace="default"} 1
```

#### `heliosapp_pipeline_duration_seconds`

**Type**: Histogram  
**Description**: Time spent on pipeline operations  
**Labels**:

- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace
- `operation`: Operation type ("create", "update", "delete")

**Buckets**: `[1, 5, 10, 30, 60, 120, 300, 600]`

**Example**:

```
heliosapp_pipeline_duration_seconds_bucket{app_name="my-app",namespace="default",operation="create",le="10"} 15
heliosapp_pipeline_duration_seconds_bucket{app_name="my-app",namespace="default",operation="create",le="30"} 28
heliosapp_pipeline_duration_seconds_count{app_name="my-app",namespace="default",operation="create"} 30
heliosapp_pipeline_duration_seconds_sum{app_name="my-app",namespace="default",operation="create"} 125.5
```

### 3. **ArgoCD Application Metrics**

#### `heliosapp_argocd_sync_status_total`

**Type**: Counter  
**Description**: Total number of ArgoCD sync operations by status  
**Labels**:

- `status`: Sync status ("synced", "outofsync", "unknown")
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_argocd_sync_status_total{status="synced",app_name="my-app",namespace="default"} 20
heliosapp_argocd_sync_status_total{status="outofsync",app_name="my-app",namespace="default"} 5
heliosapp_argocd_sync_status_total{status="unknown",app_name="my-app",namespace="default"} 1
```

#### `heliosapp_argocd_health_status_total`

**Type**: Counter  
**Description**: Total number of ArgoCD health checks by status  
**Labels**:

- `health`: Health status ("healthy", "unhealthy", "unknown")
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_argocd_health_status_total{health="healthy",app_name="my-app",namespace="default"} 22
heliosapp_argocd_health_status_total{health="unhealthy",app_name="my-app",namespace="default"} 3
heliosapp_argocd_health_status_total{health="unknown",app_name="my-app",namespace="default"} 1
```

### 4. **Deployment Metrics**

#### `heliosapp_deployment_health_total`

**Type**: Counter  
**Description**: Total number of deployment health checks by status  
**Labels**:

- `health`: Health status ("healthy", "unhealthy", "unknown")
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_deployment_health_total{health="healthy",app_name="my-app",namespace="default"} 25
heliosapp_deployment_health_total{health="unhealthy",app_name="my-app",namespace="default"} 2
heliosapp_deployment_health_total{health="unknown",app_name="my-app",namespace="default"} 1
```

#### `heliosapp_deployment_replicas`

**Type**: Gauge  
**Description**: Current number of deployment replicas  
**Labels**:

- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace
- `replica_type`: Replica type ("ready", "desired", "available")

**Example**:

```
heliosapp_deployment_replicas{app_name="my-app",namespace="default",replica_type="ready"} 3
heliosapp_deployment_replicas{app_name="my-app",namespace="default",replica_type="desired"} 3
heliosapp_deployment_replicas{app_name="my-app",namespace="default",replica_type="available"} 3
```

### 5. **Watch Event Metrics**

#### `heliosapp_watch_events_total`

**Type**: Counter  
**Description**: Total number of watch events received  
**Labels**:

- `resource_type`: Resource type ("argocd", "pipeline", "deployment")
- `event_type`: Event type ("create", "update", "delete")
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_watch_events_total{resource_type="argocd",event_type="update",app_name="my-app",namespace="default"} 15
heliosapp_watch_events_total{resource_type="pipeline",event_type="create",app_name="my-app",namespace="default"} 5
heliosapp_watch_events_total{resource_type="deployment",event_type="update",app_name="my-app",namespace="default"} 25
```

### 6. **Resource Creation Metrics**

#### `heliosapp_resources_created_total`

**Type**: Counter  
**Description**: Total number of resources created by the operator  
**Labels**:

- `resource_type`: Resource type ("pipeline", "eventlistener", "triggerbinding", "triggertemplate", "application")
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_resources_created_total{resource_type="pipeline",app_name="my-app",namespace="default"} 5
heliosapp_resources_created_total{resource_type="eventlistener",app_name="my-app",namespace="default"} 5
heliosapp_resources_created_total{resource_type="application",app_name="my-app",namespace="default"} 5
```

#### `heliosapp_resources_deleted_total`

**Type**: Counter  
**Description**: Total number of resources deleted by the operator  
**Labels**:

- `resource_type`: Resource type
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_resources_deleted_total{resource_type="pipeline",app_name="my-app",namespace="default"} 2
heliosapp_resources_deleted_total{resource_type="eventlistener",app_name="my-app",namespace="default"} 2
heliosapp_resources_deleted_total{resource_type="application",app_name="my-app",namespace="default"} 2
```

### 7. **Error Metrics**

#### `heliosapp_errors_total`

**Type**: Counter  
**Description**: Total number of errors encountered  
**Labels**:

- `error_type`: Error type ("reconciliation", "pipeline", "argocd", "deployment")
- `app_name`: HeliosApp name
- `namespace`: HeliosApp namespace

**Example**:

```
heliosapp_errors_total{error_type="reconciliation",app_name="my-app",namespace="default"} 3
heliosapp_errors_total{error_type="pipeline",app_name="my-app",namespace="default"} 1
heliosapp_errors_total{error_type="argocd",app_name="my-app",namespace="default"} 2
```

### 8. **Webhook Metrics**

#### `heliosapp_webhook_requests_total`

**Type**: Counter  
**Description**: Total number of webhook requests  
**Labels**:

- `webhook_type`: Webhook type ("validating", "mutating")
- `result`: Request result ("allowed", "denied", "error")

**Example**:

```
heliosapp_webhook_requests_total{webhook_type="validating",result="allowed"} 100
heliosapp_webhook_requests_total{webhook_type="validating",result="denied"} 5
heliosapp_webhook_requests_total{webhook_type="mutating",result="allowed"} 100
```

#### `heliosapp_webhook_duration_seconds`

**Type**: Histogram  
**Description**: Time spent processing webhook requests  
**Labels**:

- `webhook_type`: Webhook type ("validating", "mutating")

**Buckets**: `[0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5]`

**Example**:

```
heliosapp_webhook_duration_seconds_bucket{webhook_type="validating",le="0.1"} 95
heliosapp_webhook_duration_seconds_bucket{webhook_type="validating",le="0.5"} 98
heliosapp_webhook_duration_seconds_count{webhook_type="validating"} 100
heliosapp_webhook_duration_seconds_sum{webhook_type="validating"} 12.5
```

## 🔧 **Standard Kubernetes Metrics**

The operator also exposes standard Kubernetes metrics:

### Controller Runtime Metrics

- `controller_runtime_reconcile_total`
- `controller_runtime_reconcile_duration_seconds`
- `controller_runtime_max_concurrent_reconciles`
- `controller_runtime_active_workers`

### Go Runtime Metrics

- `go_gc_duration_seconds`
- `go_memstats_alloc_bytes`
- `go_memstats_heap_bytes`
- `go_goroutines`

## 📊 **Accessing Metrics**

### Port Forward

```bash
# Port forward to metrics endpoint
kubectl port-forward -n system svc/helios-operator-metrics 8443:8443

# Access metrics
curl -k https://localhost:8443/metrics
```

### ServiceMonitor (Prometheus Operator)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: helios-operator
  namespace: system
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: helios-operator
      app.kubernetes.io/component: metrics
  endpoints:
    - port: https
      interval: 30s
      path: /metrics
      scheme: https
      tlsConfig:
        insecureSkipVerify: true
```

## 🚨 **Recommended Alerts**

### High Error Rate

```yaml
- alert: HeliosOperatorHighErrorRate
  expr: rate(heliosapp_errors_total[5m]) > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High error rate in Helios Operator"
    description: "Helios Operator is experiencing {{ $value }} errors per second"
```

### Slow Reconciliation

```yaml
- alert: HeliosOperatorSlowReconciliation
  expr: histogram_quantile(0.95, rate(heliosapp_reconciliation_duration_seconds_bucket[5m])) > 30
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Slow reconciliation in Helios Operator"
    description: "95th percentile reconciliation time is {{ $value }}s"
```

### Pipeline Failures

```yaml
- alert: HeliosAppPipelineFailures
  expr: rate(heliosapp_pipeline_status_total{status="failed"}[5m]) > 0.1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "High pipeline failure rate"
    description: "Pipeline failure rate is {{ $value }} per second"
```

### ArgoCD Sync Issues

```yaml
- alert: HeliosAppArgoCDOutOfSync
  expr: heliosapp_argocd_sync_status_total{status="outofsync"} > 0
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "ArgoCD Application out of sync"
    description: "ArgoCD Application {{ $labels.app_name }} is out of sync"
```

### Unhealthy Deployments

```yaml
- alert: HeliosAppUnhealthyDeployment
  expr: heliosapp_deployment_health_total{health="unhealthy"} > 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Unhealthy deployment detected"
    description: "Deployment {{ $labels.app_name }} is unhealthy"
```

## 📈 **Recommended Dashboards**

### Grafana Dashboard Queries

#### Reconciliation Rate

```promql
rate(heliosapp_reconciliation_total[5m])
```

#### Pipeline Success Rate

```promql
rate(heliosapp_pipeline_status_total{status="succeeded"}[5m]) /
rate(heliosapp_pipeline_status_total[5m])
```

#### Average Reconciliation Time

```promql
rate(heliosapp_reconciliation_duration_seconds_sum[5m]) /
rate(heliosapp_reconciliation_duration_seconds_count[5m])
```

#### Active HeliosApps

```promql
count by (namespace) (heliosapp_reconciliation_total)
```

#### Resource Creation Rate

```promql
rate(heliosapp_resources_created_total[5m])
```

## 🔍 **Troubleshooting Metrics**

### Check Metrics Endpoint

```bash
# Verify metrics are accessible
kubectl get svc -n system helios-operator-metrics

# Test metrics endpoint
curl -k https://$(kubectl get svc -n system helios-operator-metrics -o jsonpath='{.spec.clusterIP}'):8443/metrics
```

### Debug Metrics Issues

```bash
# Check operator logs for metrics errors
kubectl logs -n system deployment/helios-operator | grep -i metric

# Verify ServiceMonitor
kubectl get servicemonitor -n system helios-operator

# Check Prometheus targets
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Visit: http://localhost:9090/targets
```

## 📚 **Integration Examples**

### Prometheus Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: "helios-operator"
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
            - system
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: helios-operator-metrics
      - source_labels: [__meta_kubernetes_endpoint_port_name]
        action: keep
        regex: https
    scheme: https
    tls_config:
      insecure_skip_verify: true
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Helios Operator",
    "panels": [
      {
        "title": "Reconciliation Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(heliosapp_reconciliation_total[5m])",
            "legendFormat": "{{result}}"
          }
        ]
      }
    ]
  }
}
```

---

**Need help with monitoring?** Check our [Troubleshooting Guide](../user-guide/03-troubleshooting.md) or open an [issue on GitHub](https://github.com/hoangphuc841/helios-operator/issues) 🆘
