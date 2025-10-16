# HeliosApp API Reference

## Overview

The `HeliosApp` custom resource provides a declarative way to manage application deployments using GitOps. It automates the creation of Tekton Pipelines for building container images and ArgoCD Applications for deployment.

## API Version

- **Group**: `platform.helios.io`
- **Version**: `v1`
- **Kind**: `HeliosApp`

## Resource Specification

### HeliosAppSpec

The specification defines the desired state of a HeliosApp.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `gitRepo` | string | Yes | - | URL of the Git repository containing application source code. Must use https://, http://, ssh://, or git:// scheme. |
| `gitBranch` | string | No | "main" | Git branch or tag to build from. |
| `imageRepo` | string | Yes | - | Container image repository in format `[registry/]namespace/repository[:tag]`. |
| `port` | int32 | Yes | - | Port the application listens on. Range: 1-65535. |
| `replicas` | int32 | No | 1 | Desired number of replicas. Range: 0-100. |
| `serviceAccount` | string | No | "default" | Kubernetes ServiceAccount for pipeline execution. |
| `webhookSecret` | string | No | "" | Name of the Secret containing GitHub webhook token. |
| `pvcName` | string | No | "" | Name of PersistentVolumeClaim for pipeline workspace. If empty, auto-generated. |
| `gitopsRepo` | string | No | Same as gitRepo | Git repository containing Kubernetes manifests. |
| `gitopsPath` | string | No | "" | Path within gitopsRepo to application manifests. |
| `gitopsBranch` | string | No | "main" | Branch of gitopsRepo to deploy from. |

### HeliosAppStatus

The status reflects the current state of the HeliosApp.

| Field | Type | Description |
|-------|------|-------------|
| `conditions` | []Condition | Standard Kubernetes conditions for resource state. |
| `lastBuildTime` | metav1.Time | Timestamp of the last build attempt. |
| `lastBuildStatus` | string | Status of last build: "Succeeded", "Failed", "Running", "Unknown". |
| `lastDeployTime` | metav1.Time | Timestamp of the last deployment. |
| `argocdSyncStatus` | string | ArgoCD sync status: "Synced", "OutOfSync", "Unknown". |
| `argocdHealthStatus` | string | ArgoCD health status: "Healthy", "Progressing", "Degraded", "Unknown". |
| `observedGeneration` | int64 | Generation of the spec that was last reconciled. |
| `deploymentReplicas` | DeploymentReplicasStatus | Current replica counts. |
| `pipelineRunName` | string | Name of the most recent PipelineRun. |

#### DeploymentReplicasStatus

| Field | Type | Description |
|-------|------|-------------|
| `desired` | int32 | Desired number of replicas from spec. |
| `ready` | int32 | Number of replicas that are ready. |
| `available` | int32 | Number of replicas that are available. |

## Examples

### Basic Application

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: simple-app
  namespace: default
spec:
  gitRepo: https://github.com/example/simple-app
  imageRepo: docker.io/example/simple-app
  port: 8080
```

### Complete Configuration

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: production-app
  namespace: production
spec:
  # Source Configuration
  gitRepo: https://github.com/example/production-app
  gitBranch: release-v1.0
  
  # Build Configuration
  imageRepo: gcr.io/my-project/production-app
  serviceAccount: build-pipeline-sa
  pvcName: production-app-cache
  
  # GitOps Configuration
  gitopsRepo: https://github.com/example/k8s-manifests
  gitopsPath: apps/production-app
  gitopsBranch: main
  
  # Application Configuration
  port: 3000
  replicas: 3
  
  # Webhook Configuration
  webhookSecret: github-webhook-token
```

### With Webhook Automation

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: webhook-app
  namespace: default
spec:
  gitRepo: https://github.com/example/webhook-app
  imageRepo: docker.io/example/webhook-app
  port: 8080
  replicas: 2
  webhookSecret: github-webhook-secret
  serviceAccount: tekton-pipeline-sa
```

## Field Validation Rules

### gitRepo

- **Format**: Must be a valid URL
- **Schemes**: https://, http://, ssh://, git://
- **Pattern**: Should follow `scheme://host/owner/repository` pattern
- **Example**: `https://github.com/example/my-app`

### imageRepo

- **Format**: `[registry/]namespace/repository[:tag]`
- **Components**:
  - Registry (optional): Domain name, can include port (e.g., `localhost:5000`)
  - Namespace: Lowercase alphanumeric with `.`, `-`, `_` separators
  - Repository: Lowercase alphanumeric with `.`, `-`, `_` separators
  - Tag (optional): Alphanumeric with `.`, `-`, `_`
- **Examples**:
  - `docker.io/mycompany/myapp`
  - `gcr.io/my-project/myapp:v1.0.0`
  - `localhost:5000/namespace/repo`

### port

- **Range**: 1-65535
- **Note**: Ports < 1024 may require special permissions

### replicas

- **Range**: 0-100
- **Zero replicas**: Pauses the application (warning issued on update)

## Immutable Fields

The following fields cannot be changed after creation:

- `gitRepo`: Source repository URL
- `imageRepo`: Image repository location

Attempting to modify these fields will result in a validation error.

## Update Warnings

