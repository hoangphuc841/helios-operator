/*
Copyright 2025.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"k8s.io/client-go/tools/record"
)

// HeliosAppReconciler reconciles a HeliosApp object
type HeliosAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// Thêm EventRecorder để bắn event ra K8s
	EventRecorder record.EventRecorder
}

// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=triggers.tekton.dev,resources=eventlisteners;triggerbindings;triggertemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods;podmetrics,verbs=get;list;watch
// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch;delete

func (r *HeliosAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// ========================================================================
	// BƯỚC 1: Lấy HeliosApp instance
	// ========================================================================
	var heliosApp heliosappv1.HeliosApp
	if err := r.Get(ctx, req.NamespacedName, &heliosApp); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("HeliosApp not found. It may have been deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to fetch HeliosApp")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	name := heliosApp.Name
	namespace := heliosApp.Namespace



	// ========================================================================
	// BƯỚC 4: Tự động tối ưu hóa tài nguyên (Smart Optimization) & Lấy Metrics
	// ========================================================================
	logger.Info("DEBUG: Checking Ready condition", "conditions", heliosApp.Status.Conditions)
	logger.Info("DEBUG: Checking Ready condition", "conditions", heliosApp.Status.Conditions)
	isReady := false
	for _, c := range heliosApp.Status.Conditions {
		if c.Type == "Ready" && c.Status == "True" {
			isReady = true
			break
		}
	}

	if isReady {
		logger.Info("DEBUG: Conditions met! Calling Optimizer...")
		optimizer := &OptimizerService{}
		optResult := optimizer.AnalyzeResourceUsage(ctx, r.Client, &heliosApp)

		// Cập nhật Metrics vào Status để Frontend hiển thị
		if optResult.MeasuredCpu != "" {
			heliosApp.Status.CurrentCPU = optResult.MeasuredCpu
		}

		// Save Status
		if err := r.Status().Update(ctx, &heliosApp); err != nil {
			logger.Error(err, "Failed to update HeliosApp status with metrics")
		}

		if heliosApp.Spec.EnableAutoOptimization && optResult.IsWasteful {
			logger.Info("Waste detected! Triggering auto-fix...", "waste", optResult.WastePercentage)

			// Thực hiện Auto-Fix (GitOps Write-Back)
			if err := optimizer.AutoFixRepository(ctx, &heliosApp, optResult); err != nil {
				logger.Error(err, "Failed to auto-fix repository")
				r.EventRecorder.Event(&heliosApp, corev1.EventTypeWarning, "AutoFixFailed", err.Error())
			} else {
				r.EventRecorder.Event(&heliosApp, corev1.EventTypeNormal, "AutoFixTriggered", "Created PR to optimize resources")
			}
		}
	}
	logger.Info("Reconciling HeliosApp", "name", name, "namespace", namespace, "generation", heliosApp.Generation)

	// ========================================================================
	// BƯỚC 2: Khởi tạo Tekton Triggers (chỉ chạy một lần)
	// ========================================================================
	logger.Info("Ensuring Tekton Triggers are present for HeliosApp", "name", name)

	pvcName := heliosApp.Spec.PVCName
	if pvcName == "" {
		pvcName = "pvc-" + name
	}

	pipelineName := heliosApp.Spec.PipelineName
	serviceAccount := heliosApp.Spec.ServiceAccount
	githubSecret := heliosApp.Spec.WebhookSecret
	workspace := map[string]any{
		"name": "shared-data",
		"persistentVolumeClaim": map[string]any{
			"claimName": pvcName,
		},
	}

	// Tạo các tài nguyên Tekton Triggers
	// Generate defaults TriggerBinding based on HeliosApp spec
	defaultsBinding, err := GenerateDefaultsTriggerBinding(name+"-defaults", namespace, &heliosApp)
	if err != nil {
		logger.Error(err, "Failed to generate defaults TriggerBinding")
		return ctrl.Result{}, err
	}

	eventListener, err := GenerateEventListener(
		name+"-el", namespace, name+"-trigger", name+"-trigger-binding", name+"-defaults", name+"-trigger-template", githubSecret,
	)
	if err != nil {
		logger.Error(err, "Failed to ensure Tekton PipelineRun")
		// UPDATE STATUS: Failed
		heliosApp.Status.Phase = "Failed"
		heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
		r.Status().Update(ctx, &heliosApp)
		return ctrl.Result{}, err
	}
	triggerBinding, err := GenerateTriggerBinding(name+"-trigger-binding", namespace)
	if err != nil {
		logger.Error(err, "Failed to generate TriggerBinding")
		// UPDATE STATUS: Failed
		heliosApp.Status.Phase = "Failed"
		heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
		r.Status().Update(ctx, &heliosApp)
		return ctrl.Result{}, err
	}
	triggerTemplate, err := GenerateTriggerTemplate(
		name+"-trigger-template", namespace, name+"-pipelinerun", pipelineName, serviceAccount, workspace,
	)
	if err != nil {
		logger.Error(err, "Failed to generate TriggerTemplate")
		// UPDATE STATUS: Failed
		heliosApp.Status.Phase = "Failed"
		heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
		r.Status().Update(ctx, &heliosApp)
		return ctrl.Result{}, err
	}

	for _, obj := range []*unstructured.Unstructured{defaultsBinding, eventListener, triggerBinding, triggerTemplate} {
		obj.SetNamespace(namespace)
		if err := controllerutil.SetControllerReference(&heliosApp, obj, r.Scheme); err != nil {
			logger.Error(err, "Failed to set owner reference", "name", obj.GetName())
			// UPDATE STATUS: Failed
			heliosApp.Status.Phase = "Failed"
			heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
			r.Status().Update(ctx, &heliosApp)
			return ctrl.Result{}, err
		}

		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(obj.GroupVersionKind())
		err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: obj.GetName()}, existing)
		if err != nil {
			if client.IgnoreNotFound(err) == nil {
				if err := r.Create(ctx, obj); err != nil {
					logger.Error(err, "Failed to create Tekton Trigger resource", "name", obj.GetName())
					// UPDATE STATUS: Failed
					heliosApp.Status.Phase = "Failed"
					heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
					r.Status().Update(ctx, &heliosApp)
					return ctrl.Result{}, err
				}
				logger.Info("Created Tekton Trigger resource", "name", obj.GetName())
			} else {
				logger.Error(err, "Failed to get Tekton Trigger resource", "name", obj.GetName())
				// UPDATE STATUS: Failed
				heliosApp.Status.Phase = "Failed"
				heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{}, err
			}
		} else {
			if !equalUnstructured(obj, existing) {
				obj.SetResourceVersion(existing.GetResourceVersion())
				if err := r.Update(ctx, obj); err != nil {
					logger.Error(err, "Failed to update Tekton Trigger resource", "name", obj.GetName())
					// UPDATE STATUS: Failed
					heliosApp.Status.Phase = "Failed"
					heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
					r.Status().Update(ctx, &heliosApp)
					return ctrl.Result{}, err
				}
				logger.Info("Updated Tekton Trigger resource", "name", obj.GetName())
			}
		}
	}

	// ------------------------------------------------------------------
	// PHASE 0: Initialize Status
	// ------------------------------------------------------------------
	if heliosApp.Status.Phase == "" {
		heliosApp.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, &heliosApp); err != nil {
			return ctrl.Result{}, err
		}
	}

	// ------------------------------------------------------------------
	// PHASE 1: Render Manifests (Tekton)
	// ------------------------------------------------------------------
	// Kiểm tra xem có cần chạy PipelineRun mới không (dựa vào generation)
	needsNewPipelineRun := heliosApp.Status.ObservedGeneration != heliosApp.Generation || heliosApp.Status.ManifestPipelineRun == ""

	if needsNewPipelineRun {
		logger.Info("Triggering Tekton PipelineRun for manifest generation", "app", name, "generation", heliosApp.Generation)

		// Use the pipeline name from HeliosApp spec (defaults to "from-code-to-cluster" if not specified)
		manifestPipeline := heliosApp.Spec.PipelineName
		if manifestPipeline == "" {
			manifestPipeline = "from-code-to-cluster" // Default pipeline name
		}
		pipelineRun, err := GeneratePipelineRunForManifestGeneration(&heliosApp, manifestPipeline)
		if err != nil {
			logger.Error(err, "Failed to generate PipelineRun for manifest generation")
			// UPDATE STATUS: Failed
			heliosApp.Status.Phase = "Failed"
			heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
			r.Status().Update(ctx, &heliosApp)
			return ctrl.Result{}, err
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(&heliosApp, pipelineRun, r.Scheme); err != nil {
			logger.Error(err, "Failed to set owner reference for PipelineRun")
			// UPDATE STATUS: Failed
			heliosApp.Status.Phase = "Failed"
			heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
			r.Status().Update(ctx, &heliosApp)
			return ctrl.Result{}, err
		}

		// Tạo PipelineRun
		if err := r.Create(ctx, pipelineRun); err != nil {
			logger.Error(err, "Failed to create PipelineRun for manifest generation")
			// UPDATE STATUS: Failed
			heliosApp.Status.Phase = "Failed"
			heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
			r.Status().Update(ctx, &heliosApp)
			return ctrl.Result{}, err
		}

		logger.Info("Successfully created PipelineRun for manifest generation", "PipelineRun", pipelineRun.GetName())

		// Cập nhật status
		heliosApp.Status.ManifestPipelineRun = pipelineRun.GetName()
		heliosApp.Status.ObservedGeneration = heliosApp.Generation
		heliosApp.Status.Phase = "ManifestGeneration"
		heliosApp.Status.Message = fmt.Sprintf("PipelineRun %s is generating manifest in GitOps repository", pipelineRun.GetName())

		// Cập nhật Condition
		meta.SetStatusCondition(&heliosApp.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "ManifestGenerationInProgress",
			Message:            fmt.Sprintf("PipelineRun %s is generating manifest in GitOps repository", pipelineRun.GetName()),
			ObservedGeneration: heliosApp.Generation,
		})

		if err := r.Status().Update(ctx, &heliosApp); err != nil {
			logger.Error(err, "Failed to update HeliosApp status after triggering PipelineRun")
			return ctrl.Result{}, err
		}

		// Requeue sau 15 giây để kiểm tra kết quả
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	// ========================================================================
	// Kiểm tra trạng thái PipelineRun
	// ========================================================================

	if heliosApp.Status.ManifestPipelineRun != "" {
		pipelineRun := &unstructured.Unstructured{}
		pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "tekton.dev",
			Version: "v1beta1",
			Kind:    "PipelineRun",
		})

		err := r.Get(ctx, types.NamespacedName{
			Name:      heliosApp.Status.ManifestPipelineRun,
			Namespace: namespace,
		}, pipelineRun)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("PipelineRun not found, may have been deleted", "PipelineRun", heliosApp.Status.ManifestPipelineRun)
				// Reset ManifestPipelineRun to trigger a new one next reconcile
				heliosApp.Status.ManifestPipelineRun = ""
				heliosApp.Status.Phase = "ManifestGenerationFailed"
				heliosApp.Status.Message = "PipelineRun not found, triggering new manifest generation."
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{Requeue: true}, nil
			} else {
				logger.Error(err, "Failed to get PipelineRun")
				// UPDATE STATUS: Failed
				heliosApp.Status.Phase = "Failed"
				heliosApp.Status.Message = fmt.Sprintf("Tekton Error: %v", err)
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{}, err
			}
		} else {
			// Đọc status của PipelineRun
			status, found, err := unstructured.NestedMap(pipelineRun.Object, "status")
			if err != nil || !found {
				logger.Info("PipelineRun status not available yet", "PipelineRun", pipelineRun.GetName())
				heliosApp.Status.Phase = "ManifestGeneration"
				heliosApp.Status.Message = fmt.Sprintf("PipelineRun %s status not available yet.", pipelineRun.GetName())
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}

			conditions, found, err := unstructured.NestedSlice(status, "conditions")
			if err != nil || !found || len(conditions) == 0 {
				logger.Info("PipelineRun conditions not available yet", "PipelineRun", pipelineRun.GetName())
				heliosApp.Status.Phase = "ManifestGeneration"
				heliosApp.Status.Message = fmt.Sprintf("PipelineRun %s conditions not available yet.", pipelineRun.GetName())
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}

			// Lấy condition cuối cùng (thường là condition mới nhất)
			lastCondition := conditions[len(conditions)-1].(map[string]any)
			conditionStatus := lastCondition["status"].(string)
			conditionType := lastCondition["type"].(string)
			reason := lastCondition["reason"].(string)

			logger.Info("PipelineRun status check",
				"PipelineRun", pipelineRun.GetName(),
				"type", conditionType,
				"status", conditionStatus,
				"reason", reason)

			// Nếu PipelineRun đang chạy
			if conditionStatus == "Unknown" {
				logger.Info("PipelineRun is still running", "PipelineRun", pipelineRun.GetName())
				heliosApp.Status.Phase = "ManifestGeneration"
				heliosApp.Status.Message = fmt.Sprintf("PipelineRun %s is running. Reason: %s", pipelineRun.GetName(), reason)
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}

			// Nếu PipelineRun thất bại
			if conditionStatus == "False" {
				logger.Error(fmt.Errorf("PipelineRun failed"), "PipelineRun failed", "PipelineRun", pipelineRun.GetName(), "reason", reason)

				meta.SetStatusCondition(&heliosApp.Status.Conditions, metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionFalse,
					Reason:             "ManifestGenerationFailed",
					Message:            fmt.Sprintf("PipelineRun %s failed: %s", pipelineRun.GetName(), reason),
					ObservedGeneration: heliosApp.Generation,
				})
				heliosApp.Status.Phase = "Failed"
				heliosApp.Status.Message = fmt.Sprintf("Manifest generation failed: %s", reason)

				if err := r.Status().Update(ctx, &heliosApp); err != nil {
					logger.Error(err, "Failed to update status after PipelineRun failure")
				}

				return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
			}

			// Nếu PipelineRun thành công, chuyển sang Giai đoạn 2
			if conditionStatus == "True" && conditionType == "Succeeded" {
				logger.Info("PipelineRun succeeded! Manifest has been generated in GitOps repo", "PipelineRun", pipelineRun.GetName())

				// ------------------------------------------------------------------
				// PHASE 2: Register ArgoCD Application
				// ------------------------------------------------------------------

				logger.Info("Manifest generation complete. Ensuring ArgoCD Application is up to date.", "app", name)

				argoApp, err := GenerateArgoApplication(&heliosApp)
				if err != nil {
					logger.Error(err, "Failed to ensure ArgoCD Application")
					// UPDATE STATUS: Failed
					heliosApp.Status.Phase = "Failed"
					heliosApp.Status.Message = fmt.Sprintf("ArgoCD Error: %v", err)
					r.Status().Update(ctx, &heliosApp)
					return ctrl.Result{}, err
				}

				// Lưu ý: ArgoCD Application được tạo trong namespace argocd, không set owner reference
				// vì owner phải cùng namespace

				// Kiểm tra xem Application đã tồn tại chưa
				existing := &unstructured.Unstructured{}
				existing.SetGroupVersionKind(argoApp.GroupVersionKind())
				err = r.Get(ctx, client.ObjectKey{
					Namespace: argoApp.GetNamespace(),
					Name:      argoApp.GetName(),
				}, existing)

				if err != nil {
					if errors.IsNotFound(err) {
						if err := r.Create(ctx, argoApp); err != nil {
							logger.Error(err, "Failed to create ArgoCD Application")
							// UPDATE STATUS: Failed
							heliosApp.Status.Phase = "Failed"
							heliosApp.Status.Message = fmt.Sprintf("ArgoCD Error: %v", err)
							r.Status().Update(ctx, &heliosApp)
							return ctrl.Result{}, err
						}
						logger.Info("Successfully created ArgoCD Application", "Application", argoApp.GetName())

						// Cập nhật status
						heliosApp.Status.ArgoApplication = argoApp.GetName()
						heliosApp.Status.Phase = "ArgoCDRegistration"
						heliosApp.Status.Message = fmt.Sprintf("ArgoCD Application %s created.", argoApp.GetName())
					} else {
						logger.Error(err, "Failed to check ArgoCD Application")
						// UPDATE STATUS: Failed
						heliosApp.Status.Phase = "Failed"
						heliosApp.Status.Message = fmt.Sprintf("ArgoCD Error: %v", err)
						r.Status().Update(ctx, &heliosApp)
						return ctrl.Result{}, err
					}
				} else {
					// Application đã tồn tại, cần update nếu có thay đổi
					logger.Info("ArgoCD Application already exists, checking if update is needed", "Application", argoApp.GetName())

					// So sánh spec của ArgoCD Application - marshal trực tiếp từ Object thay vì dùng NestedMap
					// để tránh panic "cannot deep copy []string"
					existingSpecRaw, existingOk := existing.Object["spec"]
					newSpecRaw, newOk := argoApp.Object["spec"]

					if existingOk && newOk {
						existingSpecJSON, _ := json.Marshal(existingSpecRaw)
						newSpecJSON, _ := json.Marshal(newSpecRaw)

						if string(existingSpecJSON) != string(newSpecJSON) {
							logger.Info("Updating ArgoCD Application spec", "Application", argoApp.GetName())
							argoApp.SetResourceVersion(existing.GetResourceVersion())
							if err := r.Update(ctx, argoApp); err != nil {
								logger.Error(err, "Failed to update ArgoCD Application")
								// UPDATE STATUS: Failed
								heliosApp.Status.Phase = "Failed"
								heliosApp.Status.Message = fmt.Sprintf("ArgoCD Error: %v", err)
								r.Status().Update(ctx, &heliosApp)
								return ctrl.Result{}, err
							}
							logger.Info("Successfully updated ArgoCD Application", "Application", argoApp.GetName())
							heliosApp.Status.Phase = "ArgoCDRegistration"
							heliosApp.Status.Message = fmt.Sprintf("ArgoCD Application %s updated.", argoApp.GetName())
						}
					}

					// Trigger refresh để ArgoCD nhận manifest mới từ GitOps repo
					logger.Info("Triggering ArgoCD refresh to pull latest manifests", "Application", argoApp.GetName())
					if err := r.refreshArgoApplication(ctx, argoApp.GetName(), "argocd"); err != nil {
						logger.Error(err, "Failed to trigger ArgoCD refresh")
						// Không return error, vì refresh sẽ tự động diễn ra theo chu kỳ của ArgoCD
					}
				}

				// Update status của HeliosApp để lưu tên ArgoCD Application
				if heliosApp.Status.ArgoApplication == "" {
					heliosApp.Status.ArgoApplication = argoApp.GetName()
					if err := r.Status().Update(ctx, &heliosApp); err != nil {
						logger.Error(err, "Failed to update status with ArgoCD Application name")
						return ctrl.Result{}, err
					}
				}

				logger.Info("ArgoCD Application is ready, will check its sync status next")
				// Không return ở đây, để fall through sang GIAI ĐOẠN 3 để check ArgoCD status ngay
			}
		}
	}

	// ========================================================================
	// GIAI ĐOẠN 3: Đồng bộ trạng thái từ ArgoCD Application
	// ========================================================================

	var syncStatus, healthStatus string
	if heliosApp.Status.ArgoApplication != "" {
		argoApp := &unstructured.Unstructured{}
		argoApp.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "argoproj.io",
			Version: "v1alpha1",
			Kind:    "Application",
		})

		err := r.Get(ctx, types.NamespacedName{
			Name:      heliosApp.Status.ArgoApplication,
			Namespace: "argocd",
		}, argoApp)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("ArgoCD Application not found", "Application", heliosApp.Status.ArgoApplication)
				heliosApp.Status.Phase = "ArgoCDAppNotFound"
				heliosApp.Status.Message = fmt.Sprintf("ArgoCD Application %s not found. Recreating.", heliosApp.Status.ArgoApplication)
				heliosApp.Status.ArgoApplication = "" // Reset to trigger recreation
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{Requeue: true}, nil
			} else {
				logger.Error(err, "Failed to get ArgoCD Application")
				// UPDATE STATUS: Failed
				heliosApp.Status.Phase = "Failed"
				heliosApp.Status.Message = fmt.Sprintf("ArgoCD Error: %v", err)
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{}, err
			}
		} else {
			// Đọc status từ ArgoCD Application - dùng type assertion thay vì NestedMap để tránh panic
			statusRaw, statusExists := argoApp.Object["status"]
			if !statusExists {
				logger.Info("ArgoCD Application status not available yet")
				heliosApp.Status.Phase = "ArgoCDSyncing"
				heliosApp.Status.Message = fmt.Sprintf("ArgoCD Application %s status not available yet.", argoApp.GetName())
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}

			status, ok := statusRaw.(map[string]interface{})
			if !ok {
				logger.Info("ArgoCD Application status format unexpected")
				heliosApp.Status.Phase = "ArgoCDSyncing"
				heliosApp.Status.Message = fmt.Sprintf("ArgoCD Application %s status format unexpected.", argoApp.GetName())
				r.Status().Update(ctx, &heliosApp)
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}

			// Lấy sync status
			syncStatus, _, _ = unstructured.NestedString(status, "sync", "status")
			healthStatus, _, _ = unstructured.NestedString(status, "health", "status")

			logger.Info("ArgoCD Application status",
				"Application", argoApp.GetName(),
				"syncStatus", syncStatus,
				"healthStatus", healthStatus)

			// Lấy deployed version (image từ summary hoặc revision)
			images, found, _ := unstructured.NestedSlice(status, "summary", "images")
			if found && len(images) > 0 {
				heliosApp.Status.DeployedVersion = images[0].(string)
			}

			// Cập nhật condition dựa trên trạng thái ArgoCD
			if syncStatus == "Synced" && healthStatus == "Healthy" {
				meta.SetStatusCondition(&heliosApp.Status.Conditions, metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             "SyncedAndHealthy",
					Message:            fmt.Sprintf("Application is synced and healthy. Deployed version: %s", heliosApp.Status.DeployedVersion),
					ObservedGeneration: heliosApp.Generation,
				})

				logger.Info("ArgoCD application is synced and healthy. Reconciliation complete.",
					"Application", argoApp.GetName(),
					"version", heliosApp.Status.DeployedVersion)
			} else {
				// Nếu ArgoCD chưa Synced, có thể cần trigger refresh để pull manifest mới
				if syncStatus == "OutOfSync" {
					logger.Info("ArgoCD Application is OutOfSync, triggering refresh", "Application", argoApp.GetName())
					if err := r.refreshArgoApplication(ctx, argoApp.GetName(), "argocd"); err != nil {
						logger.Error(err, "Failed to trigger ArgoCD refresh for OutOfSync app")
					}
				}

				meta.SetStatusCondition(&heliosApp.Status.Conditions, metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionFalse,
					Reason:             "Syncing",
					Message:            fmt.Sprintf("ArgoCD sync: %s, health: %s", syncStatus, healthStatus),
					ObservedGeneration: heliosApp.Generation,
				})
			}

			if err := r.Status().Update(ctx, &heliosApp); err != nil {
				logger.Error(err, "Failed to update status from ArgoCD")
				return ctrl.Result{}, err
			}

			// Tiếp tục theo dõi nếu chưa Synced và Healthy
			if syncStatus != "Synced" || healthStatus != "Healthy" {
				return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
			}
		}
	}

	// Helper to update status safely
	updateStatus := func(phase, msg string) {
		heliosApp.Status.Phase = phase
		heliosApp.Status.Message = msg
		// We ignore error here to not block flow, mostly just for visibility
		if err := r.Status().Update(ctx, &heliosApp); err != nil {
			logger.Error(err, "Failed to update HeliosApp status in updateStatus helper")
		}
	}

	// Update Healthy Status if we reached here
	isHealthy := meta.IsStatusConditionTrue(heliosApp.Status.Conditions, "Ready")
	if isHealthy {
		updateStatus("Healthy", "App is running and synced")
	} else {
		updateStatus("Syncing", fmt.Sprintf("ArgoCD Status: %s/%s", syncStatus, healthStatus))
	}

	// ------------------------------------------------------------------
	// PHASE 4: Helios Smart Rightsizer (Auto-Optimization)
	// ------------------------------------------------------------------
	// Đây là tính năng "Killer" của đồ án: Tự động tối ưu tài nguyên

	// Chỉ chạy tối ưu khi App đã Healthy và User đã BẬT tính năng này
	isOptIn := heliosApp.Spec.EnableAutoOptimization

	if isHealthy && isOptIn {
		logger.Info("Auto-Optimization Condition Met", "isHealthy", isHealthy, "isOptIn", isOptIn)
		optimizer := &OptimizerService{}
		analysis := optimizer.AnalyzeResourceUsage(ctx, r.Client, &heliosApp)

		if analysis.IsWasteful {
			logger.Info("DETECTED RESOURCE WASTE",
				"app", heliosApp.Name,
				"current", analysis.CurrentCpu,
				"suggested", analysis.SuggestedCpu,
				"waste_percent", analysis.WastePercentage)

			// Execute Auto-Fix (Closed Loop GitOps)
			err := optimizer.AutoFixRepository(ctx, &heliosApp, analysis)
			if err != nil {
				logger.Error(err, "Failed to auto-optimize repository")
				updateStatus("OptimizationFailed", fmt.Sprintf("Could not create PR: %v", err))
			} else {
				// Success
				r.EventRecorder.Event(&heliosApp, corev1.EventTypeNormal, "AutoOptimized",
					fmt.Sprintf("Reduced CPU from %s to %s via GitOps", analysis.CurrentCpu, analysis.SuggestedCpu))
				updateStatus("Optimized", fmt.Sprintf("Waste detected (%d%%). PR created to fix.", analysis.WastePercentage))
			}
			if err != nil {
				logger.Error(err, "Failed to auto-optimize repository")
				// Không return error để tránh crash loop reconcile chính, chỉ log error
			} else {
				// Nếu fix thành công, update Event để thông báo cho user
				r.EventRecorder.Event(&heliosApp, corev1.EventTypeNormal, "AutoOptimized",
					fmt.Sprintf("Reduced CPU from %s to %s via GitOps", analysis.CurrentCpu, analysis.SuggestedCpu))
			}
		}
	}

	logger.Info("Reconciliation loop completed successfully for HeliosApp", "name", name)
	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
}

// Hàm so sánh spec của hai unstructured (chỉ so sánh phần spec)
// Sử dụng JSON marshal toàn bộ object để tránh panic với nested structures
func equalUnstructured(a, b *unstructured.Unstructured) bool {
	// Marshal toàn bộ object thay vì chỉ spec để tránh deep copy issues
	jsonA, errA := json.Marshal(a.Object)
	jsonB, errB := json.Marshal(b.Object)
	if errA != nil || errB != nil {
		// Nếu không marshal được, coi như khác nhau để update
		return false
	}

	return string(jsonA) == string(jsonB)
}

// refreshArgoApplication triggers a refresh operation on the ArgoCD Application
// This forces ArgoCD to pull the latest manifests from the GitOps repository
func (r *HeliosAppReconciler) refreshArgoApplication(ctx context.Context, appName, namespace string) error {
	logger := log.FromContext(ctx)

	// Get the ArgoCD Application
	argoApp := &unstructured.Unstructured{}
	argoApp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	})

	err := r.Get(ctx, types.NamespacedName{
		Name:      appName,
		Namespace: namespace,
	}, argoApp)

	if err != nil {
		return err
	}

	// Add refresh annotation to trigger ArgoCD to pull latest changes
	annotations := argoApp.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Use current timestamp to ensure annotation changes
	annotations["argocd.argoproj.io/refresh"] = "hard"
	annotations["helios.io/refreshed-at"] = time.Now().Format(time.RFC3339)
	argoApp.SetAnnotations(annotations)

	if err := r.Update(ctx, argoApp); err != nil {
		return err
	}

	logger.Info("Triggered ArgoCD refresh via annotation", "Application", appName)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HeliosAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heliosappv1.HeliosApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

// Note: Deployment and Service creation is handled by ArgoCD via GitOps manifests.
// If you need the operator to create K8s resources directly, reintroduce helper
// methods and call them from the reconciliation loop.
