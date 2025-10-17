# Configuration Guide

This guide covers all configuration options available for Helios Operator, including environment variables, Helm values, and resource specifications.

## Table of Contents

- [Operator Configuration](#operator-configuration)
- [HeliosApp Configuration](#heliosapp-configuration)
- [Environment Variables](#environment-variables)
- [Helm Values](#helm-values)
- [Advanced Configuration](#advanced-configuration)
- [Configuration Examples](#configuration-examples)

## Operator Configuration

The Helios Operator can be configured through environment variables, Helm values, or configuration files.

### Core Settings

| Setting                   | Environment Variable        | Default               | Description                                       |
| ------------------------- | --------------------------- | --------------------- | ------------------------------------------------- |
| Reconcile Interval        | `RECONCILE_INTERVAL`        | `10s`                 | How often to reconcile resources                  |
| Max Concurrent Reconciles | `MAX_CONCURRENT_RECONCILES` | `1`                   | Maximum number of concurrent reconciliation loops |
| Watch Namespace           | `WATCH_NAMESPACE`           | `""` (all namespaces) | Specific namespace to watch, empty for all        |
| Enable Metrics            | `ENABLE_METRICS`            | `true`                | Enable Prometheus metrics                         |
| Enable Webhooks           | `ENABLE_WEBHOOKS`           | `true`                | Enable admission webhooks                         |
| Enable Leader Election    | `ENABLE_LEADER_ELECTION`    | `false`               | Enable leader election for HA                     |

### Resource Defaults

| Setting                 | Environment Variable      | Default   | Description                                |
| ----------------------- | ------------------------- | --------- | ------------------------------------------ |
| Default Replicas        | `DEFAULT_REPLICAS`        | `1`       | Default number of replicas for deployments |
| Default Port            | `DEFAULT_PORT`            | `8080`    | Default application port                   |
| Default Service Account | `DEFAULT_SERVICE_ACCOUNT` | `default` | Default service account for applications   |

### Retry and Timeout Settings

| Setting           | Environment Variable | Default | Description                              |
| ----------------- | -------------------- | ------- | ---------------------------------------- |
| Max Retries       | `MAX_RETRIES`        | `3`     | Maximum number of retry attempts         |
| Retry Backoff     | `RETRY_BACKOFF`      | `5s`    | Backoff duration between retries         |
| Reconcile Timeout | `RECONCILE_TIMEOUT`  | `30s`   | Maximum time for a single reconciliation |
| API Call Timeout  | `API_CALL_TIMEOUT`   | `30s`   | Timeout for Kubernetes API calls         |

## HeliosApp Configuration

The `HeliosApp` custom resource is the primary configuration object for applications managed by Helios Operator.

### Basic Configuration

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-app
  namespace: default
spec:
  # Required: Source Git repository
  gitRepo: "https://github.com/example/my-app"

  # Required: Container image repository
  imageRepo: "registry.example.com/my-app"

  # Required: Application port
  port: 8080

  # Optional: Service account (defaults to "default")
  serviceAccount: "my-service-account"

  # Required: Webhook secret for Git triggers
  webhookSecret: "my-webhook-secret"

  # Required: GitOps repository configuration
  gitopsRepo: "https://github.com/example/my-app-manifests"
  gitopsPath: "apps/my-app"
  gitopsBranch: "main"
```

### Advanced Configuration

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-advanced-app
  namespace: production
  labels:
    app.kubernetes.io/name: my-app
    app.kubernetes.io/version: "1.0.0"
    environment: production
  annotations:
    helios.io/description: "Production application with advanced configuration"
spec:
  # Source configuration
  gitRepo: "https://github.com/example/my-app"
  gitBranch: "main" # Optional: specific branch to watch

  # Build configuration
  imageRepo: "registry.example.com/my-app"
  imageTag: "latest" # Optional: specific image tag

  # Application configuration
  port: 8080
  serviceAccount: "my-app-sa"

  # Security configuration
  webhookSecret: "production-webhook-secret"

  # GitOps configuration
  gitopsRepo: "https://github.com/example/gitops"
  gitopsPath: "apps/production/my-app"
  gitopsBranch: "production"

  # Resource limits and requests
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"

  # Environment variables
  env:
    - name: ENVIRONMENT
      value: "production"
    - name: LOG_LEVEL
      value: "info"

  # Health checks
  healthCheck:
    path: "/health"
    port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10

  # Scaling configuration
  scaling:
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70

  # Networking configuration
  networking:
    ingress:
      enabled: true
      host: "my-app.example.com"
      tls:
        secretName: "my-app-tls"
    service:
      type: ClusterIP
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: nlb
```

### Field Validation

The HeliosApp resource includes comprehensive validation:

- **Git URLs**: Must be valid HTTP/HTTPS Git repository URLs
- **Image Repository**: Must be a valid container registry URL
- **Port**: Must be between 1 and 65535
- **Resource Limits**: CPU and memory must be valid Kubernetes resource specifications
- **Environment Variables**: Must follow Kubernetes environment variable naming conventions

## Environment Variables

### Setting Environment Variables

#### Using Helm

```bash
helm install helios-operator helios-operator/helios-operator \
  --set config.reconcileInterval=30s \
  --set config.maxConcurrentReconciles=3 \
  --set config.watchNamespace="production" \
  --set config.enableMetrics=true
```

#### Using kubectl

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helios-operator-manager
  namespace: helios-system
spec:
  template:
    spec:
      containers:
        - name: manager
          env:
            - name: RECONCILE_INTERVAL
              value: "30s"
            - name: MAX_CONCURRENT_RECONCILES
              value: "3"
            - name: WATCH_NAMESPACE
              value: "production"
```

#### Using ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: helios-operator-config
  namespace: helios-system
data:
  RECONCILE_INTERVAL: "30s"
  MAX_CONCURRENT_RECONCILES: "3"
  WATCH_NAMESPACE: "production"
  ENABLE_METRICS: "true"
```

## Helm Values

### Complete Helm Values Reference

```yaml
# Global configuration
global:
  # Image registry and tag
  imageRegistry: "quay.io"
  imageTag: "latest"

  # Pull policy
  imagePullPolicy: IfNotPresent

# Operator configuration
config:
  # Core settings
  reconcileInterval: "10s"
  maxConcurrentReconciles: 1
  watchNamespace: "" # Empty for all namespaces

  # Resource defaults
  defaultReplicas: 1
  defaultPort: 8080
  defaultServiceAccount: "default"

  # Retry and timeout settings
  maxRetries: 3
  retryBackoff: "5s"
  reconcileTimeout: "30s"
  apiCallTimeout: "30s"

  # Feature flags
  enableMetrics: true
  enableWebhooks: true
  enableLeaderElection: false

# Resource configuration
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

# Node selector and tolerations
nodeSelector: {}
tolerations: []
affinity: {}

# Security context
securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  fsGroup: 65534

# Service account
serviceAccount:
  create: true
  name: "helios-operator-manager"
  annotations: {}

# RBAC
rbac:
  create: true

# Metrics configuration
metrics:
  enabled: true
  service:
    type: ClusterIP
    port: 8080
    annotations: {}

  serviceMonitor:
    enabled: true
    namespace: "monitoring"
    labels: {}
    interval: "30s"
    scrapeTimeout: "10s"

# Webhook configuration
webhook:
  enabled: true
  port: 9443
  service:
    type: ClusterIP
    annotations: {}

# Monitoring configuration
monitoring:
  enabled: true
  grafana:
    dashboard:
      enabled: true
      namespace: "monitoring"

  prometheus:
    rules:
      enabled: true
      namespace: "monitoring"

# Network policies
networkPolicy:
  enabled: false
  ingress: []
  egress: []

# Pod disruption budget
podDisruptionBudget:
  enabled: false
  minAvailable: 1
```

## Advanced Configuration

### Multi-Environment Setup

#### Development Environment

```yaml
# values-dev.yaml
config:
  watchNamespace: "development"
  defaultReplicas: 1
  enableMetrics: false

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

#### Production Environment

```yaml
# values-prod.yaml
config:
  watchNamespace: "production"
  maxConcurrentReconciles: 5
  enableMetrics: true
  enableLeaderElection: true

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

replicaCount: 3

podDisruptionBudget:
  enabled: true
  minAvailable: 2

networkPolicy:
  enabled: true
```

### Custom Resource Definitions

#### Custom Pipeline Templates

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: custom-pipeline-app
spec:
  # ... basic configuration ...

  # Custom pipeline configuration
  pipeline:
    template: "custom-build-pipeline"
    parameters:
      - name: "BUILD_ARGS"
        value: "--build-arg NODE_ENV=production"
      - name: "TEST_COMMAND"
        value: "npm run test:ci"
```

#### Custom ArgoCD Configuration

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: custom-argocd-app
spec:
  # ... basic configuration ...

  # Custom ArgoCD configuration
  argocd:
    project: "production"
    syncPolicy:
      automated:
        prune: true
        selfHeal: true
      syncOptions:
        - "CreateNamespace=true"
        - "PrunePropagationPolicy=foreground"
    retry:
      limit: 5
      backoff:
        duration: 5m
        factor: 2
        maxDuration: 3h
```

## Configuration Examples

### Basic Web Application

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: web-app
spec:
  gitRepo: "https://github.com/example/web-app"
  imageRepo: "registry.example.com/web-app"
  port: 3000
  serviceAccount: "web-app-sa"
  webhookSecret: "web-app-webhook"
  gitopsRepo: "https://github.com/example/web-app-manifests"
  gitopsPath: "apps/web-app"
  gitopsBranch: "main"
```

### Microservice with Custom Resources

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: user-service
spec:
  gitRepo: "https://github.com/example/user-service"
  imageRepo: "registry.example.com/user-service"
  port: 8080
  serviceAccount: "user-service-sa"
  webhookSecret: "user-service-webhook"
  gitopsRepo: "https://github.com/example/microservices-manifests"
  gitopsPath: "services/user"
  gitopsBranch: "main"

  resources:
    requests:
      cpu: "200m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"

  env:
    - name: DATABASE_URL
      valueFrom:
        secretKeyRef:
          name: user-service-secrets
          key: database-url
    - name: REDIS_URL
      valueFrom:
        secretKeyRef:
          name: user-service-secrets
          key: redis-url
```

### High-Availability Production Setup

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: production-api
spec:
  gitRepo: "https://github.com/example/api"
  imageRepo: "registry.example.com/api"
  port: 8080
  serviceAccount: "api-sa"
  webhookSecret: "api-production-webhook"
  gitopsRepo: "https://github.com/example/production-manifests"
  gitopsPath: "apis/main-api"
  gitopsBranch: "production"

  resources:
    requests:
      cpu: "500m"
      memory: "512Mi"
    limits:
      cpu: "2000m"
      memory: "2Gi"

  scaling:
    minReplicas: 3
    maxReplicas: 20
    targetCPUUtilizationPercentage: 70

  healthCheck:
    path: "/health"
    port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3

  networking:
    ingress:
      enabled: true
      host: "api.example.com"
      tls:
        secretName: "api-tls"
      annotations:
        nginx.ingress.kubernetes.io/rate-limit: "100"
        nginx.ingress.kubernetes.io/ssl-redirect: "true"
```

## Best Practices

### Configuration Management

1. **Use GitOps**: Store all configurations in Git repositories
2. **Environment Separation**: Use separate configurations for different environments
3. **Secret Management**: Use Kubernetes secrets or external secret management systems
4. **Resource Limits**: Always set appropriate resource requests and limits
5. **Monitoring**: Enable metrics and monitoring for production deployments

### Security Considerations

1. **RBAC**: Use least-privilege RBAC policies
2. **Network Policies**: Implement network policies for security
3. **Image Security**: Use trusted base images and scan for vulnerabilities
4. **Secret Rotation**: Regularly rotate webhook secrets and API keys
5. **Audit Logging**: Enable audit logging for compliance

### Performance Optimization

1. **Resource Tuning**: Monitor and adjust resource allocations
2. **Scaling**: Use horizontal pod autoscaling for variable workloads
3. **Caching**: Implement appropriate caching strategies
4. **Connection Pooling**: Configure database and service connection pooling
5. **Monitoring**: Use comprehensive monitoring to identify bottlenecks

For more advanced configuration options and examples, see the [Examples](../examples/) directory.
