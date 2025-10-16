# 📚 Helios Operator Examples

This guide provides comprehensive real-world examples for using Helios Operator, from basic scenarios to advanced configurations.

## Table of Contents

1. [Basic Examples](#basic-examples)
2. [Intermediate Examples](#intermediate-examples)
3. [Advanced Examples](#advanced-examples)
4. [Complete Real-World Scenarios](#complete-real-world-scenarios)

---

## Basic Examples

### Example 1: Simple Web Application

The most basic HeliosApp configuration for a simple web application.

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: simple-webapp
  namespace: default
spec:
  # Source repository
  gitRepo: https://github.com/example/simple-webapp
  gitBranch: main

  # Container configuration
  imageRepo: docker.io/myorg/simple-webapp
  port: 8080
  replicas: 2

  # GitOps configuration
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: apps/simple-webapp
  gitopsBranch: main

  # Tekton configuration
  serviceAccount: default
  webhookSecret: github-webhook
```

### Example 2: Application with Custom PVC

Using a pre-created PersistentVolumeClaim for build workspace.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: webapp-build-cache
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: fast-ssd
---
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: webapp-with-pvc
  namespace: default
spec:
  gitRepo: https://github.com/example/webapp
  imageRepo: docker.io/myorg/webapp
  port: 8080
  replicas: 1
  serviceAccount: pipeline-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/gitops

  # Use custom PVC for better performance
  pvcName: webapp-build-cache
```

### Example 3: Multi-Environment Setup

Simple configuration for different environments using different branches.

**Production Environment:**

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: myapp-production
  namespace: production
spec:
  gitRepo: https://github.com/example/myapp
  gitBranch: main # Production from main branch
  imageRepo: docker.io/myorg/myapp
  port: 8080
  replicas: 3 # Higher replicas for production
  serviceAccount: production-sa
  webhookSecret: github-webhook-prod
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: production/myapp
  gitopsBranch: main
```

**Staging Environment:**

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: myapp-staging
  namespace: staging
spec:
  gitRepo: https://github.com/example/myapp
  gitBranch: develop # Staging from develop branch
  imageRepo: docker.io/myorg/myapp
  port: 8080
  replicas: 1 # Fewer replicas for staging
  serviceAccount: staging-sa
  webhookSecret: github-webhook-staging
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: staging/myapp
  gitopsBranch: staging
```

---

## Intermediate Examples

### Example 4: Microservices Application

Deploying multiple services that work together.

```yaml
# API Service
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: api-service
  namespace: microservices
spec:
  gitRepo: https://github.com/example/api-service
  imageRepo: docker.io/myorg/api-service
  port: 8080
  replicas: 3
  serviceAccount: microservices-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: microservices/api-service
---
# Frontend Service
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: frontend
  namespace: microservices
spec:
  gitRepo: https://github.com/example/frontend
  imageRepo: docker.io/myorg/frontend
  port: 3000
  replicas: 2
  serviceAccount: microservices-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: microservices/frontend
---
# Worker Service
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: worker
  namespace: microservices
spec:
  gitRepo: https://github.com/example/worker
  imageRepo: docker.io/myorg/worker
  port: 8081
  replicas: 2
  serviceAccount: microservices-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: microservices/worker
```

### Example 5: Application with Custom Service Account and Registry Credentials

Full setup with image registry authentication.

```yaml
# Docker Registry Secret
apiVersion: v1
kind: Secret
metadata:
  name: docker-registry-creds
  namespace: myapp
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: BASE64_ENCODED_DOCKER_CONFIG
---
# Service Account with Image Pull Secrets
apiVersion: v1
kind: ServiceAccount
metadata:
  name: myapp-pipeline-sa
  namespace: myapp
secrets:
  - name: docker-registry-creds
imagePullSecrets:
  - name: docker-registry-creds
---
# Role for Pipeline SA
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pipeline-role
  namespace: myapp
rules:
  - apiGroups: [""]
    resources: ["secrets", "configmaps"]
    verbs: ["get", "list"]
  - apiGroups: ["tekton.dev"]
    resources: ["pipelineruns", "taskruns"]
    verbs: ["get", "list", "create", "update"]
---
# RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pipeline-rolebinding
  namespace: myapp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pipeline-role
subjects:
  - kind: ServiceAccount
    name: myapp-pipeline-sa
    namespace: myapp
---
# GitHub Webhook Secret
apiVersion: v1
kind: Secret
metadata:
  name: github-webhook-secret
  namespace: myapp
type: Opaque
stringData:
  webhook-secret: "your-github-webhook-secret-here"
---
# HeliosApp
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: secure-app
  namespace: myapp
spec:
  gitRepo: https://github.com/example/secure-app
  imageRepo: my-registry.com/myorg/secure-app
  port: 8080
  replicas: 2
  serviceAccount: myapp-pipeline-sa
  webhookSecret: github-webhook-secret
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: apps/secure-app
```

### Example 6: Application with Labels and Annotations

Adding custom labels and annotations for better organization.

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: labeled-app
  namespace: default
  labels:
    app: my-application
    team: platform
    env: production
    version: v1.0.0
  annotations:
    contact: "platform-team@example.com"
    description: "Production web application"
    documentation: "https://docs.example.com/apps/labeled-app"
spec:
  gitRepo: https://github.com/example/labeled-app
  imageRepo: docker.io/myorg/labeled-app
  port: 8080
  replicas: 3
  serviceAccount: default
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: production/labeled-app
```

---

## Advanced Examples

### Example 7: Monorepo with Multiple Applications

Deploying multiple applications from a single monorepo.

```yaml
# Backend API
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: monorepo-backend
  namespace: monorepo-apps
spec:
  gitRepo: https://github.com/example/monorepo
  gitBranch: main
  imageRepo: docker.io/myorg/backend
  port: 8080
  replicas: 3
  serviceAccount: monorepo-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/monorepo-gitops
  gitopsPath: backend
---
# Frontend
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: monorepo-frontend
  namespace: monorepo-apps
spec:
  gitRepo: https://github.com/example/monorepo
  gitBranch: main
  imageRepo: docker.io/myorg/frontend
  port: 3000
  replicas: 2
  serviceAccount: monorepo-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/monorepo-gitops
  gitopsPath: frontend
---
# Admin Dashboard
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: monorepo-admin
  namespace: monorepo-apps
spec:
  gitRepo: https://github.com/example/monorepo
  gitBranch: main
  imageRepo: docker.io/myorg/admin
  port: 8081
  replicas: 1
  serviceAccount: monorepo-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/monorepo-gitops
  gitopsPath: admin
```

### Example 8: High-Availability Production Setup

Production-grade configuration with high availability and monitoring.

```yaml
# Namespace with resource quotas
apiVersion: v1
kind: Namespace
metadata:
  name: ha-production
  labels:
    environment: production
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: production-quota
  namespace: ha-production
spec:
  hard:
    requests.cpu: "50"
    requests.memory: "100Gi"
    persistentvolumeclaims: "10"
---
# ServiceAccount with proper RBAC
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ha-app-sa
  namespace: ha-production
---
# PVC with SSD storage
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ha-app-build-cache
  namespace: ha-production
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast-ssd
---
# Webhook Secret
apiVersion: v1
kind: Secret
metadata:
  name: ha-webhook-secret
  namespace: ha-production
type: Opaque
stringData:
  webhook-secret: "production-webhook-secret-change-me"
---
# HeliosApp with HA configuration
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: ha-production-app
  namespace: ha-production
  labels:
    app: critical-service
    tier: production
    sla: high
  annotations:
    monitoring: "enabled"
    alerts: "slack-channel-prod"
    oncall: "platform-team"
spec:
  # Source configuration
  gitRepo: https://github.com/example/ha-app
  gitBranch: release # Use release branch for production

  # Image configuration
  imageRepo: my-registry.com/prod/ha-app
  port: 8080
  replicas: 5 # High replica count for HA

  # GitOps configuration
  gitopsRepo: https://github.com/example/ha-gitops
  gitopsPath: production/ha-app
  gitopsBranch: production

  # Pipeline configuration
  serviceAccount: ha-app-sa
  webhookSecret: ha-webhook-secret
  pvcName: ha-app-build-cache
```

### Example 9: Blue-Green Deployment Strategy

Setting up blue-green deployments using separate HeliosApps.

```yaml
# Blue Environment (Currently Active)
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: myapp-blue
  namespace: production
  labels:
    app: myapp
    slot: blue
    active: "true"
spec:
  gitRepo: https://github.com/example/myapp
  gitBranch: v1.2.3 # Current production version
  imageRepo: docker.io/myorg/myapp
  port: 8080
  replicas: 3
  serviceAccount: production-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: production/blue
---
# Green Environment (Staging New Version)
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: myapp-green
  namespace: production
  labels:
    app: myapp
    slot: green
    active: "false"
spec:
  gitRepo: https://github.com/example/myapp
  gitBranch: v1.3.0 # New version being tested
  imageRepo: docker.io/myorg/myapp
  port: 8080
  replicas: 1 # Fewer replicas for testing
  serviceAccount: production-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/gitops
  gitopsPath: production/green
```

---

## Complete Real-World Scenarios

### Scenario 1: E-Commerce Platform

Complete setup for an e-commerce platform with multiple services.

```yaml
# 1. Namespace Setup
apiVersion: v1
kind: Namespace
metadata:
  name: ecommerce
  labels:
    environment: production
---
# 2. Shared Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ecommerce-sa
  namespace: ecommerce
---
# 3. Shared Webhook Secret
apiVersion: v1
kind: Secret
metadata:
  name: github-webhook
  namespace: ecommerce
type: Opaque
stringData:
  webhook-secret: "ecommerce-webhook-secret"
---
# 4. Product Catalog Service
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: product-catalog
  namespace: ecommerce
  labels:
    service: product-catalog
    team: backend
spec:
  gitRepo: https://github.com/example/product-catalog
  imageRepo: docker.io/myorg/product-catalog
  port: 8080
  replicas: 3
  serviceAccount: ecommerce-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/ecommerce-gitops
  gitopsPath: services/product-catalog
---
# 5. Shopping Cart Service
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: shopping-cart
  namespace: ecommerce
  labels:
    service: shopping-cart
    team: backend
spec:
  gitRepo: https://github.com/example/shopping-cart
  imageRepo: docker.io/myorg/shopping-cart
  port: 8080
  replicas: 3
  serviceAccount: ecommerce-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/ecommerce-gitops
  gitopsPath: services/shopping-cart
---
# 6. Order Service
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: order-service
  namespace: ecommerce
  labels:
    service: order-service
    team: backend
spec:
  gitRepo: https://github.com/example/order-service
  imageRepo: docker.io/myorg/order-service
  port: 8080
  replicas: 3
  serviceAccount: ecommerce-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/ecommerce-gitops
  gitopsPath: services/order-service
---
# 7. Payment Service
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: payment-service
  namespace: ecommerce
  labels:
    service: payment-service
    team: backend
    security: high
spec:
  gitRepo: https://github.com/example/payment-service
  imageRepo: docker.io/myorg/payment-service
  port: 8080
  replicas: 5 # More replicas for critical service
  serviceAccount: ecommerce-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/ecommerce-gitops
  gitopsPath: services/payment-service
---
# 8. Web Frontend
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: web-frontend
  namespace: ecommerce
  labels:
    service: web-frontend
    team: frontend
spec:
  gitRepo: https://github.com/example/web-frontend
  imageRepo: docker.io/myorg/web-frontend
  port: 3000
  replicas: 4
  serviceAccount: ecommerce-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/ecommerce-gitops
  gitopsPath: services/web-frontend
---
# 9. Mobile API Gateway
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: mobile-api
  namespace: ecommerce
  labels:
    service: mobile-api
    team: mobile
spec:
  gitRepo: https://github.com/example/mobile-api
  imageRepo: docker.io/myorg/mobile-api
  port: 8080
  replicas: 3
  serviceAccount: ecommerce-sa
  webhookSecret: github-webhook
  gitopsRepo: https://github.com/example/ecommerce-gitops
  gitopsPath: services/mobile-api
```

### Scenario 2: SaaS Multi-Tenant Application

Setting up a multi-tenant SaaS application with tenant isolation.

```yaml
# Tenant A - Enterprise Customer
apiVersion: v1
kind: Namespace
metadata:
  name: tenant-a
  labels:
    tenant: enterprise-corp
    tier: premium
---
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: saas-app
  namespace: tenant-a
  labels:
    tenant: tenant-a
    tier: premium
spec:
  gitRepo: https://github.com/example/saas-app
  gitBranch: main
  imageRepo: docker.io/myorg/saas-app
  port: 8080
  replicas: 5 # Premium tier gets more resources
  serviceAccount: tenant-a-sa
  webhookSecret: tenant-a-webhook
  gitopsRepo: https://github.com/example/saas-gitops
  gitopsPath: tenants/tenant-a
---
# Tenant B - Startup Customer
apiVersion: v1
kind: Namespace
metadata:
  name: tenant-b
  labels:
    tenant: startup-inc
    tier: standard
---
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: saas-app
  namespace: tenant-b
  labels:
    tenant: tenant-b
    tier: standard
spec:
  gitRepo: https://github.com/example/saas-app
  gitBranch: main
  imageRepo: docker.io/myorg/saas-app
  port: 8080
  replicas: 2 # Standard tier gets fewer resources
  serviceAccount: tenant-b-sa
  webhookSecret: tenant-b-webhook
  gitopsRepo: https://github.com/example/saas-gitops
  gitopsPath: tenants/tenant-b
```

---

## Tips and Best Practices

### 1. **Naming Conventions**

- Use descriptive names: `<service>-<environment>` (e.g., `api-production`)
- Include team or project names for multi-team environments
- Keep names lowercase and use hyphens

### 2. **Resource Management**

- Set appropriate replica counts based on load
- Use PVCs with appropriate storage classes for build caches
- Consider resource quotas for namespaces

### 3. **Security**

- Always use dedicated ServiceAccounts per namespace
- Store webhook secrets in Kubernetes Secrets
- Use RBAC to limit ServiceAccount permissions
- Rotate webhook secrets regularly

### 4. **GitOps Organization**

- Organize gitopsPath by environment: `production/app`, `staging/app`
- Use separate GitOps repos for different environments
- Keep GitOps manifests simple and declarative

### 5. **Monitoring**

- Add labels for easier querying and alerting
- Use annotations for documentation
- Monitor HeliosApp status conditions

### 6. **Multi-Environment**

- Use different namespaces for different environments
- Use different branches (main for prod, develop for staging)
- Scale replicas appropriately per environment

---

## Next Steps

- **Monitoring**: Set up [Monitoring and Observability](05-monitoring.md)
- **Troubleshooting**: Check the [Troubleshooting Guide](03-troubleshooting.md)
- **API Reference**: Review the [HeliosApp Spec Reference](02-helios-app-spec.md)
