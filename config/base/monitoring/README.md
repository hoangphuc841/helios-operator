# Helios Operator Monitoring

This directory contains monitoring configurations for the Helios Operator.

## Components

### ServiceMonitor (`servicemonitor.yaml`)

Prometheus ServiceMonitor for scraping metrics from the Helios Operator. This requires the Prometheus Operator to be installed in your cluster.

**Metrics exposed:**

- Controller runtime metrics (reconciliation rates, errors, latency)
- Work queue metrics (depth, latency, retries)
- Leader election metrics
- Webhook metrics
- Custom Helios metrics (app status, pipeline runs, etc.)

### Prometheus Rules (`prometheus-rules.yaml`)

Alert rules for monitoring Helios Operator health and performance:

- **HeliosOperatorDown**: Operator pod is not running
- **HeliosOperatorHighErrorRate**: High reconciliation error rate
- **HeliosReconciliationSlow**: Slow reconciliation performance
- **HeliosReconciliationQueueDepth**: High work queue depth
- **HeliosOperatorHighMemoryUsage**: Memory usage above 85%
- **HeliosOperatorHighCPUUsage**: CPU usage above 80%
- **HeliosAppPipelineFailures**: Multiple pipeline failures
- **HeliosAppNotReady**: App stuck in non-ready state
- **HeliosLeaderElectionFailure**: Leader election instability
- **HeliosWebhookHighLatency**: Webhook latency above 5s
- **HeliosWebhookFailures**: High webhook failure rate

### Grafana Dashboard (`grafana-dashboard.json`)

Pre-built Grafana dashboard with panels for:

- Operator health status
- Total and ready HeliosApps
- Reconciliation rate and errors
- Reconciliation duration (p99)
- Work queue depth
- Memory and CPU usage
- Pipeline run statistics
- Webhook latency

## Deployment

### Using Kustomize

```bash
kubectl apply -k config/monitoring/
```

### Using Helm

Enable monitoring in the Helm chart values:

```yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
  prometheusRule:
    enabled: true
  grafana:
    enabled: true
```

Then install/upgrade:

```bash
helm install helios-operator helm/helios-operator -f values.yaml
```

## Prerequisites

- **Prometheus Operator**: Required for ServiceMonitor and PrometheusRule resources
- **Grafana**: Required for dashboard visualization

### Installing Prerequisites

#### Prometheus Operator (with Helm)

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace
```

#### Import Grafana Dashboard

1. Access Grafana UI
2. Navigate to Dashboards → Import
3. Upload `grafana-dashboard.json`
4. Select Prometheus datasource
5. Click Import

## Accessing Metrics

### Port Forward to Metrics Endpoint

```bash
kubectl port-forward -n helios-system svc/helios-operator-metrics 8443:8443
curl -k https://localhost:8443/metrics
```

### Query Metrics in Prometheus

```bash
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Then open http://localhost:9090 and query metrics like:

- `controller_runtime_reconcile_total{controller="heliosapp"}`
- `helios_app_ready`
- `helios_pipeline_runs_total`

## Custom Metrics

The Helios Operator exposes custom metrics:

- `helios_app_info`: Info metric for HeliosApp resources
- `helios_app_ready`: Ready status (0 or 1) per HeliosApp
- `helios_pipeline_runs_total`: Total pipeline runs
- `helios_pipeline_runs_successful_total`: Successful pipeline runs
- `helios_pipeline_runs_failed_total`: Failed pipeline runs
- `helios_pipeline_duration_seconds`: Pipeline execution duration

## Troubleshooting

### ServiceMonitor not working

1. Check if Prometheus Operator is installed:

   ```bash
   kubectl get servicemonitors -n helios-system
   ```

2. Check Prometheus targets:

   ```bash
   kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
   ```

   Open http://localhost:9090/targets and look for `helios-operator-metrics`

3. Verify service labels match ServiceMonitor selector:
   ```bash
   kubectl get svc -n helios-system helios-operator-metrics -o yaml
   ```

### Alerts not firing

1. Check PrometheusRule is loaded:

   ```bash
   kubectl get prometheusrules -n helios-system
   ```

2. Check Prometheus rules:
   Open http://localhost:9090/rules and search for "helios"

3. Check Alertmanager:
   ```bash
   kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-alertmanager 9093:9093
   ```
   Open http://localhost:9093

### Dashboard not showing data

1. Verify Prometheus datasource is configured in Grafana
2. Check that metrics are being scraped (see "ServiceMonitor not working" above)
3. Verify time range in Grafana matches when operator was running
4. Check query expressions in dashboard panels

## Reference

- [Prometheus Operator Documentation](https://prometheus-operator.dev/)
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [Controller Runtime Metrics](https://book.kubebuilder.io/reference/metrics.html)
