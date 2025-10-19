# Getting Started with Helios Operator

Complete end-to-end guide to deploy your application using Helios Operator with full GitOps workflow.

## 🎯 What You'll Achieve

By the end of this guide, you will:

1. ✅ Have a working Kubernetes cluster with Helios Operator
2. ✅ Deploy your application by pushing code to GitHub
3. ✅ See automatic manifest generation via Tekton Pipeline
4. ✅ Watch ArgoCD automatically deploy your application
5. ✅ Understand the complete GitOps workflow

## 📋 Prerequisites

### Required Software

- **Kubernetes cluster** (v1.34+) - Minikube, Kind, K3s, or cloud cluster
- **kubectl** (v1.34+)
- **Git** and GitHub account
- **Docker Hub account** (or any container registry)

### Required GitHub Repositories

You'll need **3 repositories** (can be public or private):

1. **Application Source Code Repo** - Your actual application code

   - Example: `https://github.com/YOUR_USERNAME/my-nodejs-app`
   - Must contain: `Dockerfile`

2. **GitOps Manifests Repo** - Where rendered Kubernetes manifests are stored

   - Example: `https://github.com/YOUR_USERNAME/gitops-manifests`
   - Can be empty initially (will be populated automatically)

3. **App Templates Repo** (Optional) - Kubernetes manifest templates
   - Example: `https://github.com/YOUR_USERNAME/k8s-templates`
   - Or use existing template repos

### Development (Optional)

- Go 1.25+
- Make

### ⚠️ Important Setup Notes

Before starting, be aware of these common issues:

1. **Single-Node Clusters (Minikube/Kind):** You MUST disable Tekton's affinity assistant (covered in Step 2.1)
2. **Docker Hub Authentication:** Login to Docker Hub BEFORE creating the `docker-credentials` secret (Step 5.1)
3. **Repository Structure:** If your Dockerfile is in a subdirectory, set `contextSubpath` in your HeliosApp spec (Step 7.1)
4. **Initial Manifest Required:** You must create an initial deployment manifest in your GitOps repo (Step 0.2) - the pipeline updates it but doesn't create it from scratch

---

## 🚀 Step-by-Step Setup

### Step 0: Prepare Your GitHub Repositories

Before starting, create these repositories on GitHub:

#### 0.1. Create Application Repo

```bash
# Create a simple Node.js app example
mkdir my-nodejs-app && cd my-nodejs-app
git init

# Create app code
cat > index.js << 'EOF'
const express = require('express');
const app = express();
const PORT = process.env.PORT || 8080;

app.get('/', (req, res) => {
  res.json({
    message: 'Hello from Helios!',
    version: process.env.VERSION || 'v1.0.0',
    timestamp: new Date().toISOString()
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
EOF

# Create package.json
cat > package.json << 'EOF'
{
  "name": "my-nodejs-app",
  "version": "1.0.0",
  "main": "index.js",
  "dependencies": {
    "express": "^4.18.2"
  },
  "scripts": {
    "start": "node index.js"
  }
}
EOF

# Create Dockerfile
cat > Dockerfile << 'EOF'
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install --production
COPY . .
EXPOSE 8080
CMD ["npm", "start"]
EOF

# Push to GitHub
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/my-nodejs-app.git
git push -u origin main
```

#### 0.2. Create GitOps Manifests Repo

````bash
# Create empty GitOps repo
mkdir gitops-manifests && cd gitops-manifests
git init

# Create base structure
mkdir -p apps/dev apps/staging apps/production

# Create main README
cat > README.md << 'EOF'
# GitOps Manifests Repository

This repository stores the rendered Kubernetes manifests for all environments.

## Structure
- `apps/dev/` - Development environment
- `apps/staging/` - Staging environment
- `apps/production/` - Production environment
EOF

**⚠️ IMPORTANT: Create Initial Manifest**

The Tekton `git-update-manifest` task uses `sed` to update the `image:` field in existing manifests—it **does NOT create files from scratch**. You must create an initial manifest for each application.

**Create the initial manifest for your application:**

```bash
# Still in gitops-manifests directory
# Create directory matching your source repo name
# If your source repo is "my-nodejs-app", use apps/dev/my-nodejs-app
# If your source repo is "helios", use apps/dev/helios
mkdir -p apps/dev/helios

# Create initial manifest
# ⚠️ IMPORTANT: Replace 'hophuochoan' with YOUR Docker Hub username!
cat > apps/dev/helios/deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helios
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: helios
  template:
    metadata:
      labels:
        app: helios
    spec:
      containers:
      - name: app
        image: docker.io/YOUR_DOCKERHUB_USERNAME/helios:latest
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: helios
  namespace: default
spec:
  selector:
    app: helios
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
EOF

# Commit and push
git add apps/dev/helios/deployment.yaml
git commit -m "feat: Add initial manifest for helios app"
git push origin main
````

**📝 Important Notes:**

1. **⚠️ UPDATE THE IMAGE FIELD!** - The `image:` field in the manifest **MUST** match your Docker Hub repository:

   ```yaml
   image: docker.io/YOUR_DOCKERHUB_USERNAME/helios:latest
   ```

   - Replace `YOUR_DOCKERHUB_USERNAME` with your actual Docker Hub username
   - The image name (`helios`) should match your GitHub repository name
   - This MUST match the `imageRepo` field you'll use in your HeliosApp (Step 7.1)
   - **Why?** After the first pipeline run, this gets auto-updated with the actual image digest. If the username is wrong, deployment will fail!

2. **Match your repository name** - The directory name under `apps/dev/` MUST match your GitHub source repository name. The pipeline uses `$(body.repository.name)` from the webhook payload.

   - GitHub repo: `helios` → Directory: `apps/dev/helios/`
   - GitHub repo: `my-nodejs-app` → Directory: `apps/dev/my-nodejs-app/`

3. **Update other placeholders** - Also replace if needed:

   - App name (if different from repo name)
   - Port numbers (if your app uses different ports than 8080)
   - Namespace (if not using `default`)

4. **One-time setup per app** - You only need to do this once per application. After that, the pipeline automatically updates the `image:` field with new image digests on every build.

5. **Multiple apps** - Repeat this process for each application you want to deploy, creating separate directories like `apps/dev/app1/`, `apps/dev/app2/`, etc.

**📋 Example Configuration:**

If your setup is:

- **Docker Hub username**: `johndoe`
- **GitHub repo name**: `helios`
- **App port**: `8080`

Then your manifest should look like:

```yaml
# apps/dev/helios/deployment.yaml
spec:
  containers:
    - name: app
      image: docker.io/johndoe/helios:latest # ← Must match YOUR username!
      ports:
        - containerPort: 8080
```

And later in Step 7.1, your HeliosApp should have:

```yaml
# my-app.yaml
spec:
  imageRepo: "docker.io/johndoe/helios" # ← Same username and image name!
  gitopsPath: "apps/dev/helios" # ← Matches GitHub repo name!
```

**These MUST be consistent across all three places:**

1. ✅ Initial GitOps manifest image: `docker.io/johndoe/helios:latest`
2. ✅ HeliosApp `imageRepo`: `docker.io/johndoe/helios`
3. ✅ HeliosApp `gitopsPath`: `apps/dev/helios` (matches repo name)

#### 0.3. Create App Templates Repo (Optional)

```bash
mkdir k8s-templates && cd k8s-templates
git init
mkdir -p templates/simple-app

# Create basic Kubernetes template
cat > templates/simple-app/deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .AppName }}
  namespace: {{ .Namespace }}
spec:
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
      app: {{ .AppName }}
  template:
    metadata:
      labels:
        app: {{ .AppName }}
    spec:
      containers:
      - name: app
        image: {{ .ImageRepo }}:{{ .ImageTag }}
        ports:
        - containerPort: {{ .Port }}
        env:
        - name: VERSION
          value: "{{ .ImageTag }}"
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .AppName }}
  namespace: {{ .Namespace }}
spec:
  selector:
    app: {{ .AppName }}
  ports:
  - port: 80
    targetPort: {{ .Port }}
  type: ClusterIP
EOF

git add .
git commit -m "Add simple app template"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/k8s-templates.git
git push -u origin main
```

---

### Step 1: Install Kubernetes Cluster (if needed)

If you don't have a cluster, install Minikube:

```bash
# Install Minikube (Linux)
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Start cluster with enough resources
minikube start

# Verify
kubectl get nodes
```

Or use Kind:

```bash
# Install Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Create cluster
kind create cluster --name helios-cluster

