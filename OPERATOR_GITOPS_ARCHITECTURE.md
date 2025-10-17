# Helios Operator - GitOps Enhanced Architecture

## 🎯 Tổng Quan

Helios Operator đã được nâng cấp để triển khai kiến trúc GitOps 2 giai đoạn theo tài liệu `new_flow.pdf`. Operator giờ đây hoạt động như một "nhạc trưởng" thông minh, điều phối toàn bộ quy trình từ code commit đến ứng dụng chạy thực tế.

## 🔄 Luồng Hoạt Động Mới

### **Giai Đoạn 1: Manifest Generation (Operator-Driven)**

1. Developer tạo/cập nhật `HeliosApp` CRD
2. Operator đọc spec và khởi tạo Tekton Triggers (webhook listener)
3. **Operator kích hoạt PipelineRun đặc biệt** để:
   - Clone template từ `templateRepo`
   - Render manifest với values từ `HeliosApp.spec.values`
   - Commit manifest đã render vào `gitopsRepo` tại đường dẫn `gitopsPath`
4. Operator cập nhật status: `ManifestGenerationInProgress` → theo dõi PipelineRun

### **Giai Đoạn 2: Deployment với ArgoCD**

5. Khi PipelineRun thành công, Operator tự động tạo **ArgoCD Application**
6. ArgoCD Application được cấu hình để:
   - Sync từ `gitopsRepo` tại `gitopsPath`
   - Deploy vào namespace của ứng dụng
   - Auto-sync enabled (prune & self-heal)
7. Operator liên tục đồng bộ trạng thái từ ArgoCD về `HeliosApp.status`

### **Giai Đoạn 3: Continuous Monitoring**

8. Operator theo dõi ArgoCD Application status:
   - Sync status (Synced/OutOfSync)
   - Health status (Healthy/Degraded/Progressing)
   - Deployed version (image tag)
9. Cập nhật `HeliosApp.status.conditions` với trạng thái cuối cùng: `SyncedAndHealthy`

## 📋 Các Thay Đổi Chính

### 1. **CRD Schema Mới** (`api/v1/heliosapp_types.go`)

#### Spec Fields Mới:

```go
type HeliosAppSpec struct {
    // ... existing fields ...

    // Template Configuration
    TemplateRepo string `json:"templateRepo"`  // Repo chứa Helm/Kustomize template
    TemplatePath string `json:"templatePath"`  // Đường dẫn đến template

    // GitOps Configuration
    GitOpsRepo string `json:"gitopsRepo"`      // GitOps repo trung tâm
    GitOpsPath string `json:"gitopsPath"`      // Thư mục của app trong GitOps repo

    // Custom Values
    Values map[string]string `json:"values,omitempty"`  // Values để render template
}
```

#### Status Fields Mới:

```go
type HeliosAppStatus struct {
    // ... existing fields ...

    ObservedGeneration  int64  `json:"observedGeneration,omitempty"`
    ManifestPipelineRun string `json:"manifestPipelineRun,omitempty"`
    ArgoApplication     string `json:"argoApplication,omitempty"`
}
```

### 2. **Helper Functions Mới** (`internal/controller/tekton_resources.go`)

- `GeneratePipelineRunForManifestGeneration()`: Tạo PipelineRun cho manifest generation
- `GenerateArgoApplication()`: Tạo ArgoCD Application resource

### 3. **Reconciliation Loop Mới** (`internal/controller/heliosapp_controller.go`)

Controller giờ đây có 3 giai đoạn logic rõ ràng:

```go
func (r *HeliosAppReconciler) Reconcile(ctx, req) {
    // 1. Khởi tạo Tekton Triggers
    // 2. GIAI ĐOẠN 1: Kích hoạt PipelineRun manifest generation
    // 3. Kiểm tra trạng thái PipelineRun
    // 4. GIAI ĐOẠN 2: Tạo ArgoCD Application khi PipelineRun thành công
    // 5. GIAI ĐOẠN 3: Đồng bộ trạng thái từ ArgoCD
}
```

## 🚀 Cách Sử Dụng

### 1. Chuẩn Bị Prerequisites

Đảm bảo cluster có:

- ✅ Tekton Pipelines & Triggers
- ✅ ArgoCD
- ✅ Tekton Pipeline `manifest-generation-pipeline` (tên mặc định)
- ✅ ServiceAccount với quyền push vào GitOps repo

### 2. Tạo HeliosApp CRD

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-production-api
  namespace: production
