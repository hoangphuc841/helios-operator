# 🚀 Getting Started with Helios

This guide will walk you through deploying your first application with Helios Operator in under 15 minutes.

## 📋 **Prerequisites**

Before you begin, ensure you have:

- ✅ Kubernetes cluster (1.19+) with cluster admin access
- ✅ Helm 3.0+ installed
- ✅ `kubectl` configured to access your cluster
- ✅ A Git repository with your application source code
- ✅ A container registry (Docker Hub, GitHub Container Registry, etc.)

## 🏗️ **Step 1: Install Dependencies**

### Install Tekton Pipelines

```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
```

### Install ArgoCD

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update
kubectl create namespace argocd
helm install argocd argo/argo-cd --namespace argocd
```

### Wait for Dependencies

```bash
# Wait for Tekton to be ready
kubectl wait --for=condition=ready pod -l app=tekton-pipelines-controller -n tekton-pipelines --timeout=300s

# Wait for ArgoCD to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s
```

## 🎯 **Step 2: Install Helios Operator**

### Install via Helm

```bash
# Add Helm repository (when available)
helm repo add helios-operator https://hoangphuc841.github.io/helios-operator
helm repo update

# Install Helios Operator
helm install helios-operator helios-operator/helios-operator
```

### Verify Installation

```bash
# Check operator is running
kubectl get pods -n system -l app.kubernetes.io/name=helios-operator

# Verify CRD is installed
kubectl get crd heliosapps.platform.helios.io
```

## 📝 **Step 3: Prepare Your Application**

### Create Application Namespace

```bash
kubectl create namespace my-apps
```

### Create PVC for Build Workspace

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-app-pvc
  namespace: my-apps
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF
```

### Create Service Account for Pipeline

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pipeline-sa
  namespace: my-apps
EOF
```

## 🚀 **Step 4: Deploy Your First Application**

### Create HeliosApp Resource

```bash
kubectl apply -f - <<EOF
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-first-app
  namespace: my-apps
spec:
  # Source code repository
  gitRepo: "https://github.com/your-username/your-app.git"
  gitBranch: "main"

  # GitOps repository (where deployment manifests are stored)
  gitopsRepo: "https://github.com/your-username/your-gitops.git"
  gitopsPath: "my-first-app"  # Path in GitOps repo
  gitopsBranch: "main"

  # Container image configuration
  imageRepo: "docker.io/your-username/my-first-app"
  port: 80
  replicas: 1

  # Pipeline configuration
  serviceAccount: "pipeline-sa"
  webhookSecret: "github-webhook-secret"
  pvcName: "my-app-pvc"
EOF
```

### Watch the Magic Happen ✨

```bash
# Watch operator create resources
kubectl get heliosapp my-first-app -w

# Check created resources
kubectl get pipelines,eventlisteners,triggerbindings,triggertemplates -n my-apps
kubectl get applications -n argocd
```

## 🎉 **Step 5: Verify Your Deployment**

### Check Application Status

```bash
# Get detailed status
kubectl describe heliosapp my-first-app -n my-apps

# Check ArgoCD Application
kubectl describe application my-first-app-argocd -n argocd
```

### Access ArgoCD UI (Optional)

```bash
# Port forward to ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Get admin password
kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" | base64 -d && echo

# Access: https://localhost:8080
# Username: admin
# Password: (from above command)
```

## 🔄 **Step 6: Test the Complete Flow**

### Trigger a Build (Optional)

```bash
# Create a test PipelineRun to simulate a Git push
kubectl apply -f - <<EOF
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: my-first-app-build-test
  namespace: my-apps
  labels:
    helios.io/app-name: "my-first-app"
    helios.io/managed-by: "helios-operator"
spec:
  pipelineRef:
    name: my-first-app-pipeline
  serviceAccountName: pipeline-sa
  params:
  - name: git-repo
    value: "https://github.com/your-username/your-app.git"
  - name: git-revision
    value: "main"
  - name: image-repo
    value: "docker.io/your-username/my-first-app"
  workspaces:
  - name: source-code
    persistentVolumeClaim:
      claimName: my-app-pvc
EOF
```

### Monitor Build Progress

```bash
# Watch PipelineRun status
kubectl get pipelinerun -n my-apps -w

# Check build logs
kubectl logs -n my-apps -l tekton.dev/pipelineRun=my-first-app-build-test
```

## 🎯 **What Happened?**

When you created the `HeliosApp`, the operator automatically:

1. ✅ **Generated Tekton Pipeline** (`my-first-app-pipeline`) from template
2. ✅ **Created Tekton Triggers** (EventListener, TriggerBinding, TriggerTemplate)
3. ✅ **Created ArgoCD Application** (`my-first-app-argocd`) for GitOps deployment
4. ✅ **Set up Webhook Integration** for automatic builds on Git pushes
5. ✅ **Configured Real-time Monitoring** with Kubernetes Watches

## 🔧 **Next Steps**

### Configure Webhooks (Optional)

To enable automatic builds on Git pushes:

1. Go to your GitHub repository settings
2. Navigate to "Webhooks"
3. Add webhook URL: `http://your-cluster-ip:8080/`
4. Set secret to match `webhookSecret` in your HeliosApp

### Set up GitOps Repository

Create deployment manifests in your GitOps repository:

```yaml
# my-gitops-repo/my-first-app/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-first-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-first-app
  template:
    metadata:
      labels:
        app: my-first-app
    spec:
      containers:
        - name: app
          image: docker.io/your-username/my-first-app:latest
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: my-first-app-service
spec:
  selector:
    app: my-first-app
  ports:
    - port: 80
      targetPort: 80
  type: LoadBalancer
```

## 🆘 **Need Help?**

- **Issues**: Check our [Troubleshooting Guide](03-troubleshooting.md)
- **Configuration**: See [HeliosApp Specification](02-helios-app-spec.md)
- **Support**: Open an [issue on GitHub](https://github.com/hoangphuc841/helios-operator/issues)

---

**🎉 Congratulations!** You've successfully deployed your first application with Helios Operator. Ready to deploy more? Check out the [HeliosApp Specification](02-helios-app-spec.md) for advanced configuration options.