# Verify
kubectl cluster-info
```

---

### Step 2: Install Dependencies

#### 2.1. Install Tekton Pipelines

```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
```

> **⚠️ Important for Kind/Minikube users (single-node clusters):**
>
> Tekton Pipelines enables the "affinity assistant" by default, which can cause all PipelineRun pods to be stuck in `Pending` with errors like:
>
> `0/1 nodes are available: 1 node(s) didn't match pod affinity rules.`
>
> This happens because the affinity assistant tries to schedule pods on different nodes, but your cluster only has one node.
>
> **If you are using Kind, Minikube, or any single-node cluster, you MUST follow the troubleshooting steps in [Fix Tekton Affinity Assistant](#-required-fix-tekton-affinity-assistant-on-single-node-clusters) below after installing Tekton.**

Wait for pods to be ready:

```bash
kubectl wait --for=condition=ready pod -l app=tekton-pipelines-controller -n tekton-pipelines --timeout=90s
kubectl get pods -n tekton-pipelines
```

---

### 🔧 REQUIRED: Fix Tekton Affinity Assistant on Single-Node Clusters

> **⚠️ CRITICAL FOR KIND/MINIKUBE USERS**
>
> If you're using Kind, Minikube, or any single-node cluster, **you MUST apply this fix immediately after installing Tekton** or all PipelineRuns will fail with:
>
> ```text
> 0/1 nodes are available: 1 node(s) didn't match pod affinity rules.
> ```

#### Why This Happens

Tekton's "affinity assistant" is designed for **multi-node clusters** to optimize PVC access. On single-node clusters:

- It creates pod affinity rules that require multiple nodes
- Your cluster only has 1 node
- Pods get stuck in `Pending` forever

#### The Complete Fix (Copy-Paste This)

Run this **one-time setup script** right after installing Tekton:

```bash
# Step 1: Disable affinity assistant AND coschedule
kubectl patch configmap feature-flags -n tekton-pipelines --type merge \
  -p '{"data":{"disable-affinity-assistant":"true","coschedule":"disabled"}}'

# Step 2: Restart Tekton controller to apply the change
kubectl delete pod -n tekton-pipelines -l app=tekton-pipelines-controller

# Step 3: Wait for controller to restart
kubectl wait --for=condition=ready pod -l app=tekton-pipelines-controller -n tekton-pipelines --timeout=60s

# Step 4: Clean up any stuck PipelineRuns (if you had any)
kubectl delete pipelinerun --all 2>/dev/null || true

# Step 5: Verify the fix
echo "Checking if affinity assistant is disabled..."
kubectl get configmap feature-flags -n tekton-pipelines -o jsonpath='{.data.disable-affinity-assistant}'
echo ""
kubectl get configmap feature-flags -n tekton-pipelines -o jsonpath='{.data.coschedule}'
echo ""
echo "✅ If you see 'true' and 'disabled' above, the fix is applied!"
```

**That's it!** You don't need to test with a PipelineRun yet. The fix is permanent and will work when you create your first PipelineRun later in Step 7.

#### Why This Is Safe

- **Single-node clusters**: All pods run on the same node anyway, so affinity rules are meaningless
- **No performance impact**: The affinity assistant is only useful on multi-node clusters
- **This is the official Tekton recommendation** for single-node development environments

#### Alternative: Multi-Node Cluster (Not Recommended for Local Dev)

If you want to keep the affinity assistant enabled, you need a **multi-node cluster**:

```bash
# Kind example with 3 nodes
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
EOF
```

**However, for local development, disabling the affinity assistant is simpler and has zero downsides.**

---#### 2.2. Install Tekton Triggers

```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml
```

Verify:

```bash
kubectl get pods -n tekton-pipelines | grep trigger
```

#### 2.3. Install ArgoCD

```bash
# Create namespace
kubectl create namespace argocd

# Install ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for pods
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=180s
```

#### 2.3.1. Configure ArgoCD for Faster Syncing (Recommended)

By default, ArgoCD checks for changes in Git repositories every **3 minutes**. For development environments, you can reduce this to get faster deployments:

```bash
# Set ArgoCD to check for changes every 30 seconds (instead of 3 minutes)
kubectl patch configmap argocd-cm -n argocd --type merge \
  -p '{"data":{"timeout.reconciliation":"30s"}}'

# Restart ArgoCD components to apply the change
kubectl rollout restart deployment argocd-repo-server -n argocd
kubectl rollout restart statefulset argocd-application-controller -n argocd

# Wait for ArgoCD to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-application-controller -n argocd --timeout=60s

# Verify the setting
kubectl get configmap argocd-cm -n argocd -o jsonpath='{.data.timeout\.reconciliation}'
# Should output: 30s
```

**Note:** For production environments, the default 3-minute interval is recommended to reduce load on Git servers and ArgoCD.

Verify and access ArgoCD UI:

```bash
# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
echo ""

# Port forward to access UI (use 8081 locally to avoid conflicts with Tekton EventListener)
kubectl port-forward svc/argocd-server -n argocd 8081:443

# Access at: https://localhost:8081
# Username: admin
# Password: (from command above)
```

---

### Step 3: Install Helios Operator

#### 3.1. Clone the Helios Operator Repository

```bash
git clone https://github.com/YOUR_ORG/helios-operator
cd helios-operator
```

#### 3.2. Install CRD and Deploy Operator

First, install the CRD (required for any cluster):

```bash
make install
```

Then choose ONE path based on your cluster:

- If you started a Minikube cluster in Step 1

  ```bash
  # Build the controller image locally
  make docker-build IMG=helios-operator:latest

  # Load the image into the Minikube cluster
  minikube image load helios-operator:latest

  # Deploy using the same image reference
  make deploy IMG=helios-operator:latest
  ```

- If you started a Kind cluster in Step 1 (named "helios-cluster")

  ```bash
  # Build the controller image locally
  make docker-build IMG=helios-operator:latest

  # Load the image into the Kind cluster
  kind load docker-image helios-operator:latest --name helios-cluster

  # Deploy using the same image reference
  make deploy IMG=helios-operator:latest
  ```

Notes:

- Ensure your cluster can pull from the registry (public image or set imagePullSecrets on the manager SA/deployment as needed).
- Keep the IMG value consistent between build/push and deploy.

Verify operator is running:

```bash
kubectl get pods -n helios-operator-system
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager -f
```

Check CRD is installed:

```bash
kubectl get crd heliosapps.platform.helios.io
```

---

### Step 4: Setup Tekton Pipeline Resources

#### 4.1. Create Persistent Volume for Workspace

```bash
kubectl apply -f tekton/workspace-pvc.yaml
```

Verify:

```bash
kubectl get pvc shared-workspace-pvc
```

#### 4.2. Install Tekton Tasks

```bash
# Install all tasks
kubectl apply -f tekton/task-git-clone.yaml
kubectl apply -f tekton/task-kaniko-build.yaml
kubectl apply -f tekton/task-git-update-manifest.yaml

# Verify
kubectl get task
```

#### 4.3. Install Pipeline

```bash
kubectl apply -f tekton/pipeline.yaml

# Verify
kubectl get pipeline from-code-to-cluster
```

#### 4.4. Configure Default Values for Webhooks

**⚠️ IMPORTANT:** Before installing trigger resources, you must configure the default values with your actual repository information.

Edit `tekton/triggerbinding-defaults.yaml` and replace the placeholder values:

```bash
# Edit the file
vi tekton/triggerbinding-defaults.yaml

# OR use sed to replace values
sed -i 's|docker.io/yourname|docker.io/YOUR_DOCKERHUB_USERNAME|g' tekton/triggerbinding-defaults.yaml
sed -i 's|https://github.com/YOUR_ORG/your-gitops-repo|https://github.com/YOUR_USERNAME/gitops-manifests|g' tekton/triggerbinding-defaults.yaml
```

**Example of correct values:**

```yaml
spec:
  params:
    - name: image-registry
      value: docker.io/phuochoan # <-- YOUR Docker Hub username
    - name: gitops-repo-url
      value: https://github.com/PhuocHoan/gitops-manifests # <-- YOUR GitOps repo
    - name: gitops-repo-branch
      value: main
    - name: manifest-path-in-gitops-repo
      value: apps/dev/$(body.repository.name)/deployment.yaml
    - name: pvc-name
      value: shared-workspace-pvc
    - name: context-subpath
      value: "" # <-- Empty if Dockerfile is at repo root, or "subdirectory-name" if in a subfolder
```

**📝 About `context-subpath`:**

- If your Dockerfile is **at the root** of your repository, leave this **empty** (`value: ""`)
- If your Dockerfile is in a **subdirectory** (e.g., `my-app/Dockerfile`), set it to that directory name (`value: "my-app"`)
- The pipeline will automatically look for the Dockerfile in the correct location

**This configuration is applied to ALL applications** that trigger via webhook. This is a one-time setup per cluster/environment.

#### 4.5. Install Tekton Trigger Resources (for webhooks)

Create the EventListener, TriggerBindings (GitHub payload + environment defaults), TriggerTemplate and its RBAC now (they live in the same `tekton/` folder):

```bash
kubectl apply -f tekton/tekton-triggers-sa.yaml
kubectl apply -f tekton/triggerbinding-github.yaml
kubectl apply -f tekton/triggerbinding-defaults.yaml   # Your customized defaults
kubectl apply -f tekton/triggertemplate-helios.yaml
kubectl apply -f tekton/eventlistener-helios.yaml
```

---

### Step 5: Setup Credentials and Service Account

#### 5.1. Create Container Registry Credentials (Docker Hub or GHCR)

**⚠️ CRITICAL:** The docker-credentials secret must contain valid Docker Hub authentication. If you only have Google Cloud credentials in `~/.docker/config.json`, the pipeline will fail at the image push step.

**Step 1: Login to Docker Hub first**

```bash
# Login to Docker Hub (this updates ~/.docker/config.json)
docker login -u YOUR_DOCKERHUB_USERNAME
# Enter your Docker Hub Access Token when prompted (NOT your password!)
```

**To get Docker Hub Access Token:**

1. Go to <https://hub.docker.com/settings/security>
2. Click "New Access Token"
3. Name: "tekton-pipeline"
4. Permissions: Read & Write
5. Generate and COPY the token
6. Paste it when `docker login` prompts for password

**Step 2: Verify your Docker config**

```bash
# Check that Docker Hub auth is present
cat ~/.docker/config.json

# You should see something like:
# {
#   "auths": {
#     "https://index.docker.io/v1/": {
#       "auth": "base64encodedtoken"
#     }
#   }
# }
#
# If you only see "credHelpers" for gcloud, you need to login to Docker Hub first!
```

**Step 3: Create the secret**

```bash
# Create docker credentials secret from your config
kubectl create secret generic docker-credentials \
  --from-file=.dockerconfigjson=$HOME/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson
```

**Step 4: Verify the secret**

```bash
# Verify the secret contains Docker Hub credentials
kubectl get secret docker-credentials -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | jq '.'

# You should see "auths" with "https://index.docker.io/v1/" entry
```

**Common Issue:** If you see `UNAUTHORIZED: authentication required` during pipeline runs, it means your secret doesn't have valid Docker Hub credentials. Delete and recreate it:

```bash
kubectl delete secret docker-credentials
# Login to Docker Hub again
docker login -u YOUR_DOCKERHUB_USERNAME
# Recreate the secret
kubectl create secret generic docker-credentials \
  --from-file=.dockerconfigjson=$HOME/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson
```

#### 5.2. Create GitHub Credentials (HTTPS with Personal Access Token)

Use a GitHub Personal Access Token (PAT) so Tekton can clone your repos and push updates to the GitOps repo. The pipeline reads it from the `git-credentials` Secret—no tokens are embedded in YAML.

```bash
# Create GitHub Personal Access Token
# Classic token: GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
#   - Scopes: repo (for private repos) — includes read and write
# Fine-grained token: Settings > Developer settings > Fine-grained tokens
#   - Repository access: your app repo (read) and GitOps repo (contents: read/write)

kubectl create secret generic git-credentials \
  --from-literal=username=YOUR_GITHUB_USERNAME \
  --from-literal=password=YOUR_GITHUB_TOKEN \
  --type=kubernetes.io/basic-auth

# Tell Tekton which host this secret should be used for (HTTPS)
kubectl annotate secret git-credentials tekton.dev/git-0=https://github.com
```

What this does:

- Creates a Kubernetes Secret named `git-credentials` with basic-auth fields that Tekton uses for git operations.
- The annotation `tekton.dev/git-0=https://github.com` maps this secret to requests made to GitHub over HTTPS so tasks like clone and push automatically use it.
- Make sure the PAT has permission to read your source repo and write to your GitOps repo (for committing rendered manifests).

#### 5.3. Create GitHub Webhook Secret

This secret will be used by Tekton Triggers to validate webhook requests from GitHub:

```bash
# Create a random webhook secret
kubectl create secret generic github-webhook-secret \
  --from-literal=secretToken=$(openssl rand -base64 32)

# Or use your own secret value
kubectl create secret generic github-webhook-secret \
  --from-literal=secretToken="YOUR_WEBHOOK_SECRET"
```

Verify:

```bash
kubectl get secret github-webhook-secret
```

**Note:** If you plan to set up GitHub webhooks later, save this secret value. You'll need it when configuring the webhook in GitHub repository settings.

To retrieve the secret value later:

```bash
kubectl get secret github-webhook-secret -o jsonpath='{.data.secretToken}' | base64 -d
echo ""
```

#### 5.4. Create Service Account

```bash
kubectl apply -f tekton/service-account.yaml
```

Verify:

```bash
kubectl get sa pipeline-sa
kubectl describe sa pipeline-sa
```

---

### Step 6: Setup GitHub Webhooks (Quick)

Set this up now so pushes to your repo automatically trigger the pipeline. You can still deploy manually without it.

#### 6.1. Verify EventListener is Running

If you completed Step 4.4, the EventListener should already exist. Verify it:

```bash
kubectl wait --for=condition=ready pod -l eventlistener=helios-listener --timeout=60s
kubectl get pods -l eventlistener=helios-listener
```

#### 6.2. Expose the Tekton EventListener

Choose one method based on your environment:

##### Option A: Development (ngrok)

```bash
# Start port-forward for Tekton EventListener (use 8080 for ngrok)
kubectl port-forward -n default svc/el-helios-listener 8080:8080

# In another terminal, start ngrok (must match port above)
ngrok http 8080
```

Copy the HTTPS URL from ngrok, e.g. `https://abc123.ngrok-free.app`

##### Option B: Production (Ingress)

Point a domain to your cluster and expose `el-helios-listener` at `https://webhook.yourdomain.com`.

#### 6.3. Get the Webhook Secret

Retrieve the webhook secret you created in Step 5.3:

```bash
kubectl get secret github-webhook-secret -o jsonpath='{.data.secretToken}' | base64 -d
echo ""
```

Save this value - you'll need it in the next step.

#### 6.4. Configure GitHub Webhook

Add a webhook in your GitHub application repository:

1. Go to your repository on GitHub
2. Click **Settings** > **Webhooks** > **Add webhook**
3. Configure the webhook:
   - **Payload URL**: `https://YOUR_NGROK_URL/` (note: root path `/`, **NOT** `/hooks`)
   - **Content type**: `application/json`
   - **Secret**: paste the value from step 6.3
   - **Which events**: Select "Just the push event"
   - **Active**: ✓ checked
4. Click **Add webhook**

**Important:** The EventListener listens on the root path `/`, not `/hooks`. Using `/hooks` will result in a 307 redirect error.

#### 6.5. Verify Webhook Setup

Test that the webhook was configured correctly:

```bash
# Check EventListener pod is running
kubectl get pods -l eventlistener=helios-listener

# Check EventListener logs
kubectl logs -l eventlistener=helios-listener --tail=50

# Verify the service
kubectl get svc el-helios-listener
```

On GitHub, check the webhook delivery:

1. Go to **Settings** > **Webhooks**
2. Click on your webhook
3. Check **Recent Deliveries**
4. You should see a "ping" event with a ✅ green checkmark (200 response)

#### 6.6. Test with a Push Event

Push a trivial change to your repo and watch for a new PipelineRun:

```bash
# Watch for new PipelineRuns
kubectl get pipelinerun -w

# Or filter by specific app
kubectl get pipelinerun -l app=YOUR_REPO_NAME -w
```

If the webhook is working correctly, you should see a new PipelineRun created automatically after pushing to GitHub.

#### 6.7. Troubleshooting Webhooks

##### Problem: 307 Temporary Redirect

- **Cause**: Incorrect webhook URL path
- **Solution**: Use `https://YOUR_NGROK_URL/` (root path), not `/hooks`

##### Problem: 401 Unauthorized or 403 Forbidden

- **Cause**: Webhook secret mismatch
- **Solution**: Verify the secret matches:

```bash
# Get the secret from Kubernetes
kubectl get secret github-webhook-secret -o jsonpath='{.data.secretToken}' | base64 -d

# Update in GitHub webhook settings if needed
```

##### Problem: Webhook succeeds but no PipelineRun created

- **Cause**: TriggerBinding/TriggerTemplate issues or RBAC permissions
- **Solution**: Check EventListener logs and verify resources exist:

```bash
kubectl logs -l eventlistener=helios-listener --tail=100
kubectl get triggerbinding,triggertemplate,eventlistener
kubectl describe eventlistener helios-listener
```

##### Problem: EventListener pod crashes (permission errors)

- **Cause**: Missing RBAC permissions for Tekton Triggers resources
- **Solution**: Reapply the service account with updated permissions:

```bash
kubectl apply -f tekton/tekton-triggers-sa.yaml
kubectl delete pod -l eventlistener=helios-listener  # Restart the pod
```

Need more detailed troubleshooting? See WEBHOOK_SETUP_GUIDE.md.

---

### Step 7: Deploy Your First Application with Helios

Now the exciting part - deploy your application!

#### 7.1. Prepare Your HeliosApp Configuration

The HeliosApp Custom Resource is the heart of the Helios Operator. When you apply it, the operator automatically:

1. Creates Tekton Triggers for webhook-based automation
2. Triggers a PipelineRun to build your image and update GitOps repo
3. Creates an ArgoCD Application to deploy your app
4. Monitors the entire workflow and updates status

Create `my-app.yaml` with your actual repository URLs:

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  # ===== Your Application Source Code =====
  gitRepo: "https://github.com/YOUR_USERNAME/my-nodejs-app"
  gitBranch: "main"

  # ===== Container Image =====
  imageRepo: "docker.io/YOUR_DOCKERHUB_USERNAME/my-nodejs-app"

  # ===== Application Settings =====
  port: 8080
  replicas: 2

  # ===== CI/CD Pipeline =====
  pipelineName: "from-code-to-cluster"
  serviceAccount: "pipeline-sa"
  webhookSecret: "github-webhook-secret"
  pvcName: "shared-workspace-pvc"

  # ===== GitOps Template (optional - use default or custom) =====
  templateRepo: "https://github.com/YOUR_USERNAME/k8s-templates"
  templatePath: "templates/simple-app"

  # ===== GitOps Destination =====
  gitopsRepo: "https://github.com/YOUR_USERNAME/gitops-manifests"
  gitopsPath: "apps/dev/my-nodejs-app"

  # ===== Context Subpath (IMPORTANT if Dockerfile is in a subdirectory) =====
  # Leave empty or omit if Dockerfile is at repo root
  # Set to subdirectory name if Dockerfile is in a subfolder
  contextSubpath: "" # e.g., "my-app" if Dockerfile is in my-app/Dockerfile

  # ===== Custom Values =====
  values:
    environment: "dev"
    replicaCount: "2"
    imageTag: "latest"
```

**Important Configuration Notes:**

1. **`gitRepo`**: Your application source code repository (must contain a Dockerfile)
2. **`imageRepo`**: Where to push built images (e.g., `docker.io/username/app` or `ghcr.io/org/app`)
3. **`gitopsPath`**: Must match your repo name! If repo is "my-nodejs-app", use `apps/dev/my-nodejs-app`
4. **`contextSubpath`**:
   - Leave empty (`""`) if Dockerfile is at repo root
   - Set to subdirectory (e.g., `"my-app"`) if Dockerfile is in a subfolder
   - Example: If your repo structure is `repo/my-app/Dockerfile`, set `contextSubpath: "my-app"`

**Example for monorepo structure:**

```yaml
# If your repo has: my-monorepo/backend/Dockerfile
spec:
  gitRepo: "https://github.com/YOUR_USERNAME/my-monorepo"
  contextSubpath: "backend" # <-- Points to the subdirectory
  gitopsPath: "apps/dev/my-monorepo" # <-- Matches repo name
```

All credentials are provided via Kubernetes Secrets. No sensitive data is hardcoded in the HeliosApp.

#### 7.2. Apply the HeliosApp

This single command triggers the entire automated workflow:

```bash
kubectl apply -f my-app.yaml
```

#### 7.3. Watch the Operator Orchestrate Everything

The Helios Operator now automatically handles the entire deployment workflow. Open multiple terminals to watch each phase:

**Terminal 1 - Watch HeliosApp status (Overall Progress):**

```bash
kubectl get heliosapp my-nodejs-app -w
```

You'll see the operator update the status as it progresses through each phase.

**Terminal 2 - Watch PipelineRuns (Build & GitOps Update):**

```bash
# Watch PipelineRuns created by the operator
kubectl get pipelinerun --sort-by=.metadata.creationTimestamp

# Or watch in real-time
kubectl get pipelinerun -w
```

**Terminal 3 - Watch ArgoCD Applications (Deployment):**

```bash
# The operator automatically creates this!
kubectl get application -n argocd -w
```

**Terminal 4 - Watch Pods Being Deployed:**

```bash
kubectl get pods -l app=my-nodejs-app -w
```

**What the Operator Does Automatically:**

1. ✅ **Creates Tekton Triggers** - EventListener, TriggerBindings, TriggerTemplate
2. ✅ **Triggers PipelineRun** - Builds image and updates GitOps repo
3. ✅ **Creates ArgoCD Application** - Named `{app-name}-argocd`
4. ✅ **Monitors Progress** - Tracks pipeline → ArgoCD sync → deployment health
5. ✅ **Updates Status** - HeliosApp status shows current state and deployed version

#### 7.4. Monitor Pipeline Execution

```bash
# List all pipeline runs
kubectl get pipelinerun

# Get detailed status
kubectl describe pipelinerun <pipelinerun-name>

# View logs (requires tkn CLI)
tkn pipelinerun logs <pipelinerun-name> -f

# Or use kubectl
kubectl logs -l tekton.dev/pipelineRun=<pipelinerun-name> --all-containers=true -f
```

#### 7.5. Check ArgoCD Application (Created by Operator)

The operator automatically creates an ArgoCD Application for you:

```bash
# Check if ArgoCD app was created (named {app-name}-argocd)
kubectl get application -n argocd my-nodejs-app-argocd

# Get detailed status
kubectl get application -n argocd my-nodejs-app-argocd -o wide

# Get sync status
kubectl get application -n argocd my-nodejs-app-argocd -o jsonpath='{.status.sync.status}'

# Get health status
kubectl get application -n argocd my-nodejs-app-argocd -o jsonpath='{.status.health.status}'
```

**Expected Status Progression:**

1. **Initial**: `Synced` + `Progressing` - ArgoCD is deploying
2. **Final**: `Synced` + `Healthy` - Application is fully deployed

#### 7.6. Verify Deployment and Check HeliosApp Status

Once everything is deployed, verify the complete status:

```bash
# Check HeliosApp status (comprehensive view)
kubectl describe heliosapp my-nodejs-app

# Get status in JSON format
kubectl get heliosapp my-nodejs-app -o jsonpath='{.status}' | jq '.'
```

**Expected HeliosApp Status:**

```json
{
  "argoApplication": "my-nodejs-app-argocd",
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "reason": "SyncedAndHealthy",
      "message": "Application is synced and healthy. Deployed version: docker.io/username/app:sha256:..."
    }
  ],
  "deployedVersion": "docker.io/username/app:sha256:...",
  "manifestPipelineRun": "my-nodejs-app-manifest-20251019-132402",
  "observedGeneration": 1
}
```

**Verify the deployed application:**

```bash
# Check pods
kubectl get pods -l app=my-nodejs-app

