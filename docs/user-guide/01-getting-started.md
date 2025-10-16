# 🚀 Getting Started with Helios Operator

This guide will help you deploy your first application with Helios Operator in under 15 minutes. While this guide focuses on **local deployment** using Minikube, Kind, or K3s, Helios Operator is also fully compatible with cloud environments like **Google Kubernetes Engine (GKE)**, **Amazon Elastic Kubernetes Service (EKS)**, **Azure Kubernetes Service (AKS)**, and other managed Kubernetes services.

## 📋 **Prerequisites**

### **Step 1: Install Kubernetes Cluster**

You need a Kubernetes cluster to run Helios Operator. Choose one of the following options:

#### **Option A: Minikube (Recommended for Beginners)**

```bash
# Install Minikube
# On macOS:
brew install minikube

# On Linux:
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# On Windows (with Chocolatey):
choco install minikube

# Start cluster
minikube start --memory=4096 --cpus=2

# Check cluster status
minikube status
```

#### **Option B: Kind (Kubernetes in Docker)**

```bash
# Install Kind
# On macOS:
brew install kind

# On Linux:
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.30.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Create cluster
kind create cluster --name helios-cluster --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        system-reserved: memory=1Gi
  extraPortMappings:
  - containerPort: 30000
    hostPort: 30000
    protocol: TCP
- role: worker
- role: worker
EOF
```

#### **Option C: K3s (Lightweight Kubernetes)**

```bash
# Install K3s
curl -sfL https://get.k3s.io | sh -

# Configure kubectl
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
```

### **Step 2: Install kubectl**

```bash
# On macOS:
brew install kubectl

# On Linux:
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# On Windows (with Chocolatey):
choco install kubernetes-cli

# Verify installation
kubectl version --client
```

### **Step 3: Install Helm**

```bash
# On macOS:
brew install helm

# On Linux:
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# On Windows (with Chocolatey):
choco install kubernetes-helm

# Verify installation
helm version
```

### **Step 4: Prepare Repositories**

You need to prepare 2 GitHub repositories:

1. **Source Code Repository** (contains your application code)
2. **GitOps Repository** (empty, Helios will automatically create manifests)

### **Step 5: Container Registry**

Create an account at one of these container registries:

- Docker Hub
- GitHub Container Registry
- Google Container Registry
- Amazon ECR

## ☁️ **Cloud Deployment Options**

While this guide focuses on local development, Helios Operator works seamlessly on cloud platforms:

### **Google Kubernetes Engine (GKE)**

```bash
# Create GKE cluster
gcloud container clusters create helios-cluster \
  --zone=us-central1-a \
  --num-nodes=3 \
  --machine-type=e2-medium

# Get credentials
gcloud container clusters get-credentials helios-cluster --zone=us-central1-a
```

### **Amazon Elastic Kubernetes Service (EKS)**

```bash
# Create EKS cluster
eksctl create cluster \
  --name helios-cluster \
  --region us-west-2 \
  --nodegroup-name workers \
  --node-type t3.medium \
  --nodes 3
```

### **Azure Kubernetes Service (AKS)**

```bash
# Create AKS cluster
az aks create \
  --resource-group myResourceGroup \
  --name helios-cluster \
  --node-count 3 \
  --node-vm-size Standard_B2s

# Get credentials
az aks get-credentials --resource-group myResourceGroup --name helios-cluster
```

**Note:** For cloud deployments, ensure your cluster has:

- Minimum 2 CPU cores and 4GB RAM per node
- LoadBalancer service support (for webhooks)
- Container registry access permissions
- Sufficient storage for PVCs

## 🏗️ **Step 2: Install Dependencies**

### **Install Tekton Pipelines**

Tekton will handle building and pushing Docker images:

```bash
# Install Tekton Pipelines
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

# Install Tekton Triggers (to handle webhooks)
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml

# Wait for Tekton to be ready
kubectl wait --for=condition=ready pod -l app=tekton-pipelines-controller -n tekton-pipelines --timeout=300s
```

### **Install ArgoCD**

ArgoCD will automatically deploy applications from the GitOps repository:

```bash
# Add ArgoCD Helm repository
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

# Create namespace for ArgoCD
kubectl create namespace argocd

# Install ArgoCD
helm install argocd argo/argo-cd --namespace argocd

# Wait for ArgoCD to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s
```

### **Verify All Dependencies**

```bash
# Check Tekton
kubectl get pods -n tekton-pipelines

# Check ArgoCD
kubectl get pods -n argocd

# Check overall cluster
kubectl get nodes
```