The admission webhook provides warnings for potentially disruptive changes:

### Port Changes

Changing the port triggers a rolling update of all pods.

```yaml
# Warning: Changing port from 8080 to 3000 will trigger a rolling update
```

### Large Replica Increases

Scaling to more than 2x current replicas warns about resource requirements.

```yaml
# Warning: Scaling from 1 to 10 replicas is a large increase
```

### Scaling to Zero

Setting replicas to 0 pauses the application.

```yaml
# Warning: Setting replicas to 0 will stop all pods
```

## Status Conditions

### Condition Types

| Type | Description |
|------|-------------|
| `Ready` | Application is fully reconciled and healthy |
| `PipelineReady` | Tekton Pipeline is created and ready |
| `TriggersReady` | Tekton Triggers are configured |
| `ArgoCDReady` | ArgoCD Application is synced |

### Condition Status

- `True`: Condition is satisfied
- `False`: Condition is not satisfied
- `Unknown`: Condition state cannot be determined

### Example Status

```yaml
status:
  conditions:
  - type: Ready
    status: "True"
    reason: ReconciliationSucceeded
    message: All resources created successfully
    lastTransitionTime: "2025-10-16T10:00:00Z"
  - type: PipelineReady
    status: "True"
    reason: PipelineCreated
    message: Tekton Pipeline is ready
  - type: ArgoCDReady
    status: "True"
    reason: ApplicationSynced
    message: ArgoCD Application is synced and healthy
  lastBuildTime: "2025-10-16T09:55:00Z"
  lastBuildStatus: Succeeded
  argocdSyncStatus: Synced
  argocdHealthStatus: Healthy
  observedGeneration: 1
  deploymentReplicas:
    desired: 3
    ready: 3
    available: 3
  pipelineRunName: webhook-app-run-abc123
```

## Resource Ownership

All resources created by the operator have:

- **OwnerReferences**: Links to the parent HeliosApp
- **Labels**:
  - `app.kubernetes.io/name`: HeliosApp name
  - `app.kubernetes.io/managed-by`: "helios-operator"
  - `helios.io/app`: HeliosApp name

This ensures:
- Automatic cleanup when HeliosApp is deleted
- Easy resource discovery
- Proper garbage collection

## RBAC Requirements

To create HeliosApps, users need:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: heliosapp-user
rules:
- apiGroups: ["platform.helios.io"]
  resources: ["heliosapps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["platform.helios.io"]
  resources: ["heliosapps/status"]
  verbs: ["get"]
```

## Related Resources

When a HeliosApp is created, the operator automatically generates:

### Tekton Resources

- **Pipeline**: `<app-name>-pipeline`
- **EventListener**: `<app-name>-listener`
- **TriggerBinding**: `<app-name>-binding`
- **TriggerTemplate**: `<app-name>-template`

### ArgoCD Resources

- **Application**: `<app-name>-app`

## Metrics

The operator exposes Prometheus metrics for HeliosApp resources:

- `heliosapp_reconciliation_duration_seconds`: Time to reconcile
- `heliosapp_reconciliation_phase_duration_seconds`: Time per phase
- `heliosapp_builds_total`: Build count by status
- `heliosapp_deployment_health`: Current health status
- `heliosapp_argocd_sync_status`: ArgoCD sync status

See [Prometheus Metrics](../reference/prometheus-metrics.md) for details.

## Troubleshooting

### Common Issues

#### Pipeline Not Created

**Symptom**: Pipeline resource doesn't exist

**Solutions**:
1. Check operator logs: `kubectl logs -n helios-system deployment/helios-operator`
2. Verify RBAC permissions
3. Check HeliosApp status conditions

#### Build Failures

**Symptom**: `lastBuildStatus: Failed`

**Solutions**:
1. Check PipelineRun logs: `kubectl logs <pipelinerun-name>`
2. Verify ServiceAccount has image push permissions
3. Check git repository access

#### ArgoCD Not Syncing

**Symptom**: `argocdSyncStatus: OutOfSync`

**Solutions**:
1. Check ArgoCD Application: `kubectl get application <app-name>-app`
2. Verify GitOps repository access
3. Check manifest validity in gitopsRepo

## Best Practices

### 1. Use Separate GitOps Repository

```yaml
spec:
  gitRepo: https://github.com/example/app-source
  gitopsRepo: https://github.com/example/k8s-manifests
  gitopsPath: apps/my-app
```

### 2. Configure Resource Limits

Use PVC for build cache to improve performance:

```yaml
spec:
  pvcName: my-app-build-cache  # Reuse build artifacts
```

### 3. Use Meaningful Names

```yaml
metadata:
  name: production-payment-service  # Clear, descriptive name
  namespace: production             # Appropriate namespace
```

### 4. Enable Webhooks for Automation

```yaml
spec:
  webhookSecret: github-webhook-secret  # Automate builds on push
```

### 5. Monitor Status

```bash
# Watch application status
kubectl get heliosapp my-app -w

# Check detailed status
kubectl describe heliosapp my-app
```

## API Stability

- **v1**: Stable API, backward compatible changes only
- Changes follow Kubernetes API deprecation policy
- Breaking changes will be introduced in new API versions (v2, etc.)
