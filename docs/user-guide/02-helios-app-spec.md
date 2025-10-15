# 📋 HeliosApp Specification

The `HeliosApp` Custom Resource is the heart of Helios Operator. This document explains every field and configuration option available.

## 🏗️ **Basic Structure**

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-app
  namespace: my-namespace
spec:
  # Configuration fields go here
```

## 📝 **Specification Fields**

### **Source Code Configuration**

#### `gitRepo` (Required)

**Type**: `string`  
**Description**: URL of your source code repository

```yaml
spec:
  gitRepo: "https://github.com/your-org/your-app.git"
```

**Supported formats**:

- `https://github.com/owner/repo.git`
- `https://gitlab.com/owner/repo.git`
- `https://bitbucket.org/owner/repo.git`

#### `gitBranch` (Optional)

**Type**: `string`  
**Default**: `"main"`  
**Description**: Branch to monitor for changes

```yaml
spec:
  gitBranch: "develop" # Defaults to "main" if not specified
```

### **GitOps Configuration**

#### `gitopsRepo` (Required)

**Type**: `string`  
**Description**: URL of your GitOps repository containing deployment manifests

```yaml
spec:
  gitopsRepo: "https://github.com/your-org/your-gitops.git"
```

#### `gitopsPath` (Optional)

**Type**: `string`  
**Default**: `{metadata.name}` (app name)  
**Description**: Path within GitOps repository where manifests are stored

```yaml
spec:
  gitopsPath: "apps/my-app" # Defaults to "my-app" if not specified
```

**Example GitOps repository structure**:

```
my-gitops-repo/
├── my-app/              # <- gitopsPath
│   ├── deployment.yaml
│   ├── service.yaml
│   └── configmap.yaml
└── another-app/
    ├── deployment.yaml
    └── service.yaml
```

#### `gitopsBranch` (Optional)

**Type**: `string`  
**Default**: `"main"`  
**Description**: Branch in GitOps repository to monitor

```yaml
spec:
  gitopsBranch: "production" # Defaults to "main" if not specified
```

### **Container Image Configuration**

#### `imageRepo` (Required)

**Type**: `string`  
**Description**: Container image repository URL

```yaml
spec:
  imageRepo: "docker.io/your-org/my-app"
```

**Examples**:

- `docker.io/your-org/my-app`
- `ghcr.io/your-org/my-app`
- `your-registry.com/your-org/my-app`

#### `port` (Optional)

**Type**: `integer`  
**Default**: `80`  
**Description**: Port your application listens on

```yaml
spec:
  port: 8080 # Defaults to 80 if not specified
```

#### `replicas` (Optional)

**Type**: `integer`  
**Default**: `1`  
**Description**: Number of application replicas

```yaml
spec:
  replicas: 3 # Defaults to 1 if not specified
```

### **Pipeline Configuration**

#### `serviceAccount` (Required)

**Type**: `string`  
**Description**: Kubernetes ServiceAccount for Tekton Pipeline execution

```yaml
spec:
  serviceAccount: "pipeline-sa"
```

**Requirements**:

- Must exist in the same namespace as HeliosApp
- Needs permissions to push to your container registry
- Should have access to the specified PVC

#### `webhookSecret` (Required)

**Type**: `string`  
**Description**: Secret for validating GitHub/GitLab webhooks

```yaml
spec:
  webhookSecret: "my-webhook-secret"
```

**Security**: Use a strong, random secret. Generate one with:

```bash
openssl rand -hex 32
```

#### `pvcName` (Required)

**Type**: `string`  
**Description**: PersistentVolumeClaim for Tekton workspace

```yaml
spec:
  pvcName: "my-app-pvc"
```

**Requirements**:

- Must exist in the same namespace as HeliosApp
- Should have sufficient storage for your build process
- Access mode: `ReadWriteOnce`

## 🎯 **Complete Example**

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: e-commerce-api
  namespace: production
  labels:
    app.kubernetes.io/name: e-commerce-api
    app.kubernetes.io/version: "1.0.0"
    environment: production
spec:
  # Source code
  gitRepo: "https://github.com/acme-corp/e-commerce-api.git"
  gitBranch: "main"

  # GitOps deployment
  gitopsRepo: "https://github.com/acme-corp/gitops.git"
  gitopsPath: "production/e-commerce-api"
  gitopsBranch: "production"

  # Container configuration
  imageRepo: "ghcr.io/acme-corp/e-commerce-api"
  port: 3000
  replicas: 5

  # Pipeline configuration
  serviceAccount: "e-commerce-pipeline-sa"
  webhookSecret: "e-commerce-webhook-secret-xyz"
  pvcName: "e-commerce-build-workspace"