# Check service
kubectl get svc my-nodejs-app

# Test the application
kubectl port-forward svc/my-nodejs-app 3000:80

# In another terminal
curl http://localhost:3000
```

**You should see your application responding!** 🎉

---

### Step 8: Test the Full Automated GitOps Flow

Now let's test the complete end-to-end automation! The operator handles everything automatically.

#### 8.1. Make a Code Change and Push

```bash
cd my-nodejs-app

# Update the message
sed -i 's/Hello from Helios!/Hello from Helios v2.0!/' index.js

# Commit and push
git add index.js
git commit -m "Update welcome message to v2.0"
git push origin main
```

**🎉 What Happens Next (Fully Automated):**

1. **GitHub Webhook** triggers the Tekton EventListener
2. **Tekton Pipeline** runs automatically:
   - Clones your code
   - Builds new Docker image
   - Pushes to registry
   - Updates GitOps repo with new image digest (~2-3 minutes)
3. **ArgoCD** detects the change and syncs automatically:
   - With default config: within 3 minutes of GitOps update
   - With fast config (Step 2.3.1): within 30 seconds
4. **HeliosApp status** updates automatically
5. **Pods** roll out with new version

**The operator orchestrates this entire flow without any manual intervention!**

**Timeline Example:**

- T+0s: Push commit to GitHub
- T+5s: Webhook triggers PipelineRun
- T+180s: Pipeline completes, GitOps repo updated
- T+210s: ArgoCD detects change (with 30s interval)
- T+240s: New pods running with updated version ✅

#### 8.2. Watch the Automated Flow

Monitor the operator managing the entire deployment:

```bash
# Watch HeliosApp status update
kubectl get heliosapp my-nodejs-app -w

