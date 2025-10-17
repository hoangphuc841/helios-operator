# API Reference

This document provides a comprehensive reference for the Helios Operator API, including all custom resources, their fields, and usage examples.

## Table of Contents

- [HeliosApp Resource](#heliosapp-resource)
- [Status Fields](#status-fields)
- [Condition Types](#condition-types)
- [Events](#events)
- [Webhooks](#webhooks)
- [RBAC Requirements](#rbac-requirements)

## HeliosApp Resource

The `HeliosApp` is the primary custom resource that defines a GitOps-based application managed by Helios Operator.

### API Version and Kind

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
```

### Schema Definition

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  # Standard Kubernetes metadata
  name: string                    # Required: Name of the application
  namespace: string               # Required: Namespace where the app is deployed
  labels:                         # Optional: Additional labels
    string: string
  annotations:                    # Optional: Additional annotations
    string: string
spec:
  # Source repository configuration
  gitRepo: string                 # Required: Git repository URL
  gitBranch: string               # Optional: Git branch to watch (default: "main")

  # Container image configuration
  imageRepo: string               # Required: Container registry URL
  imageTag: string                # Optional: Image tag (default: "latest")

  # Application configuration
  port: int32                     # Required: Application port (1-65535)
  serviceAccount: string          # Optional: Service account (default: "default")

  # Security configuration
  webhookSecret: string           # Required: Secret for Git webhook authentication

  # GitOps repository configuration
  gitopsRepo: string              # Required: GitOps repository URL
  gitopsPath: string              # Required: Path to application manifests
  gitopsBranch: string            # Required: GitOps repository branch

  # Optional: Resource configuration
  resources:                      # Optional: Resource requests and limits
    requests:
      cpu: string
      memory: string
    limits:
      cpu: string
      memory: string

  # Optional: Environment variables
  env:                            # Optional: Environment variables
    - name: string
      value: string
      valueFrom:
        secretKeyRef:
          name: string
          key: string
        configMapKeyRef:
          name: string
          key: string

  # Optional: Health check configuration
  healthCheck:                    # Optional: Health check configuration
    path: string                  # Health check endpoint path
    port: int32                   # Health check port
    initialDelaySeconds: int32    # Initial delay before first check
    periodSeconds: int32          # Period between checks
    timeoutSeconds: int32         # Timeout for each check
    failureThreshold: int32       # Number of failures before marking unhealthy

  # Optional: Scaling configuration
  scaling:                        # Optional: Horizontal Pod Autoscaler configuration
    minReplicas: int32            # Minimum number of replicas
    maxReplicas: int32            # Maximum number of replicas
    targetCPUUtilizationPercentage: int32  # Target CPU utilization

  # Optional: Networking configuration
  networking:                     # Optional: Networking configuration
    ingress:                      # Optional: Ingress configuration
      enabled: bool               # Enable ingress
      host: string                # Ingress host
      tls:                        # TLS configuration
        secretName: string        # TLS secret name
      annotations:                # Ingress annotations
        string: string
    service:                      # Optional: Service configuration
      type: string                # Service type (ClusterIP, NodePort, LoadBalancer)
      annotations:                # Service annotations
        string: string

  # Optional: Custom pipeline configuration
  pipeline:                       # Optional: Custom pipeline configuration
    template: string              # Pipeline template name
    parameters:                   # Pipeline parameters
      - name: string
        value: string

  # Optional: Custom ArgoCD configuration
  argocd:                         # Optional: Custom ArgoCD configuration
    project: string               # ArgoCD project name
    syncPolicy:                   # Sync policy configuration
      automated:
        prune: bool               # Enable automatic pruning
        selfHeal: bool            # Enable self-healing
      syncOptions:                # Additional sync options
        - string
    retry:                        # Retry configuration
      limit: int32                # Maximum retry attempts
      backoff:                    # Backoff configuration
        duration: string          # Initial backoff duration
        factor: int32             # Backoff factor
        maxDuration: string       # Maximum backoff duration

status:
  # Application status information
  phase: string                   # Current phase of the application
  message: string                 # Human-readable status message
  reason: string                  # Reason for current status

  # Component statuses
  pipelineStatus:                 # Tekton pipeline status
    status: string                # Pipeline status (Pending, Running, Succeeded, Failed)
    buildVersion: string          # Version of the built image
    lastBuildTime: *Time          # Timestamp of last build

  deploymentStatus:               # Deployment status
    health: string                # Deployment health (Healthy, Progressing, Degraded)
    readyReplicas: int32          # Number of ready replicas
    desiredReplicas: int32        # Number of desired replicas
    lastHealthyTime: *Time        # Timestamp when deployment was last healthy

  argocdStatus:                   # ArgoCD application status
    syncStatus: string            # ArgoCD sync status
    healthStatus: string          # ArgoCD health status
    lastSyncTime: *Time           # Timestamp of last sync

  # Conditions array
  conditions:                     # Array of condition objects
    - type: string                # Condition type
      status: string              # Condition status (True, False, Unknown)
      lastTransitionTime: *Time   # When condition last changed
      reason: string              # Reason for condition
      message: string             # Human-readable message

  # Observed generation
  observedGeneration: int64       # Generation of the spec that was last processed
```

### Field Validation Rules

#### Required Fields

- `spec.gitRepo`: Must be a valid HTTP/HTTPS Git repository URL
- `spec.imageRepo`: Must be a valid container registry URL
- `spec.port`: Must be between 1 and 65535
- `spec.webhookSecret`: Must be a non-empty string
- `spec.gitopsRepo`: Must be a valid HTTP/HTTPS Git repository URL
- `spec.gitopsPath`: Must be a non-empty string
- `spec.gitopsBranch`: Must be a non-empty string

#### Optional Fields with Defaults

- `spec.gitBranch`: Defaults to "main"
- `spec.imageTag`: Defaults to "latest"
- `spec.serviceAccount`: Defaults to "default"
- `spec.resources.requests.cpu`: Defaults to "100m"
- `spec.resources.requests.memory`: Defaults to "128Mi"
- `spec.resources.limits.cpu`: Defaults to "500m"
- `spec.resources.limits.memory`: Defaults to "512Mi"

#### Validation Rules

- Git URLs must use HTTP or HTTPS protocol
- Image repository URLs must be valid container registry URLs
- Port numbers must be valid integers between 1 and 65535
- Resource specifications must follow Kubernetes resource format
- Environment variable names must follow Kubernetes naming conventions

## Status Fields

### Phase

The `status.phase` field indicates the overall state of the HeliosApp:

- **Pending**: Initial state, waiting for processing
- **Building**: Tekton pipeline is running
- **Deploying**: ArgoCD is syncing the application
- **Running**: Application is deployed and healthy
- **Failed**: Application deployment failed
- **Unknown**: Status cannot be determined

### Message and Reason

- `status.message`: Human-readable description of the current status
- `status.reason`: Machine-readable reason code for the current status

### Component Statuses

#### Pipeline Status

```yaml
pipelineStatus:
  status: string        # Pending, Running, Succeeded, Failed
  buildVersion: string  # Version of the built image
  lastBuildTime: *Time  # Timestamp of last build
```

#### Deployment Status

```yaml
deploymentStatus:
  health: string           # Healthy, Progressing, Degraded
  readyReplicas: int32     # Number of ready replicas
  desiredReplicas: int32   # Number of desired replicas
  lastHealthyTime: *Time   # Timestamp when deployment was last healthy
```

#### ArgoCD Status

```yaml
argocdStatus:
  syncStatus: string    # Synced, OutOfSync, Unknown
  healthStatus: string  # Healthy, Progressing, Degraded, Suspended
  lastSyncTime: *Time   # Timestamp of last sync
```

## Condition Types

The `status.conditions` array contains detailed information about various aspects of the application lifecycle:

### Standard Conditions

#### Ready

Indicates whether the application is fully reconciled and ready for use.

```yaml
- type: "Ready"
  status: "True" | "False" | "Unknown"
  reason: "ApplicationReady" | "ApplicationNotReady" | "ReconciliationInProgress"
  message: "Application is ready and healthy"
```

#### Synced

Indicates whether the ArgoCD Application is in sync with the GitOps repository.

```yaml
- type: "Synced"
  status: "True" | "False" | "Unknown"
  reason: "Synced" | "OutOfSync" | "SyncFailed"
  message: "Application is in sync with GitOps repository"
```

#### BuildSucceeded

Indicates the status of the last build pipeline.

```yaml
- type: "BuildSucceeded"
  status: "True" | "False" | "Unknown"
  reason: "BuildCompleted" | "BuildFailed" | "BuildInProgress"
  message: "Build pipeline completed successfully"
```

#### Deployed

Indicates whether the application has been deployed.

```yaml
- type: "Deployed"
  status: "True" | "False" | "Unknown"
  reason: "DeploymentSucceeded" | "DeploymentFailed" | "DeploymentInProgress"
  message: "Application deployment completed successfully"
```

#### PipelineReady

Indicates whether the Tekton Pipeline is ready.

```yaml
- type: "PipelineReady"
  status: "True" | "False" | "Unknown"
  reason: "PipelineCreated" | "PipelineCreationFailed"
  message: "Tekton Pipeline is ready for execution"
```

#### WebhookReady

Indicates whether the webhook is ready.

```yaml
- type: "WebhookReady"
  status: "True" | "False" | "Unknown"
  reason: "WebhookCreated" | "WebhookCreationFailed"
  message: "Git webhook is ready for triggers"
```

## Events

Helios Operator emits Kubernetes events for important state changes:

### Event Types

#### Normal Events

- **HeliosAppCreated**: HeliosApp resource was created
- **PipelineCreated**: Tekton Pipeline was created
- **WebhookCreated**: Git webhook was created
- **ArgoAppCreated**: ArgoCD Application was created
- **BuildSucceeded**: Build pipeline completed successfully
- **DeploymentSucceeded**: Application deployment completed
- **SyncSucceeded**: ArgoCD sync completed successfully

#### Warning Events

- **BuildFailed**: Build pipeline failed
- **DeploymentFailed**: Application deployment failed
- **SyncFailed**: ArgoCD sync failed
- **WebhookFailed**: Git webhook creation failed
- **PipelineFailed**: Tekton Pipeline creation failed
- **ArgoAppFailed**: ArgoCD Application creation failed

### Event Examples

```bash
# View events for a specific HeliosApp
kubectl get events --field-selector involvedObject.name=my-app

# View all events in a namespace
kubectl get events --sort-by='.lastTimestamp'
```

## Webhooks

Helios Operator supports admission webhooks for validating HeliosApp resources.

### Validation Webhook

The validation webhook ensures that HeliosApp resources are valid before they are stored in etcd.

#### Validation Rules

1. **Required Fields**: All required fields must be present
2. **Field Format**: Fields must match expected formats (URLs, ports, etc.)
3. **Resource Limits**: Resource specifications must be valid
4. **Environment Variables**: Environment variable names must be valid

#### Webhook Configuration

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionWebhook
metadata:
  name: heliosapp-validation.helios.io
webhooks:
  - name: heliosapp-validation.helios.io
    clientConfig:
      service:
        name: helios-operator-webhook-service
        namespace: helios-system
        path: "/validate-platform-helios-io-v1-heliosapp"
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["platform.helios.io"]
        apiVersions: ["v1"]
        resources: ["heliosapps"]
    failurePolicy: Fail
    sideEffects: None
```

### Webhook Examples

#### Valid HeliosApp

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: valid-app
spec:
  gitRepo: "https://github.com/example/valid-app"
  imageRepo: "registry.example.com/valid-app"
  port: 8080
  webhookSecret: "valid-secret"
  gitopsRepo: "https://github.com/example/gitops"
  gitopsPath: "apps/valid-app"
  gitopsBranch: "main"
```

#### Invalid HeliosApp (will be rejected)

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: invalid-app
spec:
  gitRepo: "invalid-url" # Invalid Git URL
  imageRepo: "registry.example.com/invalid-app"
  port: 99999 # Invalid port number
  webhookSecret: "" # Empty webhook secret
  gitopsRepo: "https://github.com/example/gitops"
  gitopsPath: "apps/invalid-app"
  gitopsBranch: "main"
```

## RBAC Requirements

Helios Operator requires specific RBAC permissions to manage resources.

### ClusterRole Permissions

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: helios-operator-manager-role
rules:
# HeliosApp resources
  - apiGroups: ["platform.helios.io"]
    resources: ["heliosapps"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["platform.helios.io"]
  resources: ["heliosapps/status", "heliosapps/finalizers"]
  verbs: ["get", "update", "patch"]

# Core resources
- apiGroups: [""]
  resources: ["events", "pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# Apps resources
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]

# Tekton resources
- apiGroups: ["triggers.tekton.dev"]
  resources: ["eventlisteners", "triggerbindings", "triggertemplates"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# ArgoCD resources
- apiGroups: ["argoproj.io"]
  resources: ["applications"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

### ServiceAccount

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: helios-operator-manager
  namespace: helios-system
```

### ClusterRoleBinding

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: helios-operator-manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: helios-operator-manager-role
subjects:
  - kind: ServiceAccount
    name: helios-operator-manager
    namespace: helios-system
```

## API Examples

### Complete Example

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: complete-example
  namespace: production
  labels:
    app.kubernetes.io/name: complete-example
    app.kubernetes.io/version: "1.0.0"
    environment: production
  annotations:
    helios.io/description: "Complete example with all optional fields"
spec:
  # Source configuration
  gitRepo: "https://github.com/example/complete-app"
  gitBranch: "main"

  # Build configuration
  imageRepo: "registry.example.com/complete-app"
  imageTag: "v1.0.0"

  # Application configuration
  port: 8080
  serviceAccount: "complete-app-sa"

  # Security configuration
  webhookSecret: "complete-app-webhook-secret"

  # GitOps configuration
  gitopsRepo: "https://github.com/example/production-manifests"
  gitopsPath: "apps/complete-app"
  gitopsBranch: "production"

  # Resource configuration
  resources:
    requests:
      cpu: "200m"
      memory: "256Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"

  # Environment variables
  env:
    - name: ENVIRONMENT
      value: "production"
    - name: LOG_LEVEL
      value: "info"
    - name: DATABASE_URL
      valueFrom:
        secretKeyRef:
          name: complete-app-secrets
          key: database-url

  # Health check configuration
  healthCheck:
    path: "/health"
    port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3

  # Scaling configuration
  scaling:
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70

  # Networking configuration
  networking:
    ingress:
      enabled: true
      host: "complete-app.example.com"
      tls:
        secretName: "complete-app-tls"
      annotations:
        nginx.ingress.kubernetes.io/rate-limit: "100"
    service:
      type: ClusterIP
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: nlb

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

This API reference provides comprehensive documentation for all aspects of the Helios Operator API. For more examples and use cases, see the [Examples](../examples/) directory.