```

## 📊 **Status Fields**

The operator automatically updates these status fields:

### `status.conditions`

Array of condition objects indicating the current state:

```yaml
status:
  conditions:
    - type: "ApplicationSynced"
      status: "True"
      reason: "Synced"
      message: "ArgoCD Application is synced"
      lastTransitionTime: "2023-10-15T10:30:00Z"
    - type: "PipelineReady"
      status: "True"
      reason: "Created"
      message: "Tekton Pipeline created successfully"
      lastTransitionTime: "2023-10-15T10:29:00Z"
```

### `status.deployedVersion`

Current deployed version (updated when ArgoCD syncs):

```yaml
status:
  deployedVersion: "abc123def456"
```

### `status.lastBuildStatus`

Status of the last build:

```yaml
status:
  lastBuildStatus: "Succeeded"
```

## 🔧 **Advanced Configuration**

### Labels and Annotations

Add custom labels and annotations to your HeliosApp:

```yaml
metadata:
  name: my-app
  namespace: my-namespace
  labels:
    app.kubernetes.io/name: my-app
    app.kubernetes.io/version: "1.0.0"
    team: backend
    environment: staging
  annotations:
    helios.io/description: "Main application backend"
    helios.io/contact: "backend-team@company.com"
```

### Multiple Environments

Deploy the same application to multiple environments:

```yaml
# staging.yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-app-staging
  namespace: staging
spec:
  gitRepo: "https://github.com/your-org/your-app.git"
  gitopsRepo: "https://github.com/your-org/gitops.git"
  gitopsPath: "staging/my-app"
  imageRepo: "docker.io/your-org/my-app"
  # ... other fields

---
# production.yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-app-production
  namespace: production
spec:
  gitRepo: "https://github.com/your-org/your-app.git"
  gitopsRepo: "https://github.com/your-org/gitops.git"
  gitopsPath: "production/my-app"
  imageRepo: "docker.io/your-org/my-app"
  replicas: 5 # More replicas for production
  # ... other fields
```

## 🚨 **Validation Rules**

The operator validates your HeliosApp before creating resources:

### Required Fields

- `gitRepo` - Must be a valid Git URL
- `gitopsRepo` - Must be a valid Git URL
- `imageRepo` - Must be a valid container image name
- `serviceAccount` - Must reference an existing ServiceAccount
- `webhookSecret` - Must not be empty
- `pvcName` - Must reference an existing PVC

### Field Validation

- `port` - Must be between 1 and 65535
- `replicas` - Must be a positive integer
- `gitBranch` - Must be a valid Git branch name
- `gitopsBranch` - Must be a valid Git branch name

## 🎯 **Best Practices**

### 1. Naming Conventions

```yaml
# Use descriptive, consistent names
metadata:
  name: user-service-api
  namespace: backend-services
```

### 2. Resource Organization

```yaml
# Group related applications in the same namespace
metadata:
  namespace: backend-services # All backend services together
```

### 3. GitOps Repository Structure

```
gitops-repo/
├── environments/
│   ├── staging/
│   │   ├── user-service/
│   │   └── payment-service/
│   └── production/
│       ├── user-service/
│       └── payment-service/
└── shared/
    ├── monitoring/
    └── networking/
```

### 4. Security

```yaml
# Use strong webhook secrets
spec:
  webhookSecret: "super-secure-random-string-here"

# Use dedicated service accounts
spec:
  serviceAccount: "my-app-pipeline-sa"
```

## 🆘 **Troubleshooting**

### Common Issues

#### Invalid Git URL

```bash
# Check if repository is accessible
git ls-remote https://github.com/your-org/your-repo.git
```

#### Missing ServiceAccount

```bash
# Create the required ServiceAccount
kubectl create serviceaccount pipeline-sa -n my-namespace
```

#### PVC Not Found

```bash
# Create the required PVC
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-app-pvc
  namespace: my-namespace
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF
```

### Validation Errors

Check the HeliosApp status for validation errors:

```bash
kubectl describe heliosapp my-app -n my-namespace
```

---

**Need more help?** Check our [Troubleshooting Guide](03-troubleshooting.md) or open an [issue on GitHub](https://github.com/hoangphuc841/helios-operator/issues) 🆘
