# ⚙️ Helm Chart Values Reference

Complete reference for all configurable parameters in the Helios Operator Helm chart.

## 📋 **Overview**

The Helm chart supports extensive customization through the `values.yaml` file. This document provides detailed information about every configurable parameter.

## 🎯 **Core Configuration**

### Image Configuration

```yaml
image:
  repository: helios-operator # Container image repository
  tag: "2.0.0" # Container image tag
  pullPolicy: IfNotPresent # Image pull policy (IfNotPresent, Always, Never)
```

### Replica Configuration

```yaml
replicaCount: 1 # Number of operator replicas
```

### Name Overrides

```yaml
nameOverride: "" # Override chart name
fullnameOverride: "" # Override full name
```

## 🔐 **Security Configuration**

### Service Account

```yaml
serviceAccount:
  create: true # Create service account
  annotations: {} # Additional annotations
  name: "" # Service account name (auto-generated if empty)
```

### Pod Security Context

```yaml
podSecurityContext:
  fsGroup: 65532 # Filesystem group
  runAsNonRoot: true # Run as non-root user
  runAsUser: 65532 # User ID
  seccompProfile:
    type: RuntimeDefault # Seccomp profile type
```

### Container Security Context

```yaml
securityContext:
  allowPrivilegeEscalation: false # Disable privilege escalation
  capabilities:
    drop:
      - ALL # Drop all capabilities
  readOnlyRootFilesystem: true # Read-only root filesystem
```

### RBAC

```yaml
rbac:
  create: true # Create RBAC resources
  rules: [] # Additional RBAC rules
```

## 📊 **Resource Management**

### Resource Limits and Requests

```yaml
resources:
  limits:
    cpu: 500m # CPU limit
    memory: 512Mi # Memory limit
  requests:
    cpu: 100m # CPU request
    memory: 128Mi # Memory request
```

### Node Selector

```yaml
nodeSelector: {} # Node selection constraints
```

### Tolerations

```yaml
tolerations: [] # Pod tolerations
```

### Affinity Rules

```yaml
affinity: {} # Pod affinity/anti-affinity rules
```

## 🏗️ **High Availability**

### Pod Disruption Budget

```yaml
podDisruptionBudget:
  enabled: false # Enable PDB
  minAvailable: 1 # Minimum available pods
  maxUnavailable: 1 # Maximum unavailable pods
```

### Horizontal Pod Autoscaler

```yaml
autoscaling:
  enabled: false # Enable HPA
  minReplicas: 1 # Minimum replicas
  maxReplicas: 10 # Maximum replicas
  targetCPUUtilizationPercentage: 80 # CPU utilization target
```

## 🔍 **Health Checks**

### Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /healthz # Health check path
    port: 8081 # Health check port
  initialDelaySeconds: 15 # Initial delay
  periodSeconds: 20 # Check period
```

### Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /readyz # Readiness check path
    port: 8081 # Readiness check port
  initialDelaySeconds: 5 # Initial delay
  periodSeconds: 10 # Check period
```

## 🌐 **Service Configuration**

### Main Service

```yaml
service:
  enabled: true # Enable service
  type: ClusterIP # Service type (ClusterIP, NodePort, LoadBalancer)
  port: 443 # Service port
  annotations: {} # Service annotations
```

### Metrics Service

```yaml
metrics:
  enabled: true # Enable metrics
  service:
    type: ClusterIP # Metrics service type
    port: 8443 # Metrics service port
    annotations: {} # Metrics service annotations
  serviceMonitor:
    enabled: false # Enable ServiceMonitor
    interval: 30s # Scrape interval
    scrapeTimeout: 10s # Scrape timeout
    labels: {} # ServiceMonitor labels
    annotations: {} # ServiceMonitor annotations
```

## 🔧 **Webhook Configuration**

### Admission Webhooks

```yaml
webhook:
  enabled: true # Enable webhooks
  failurePolicy: Fail # Failure policy (Fail, Ignore)
  timeoutSeconds: 10 # Webhook timeout
  annotations: {} # Webhook annotations
```

## 🏢 **Namespace Configuration**

### Target Namespace

```yaml
namespace:
  create: true # Create namespace
  name: "system" # Namespace name
  annotations: {} # Namespace annotations
  labels: {} # Namespace labels
```

## ⚙️ **Feature Flags**

### Core Features

```yaml
features:
  autoPipelineGeneration: true # Automatic Pipeline generation
  realTimeStatusUpdates: true # Real-time status updates
  webhookValidation: true # Webhook validation
  metricsCollection: true # Metrics collection
```

### Integration Features

```yaml
tekton:
  enabled: true # Enable Tekton integration
  createDefaultResources: true # Create default Tekton resources
  namespace: "tekton-pipelines" # Tekton namespace

argocd:
  enabled: true # Enable ArgoCD integration
  namespace: "argocd" # ArgoCD namespace
  defaultProject: "default" # Default ArgoCD project
```

## 🔄 **Leader Election**

### Leader Election Configuration

```yaml
leaderElection:
  enabled: true # Enable leader election
  resourceName: "501981a2.helios.dev" # Leader election resource name
  resourceNamespace: "system" # Leader election namespace
  leaseDuration: "15s" # Lease duration
  renewDeadline: "10s" # Renew deadline
  retryPeriod: "2s" # Retry period
```