# Watch new PipelineRun (triggered by webhook)
kubectl get pipelinerun --sort-by=.metadata.creationTimestamp | tail -5

# Watch ArgoCD sync the new version
kubectl get application -n argocd my-nodejs-app-argocd -w

# Watch pods rolling update
kubectl get pods -l app=my-nodejs-app -w

# Check EventListener logs to see webhook received
kubectl logs -l eventlistener=helios-listener --tail=20
```

**Status Progression You'll See:**

1. HeliosApp: `ManifestGenerationInProgress` → `DeployingWithArgoCD` → `SyncedAndHealthy`
2. PipelineRun: `Running` → `Succeeded` (2-3 minutes)
3. ArgoCD: `OutOfSync` → `Synced` + `Healthy` (30s-3min after GitOps update)
4. Pods: Old version → `Terminating` → New version `Running`

**⏱️ Sync Timing:**

- ArgoCD uses automated sync with a polling interval
- Default: Checks Git every **3 minutes**
- Fast config: Checks Git every **30 seconds** (if you completed Step 2.3.1)
- The operator has already configured `automated: true`, `prune: true`, `selfHeal: true`

#### 8.3. Verify on GitHub

1. Go to your GitHub repository
2. Click **Settings** > **Webhooks**
3. Click on your webhook
4. Check **Recent Deliveries**
5. You should see your push event with ✅ green checkmark

#### 8.4. Manual Trigger (Without Webhooks)

If you **skipped webhook setup**, you can manually trigger a new deployment by updating the HeliosApp:

```bash
# Option 1: Edit the HeliosApp (triggers new PipelineRun)
kubectl edit heliosapp my-nodejs-app
# Change spec.gitBranch or any spec field to trigger reconciliation

# Option 2: Update via annotation (triggers reconciliation)
kubectl annotate heliosapp my-nodejs-app reconcile=$(date +%s) --overwrite

# Option 3: Change a value
kubectl patch heliosapp my-nodejs-app --type merge -p '{"spec":{"values":{"imageTag":"v2.0"}}}'
```

**Any change to the HeliosApp spec triggers the operator to:**

1. Create a new PipelineRun
2. Wait for it to complete
3. Let ArgoCD sync the updated manifests

This gives you full control even without webhooks configured.

#### 8.5. Monitor Operator-Managed Workflow

Watch the operator manage the entire workflow automatically:

```bash
# Check HeliosApp status (shows complete workflow state)
kubectl get heliosapp my-nodejs-app -o jsonpath='{.status.conditions[0]}' | jq '.'

# Watch the operator's logs
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager -f

# Check latest PipelineRun created by operator
kubectl get pipelinerun --sort-by=.metadata.creationTimestamp | tail -1

# Watch ArgoCD sync (managed by operator)
kubectl get application -n argocd my-nodejs-app-argocd -o jsonpath='{.status.sync.status}'

