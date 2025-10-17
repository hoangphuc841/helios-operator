# Helios Operator Helm Chart

This Helm chart deploys the Helios Operator, a Kubernetes operator for automated GitOps-based application deployment using Tekton and ArgoCD.

## Prerequisites

- Kubernetes 1.34+
- Helm 3.19+
- Tekton Pipelines (installed in `tekton-pipelines` namespace)
- ArgoCD (installed in `argocd` namespace)

## Installing the Chart

To install the chart with the release name `helios-operator`:

```bash
helm repo add helios-operator https://hoangphuc841.github.io/helios-operator
helm repo update
helm install helios-operator helios-operator/helios-operator
```

Or install from local chart:

```bash
helm install helios-operator ./helm/helios-operator
```

## Uninstalling the Chart

To uninstall/delete the `helios-operator` deployment:

```bash
helm uninstall helios-operator
```

## Configuration

The following table lists the configurable parameters and their default values.

| Parameter                         | Description                     | Default           |
| --------------------------------- | ------------------------------- | ----------------- |
| `image.repository`                | Image repository                | `helios-operator` |
| `image.tag`                       | Image tag                       | `2.0.0`           |
| `image.pullPolicy`                | Image pull policy               | `IfNotPresent`    |
| `replicaCount`                    | Number of replicas              | `1`               |
| `serviceAccount.create`           | Create service account          | `true`            |
| `rbac.create`                     | Create RBAC resources           | `true`            |
| `metrics.enabled`                 | Enable metrics collection       | `true`            |
| `webhook.enabled`                 | Enable webhook validation       | `true`            |
| `leaderElection.enabled`          | Enable leader election          | `true`            |
| `features.autoPipelineGeneration` | Enable auto Pipeline generation | `true`            |
| `features.realTimeStatusUpdates`  | Enable real-time status updates | `true`            |
| `features.webhookValidation`      | Enable webhook validation       | `true`            |
| `features.metricsCollection`      | Enable metrics collection       | `true`            |

## Features

### Automatic Pipeline Generation

The operator automatically creates Tekton Pipelines based on the HeliosApp name, eliminating the need for users to manage Pipeline resources manually.

### Real-time Status Updates

Uses Kubernetes Watches to provide real-time status updates for:

- ArgoCD Application sync status
- Tekton PipelineRun build status
- Kubernetes Deployment health

### Webhook Validation

Validates HeliosApp resources before they are applied, ensuring data integrity and providing clear error messages.

### Comprehensive Metrics

Provides Prometheus metrics for:

- Reconciliation duration
- Build status tracking
- Deployment health monitoring
- Watch event counts

## Examples

### Basic Installation

```bash
helm install helios-operator ./helm/helios-operator
```

### Custom Configuration

```bash
helm install helios-operator ./helm/helios-operator \
  --set image.tag=latest \
  --set metrics.enabled=true \
  --set webhook.enabled=true \
  --set features.autoPipelineGeneration=true
```

### Production Configuration

```bash
helm install helios-operator ./helm/helios-operator \
  --set replicaCount=2 \
  --set resources.limits.cpu=500m \
  --set resources.limits.memory=512Mi \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=128Mi \
  --set podDisruptionBudget.enabled=true \
  --set metrics.serviceMonitor.enabled=true
```

## Troubleshooting

### Check Operator Status

```bash
kubectl get pods -n system -l app.kubernetes.io/name=helios-operator
```

### View Operator Logs

```bash
kubectl logs -n system deployment/helios-operator
```

### Check CRD Installation

```bash
kubectl get crd heliosapps.platform.helios.io
```

### Verify Webhook

```bash
kubectl get validatingwebhookconfiguration helios-operator-validating-webhook-configuration
kubectl get mutatingwebhookconfiguration helios-operator-mutating-webhook-configuration
```

### Access Metrics

```bash
kubectl port-forward -n system svc/helios-operator-metrics 8443:8443
# Visit: https://localhost:8443/metrics
```

## Development

### Testing the Chart

```bash
# Lint the chart
helm lint ./helm/helios-operator

# Test template rendering
helm template helios-operator ./helm/helios-operator

# Dry run installation
helm install helios-operator ./helm/helios-operator --dry-run --debug
```

### Updating the Chart

1. Update version in `Chart.yaml`
2. Update values in `values.yaml`
3. Test with `helm lint` and `helm template`
4. Package with `helm package ./helm/helios-operator`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test the chart
5. Submit a pull request

## License

This chart is licensed under the Apache License 2.0.
