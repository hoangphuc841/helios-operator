# 🏗️ Helios Operator Architecture

This document provides a comprehensive overview of the Helios Operator architecture, components, and design decisions.

## Table of Contents

1. [Overview](#overview)
2. [High-Level Architecture](#high-level-architecture)
3. [Core Components](#core-components)
4. [Reconciliation Flow](#reconciliation-flow)
5. [Metrics and Observability](#metrics-and-observability)
6. [Configuration](#configuration)
7. [Security](#security)
8. [Testing](#testing)

## Overview

Helios Operator is a Kubernetes operator that provides a simplified interface for deploying applications using GitOps. It orchestrates Tekton Pipelines for CI/CD and ArgoCD Applications for continuous deployment, abstracting away the complexity of managing these tools directly.

**Key Features:**

- Declarative application deployment via HeliosApp CRD
- Automated CI/CD pipeline creation with Tekton
- GitOps-based deployment with ArgoCD
- Comprehensive metrics and observability
- Webhook-based Git integration
- Extensive validation and error handling

## High-Level Architecture

### System Overview

```mermaid
graph TB
    subgraph "Developer Workflow"
        Dev[Developer] -->|Git Push| GitRepo[Git Repository]
    end

    subgraph "Helios Operator"
        CRD[HeliosApp CRD] --> Controller[Controller/Reconciler]
        Controller --> PipelineGen[Pipeline Generator]
        Controller --> TriggerGen[Trigger Generator]
        Controller --> ArgoCDGen[ArgoCD Generator]
        Controller --> StatusMgr[Status Manager]
        Controller --> Metrics[Metrics Collector]
    end

    subgraph "Tekton Pipelines"
        EventListener[EventListener] --> TriggerBinding[TriggerBinding]
        TriggerBinding --> TriggerTemplate[TriggerTemplate]
        TriggerTemplate --> PipelineRun[PipelineRun]
        PipelineRun --> BuildPush[Build & Push Image]
    end

    subgraph "ArgoCD GitOps"
        ArgoCDApp[ArgoCD Application] --> GitSync[Git Sync]
        GitSync --> K8sDeploy[Kubernetes Deployment]
    end

    subgraph "Monitoring"
        Metrics --> Prometheus[Prometheus]
        Prometheus --> Grafana[Grafana Dashboards]
    end

    GitRepo -->|Webhook| EventListener
    PipelineGen --> Pipeline[Tekton Pipeline]
    TriggerGen --> EventListener
    ArgoCDGen --> ArgoCDApp
    BuildPush -->|Update Manifest| GitOpsRepo[GitOps Repository]
    GitOpsRepo --> GitSync
    StatusMgr -->|Update Status| CRD

    style Controller fill:#e1f5ff
    style Metrics fill:#fff4e1
    style BuildPush fill:#e8f5e9
```

### Component Interaction

```mermaid
sequenceDiagram
    participant User
    participant K8s as Kubernetes API
    participant Ctrl as Helios Controller
    participant Tekton
    participant ArgoCD
    participant Git as Git Repository

    User->>K8s: Create HeliosApp
    K8s->>Ctrl: Reconcile Event

    Ctrl->>K8s: Create Tekton Pipeline
    Ctrl->>K8s: Create EventListener
    Ctrl->>K8s: Create TriggerBinding
    Ctrl->>K8s: Create TriggerTemplate
    Ctrl->>K8s: Create ArgoCD Application
    Ctrl->>K8s: Update Status (Ready)

    Git->>Tekton: Webhook (Push Event)
    Tekton->>Tekton: Run PipelineRun
    Tekton->>Git: Push Updated Manifest

    ArgoCD->>Git: Sync GitOps Repo
    ArgoCD->>K8s: Deploy Application

    Ctrl->>K8s: Update Status (Deployed)
```

## Core Components

### 1. HeliosApp Controller

**Location:** `internal/controller/heliosapp_controller.go`

The main reconciliation controller that manages the entire lifecycle of HeliosApp resources with intelligent error handling and specialized phase management.

**Key Responsibilities:**

- Fetch and validate HeliosApp resources
- Manage finalizers for proper resource cleanup
- Reconcile Tekton Pipelines with intelligent error handling
- Reconcile Tekton Triggers (EventListener, TriggerBinding, TriggerTemplate)
- Reconcile ArgoCD Applications with smart retry logic
- Update comprehensive status with build and deployment information
- Handle transient errors with exponential backoff

**Enhanced Reconciliation Phases:**

1. **Fetch Phase**: Retrieve HeliosApp resource from cluster
2. **Finalizer Phase**: Handle finalizer logic for resource cleanup
3. **Pipeline Phase**: Create/update Tekton Pipeline with error recovery
4. **Triggers Phase**: Create/update webhook triggers with retry logic
5. **ArgoCD Phase**: Create/update ArgoCD Application with smart requeuing
6. **Status Phase**: Update status with build and deployment health

**Intelligent Error Handling:**

- **Transient Error Detection**: Automatically identifies temporary failures
- **Exponential Backoff**: Smart retry timing (30s → 60s → 120s → 5min max)
- **Graceful Degradation**: Continues processing other phases when one fails
- **Comprehensive Logging**: Phase-specific error context and metrics

### 2. Specialized Reconcile Functions

**Location:** `internal/controller/reconcile_phases.go`

Specialized functions for each reconciliation phase with intelligent error handling and detailed status reporting.

**Phase Functions:**

- **`ReconcilePipeline()`**: Handles Tekton Pipeline creation/updates with error recovery
- **`ReconcileTriggers()`**: Manages EventListener, TriggerBinding, TriggerTemplate resources
- **`ReconcileArgoCD()`**: Creates/updates ArgoCD Applications with smart retry logic
- **`ReconcileStatus()`**: Comprehensive status aggregation from all child resources

**Key Features:**

- **ReconcileResult Structure**: Each function returns detailed status information
- **Transient Error Handling**: Automatic retry with exponential backoff
- **Status Aggregation**: Comprehensive status reporting from all phases
- **Independent Testing**: Each phase can be tested in isolation

### 3. Resource Generators

**Location:** `internal/resources/`

Modules responsible for generating Kubernetes resources:

- **`argocd.go`**: ArgoCD Application resource generation
- **`tekton.go`**: Tekton Pipeline resource generation
- **`pipeline.go`**: Pipeline task definitions

**Key Functions:**

- Template-based resource generation
- Parameter substitution
- Ownership reference management

### 4. Webhook Validation

**Location:** `api/v1/heliosapp_webhook.go`

Admission webhook for validating HeliosApp resources before they're created or updated.

**Validations:**

- Git URL format and scheme validation
- Container image repository format
- Port range (1-65535)
- Replica limits (0-100)
- Immutable field enforcement (gitRepo, imageRepo)
- Update impact warnings

### 5. Metrics System

**Location:** `internal/controller/metrics.go`

Comprehensive Prometheus metrics for observability.

**Metric Categories:**

**Reconciliation Metrics:**

- `heliosapp_reconciliation_duration_seconds`: Overall reconciliation duration
- `heliosapp_reconciliation_phase_duration_seconds`: Per-phase duration
- `heliosapp_reconciliations_total`: Total reconciliation count
- `heliosapp_reconciliation_errors_total`: Errors by type and phase

**Build Metrics:**

- `heliosapp_builds_total`: Build count by status
- `heliosapp_argocd_sync_status`: ArgoCD sync status

**Health Metrics:**

- `heliosapp_deployment_health`: Deployment health status
- `heliosapp_replicas`: Replica counts (desired vs ready)

**API Metrics:**

- `heliosapp_api_calls_total`: Kubernetes API call count
- `heliosapp_api_call_duration_seconds`: API call latency

**Webhook Metrics:**

- `heliosapp_webhook_validations_total`: Validation attempts
- `heliosapp_webhook_validation_duration_seconds`: Validation duration

### 6. Health Checks

**Location:** `internal/health/health.go`

Health check endpoints for Kubernetes probes.

**Endpoints:**

- `/healthz`: Liveness probe (process responsiveness)
- `/readyz`: Readiness probe (API connectivity, webhook status)

**Checks:**

- Kubernetes API server connectivity
- Webhook certificate validity
- Leader election status

### 7. Configuration Management

**Location:** `internal/config/config.go`

Centralized configuration management with environment variable support.

**Configuration Options:**

- Reconciliation settings (interval, concurrency)
- Resource defaults (replicas, port, service account)
- Retry settings (max retries, backoff)
- Feature flags (metrics, webhooks, leader election)
- Timeouts (reconciliation, API calls)

## Reconciliation Flow

### High-Level Reconciliation Process

````mermaid
graph TD
    A[Reconcile Triggered] --> B{Resource Exists?}
    B -->|No| C[Handle Deletion]
    B -->|Yes| D[Fetch Phase]

    D --> E[Pipeline Phase]

### Detailed Reconciliation Phases

```mermaid
sequenceDiagram
    participant R as Reconciler
    participant K as Kubernetes API
    participant M as Metrics
    participant T as Tekton
    participant A as ArgoCD

    Note over R: Start Reconciliation
    R->>M: Start Overall Timer

    R->>K: Fetch HeliosApp
    K-->>R: HeliosApp Resource
    R->>M: Record Fetch Duration

    R->>K: Generate Pipeline
    R->>T: CreateOrUpdate Pipeline
    T-->>R: Pipeline Created/Updated
    R->>M: Record Pipeline Duration

    R->>K: Generate EventListener
    R->>T: CreateOrUpdate EventListener
    T-->>R: EventListener Ready

    R->>K: Generate TriggerBinding
    R->>T: CreateOrUpdate TriggerBinding
    T-->>R: TriggerBinding Ready

    R->>K: Generate TriggerTemplate
    R->>T: CreateOrUpdate TriggerTemplate
    T-->>R: TriggerTemplate Ready
    R->>M: Record Triggers Duration

    R->>K: Generate ArgoCD Application
    R->>A: CreateOrUpdate Application
    A-->>R: Application Created
    R->>M: Record ArgoCD Duration

    R->>K: Update HeliosApp Status
    R->>K: Set Conditions (Ready, Synced, BuildSucceeded)
    K-->>R: Status Updated
    R->>M: Record Status Duration

    R->>M: Record Overall Duration
    R->>M: Increment Success Counter
    Note over R: Reconciliation Complete
````

### Phase Breakdown

1. **Fetch Phase** (`fetch`)

   - Retrieve HeliosApp resource from Kubernetes API
   - Return if resource not found (deletion case)
   - Log fetch errors with context

2. **Pipeline Phase** (`pipeline`)

   - Generate Tekton Pipeline from template
   - Apply owner references
   - CreateOrUpdate Pipeline resource
   - Handle pipeline creation errors

3. **Triggers Phase** (`triggers`)

   - Generate EventListener for webhook reception
   - Generate TriggerBinding for parameter extraction
   - Generate TriggerTemplate for PipelineRun instantiation
   - CreateOrUpdate all trigger resources
   - Handle trigger creation errors

4. **ArgoCD Phase** (`argocd`)

   - Generate ArgoCD Application manifest
   - Configure GitOps repository sync
   - Apply sync policy and health checks
   - CreateOrUpdate Application resource
   - Handle ArgoCD errors

5. **Status Phase** (`status`)
   - Query ArgoCD Application sync status
   - Query latest Tekton PipelineRun status
   - Calculate deployment health
   - Update HeliosApp status conditions:
     - `Ready`: Overall health
     - `Synced`: ArgoCD sync status
     - `BuildSucceeded`: Last build status
   - Update status fields (deployedVersion, lastBuild, etc.)

### Watch Configuration

The controller watches multiple resource types and uses predicates to filter events:

```mermaid
graph LR
    A[HeliosApp Changes] --> C[Reconciler]
    B[ArgoCD App Changes] --> C
    D[PipelineRun Changes] --> C

    C --> E{Predicate Filter}
    E -->|Generation Changed| F[Full Reconcile]
    E -->|Status Only| G[Status Update Only]
    E -->|Irrelevant| H[Skip]
```

### Watch Triggers

The controller watches multiple resource types for changes:

**Primary Resource:**

- `HeliosApp`: Direct changes trigger reconciliation

**Secondary Resources (with predicates):**

- `ArgoCD Application`: Changes to sync/health status
- `Tekton PipelineRun`: Build completion or failure
- `Deployment`: Pod readiness changes

**Predicate Filtering:**
Each watch includes predicates to reduce unnecessary reconciliations:

- ArgoCD: Only apps with managed-by label
- PipelineRun: Only runs with app label
- Deployment: Only deployments with app label

## Metrics and Observability

### Prometheus Integration

All metrics are automatically registered with the controller-runtime metrics registry and exposed on the `/metrics` endpoint (default port 8443).

### Grafana Dashboards

Recommended dashboard panels:

1. **Reconciliation Performance**

   - Average reconciliation duration
   - P95/P99 reconciliation latency
   - Phase-by-phase breakdown

2. **Error Rates**

   - Errors by phase
   - Error rate over time
   - Error types distribution

3. **Application Health**

   - Deployment health by application
   - ArgoCD sync status
   - Build success rate

4. **Resource Metrics**
   - Managed resources count
   - API call rate and latency
   - Webhook validation performance

## Configuration

### Environment Variables

```bash
# Reconciliation Settings
RECONCILE_INTERVAL=5m
MAX_CONCURRENT_RECONCILES=3

# Resource Defaults
DEFAULT_REPLICAS=1
DEFAULT_PORT=8080
DEFAULT_SERVICE_ACCOUNT=default

# Retry Settings
MAX_RETRIES=3
RETRY_BACKOFF=30s

# Feature Flags
ENABLE_METRICS=true
ENABLE_WEBHOOKS=true
ENABLE_LEADER_ELECTION=true

# Timeouts
RECONCILE_TIMEOUT=10m
API_CALL_TIMEOUT=30s

# Namespace Watching
WATCH_NAMESPACE=  # Empty = all namespaces
```

### ConfigMap Support

Configuration can also be loaded from a ConfigMap (future enhancement).

## Security

### RBAC

The operator requires specific permissions:

**Core Permissions:**

- HeliosApp resources: Full access
- Tekton resources: Full access (Pipelines, PipelineRuns, Triggers)
- ArgoCD resources: Full access (Applications)
- Deployments: Read access for status

**Webhook Permissions:**

- ValidatingWebhookConfiguration: Update for webhook setup

### Pod Security Standards

The operator implements **restricted** Pod Security Standards compliance:

- **Non-root execution**: Runs as user 65532 with fsGroup 65532
- **Read-only root filesystem**: Container root filesystem is read-only with tmp volume
- **Dropped capabilities**: All Linux capabilities are dropped
- **Seccomp profile**: Runtime default seccomp profile enabled
- **No privilege escalation**: Privilege escalation is disabled
- **Namespace security labels**: All relevant namespaces enforce restricted security

For detailed security implementation, see [Pod Security Standards Documentation](../security/pod-security-standards.md).

### Admission Control

Webhook validation prevents:

- Invalid Git URLs
- Malformed image repositories
- Out-of-range ports or replicas
- Modification of immutable fields

## Error Handling

### Enhanced Error Handling System

**Location:** `internal/common/errors.go`

The operator implements intelligent error handling with automatic transient error detection and smart retry logic.

### Custom Error Types

- `ReconciliationError`: General reconciliation failures
- `ResourceGenerationError`: Resource generation issues
- `ValidationError`: Validation failures
- `StatusUpdateError`: Status update problems
- `ReconcileResult`: Comprehensive result structure with status information

All errors support wrapping with `Unwrap()` for error chain inspection.

### Transient Error Detection

**Function:** `IsTransientError()`

Automatically identifies errors that should be retried:

- **Kubernetes API Errors**: Conflict, ServerTimeout, ServiceUnavailable, Timeout, TooManyRequests, InternalError
- **Network Errors**: Temporary network failures, connection issues
- **Pattern Matching**: Detects common transient error messages

### Smart Retry Strategy

**Function:** `GetRequeueDelay()`

- **Exponential Backoff**: 30s → 60s → 120s → 240s → 5min max
- **Jitter**: ±25% randomness to prevent thundering herd
- **Phase-Specific Handling**: Each reconciliation phase handles errors independently
- **Graceful Degradation**: Continues processing other phases when one fails

### ReconcileResult Structure

Each reconciliation phase returns a detailed result:

```go
type ReconcileResult struct {
    Success      bool
    Error        error
    RequeueAfter time.Duration
    Status       ReconcileStatus
}
```

This enables:

- **Intelligent Requeuing**: Only requeue when necessary
- **Status Aggregation**: Comprehensive status from all phases
- **Error Context**: Detailed error information with phase context

## Testing

### Unit Tests

**Location:** `internal/controller/reconcile_phases_test.go`

- **Error Detection Tests**: Verify transient vs permanent error classification
- **Backoff Logic Tests**: Validate exponential backoff with jitter
- **Reconcile Result Tests**: Test result structure and behavior
- **Phase-Specific Tests**: Individual reconciliation phase testing
- **Controller Logic Tests**: Main reconciliation loop testing
- **Resource Generation Tests**: Resource creation and updates
- **Validation Tests**: Input validation and error handling
- **Metrics Tests**: Prometheus metrics collection

### Integration Tests

- **envtest**: Kubernetes API simulation for full workflow testing
- **Reconciliation Workflow**: End-to-end reconciliation testing
- **Watch Predicate Tests**: Event filtering and triggering
- **Error Recovery Tests**: Transient error handling and retry logic

### Test Coverage

- **Minimum 80% code coverage** across all packages
- **Critical paths: 100% coverage** (error handling, reconciliation logic)
- **Phase-specific coverage**: Each reconcile function thoroughly tested
- **Error scenario coverage**: All error types and retry scenarios tested

## Future Enhancements

1. **Multi-Tenancy**: Namespace isolation and resource quotas
2. **Custom Builders**: Support for different build tools (Maven, Gradle, etc.)
3. **Canary Deployments**: Progressive delivery with ArgoCD Rollouts
4. **Cost Optimization**: Resource recommendations and auto-scaling
5. **Backup/Restore**: HeliosApp resource backup and disaster recovery