# Check pods rolling update
kubectl get pods -l app=my-nodejs-app -w
```

**Operator Log Messages You'll See:**

```text
Reconciling HeliosApp
Triggering Tekton PipelineRun for manifest generation
Successfully created PipelineRun
PipelineRun succeeded! Manifest has been generated
Creating ArgoCD Application
ArgoCD application is synced and healthy
```

#### 8.6. Verify Updated Application

```bash
# Port forward
kubectl port-forward svc/my-nodejs-app 3000:80

# Test
curl http://localhost:3000
# Should see: {"message":"Hello from Helios v2.0!",...}
```

#### 8.7. Check GitOps Repository

```bash
# Clone your gitops repo to see the changes
cd ~/
git clone https://github.com/YOUR_USERNAME/gitops-manifests
cd gitops-manifests/apps/dev/my-nodejs-app

# View the generated manifests
cat deployment.yaml
# You should see the new image tag!
```

---

### Step 9: Advanced Webhook Troubleshooting

If webhooks aren't working as expected after following Step 6, use these advanced troubleshooting steps:

#### 9.1. Debug 307 Redirect Error

If you're seeing a 307 Temporary Redirect status:

**Cause:** Using incorrect webhook URL path (e.g., `/hooks` instead of `/`)

**Solution:**

```bash
# ❌ Wrong - causes 307 redirect
Payload URL: https://YOUR_NGROK_URL/hooks

# ✅ Correct - use root path
Payload URL: https://YOUR_NGROK_URL/

# Update your GitHub webhook settings to use the root path
```

The EventListener listens on the root path `/`, not `/hooks`.

#### 9.2. Debug Webhook Connectivity

If webhook shows "Failed to connect" in GitHub:

```bash
# Check if EventListener is running
kubectl get pods -l eventlistener=helios-listener

# Check EventListener logs
kubectl logs -l eventlistener=helios-listener --tail=50

# Verify service is accessible
kubectl get svc el-helios-listener

# For ngrok setup, check if port-forward is running
ps aux | grep "port-forward.*el-helios-listener"

# If not running, restart it
kubectl port-forward -n default svc/el-helios-listener 8080:8080 > /tmp/port-forward.log 2>&1 &

# Test locally (note: root path /, not /hooks)
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-Hub-Signature-256: sha256=test" \
  -d '{"ref":"refs/heads/main","repository":{"clone_url":"https://github.com/YOUR_USERNAME/my-nodejs-app","name":"my-nodejs-app"},"after":"abc123"}'
```

#### 9.3. Debug Authentication Issues

If webhook returns 401 or 403:

```bash
# Verify webhook secret exists and matches
kubectl get secret github-webhook-secret -o jsonpath='{.data.secretToken}' | base64 -d
echo ""

# Check EventListener references the correct secret
kubectl describe eventlistener helios-listener | grep -A 5 "secretRef"

# If secret is wrong, recreate it
kubectl delete secret github-webhook-secret
kubectl create secret generic github-webhook-secret \
  --from-literal=secretToken="YOUR_NEW_SECRET"

# Then update the secret in GitHub webhook settings
```

#### 9.4. Debug PipelineRun Not Created

If webhook succeeds (200 OK) but no PipelineRun is created:

```bash
# Check all Tekton Triggers resources exist
kubectl get triggerbinding,triggertemplate,eventlistener

# Check EventListener logs for detailed errors
kubectl logs -l eventlistener=helios-listener --tail=100 -f

# Verify GitHub interceptor is running
kubectl get pods -n tekton-pipelines | grep github

# Check service account permissions
kubectl describe sa tekton-triggers-sa

# Verify TriggerTemplate configuration
kubectl describe triggertemplate helios-trigger-template

# Check recent events
kubectl get events --field-selector involvedObject.name=helios-listener --sort-by='.lastTimestamp'
```

---

## 🔄 Managing Your HeliosApp

### Step 9: Update Your Application Configuration

The HeliosApp Custom Resource is declarative - any changes to the spec will trigger the operator to reconcile and update your application. You don't need to delete and recreate; simply edit the YAML and reapply.

#### 9.1. Update Application Settings (Replicas, Port, etc.)

**Example: Scale your application**

```bash
# Edit your my-app.yaml file
vi my-app.yaml
```

Change the replicas:

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  # ... other fields ...
  replicas: 5 # Changed from 2 to 5
  # ... rest of spec ...
```

Apply the changes:

```bash
# Apply the updated HeliosApp
kubectl apply -f my-app.yaml

# Watch the operator reconcile
kubectl get heliosapp my-nodejs-app -w
```

**What happens:**

1. Operator detects the spec change
2. Triggers a new PipelineRun to regenerate manifests with new values
3. Updates the GitOps repository with new manifests
4. ArgoCD automatically syncs the changes
5. Kubernetes scales your deployment to 5 replicas

**Watch the update flow:**

```bash
# Terminal 1: Watch HeliosApp status
kubectl get heliosapp my-nodejs-app -w

# Terminal 2: Watch new PipelineRun
kubectl get pipelinerun --sort-by=.metadata.creationTimestamp | tail -5

# Terminal 3: Watch ArgoCD sync
kubectl get application -n argocd my-nodejs-app-argocd -w

# Terminal 4: Watch pods scale
kubectl get pods -l app=my-nodejs-app -w
```

#### 9.2. Update Environment Variables or Custom Values

```bash
# Edit my-app.yaml to add/update values
vi my-app.yaml
```

Add custom environment variables:

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  # ... existing fields ...
  values:
    environment: "dev"
    replicaCount: "5"
    imageTag: "latest"
    # Add new custom values
    logLevel: "debug"
    featureFlags: "new-feature-enabled"
    databaseUrl: "postgres://db.example.com:5432/mydb"
```

Apply and watch:

```bash
kubectl apply -f my-app.yaml

# Monitor the update
kubectl describe heliosapp my-nodejs-app
```

**Note:** These values are passed to your manifest templates. Make sure your templates (in `templateRepo`) use these values correctly.

#### 9.3. Update Git Repository or Branch

**Example: Switch to a different branch**

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  gitRepo: "https://github.com/YOUR_USERNAME/my-nodejs-app"
  gitBranch: "staging" # Changed from "main" to "staging"
  # ... rest of spec ...
```

Apply the change:

```bash
kubectl apply -f my-app.yaml
```

**What happens:**

1. Next build will pull from the `staging` branch
2. New image is built from staging code
3. GitOps repo is updated with the new image
4. ArgoCD deploys the staging version

#### 9.4. Update Image Registry or Repository

**Example: Switch to a different container registry**

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  # Change from Docker Hub to GitHub Container Registry
  imageRepo: "ghcr.io/YOUR_USERNAME/my-nodejs-app" # Changed from docker.io
  # ... rest of spec ...
```

**Important:** Make sure you have credentials configured for the new registry:

```bash
# Create new registry credentials
kubectl create secret docker-registry ghcr-credentials \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN \
  --docker-email=YOUR_EMAIL

# Update the docker-credentials secret or patch service account
kubectl patch serviceaccount pipeline-sa -p '{"imagePullSecrets": [{"name": "ghcr-credentials"}]}'
```

Then apply:

```bash
kubectl apply -f my-app.yaml
```

#### 9.5. Update GitOps Repository or Path

**Example: Move to a different environment path**

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  # ... existing fields ...
  gitopsRepo: "https://github.com/YOUR_USERNAME/gitops-manifests"
  gitopsPath: "apps/staging/my-nodejs-app" # Changed from apps/dev/my-nodejs-app
  # ... rest of spec ...
```

**Important:** Make sure the new path exists in your GitOps repository with an initial manifest file!

```bash
# In your gitops-manifests repo
mkdir -p apps/staging/my-nodejs-app
cp apps/dev/my-nodejs-app/deployment.yaml apps/staging/my-nodejs-app/
git add apps/staging/my-nodejs-app/
git commit -m "Add staging environment for my-nodejs-app"
git push
```

Then apply the HeliosApp update:

```bash
kubectl apply -f my-app.yaml
```

#### 9.6. Quick Update via kubectl patch

For quick changes without editing the YAML file:

```bash
# Scale replicas
kubectl patch heliosapp my-nodejs-app --type merge \
  -p '{"spec":{"replicas":3}}'

# Update git branch
kubectl patch heliosapp my-nodejs-app --type merge \
  -p '{"spec":{"gitBranch":"develop"}}'

# Add/update a custom value
kubectl patch heliosapp my-nodejs-app --type merge \
  -p '{"spec":{"values":{"logLevel":"info"}}}'

# Force reconciliation (useful for debugging)
kubectl annotate heliosapp my-nodejs-app reconcile=$(date +%s) --overwrite
```

#### 9.7. View Update Progress

**Check HeliosApp status:**

```bash
# Quick status
kubectl get heliosapp my-nodejs-app

# Detailed status with conditions
kubectl get heliosapp my-nodejs-app -o jsonpath='{.status}' | jq '.'

# Human-readable description
kubectl describe heliosapp my-nodejs-app
```

**Expected status during update:**

```json
{
  "conditions": [
    {
      "type": "Ready",
      "status": "False",
      "reason": "ManifestGenerationInProgress",
      "message": "PipelineRun my-nodejs-app-manifest-... is running"
    }
  ],
  "manifestPipelineRun": "my-nodejs-app-manifest-20251019-143022",
  "observedGeneration": 2
}
```

**After update completes:**