## 📝 **Logging Configuration**

### Logging Settings

```yaml
logging:
  level: info # Log level (debug, info, warn, error)
  format: json # Log format (json, console)
  development: false # Development mode
```

## 🐛 **Debug Configuration**

### Debug Settings

```yaml
debug:
  enabled: false # Enable debug mode
  logLevel: debug # Debug log level
  enableProfiling: false # Enable profiling
```

## 🌍 **Environment Variables**

### Custom Environment Variables

```yaml
env: [] # Additional environment variables
```

### Extra Volumes

```yaml
extraVolumes: [] # Additional volumes
```

### Extra Volume Mounts

```yaml
extraVolumeMounts: [] # Additional volume mounts
```

## 🏷️ **Labels and Annotations**

### Global Labels

```yaml
globalLabels: {} # Labels applied to all resources
```

### Global Annotations

```yaml
globalAnnotations: {} # Annotations applied to all resources
```

### Pod Labels

```yaml
podLabels: {} # Labels applied to pods
```

### Pod Annotations

```yaml
podAnnotations: {} # Annotations applied to pods
```

## 📦 **CRD Management**

### CRD Configuration

```yaml
crd:
  install: true # Install CRDs
  upgrade: true # Upgrade CRDs
  cleanup: true # Cleanup CRDs on uninstall
```

## 🎯 **Common Configuration Examples**

### Development Configuration

```yaml
# Development setup with debug logging
image:
  tag: latest
  pullPolicy: Always

debug:
  enabled: true
  logLevel: debug

logging:
  level: debug
  development: true

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 50m
    memory: 64Mi
```

### Production Configuration

```yaml
# Production setup with high availability
replicaCount: 3

podDisruptionBudget:
  enabled: true
  minAvailable: 2

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi

metrics:
  serviceMonitor:
    enabled: true
    interval: 15s

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app.kubernetes.io/name
                operator: In
                values:
                  - helios-operator
          topologyKey: kubernetes.io/hostname
```

### Security Hardened Configuration

```yaml
# Security hardened setup
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 65532

podSecurityContext:
  fsGroup: 65532
  runAsNonRoot: true
  runAsUser: 65532
  seccompProfile:
    type: RuntimeDefault

serviceAccount:
  create: true
  annotations:
    iam.gke.io/gcp-service-account: helios-operator@project.iam.gserviceaccount.com

webhook:
  enabled: true
  failurePolicy: Fail
  timeoutSeconds: 10
```

### Monitoring Configuration

```yaml
# Monitoring setup with Prometheus
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
    labels:
      prometheus: kube-prometheus
    annotations:
      prometheus.io/scrape: "true"

logging:
  level: info
  format: json

globalAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8443"
```

## 🔧 **Advanced Configuration**

### Custom Resource Templates

```yaml
# Customize resource templates
extraVolumes:
  - name: config-volume
    configMap:
      name: helios-config

extraVolumeMounts:
  - name: config-volume
    mountPath: /etc/helios
    readOnly: true

env:
  - name: CUSTOM_CONFIG
    value: "production"
  - name: DEBUG_MODE
    value: "false"
```

### Network Policies

```yaml
# Network policy configuration
globalLabels:
  app.kubernetes.io/component: operator
  app.kubernetes.io/part-of: helios

globalAnnotations:
  networking.helios.io/allow-ingress: "true"
  networking.helios.io/allow-egress: "true"
```

## 🚀 **Installation Examples**

### Basic Installation

```bash
helm install helios-operator ./helm/helios-operator
```

### Custom Configuration

```bash
helm install helios-operator ./helm/helios-operator \
  --set image.tag=latest \
  --set replicaCount=2 \
  --set metrics.serviceMonitor.enabled=true \
  --set webhook.enabled=true
```

### Values File Installation

```bash
# Create custom values file
cat > my-values.yaml << EOF
replicaCount: 3
metrics:
  serviceMonitor:
    enabled: true
features:
  autoPipelineGeneration: true
EOF

# Install with custom values
helm install helios-operator ./helm/helios-operator -f my-values.yaml
```

### Production Installation

```bash
helm install helios-operator ./helm/helios-operator \
  --namespace system \
  --create-namespace \
  --set replicaCount=3 \
  --set podDisruptionBudget.enabled=true \
  --set metrics.serviceMonitor.enabled=true \
  --set webhook.enabled=true \
  --set resources.limits.cpu=1000m \
  --set resources.limits.memory=1Gi \
  --set resources.requests.cpu=200m \
  --set resources.requests.memory=256Mi
```

## 🔍 **Validation**

### Validate Configuration

```bash
# Lint chart
helm lint ./helm/helios-operator

# Template with custom values
helm template helios-operator ./helm/helios-operator -f my-values.yaml

# Dry run installation
helm install helios-operator ./helm/helios-operator --dry-run --debug
```

### Check Values

```bash
# Show current values
helm get values helios-operator

# Show all values (including defaults)
helm get values helios-operator --all
```

---

**Need help with configuration?** Check our [Troubleshooting Guide](../user-guide/03-troubleshooting.md) or open an [issue on GitHub](https://github.com/hoangphuc841/helios-operator/issues) 🆘