## 🎯 **Step 3: Install Helios Operator**

### **Install via Helm**

```bash
# Add Helios Operator Helm repository
helm repo add helios-operator https://hoangphuc841.github.io/helios-operator
helm repo update

# Install Helios Operator
helm install helios-operator helios-operator/helios-operator

# Check operator is running
kubectl get pods -l app.kubernetes.io/name=helios-operator

# Verify CRD is installed
kubectl get crd heliosapps.platform.helios.io
```

### **Verify Installation**

```bash
# Check operator pod
kubectl get pods -l app.kubernetes.io/name=helios-operator

# Check operator logs
kubectl logs -l app.kubernetes.io/name=helios-operator

# Check CRD
kubectl get crd | grep helios
```

## 📝 **Step 4: Prepare Your Application**

### **Create Application Namespace**

```bash
# Create namespace for your application
kubectl create namespace my-apps

# Verify namespace was created
kubectl get namespace my-apps
```

### **Create PVC for Build Workspace**

Tekton needs storage space to build your application:

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
      storage: 2Gi
EOF
```

### **Create Service Account and Secrets**

```bash
# Create Service Account for pipeline
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pipeline-sa
  namespace: my-apps
EOF

# Create secret for GitHub (to clone code)
kubectl create secret generic github-secret \
  --from-literal=username=your-github-username \
  --from-literal=password=your-github-token \
  -n my-apps

# Create secret for Docker registry (to push image)
kubectl create secret docker-registry docker-secret \
  --docker-server=docker.io \
  --docker-username=your-docker-username \
  --docker-password=your-docker-password \
  --docker-email=your-email@example.com \
  -n my-apps

# Attach secrets to service account
kubectl patch serviceaccount pipeline-sa -n my-apps -p '{"imagePullSecrets": [{"name": "docker-secret"}]}'
```

## 🚀 **Step 5: Deploy Your First Application**

### **Create HeliosApp Resource**

This is the most important step - creating the HeliosApp resource to activate the entire CI/CD flow:

```bash
kubectl apply -f - <<EOF
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-first-app
  namespace: my-apps
spec:
  # Source code repository (contains your application code)
  gitRepo: "https://github.com/your-username/your-app.git"
  gitBranch: "main"

  # GitOps repository (empty, Helios will automatically create manifests)
  gitopsRepo: "https://github.com/your-username/your-gitops.git"
  gitopsPath: "my-first-app"  # Directory in GitOps repo
  gitopsBranch: "main"

  # Container image configuration
  imageRepo: "docker.io/your-username/my-first-app"
  port: 80
  replicas: 1

  # Pipeline configuration
  serviceAccount: "pipeline-sa"
  webhookSecret: "github-webhook-secret"
  pvcName: "my-app-pvc"

  # Build configuration
  buildContext: "."
  dockerfilePath: "Dockerfile"

  # GitOps configuration
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
EOF
```

### **Parameter Explanation**

- **`gitRepo`**: URL of the repository containing your source code
- **`gitopsRepo`**: URL of the GitOps repository (empty, Helios will automatically create manifests)
- **`imageRepo`**: Name of the Docker image that will be built and pushed
- **`webhookSecret`**: Secret to authenticate webhooks from GitHub
- **`syncPolicy`**: ArgoCD configuration for automatic sync and healing

### **Watch the Magic Happen ✨**

After creating the HeliosApp, the operator will automatically create all necessary resources:

```bash
# Watch HeliosApp status
kubectl get heliosapp my-first-app -n my-apps -w

# Check Tekton resources created
kubectl get pipelines,eventlisteners,triggerbindings,triggertemplates -n my-apps

# Check ArgoCD Application created
kubectl get applications -n argocd

# View detailed status
kubectl describe heliosapp my-first-app -n my-apps
```

### **Configure GitHub Webhook**

To activate the automatic CI/CD flow when pushing code:

#### **For Local Deployment (Minikube/Kind/K3s):**

```bash
# Get cluster IP (for Minikube)
minikube ip

# Or get EventListener service IP
kubectl get svc -n my-apps | grep el-

# Port forward to test webhook (if needed)
kubectl port-forward svc/el-my-first-app-el -n my-apps 8080:8080
```

#### **For Cloud Deployment (GKE/EKS/AKS):**

```bash
# Get LoadBalancer external IP
kubectl get svc -n my-apps el-my-first-app-el