```json
{
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "reason": "SyncedAndHealthy",
      "message": "Application is synced and healthy. Deployed version: docker.io/..."
    }
  ],
  "argoApplication": "my-nodejs-app-argocd",
  "deployedVersion": "docker.io/username/app:sha256:...",
  "manifestPipelineRun": "my-nodejs-app-manifest-20251019-143022",
  "observedGeneration": 2
}
```

#### 9.8. Rollback to Previous Configuration

If you need to revert changes:

```bash
# Option 1: Revert your my-app.yaml file
git checkout HEAD~1 my-app.yaml
kubectl apply -f my-app.yaml

# Option 2: Use kubectl rollout (on the actual deployment in cluster)
kubectl rollout undo deployment/my-nodejs-app

# Option 3: Edit HeliosApp directly
kubectl edit heliosapp my-nodejs-app
# Change spec back to previous values, save and exit
```

**Note:** The operator manages the deployment through GitOps, so using `kubectl rollout undo` directly on the deployment will be overridden by ArgoCD. It's better to update the HeliosApp spec or revert the GitOps repository.

---

## 🗑️ Deleting Your HeliosApp

### Step 10: Clean Up Your Application

When you want to remove your application completely, the Helios Operator provides automatic cleanup of most resources.

#### 10.1. Delete the HeliosApp

Simply delete the HeliosApp Custom Resource:

```bash
# Delete using the YAML file
kubectl delete -f my-app.yaml

# Or delete by name
kubectl delete heliosapp my-nodejs-app

# Watch the deletion process
kubectl get heliosapp my-nodejs-app -w
```

#### 10.2. What Gets Deleted Automatically

The operator automatically cleans up these resources (via owner references):

✅ **Automatically Deleted:**

- **Tekton Triggers Resources:**
  - EventListener (`el-helios-listener-my-nodejs-app`)
  - TriggerBinding (`github-binding-my-nodejs-app`)
  - TriggerBinding for defaults
  - TriggerTemplate (`helios-trigger-template-my-nodejs-app`)
- **PipelineRuns:** All PipelineRuns created by this HeliosApp
- **ConfigMaps/Secrets:** Any resources owned by the HeliosApp

**Verify automatic cleanup:**

```bash
# Check EventListener is gone
kubectl get eventlistener -l heliosapp=my-nodejs-app

# Check TriggerBindings are gone
kubectl get triggerbinding -l heliosapp=my-nodejs-app

# Check TriggerTemplates are gone
kubectl get triggertemplate -l heliosapp=my-nodejs-app

# Check PipelineRuns are gone
kubectl get pipelinerun -l heliosapp=my-nodejs-app
```

#### 10.3. What Needs Manual Deletion

⚠️ **NOT Automatically Deleted (requires manual cleanup):**

1. **ArgoCD Application** - Intentionally preserved for safety
2. **Deployed Application Resources** - Your actual pods, services, deployments
3. **GitOps Repository Content** - Manifest files remain in Git
4. **Container Images** - Images remain in your registry

**Why ArgoCD Application is NOT auto-deleted:**

The operator does **not** automatically delete the ArgoCD Application to prevent accidental deletion of running workloads. This is a safety feature to protect your production applications.

#### 10.4. Manual Cleanup Steps

**Step 1: Delete the ArgoCD Application**

```bash
# Check if ArgoCD Application exists
kubectl get application -n argocd my-nodejs-app-argocd

# Delete the ArgoCD Application
kubectl delete application -n argocd my-nodejs-app-argocd
```

**What happens when you delete ArgoCD Application:**

- By default, ArgoCD will **NOT** delete the deployed resources (pods, services, etc.)
- The application is just removed from ArgoCD's management
- Your workloads continue running

**Step 2: Delete the Deployed Application Resources**

If you want to remove the actual deployed application:

```bash
# Option A: Let ArgoCD cascade delete (if configured)
kubectl delete application -n argocd my-nodejs-app-argocd --cascade=true

# Option B: Manually delete resources
kubectl delete deployment my-nodejs-app
kubectl delete service my-nodejs-app
kubectl delete all -l app=my-nodejs-app

# Option C: Delete the entire namespace (if dedicated)
kubectl delete namespace my-app-namespace
```

**Step 3: Clean Up GitOps Repository (Optional)**

```bash
# Clone your gitops repo
cd ~/gitops-manifests
git pull

# Remove the application directory
rm -rf apps/dev/my-nodejs-app

# Commit and push
git add .
git commit -m "Remove my-nodejs-app manifests"
git push origin main
```

**Step 4: Remove Container Images (Optional)**

Images in your registry are not automatically deleted. Remove them manually if needed:

```bash
# For Docker Hub: use the web UI or Docker Hub API
# For GHCR: use GitHub Packages UI or API
# For other registries: use their respective tools

# Example: List images
docker images | grep my-nodejs-app

# Remove local images
docker rmi docker.io/YOUR_USERNAME/my-nodejs-app:TAG
```

#### 10.5. Complete Cleanup Script

For convenience, here's a complete cleanup script:

```bash
#!/bin/bash
APP_NAME="my-nodejs-app"
NAMESPACE="default"
ARGOCD_NAMESPACE="argocd"

echo "🗑️  Starting cleanup for ${APP_NAME}..."

# 1. Delete HeliosApp (triggers automatic cleanup)
echo "Deleting HeliosApp..."
kubectl delete heliosapp ${APP_NAME} -n ${NAMESPACE} --wait=true

# 2. Wait for automatic cleanup
echo "Waiting for automatic cleanup..."
sleep 5

# 3. Verify automatic cleanup
echo "Verifying automatic cleanup..."
kubectl get eventlistener,triggerbinding,triggertemplate -l heliosapp=${APP_NAME} 2>/dev/null
kubectl get pipelinerun -l heliosapp=${APP_NAME} 2>/dev/null

# 4. Delete ArgoCD Application (manual step)
echo "Deleting ArgoCD Application..."
kubectl delete application ${APP_NAME}-argocd -n ${ARGOCD_NAMESPACE} 2>/dev/null

# 5. Delete deployed resources (optional - uncomment if needed)
# echo "Deleting deployed resources..."
# kubectl delete deployment ${APP_NAME} -n ${NAMESPACE}
# kubectl delete service ${APP_NAME} -n ${NAMESPACE}
# kubectl delete all -l app=${APP_NAME} -n ${NAMESPACE}

echo "✅ Cleanup complete!"
echo ""
echo "📝 Manual steps remaining:"
echo "  1. Remove manifests from GitOps repo: apps/dev/${APP_NAME}/"
echo "  2. Remove container images from registry (if desired)"
```

Save as `cleanup-app.sh`, make executable, and run:

```bash
chmod +x cleanup-app.sh
./cleanup-app.sh
```

#### 10.6. Verify Complete Cleanup

```bash
# Check HeliosApp is gone
kubectl get heliosapp my-nodejs-app
# Should show: Error from server (NotFound)

# Check Tekton resources are gone
kubectl get eventlistener,triggerbinding,triggertemplate -l heliosapp=my-nodejs-app
# Should show: No resources found

# Check ArgoCD Application is gone
kubectl get application -n argocd my-nodejs-app-argocd
# Should show: Error from server (NotFound)

# Check deployed resources are gone
kubectl get all -l app=my-nodejs-app
# Should show: No resources found (if you deleted them)

# Check PipelineRuns are gone
kubectl get pipelinerun -l heliosapp=my-nodejs-app
# Should show: No resources found
```

#### 10.7. Delete vs. Pause

If you want to temporarily stop the automation without deleting everything, you can pause instead:

```bash
# Pause automation by scaling operator to 0
kubectl scale deployment helios-operator-controller-manager \
  -n helios-operator-system --replicas=0

# Your application keeps running, but no new reconciliation happens

# Resume later
kubectl scale deployment helios-operator-controller-manager \
  -n helios-operator-system --replicas=1
```

Or simply disable webhooks:

```bash
# Delete the webhook in GitHub UI
# EventListener stays running but receives no events
# Your app continues running normally
```

---

## 🎯 Understanding the Operator-Managed Flow

Here's what happens when you create/update a HeliosApp with the Helios Operator:

````mermaid
sequenceDiagram
    participant Dev as Developer
    participant K8s as Kubernetes
    participant Helios as Helios Operator
    participant Tekton as Tekton Pipeline
    participant Registry as Docker Registry
    participant GitOps as GitOps Repo
    participant ArgoCD as ArgoCD

    Dev->>K8s: kubectl apply -f heliosapp.yaml
    K8s->>Helios: HeliosApp CR Created/Updated

    Note over Helios: Reconciliation Loop Starts
    Helios->>Helios: Create Tekton Triggers
    Helios->>Tekton: Create PipelineRun

    Note over Tekton: Phase 1: Build & Push
    Tekton->>GitHub: Clone source code
    Tekton->>Tekton: Build Docker image
    Tekton->>Registry: Push image with digest

    Note over Tekton: Phase 2: Update GitOps
    Tekton->>GitOps: Clone GitOps repo
    Tekton->>GitOps: Update image digest
    Tekton->>GitOps: Commit & push changes

    Note over Helios: Monitor PipelineRun
    Helios->>Tekton: Check PipelineRun status
    Tekton-->>Helios: Succeeded

    Note over Helios: Create ArgoCD Resources
    Helios->>ArgoCD: Create Application

    Note over ArgoCD: Phase 3: Deploy
    ArgoCD->>GitOps: Pull updated manifests
    ArgoCD->>K8s: Apply manifests
    ArgoCD-->>Helios: Report sync status

    Note over Helios: Update Status
    Helios->>K8s: Update HeliosApp status
    K8s-->>Dev: Status: Ready (Synced & Healthy)
