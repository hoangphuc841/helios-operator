# Helios Operator Examples

This directory contains example configurations for the Helios Operator, demonstrating different use cases and deployment scenarios.

## Examples

### Simple Application (`simple-heliosapp.yaml`)

A basic HeliosApp configuration suitable for development and testing:

- Single repository setup
- Default configuration values
- Minimal resource requirements
- Perfect for getting started

**Usage:**

```bash
kubectl apply -f examples/simple-heliosapp.yaml
```

### Production Application (`production-heliosapp.yaml`)

A production-ready HeliosApp configuration with:

- Custom labels and annotations
- High availability (5 replicas)
- Custom PVC for larger builds
- Structured GitOps path
- Production namespace

**Usage:**

```bash
kubectl apply -f examples/production-heliosapp.yaml
```

## Prerequisites

Before applying these examples, ensure you have:

1. **Helios Operator installed** in your cluster
2. **Tekton Pipelines** installed and running
3. **ArgoCD** installed and running
4. **Service Account** with appropriate permissions
5. **GitHub Webhook Secret** created

## Required Resources

### Service Account

Create a service account for Tekton pipelines:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: simple-app-sa
  namespace: default
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: production-app-sa
  namespace: production
```

### GitHub Webhook Secret

Create a secret containing your GitHub webhook secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-webhook-secret
  namespace: default
type: Opaque
stringData:
  webhook-secret: "your-github-webhook-secret"
```

## Customization

You can customize these examples by modifying:

- **Repository URLs**: Update `gitRepo` and `gitopsRepo` to point to your repositories
- **Image Registry**: Change `imageRepo` to your container registry
- **Port and Replicas**: Adjust based on your application requirements
- **Namespace**: Deploy to your preferred namespace
- **Labels and Annotations**: Add metadata for better organization

## Troubleshooting

If you encounter issues:

1. Check the HeliosApp status: `kubectl describe heliosapp <name>`
2. Verify Tekton resources: `kubectl get pipelines,pipelineruns`
3. Check ArgoCD Application: `kubectl get applications -n argocd`
4. Review operator logs: `kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager`

For more information, see the [troubleshooting guide](../../docs/user-guide/03-troubleshooting.md).