# Or expose via Ingress (recommended for production)
kubectl expose service el-my-first-app-el -n my-apps --type=LoadBalancer --name=webhook-external
```

**Configure on GitHub:**

1. Go to your source code repository
2. Settings → Webhooks → Add webhook
3. **Payload URL**:
   - Local: `http://your-cluster-ip:8080`
   - Cloud: `http://your-loadbalancer-ip:8080` or `https://your-domain.com/webhook`
4. **Content type**: `application/json`
5. **Secret**: `github-webhook-secret`
6. **Events**: Select "Just the push event"
7. **Active**: ✅

## 🎉 **Step 6: Verify Your Deployment**

### **Check Application Status**

```bash
# Get detailed status
kubectl describe heliosapp my-first-app -n my-apps

# Check ArgoCD Application
kubectl describe application my-first-app-argocd -n argocd
```

### **Access ArgoCD UI (Optional)**

```bash
# Port forward to ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Get admin password
kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" | base64 -d && echo

# Access: https://localhost:8080
# Username: admin
# Password: (from above command)
```

## 🔄 **Step 7: Test the Complete Flow**

### **Push Code to Trigger Build**

```bash
# Make changes to your source code
echo "print('Hello from updated app!')" > src/main.py

# Commit and push
git add .
git commit -m "Update application"
git push origin main
```

### **Monitor the Complete Workflow**

```bash
# 1. Check webhook was triggered
kubectl logs -n my-apps -l app=el-my-first-app-el

# 2. Watch PipelineRun
kubectl get pipelinerun -n my-apps -w

# 3. View build logs
kubectl logs -n my-apps -l tekton.dev/pipelineRun=my-first-app-pipelinerun-xxx

# 4. Check GitOps repo was updated
git clone https://github.com/your-username/your-gitops-repo.git
cd your-gitops-repo
ls -la my-first-app/  # You'll see deployment.yaml, service.yaml

# 5. Watch ArgoCD sync
kubectl get application my-first-app-argocd -n argocd -w

# 6. Check deployment
kubectl get pods -n my-apps -l app=my-first-app
```

## 🎯 **What Happened?**

When you created the `HeliosApp`, the operator automatically:

1. ✅ **Generated Tekton Pipeline** (`my-first-app-pipeline`) from template
2. ✅ **Created Tekton Triggers** (EventListener, TriggerBinding, TriggerTemplate)
3. ✅ **Created ArgoCD Application** (`my-first-app-argocd`) for GitOps deployment
4. ✅ **Set up Webhook Integration** for automatic builds on Git pushes
5. ✅ **Configured Real-time Monitoring** with Kubernetes Watches

When you pushed code:

1. ✅ **GitHub Webhook** triggered Tekton EventListener
2. ✅ **Tekton Pipeline** built Docker image and pushed to registry
3. ✅ **Tekton Pipeline** automatically created/updated manifests in GitOps repo
4. ✅ **ArgoCD** detected changes and synced to Kubernetes
5. ✅ **Application** was deployed with the new image

## 🔧 **Next Steps**

### **Configure Webhooks (Optional)**

To enable automatic builds on Git pushes:

1. Go to your GitHub repository settings
2. Navigate to "Webhooks"
3. Add webhook URL: `http://your-cluster-ip:8080/`
4. Set secret to match `webhookSecret` in your HeliosApp

### **Set up GitOps Repository**

The GitOps repository will be automatically populated with deployment manifests:

```yaml
# your-gitops-repo/my-first-app/deployment.yaml (auto-generated)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-first-app
  labels:
    app: my-first-app
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
# your-gitops-repo/my-first-app/service.yaml (auto-generated)
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

## 🆘 **Troubleshooting**

### **Common Issues**

```bash
# Check webhook logs
kubectl logs -n my-apps -l app=el-my-first-app-el

# Check pipeline status
kubectl describe pipelinerun -n my-apps

# Check ArgoCD status
kubectl describe application my-first-app-argocd -n argocd

# Check deployment status
kubectl get pods -n my-apps
kubectl describe pod -n my-apps -l app=my-first-app
```

### **Need Help?**

- **Issues**: Check our [Troubleshooting Guide](03-troubleshooting.md)
- **Configuration**: See [HeliosApp Specification](02-helios-app-spec.md)
- **Support**: Open an [issue on GitHub](https://github.com/hoangphuc841/helios-operator/issues)

---

**🎉 Congratulations!** You've successfully deployed your first application with Helios Operator. The system will now automatically build and deploy your application whenever you push code to your source repository. Ready to deploy more? Check out the [HeliosApp Specification](02-helios-app-spec.md) for advanced configuration options.