spec:
  # Source Code
  gitRepo: "https://github.com/myorg/my-api"
  gitBranch: "main"
  imageRepo: "ghcr.io/myorg/my-api"

  # App Config
  port: 8080
  replicas: 3

  # CI/CD
  pipelineName: "build-push-pipeline"
  serviceAccount: "tekton-bot-sa"
  webhookSecret: "github-webhook-secret"
  pvcName: "shared-workspace-pvc"

  # GitOps - Template (Giai đoạn 1)
  templateRepo: "https://github.com/myorg/app-templates"
  templatePath: "charts/nodejs-app"

  # GitOps - Deployment (Giai đoạn 2)
  gitopsRepo: "https://github.com/myorg/gitops-manifests"
  gitopsPath: "apps/production/my-api"

  # Custom Values
  values:
    replicaCount: "3"
    ingress.enabled: "true"
    ingress.host: "api.mycompany.com"
    resources.requests.cpu: "200m"
    resources.requests.memory: "256Mi"
```

### 3. Apply và Theo Dõi

```bash
# Apply HeliosApp
kubectl apply -f my-heliosapp.yaml

# Xem logs của Operator
kubectl logs -f <helios-operator-pod> -n helios-system

# Kiểm tra trạng thái
kubectl get heliosapp my-production-api -o yaml

# Xem PipelineRun
kubectl get pipelinerun -n production

# Xem ArgoCD Application
kubectl get application -n argocd
```

### 4. Status Conditions

Operator sẽ cập nhật các trạng thái sau:

1. **ManifestGenerationInProgress**: PipelineRun đang chạy
2. **ManifestGenerationFailed**: PipelineRun thất bại
3. **DeployingWithArgoCD**: ArgoCD Application đã được tạo
4. **SyncedAndHealthy**: Ứng dụng đã được deploy thành công ✅

## 📊 Monitoring & Debugging

### Xem Status Chi Tiết

```bash
kubectl get heliosapp my-production-api -o jsonpath='{.status}' | jq
```

Output mẫu:

```json
{
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "reason": "SyncedAndHealthy",
      "message": "Application is synced and healthy. Deployed version: ghcr.io/myorg/my-api:v1.2.3"
    }
  ],
  "deployedVersion": "ghcr.io/myorg/my-api:v1.2.3",
  "observedGeneration": 5,
  "manifestPipelineRun": "my-production-api-manifest-20250118-120530",
  "argoApplication": "my-production-api-argocd"
}
```

### Debug PipelineRun

```bash
# Xem logs của PipelineRun
tkn pipelinerun logs <pipelinerun-name> -f -n production

# Xem chi tiết PipelineRun
kubectl describe pipelinerun <pipelinerun-name> -n production
```

### Debug ArgoCD Application

```bash
# Xem status ArgoCD App
kubectl get application my-production-api-argocd -n argocd -o yaml

# Sync thủ công nếu cần
argocd app sync my-production-api-argocd
```

## 🔐 RBAC Permissions

Operator cần các quyền sau (đã được cấu hình sẵn):

```yaml
# Tekton Resources
- tekton.dev/pipelineruns: get, list, watch, create, update, patch, delete
- triggers.tekton.dev/*: get, list, watch, create, update, patch, delete

# ArgoCD Resources
- argoproj.io/applications: get, list, watch, create, update, patch, delete

# Core Resources
- platform.helios.io/heliosapps: get, list, watch, create, update, patch, delete
- platform.helios.io/heliosapps/status: get, update, patch
```

## 📝 Notes cho Pipeline Developer (Người 2)

Để hệ thống hoạt động, bạn cần tạo Tekton Pipeline với tên `manifest-generation-pipeline` với các params sau:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: manifest-generation-pipeline
spec:
  params:
    - name: app-name
    - name: app-namespace
    - name: git-repo
    - name: git-branch
    - name: image-repo
    - name: template-repo
    - name: template-path
    - name: gitops-repo
    - name: gitops-path
    - name: port
    - name: replicas
    # ... + các params từ values
  tasks:
    - name: clone-template
    - name: render-manifest
    - name: commit-to-gitops
```

## 🎓 Tài Liệu Tham Khảo

- **new_flow.pdf**: Kiến trúc chi tiết của luồng GitOps 2 giai đoạn
- **From Code to Cluster**: Tổng quan về quy trình CI/CD
- **Operator Pattern**: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
- **ArgoCD**: https://argo-cd.readthedocs.io/

## ✅ Checklist Triển Khai

- [x] Cập nhật CRD với các trường GitOps mới
- [x] Thêm helper functions cho PipelineRun và ArgoCD Application
- [x] Viết lại Reconciliation Loop với 3 giai đoạn
- [x] Thêm RBAC permissions cho Tekton và ArgoCD
- [x] Tạo file example HeliosApp hoàn chỉnh
- [x] Regenerate CRD manifests và deep copy code
- [ ] Test end-to-end với cluster thực
- [ ] Tạo Tekton Pipeline `manifest-generation-pipeline`
- [ ] Cấu hình ServiceAccount với credentials cho GitOps repo

---

**Người phụ trách**: Operator Logic Lead  
**Ngày hoàn thành**: 18/10/2025  
**Trạng thái**: ✅ Ready for Integration Testing