```text

**The operator manages three phases automatically:**

1. **Build Phase** (Tekton Pipeline):
   - Operator creates PipelineRun with correct parameters
   - Clones your source code (with `contextSubpath` if specified)
   - Builds Docker image using Kaniko
   - Pushes to registry with SHA digest
   - Operator monitors PipelineRun status

2. **GitOps Update Phase** (Tekton Pipeline):
   - Updates manifest in GitOps repo with new image digest
   - Commits and pushes changes
   - Operator waits for completion

3. **Deployment Phase** (ArgoCD):
   - Operator creates ArgoCD Application with automated sync enabled (if not exists)
   - ArgoCD polls GitOps repo at regular intervals (default: 3min, fast: 30s)
   - When changes detected, ArgoCD automatically syncs manifests to cluster
   - Auto-sync, prune, and self-heal are enabled by default
   - Operator monitors ArgoCD sync status
   - Updates HeliosApp status based on deployment health

**About ArgoCD Sync Speed:**
- ArgoCD uses **polling** to detect Git changes (not webhooks by default)
- Sync happens automatically once changes are detected
- If you configured Step 2.3.1, sync occurs within 30 seconds
- For instant sync, you can manually trigger: `kubectl patch application my-nodejs-app-argocd -n argocd --type merge -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"hard"}}}'`

**Status Lifecycle:**

````

HeliosApp Applied
↓
ManifestGenerationInProgress (PipelineRun running)
↓
DeployingWithArgoCD (ArgoCD syncing)
↓
SyncedAndHealthy (Application deployed ✅)

````

---

## 📊 Monitoring and Debugging

### Check Overall Status

```bash
# Quick health check
./scripts/health-check.sh  # If available

# Or manually:
echo "=== Helios Operator ==="
kubectl get pods -n helios-operator-system

echo "=== Tekton ==="
kubectl get pods -n tekton-pipelines

echo "=== ArgoCD ==="
kubectl get pods -n argocd

echo "=== HeliosApps ==="
kubectl get heliosapp -A

echo "=== Recent PipelineRuns ==="
kubectl get pipelinerun --sort-by=.metadata.creationTimestamp | tail -n 5

echo "=== ArgoCD Applications ==="
kubectl get application -n argocd
```text

### Debug Failed Pipeline

```bash
# Get failed pipelineruns
kubectl get pipelinerun -l heliosapp=my-nodejs-app --field-selector status.conditions[0].status=False

# Check logs
kubectl logs -l tekton.dev/pipelineRun=<pipelinerun-name> --all-containers=true

# Describe for events
kubectl describe pipelinerun <pipelinerun-name>
````

### Debug ArgoCD Sync Issues

```bash
# Check application details
kubectl describe application -n argocd my-nodejs-app

# View events
kubectl get events -n argocd --field-selector involvedObject.name=my-nodejs-app

# Check ArgoCD logs
kubectl logs -n argocd deployment/argocd-application-controller
```

### Debug Operator Issues

```bash
# Check operator logs
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager -f

# Check events for HeliosApp
kubectl describe heliosapp my-nodejs-app

# Check if reconciliation is happening
kubectl get events --field-selector involvedObject.name=my-nodejs-app
```

---

## 🔧 Common Issues and Solutions

### Issue 1: PipelineRun Fails - Git Clone Error

**Symptoms:**

```text
fatal: destination path '/workspace/output' already exists and is not an empty directory.
```

**Solution:**

This issue has been **fixed** in the latest version of `tekton/task-git-clone.yaml`. The git-clone task now automatically cleans the workspace before cloning, making it safe to retry failed pipelines.

If you installed an older version, update your git-clone task:

```bash
kubectl apply -f tekton/task-git-clone.yaml
```

The task now includes:

- Automatic workspace cleanup before cloning
- Better error messages
- Idempotent behavior (safe to retry)

**For authentication issues:**

```bash
# Verify git credentials exist
kubectl get secret git-credentials
kubectl describe secret git-credentials

# Recreate if needed
kubectl delete secret git-credentials
kubectl create secret generic git-credentials \
  --from-literal=username=YOUR_GITHUB_USERNAME \
  --from-literal=password=YOUR_GITHUB_TOKEN \
  --type=kubernetes.io/basic-auth

kubectl annotate secret git-credentials tekton.dev/git-0=https://github.com

# Verify service account has the secret
kubectl get sa pipeline-sa -o yaml | grep -A 5 secrets
```

### Issue 2: PipelineRun Fails - Docker Push Error

**Symptoms:**

```text
UNAUTHORIZED: authentication required
error checking push permissions -- make sure you entered the correct tag name, and that you are authenticated correctly
```

**Root Cause:**

Your `docker-credentials` secret doesn't contain valid Docker Hub authentication. This commonly happens when:

- You only have Google Cloud credentials in `~/.docker/config.json`
- Your Docker Hub access token expired
- You never logged into Docker Hub locally

**Solution:**

```bash
# Step 1: Login to Docker Hub first
docker login -u YOUR_DOCKERHUB_USERNAME
# Enter your Docker Hub Access Token (get it from https://hub.docker.com/settings/security)

# Step 2: Verify your config has Docker Hub auth
cat ~/.docker/config.json
# Should show "auths" with "https://index.docker.io/v1/" entry

# Step 3: Delete the old secret
kubectl delete secret docker-credentials

# Step 4: Recreate with proper credentials
kubectl create secret generic docker-credentials \
  --from-file=.dockerconfigjson=$HOME/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson

# Step 5: Verify
kubectl get secret docker-credentials -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | jq '.'
# Should see "auths" section with Docker Hub credentials

# Step 6: Delete failed PipelineRun and trigger a new one
kubectl delete pipelinerun <failed-pipelinerun-name>
# Push a new commit to trigger the pipeline again
```

**Prevention:**

Always login to Docker Hub **before** creating the `docker-credentials` secret during initial setup (Step 5.1).

### Issue 3: PipelineRun Fails - GitOps Manifest Update Error

**Symptoms:**

```text
sed: apps/dev/my-app/deployment.yaml: No such file or directory
```

**Root Cause:**

The `git-update-manifest` task expects the manifest file to already exist in your gitops repository. It updates the `image:` field but doesn't create the file from scratch.

**Solution:**

You must create the initial manifest file in your gitops repository (this should have been done in Step 0.2):

```bash
# Clone your gitops repo
git clone https://github.com/YOUR_USERNAME/gitops-manifests
cd gitops-manifests

# Create directory matching your GitHub repo name
# If your repo is "helios", use apps/dev/helios
# If your repo is "my-nodejs-app", use apps/dev/my-nodejs-app
mkdir -p apps/dev/my-nodejs-app

# Create initial manifest
cat > apps/dev/my-nodejs-app/deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: my-nodejs-app
  template:
    metadata:
      labels:
        app: my-nodejs-app
    spec:
      containers:
      - name: app
        image: docker.io/YOUR_DOCKERHUB_USERNAME/my-nodejs-app:latest
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: my-nodejs-app
  namespace: default
spec:
  selector:
    app: my-nodejs-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
EOF

# Commit and push
git add .
git commit -m "feat: Add initial manifest for my-nodejs-app"
git push origin main
```

**Important:** The directory name must match your GitHub repository name because the TriggerBinding uses `$(body.repository.name)` from the webhook payload.

### Issue 4: ArgoCD Application Not Created by Operator

**Symptoms:**

```text
kubectl get application -n argocd
No resources found in argocd namespace.
# OR
# Only manual application exists, not the operator-created one
```

**Root Cause:**

The operator creates ArgoCD Applications only after the PipelineRun succeeds. If the PipelineRun fails, the operator won't proceed to create the ArgoCD Application.

**Solution:**

```bash
# 1. Check HeliosApp status
kubectl get heliosapp my-nodejs-app -o jsonpath='{.status.conditions[0]}' | jq '.'

# 2. Check if PipelineRun succeeded
kubectl get pipelinerun --sort-by=.metadata.creationTimestamp | tail -5

# 3. Check operator logs
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager | grep -i argocd

# 4. Verify operator has permissions
kubectl auth can-i create applications --as=system:serviceaccount:helios-operator-system:helios-operator-controller-manager -n argocd

# 5. If PipelineRun failed, check why
kubectl describe pipelinerun <failed-pipelinerun-name>

# 6. Fix the issue and trigger reconciliation
kubectl annotate heliosapp my-nodejs-app reconcile=$(date +%s) --overwrite
```

**Common Causes:**

- PipelineRun failed (check Issues 1-3)
- Initial manifest doesn't exist in GitOps repo (see Issue 3)
- `contextSubpath` misconfigured
- Operator doesn't have RBAC permissions for ArgoCD namespace

### Issue 5: ArgoCD Not Syncing Immediately

**Symptoms:**

```text
# GitOps repo has been updated but pods are still running old version
kubectl get pods -l app=my-app  # Shows old image
```

**Root Cause:**

ArgoCD uses **polling** to detect changes in Git repositories. By default, it checks every **3 minutes**, which means there's a delay between:

1. Tekton updating the GitOps repo (happens immediately after build)
2. ArgoCD detecting the change (happens at next polling interval)

This is **normal behavior** and not a bug!

**Solution 1: Wait for Automatic Sync (Recommended)**

Simply wait up to 3 minutes (or 30 seconds if you configured Step 2.3.1). ArgoCD will automatically sync when it detects the change.

```bash
# Watch ArgoCD Application status
kubectl get application -n argocd my-nodejs-app-argocd -w

# Within 3 minutes (or 30s), you should see:
# - revision changes to latest commit
# - sync happens automatically
# - pods roll out with new version
```

**Solution 2: Manual Refresh (For Instant Sync)**

If you need immediate deployment, manually trigger ArgoCD to refresh:

```bash
# Force ArgoCD to check Git immediately
kubectl patch application my-nodejs-app-argocd -n argocd --type merge \
  -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"hard"}}}'

# Watch sync happen
kubectl get application -n argocd my-nodejs-app-argocd -w
```

**Solution 3: Reduce Polling Interval (Development Only)**

For faster development cycles, reduce ArgoCD's polling interval (see Step 2.3.1):

```bash
# Set to 30 seconds (already done if you followed Step 2.3.1)
kubectl patch configmap argocd-cm -n argocd --type merge \
  -p '{"data":{"timeout.reconciliation":"30s"}}'

# Restart ArgoCD
kubectl rollout restart deployment argocd-repo-server -n argocd
kubectl rollout restart statefulset argocd-application-controller -n argocd
```

**Prevention:**

- For development: Configure Step 2.3.1 during initial setup
- For production: Accept the 3-minute delay as normal behavior
- The operator already enables `automated: true`, `prune: true`, `selfHeal: true` - no additional configuration needed

### Issue 6: Image Pull Errors

**Symptoms:**

```text
Failed to pull image: unauthorized
```

**Solution:**

```bash
# Create image pull secret in application namespace
kubectl create secret docker-registry regcred \
  --docker-server=https://index.docker.io/v1/ \
  --docker-username=YOUR_USERNAME \
  --docker-password=YOUR_TOKEN \
  --docker-email=YOUR_EMAIL

# Add to service account or deployment
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "regcred"}]}'
```

### Issue 7: Git Push Fails in Manifest Update

**Solution:**

```bash
# Check git-update-manifest task logs
kubectl logs -l tekton.dev/pipelineTask=update-gitops-manifest

# Verify git credentials for push
kubectl get secret git-credentials

# Check your repo URL uses HTTPS
# HTTPS: https://github.com/username/repo.git

# Ensure your PAT has write access to the GitOps repo (and is not expired)
# For fine-grained PATs: grant contents: read/write on the GitOps repo
```

---

## 🎓 Next Steps

Congratulations! You now have a fully working GitOps pipeline with Helios Operator.

### Learn More

- 📖 **[Architecture Guide](./ARCHITECTURE.md)** - Understand internal design
- 🔧 **[API Reference](./API_REFERENCE.md)** - Complete HeliosApp CRD spec
- 🚀 **[Production Deployment](./PRODUCTION_DEPLOYMENT.md)** - Deploy to production
- 📊 **[Monitoring Guide](./MONITORING.md)** - Setup metrics and alerts
- 🧪 **[Testing Guide](./TESTING_GUIDE.md)** - Write tests for your pipelines

### Advanced Scenarios

- **Multi-environment deployment** - Deploy to dev, staging, prod
- **Custom pipelines** - Use your own Tekton pipelines
- **Helm charts** - Use Helm instead of templates
- **Auto-scaling** - Configure HPA with custom metrics
- **Security scanning** - Add Trivy/Snyk to pipeline
- **GitOps with Webhooks** - Auto-trigger on git push

### Real-World Example

```yaml
# Production-ready configuration
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: production-api
  namespace: production
spec:
  gitRepo: "git@github.com:mycompany/api-service.git"
  gitBranch: "release"
  imageRepo: "mycompany.azurecr.io/api-service"
  port: 8080
  replicas: 5
  pipelineName: "secure-build-pipeline"
  serviceAccount: "prod-pipeline-sa"
  pvcName: "prod-workspace-pvc"
  templateRepo: "git@github.com:mycompany/helm-charts.git"
  templatePath: "charts/microservice"
  gitopsRepo: "git@github.com:mycompany/production-gitops.git"
  gitopsPath: "services/api-service"
  values:
    environment: "production"
    replicaCount: "5"
    autoscaling.enabled: "true"
    autoscaling.minReplicas: "3"
    autoscaling.maxReplicas: "20"
    autoscaling.targetCPU: "70"
    resources.requests.cpu: "500m"
    resources.requests.memory: "1Gi"
    resources.limits.cpu: "2000m"
    resources.limits.memory: "4Gi"
    ingress.enabled: "true"
    ingress.host: "api.mycompany.com"
    ingress.tls: "true"
    monitoring.enabled: "true"
    tracing.enabled: "true"
    securityScanning: "true"
```

---

## 🧹 Quick Cleanup

When you're done experimenting, see **Step 10: Deleting Your HeliosApp** below for detailed cleanup instructions.

**Quick cleanup summary:**

```bash
# 1. Delete HeliosApp (auto-cleans Tekton resources)
kubectl delete heliosapp my-nodejs-app

# 2. Manually delete ArgoCD Application (safety measure)
kubectl delete application -n argocd my-nodejs-app-argocd

# 3. Delete deployed resources (if desired)
kubectl delete all -l app=my-nodejs-app

# 4. Clean GitOps repo (optional)
# Remove apps/dev/my-nodejs-app/ directory from your gitops-manifests repo

# 5. Remove container images from registry (optional)
```

**Complete cluster cleanup (if using Minikube/Kind):**

```bash
# Delete Helios Operator
make undeploy
make uninstall

# Delete dependencies
kubectl delete namespace argocd
kubectl delete -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
kubectl delete -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml

# Delete cluster
minikube delete
# or
kind delete cluster --name helios-cluster
```

For detailed explanations of what gets deleted automatically vs. manually, see **Step 10** above.

---

## 💡 Tips and Best Practices

### 0. Understanding ArgoCD Sync Timing

**How Auto-Sync Works:**

The Helios Operator configures ArgoCD with `automated: true`, which means ArgoCD will automatically deploy changes **as soon as it detects them**. However, ArgoCD uses **polling** (not webhooks) to check Git repositories:

```text
Tekton Updates GitOps → ArgoCD Polls Git → ArgoCD Detects Change → Auto-Sync Deploys
                         (every 30s-3min)
```

**Timeline Example:**

- **T+0s**: You push code to GitHub
- **T+5s**: Webhook triggers Tekton Pipeline
- **T+180s**: Pipeline completes, updates GitOps repo with new image digest
- **T+210s**: ArgoCD's next poll cycle detects the change (30s after update)
- **T+220s**: ArgoCD automatically syncs and deploys new version ✅

**Why Not Instant?**

- ArgoCD polls at intervals to reduce load on Git servers
- Default: 3 minutes (production-safe)
- Fast config: 30 seconds (development-friendly)
- Both have automated sync enabled - the delay is only in **detection**, not deployment

**How to Speed It Up:**

1. **Recommended**: Configure Step 2.3.1 during setup (30-second polling)
2. **Alternative**: Manually refresh when needed (instant, but manual)
3. **Not Recommended**: Set polling to <10 seconds (excessive Git server load)

**Remember:** Once ArgoCD detects the change, deployment is automatic and fast (10-30 seconds). The perceived "delay" is just the polling interval.

### 1. Repository Structure

```text
my-nodejs-app/               # Source code repo
├── Dockerfile
├── src/
└── package.json

gitops-manifests/            # GitOps repo
├── apps/
│   ├── dev/
│   │   └── my-app/
│   │       ├── deployment.yaml
│   │       └── service.yaml
│   ├── staging/
│   └── production/

k8s-templates/               # Templates repo (optional)
└── templates/
    ├── simple-app/
    ├── helm-app/
    └── kustomize-app/
```

### 2. Use Separate Branches for Environments

```yaml
# dev environment
spec:
  gitBranch: "develop"
  gitopsPath: "apps/dev/my-app"

# staging
spec:
  gitBranch: "staging"
  gitopsPath: "apps/staging/my-app"

# production
spec:
  gitBranch: "main"
  gitopsPath: "apps/production/my-app"
```

### 3. Tag Your Images Properly

```yaml
# Use commit SHA for traceability
values:
  imageTag: "{{ .GitCommitSHA }}"

# Or use semantic versioning
values:
  imageTag: "v1.2.3"
```

### 4. Use Namespaces for Isolation

```bash
# Create namespaces
kubectl create namespace dev
kubectl create namespace staging
kubectl create namespace production

# Deploy to different namespaces
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-app
  namespace: production  # <-- Important
```

### 5. Monitor Everything

```bash
# Install Prometheus + Grafana
kubectl apply -f config/prometheus/

# View metrics
kubectl port-forward -n helios-operator-system svc/prometheus 9090:9090

# Setup alerts for failed pipelines
```

---

## 📚 Additional Resources

- **Tekton Documentation:** <https://tekton.dev/docs/>
- **ArgoCD Documentation:** <https://argo-cd.readthedocs.io/>
- **GitOps Principles:** <https://opengitops.dev/>
- **Kubernetes Operators:** <https://kubernetes.io/docs/concepts/extend-kubernetes/operator/>

---

## 🤝 Get Help

- 📖 **Documentation:** Check other docs in this folder
- 🐛 **Issues:** Report bugs on GitHub Issues
- 💬 **Discussions:** Use GitHub Discussions for questions
- 📧 **Contact:** Reach out to the team

---

Happy Deploying with Helios! 🚀
