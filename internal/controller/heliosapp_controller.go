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
)

// HeliosAppReconciler reconciles a HeliosApp object
type HeliosAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
		logger.Error(err, "Failed to generate EventListener")
		return ctrl.Result{}, err
	}
	triggerBinding, err := GenerateTriggerBinding(name+"-trigger-binding", namespace)
	if err != nil {
		logger.Error(err, "Failed to generate TriggerBinding")
		return ctrl.Result{}, err
	}
	triggerTemplate, err := GenerateTriggerTemplate(
		name+"-trigger-template", namespace, name+"-pipelinerun", pipelineName, serviceAccount, workspace,
	)
	if err != nil {
		logger.Error(err, "Failed to generate TriggerTemplate")
		return ctrl.Result{}, err
	}

	for _, obj := range []*unstructured.Unstructured{defaultsBinding, eventListener, triggerBinding, triggerTemplate} {
		obj.SetNamespace(namespace)
		if err := controllerutil.SetControllerReference(&heliosApp, obj, r.Scheme); err != nil {
			logger.Error(err, "Failed to set owner reference", "name", obj.GetName())
			return ctrl.Result{}, err
		}

		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(obj.GroupVersionKind())
		err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: obj.GetName()}, existing)
		if err != nil {
			if client.IgnoreNotFound(err) == nil {
				if err := r.Create(ctx, obj); err != nil {
					logger.Error(err, "Failed to create Tekton Trigger resource", "name", obj.GetName())
					return ctrl.Result{}, err
				}
				logger.Info("Created Tekton Trigger resource", "name", obj.GetName())
			} else {
				logger.Error(err, "Failed to get Tekton Trigger resource", "name", obj.GetName())
				return ctrl.Result{}, err
			}
		} else {
			if !equalUnstructured(obj, existing) {
				obj.SetResourceVersion(existing.GetResourceVersion())
				if err := r.Update(ctx, obj); err != nil {
					logger.Error(err, "Failed to update Tekton Trigger resource", "name", obj.GetName())
					return ctrl.Result{}, err
				}
				logger.Info("Updated Tekton Trigger resource", "name", obj.GetName())
			}
		}
	}

	// ========================================================================
	// GIAI ĐOẠN 1: Kích hoạt Tekton PipelineRun để sinh Manifest
	// ========================================================================

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
			return ctrl.Result{}, err
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(&heliosApp, pipelineRun, r.Scheme); err != nil {
			logger.Error(err, "Failed to set owner reference for PipelineRun")
			return ctrl.Result{}, err
		}

		// Tạo PipelineRun
		if err := r.Create(ctx, pipelineRun); err != nil {
			logger.Error(err, "Failed to create PipelineRun for manifest generation")
			return ctrl.Result{}, err
		}

		logger.Info("Successfully created PipelineRun for manifest generation", "PipelineRun", pipelineRun.GetName())

		// Cập nhật status
		heliosApp.Status.ManifestPipelineRun = pipelineRun.GetName()
		heliosApp.Status.ObservedGeneration = heliosApp.Generation

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
			} else {
				logger.Error(err, "Failed to get PipelineRun")
				return ctrl.Result{}, err
			}
		} else {
			// Đọc status của PipelineRun
			status, found, err := unstructured.NestedMap(pipelineRun.Object, "status")
			if err != nil || !found {
				logger.Info("PipelineRun status not available yet", "PipelineRun", pipelineRun.GetName())
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}

			conditions, found, err := unstructured.NestedSlice(status, "conditions")
			if err != nil || !found || len(conditions) == 0 {
				logger.Info("PipelineRun conditions not available yet", "PipelineRun", pipelineRun.GetName())
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

				if err := r.Status().Update(ctx, &heliosApp); err != nil {
					logger.Error(err, "Failed to update status after PipelineRun failure")
				}

				return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
			}

			// Nếu PipelineRun thành công, chuyển sang Giai đoạn 2
			if conditionStatus == "True" && conditionType == "Succeeded" {
				logger.Info("PipelineRun succeeded! Manifest has been generated in GitOps repo", "PipelineRun", pipelineRun.GetName())

				// ========================================================================
				// GIAI ĐOẠN 2: Tạo ArgoCD Application
				// ========================================================================

				// Kiểm tra xem ArgoCD Application đã tồn tại chưa
				if heliosApp.Status.ArgoApplication == "" {
					logger.Info("Manifest generation complete. Creating ArgoCD Application.", "app", name)

					argoApp, err := GenerateArgoApplication(&heliosApp)
					if err != nil {
						logger.Error(err, "Failed to generate ArgoCD Application")
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
								return ctrl.Result{}, err
							}
							logger.Info("Successfully created ArgoCD Application", "Application", argoApp.GetName())
						} else {
							logger.Error(err, "Failed to check ArgoCD Application")
							return ctrl.Result{}, err
						}
					}

					// Cập nhật status
					heliosApp.Status.ArgoApplication = argoApp.GetName()

					meta.SetStatusCondition(&heliosApp.Status.Conditions, metav1.Condition{
						Type:               "Ready",
						Status:             metav1.ConditionFalse,
						Reason:             "DeployingWithArgoCD",
						Message:            fmt.Sprintf("ArgoCD Application %s created. Starting deployment...", argoApp.GetName()),
						ObservedGeneration: heliosApp.Generation,
					})

					if err := r.Status().Update(ctx, &heliosApp); err != nil {
						logger.Error(err, "Failed to update status after creating ArgoCD Application")
						return ctrl.Result{}, err
					}

					return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
				}
			}
		}
	}

	// ========================================================================
	// GIAI ĐOẠN 3: Đồng bộ trạng thái từ ArgoCD Application
	// ========================================================================

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
			} else {
				logger.Error(err, "Failed to get ArgoCD Application")
				return ctrl.Result{}, err
			}
		} else {
			// Đọc status từ ArgoCD Application
			status, found, err := unstructured.NestedMap(argoApp.Object, "status")
			if err != nil || !found {
				logger.Info("ArgoCD Application status not available yet")
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}

			// Lấy sync status
			syncStatus, _, _ := unstructured.NestedString(status, "sync", "status")
			healthStatus, _, _ := unstructured.NestedString(status, "health", "status")

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

	logger.Info("Reconciliation loop completed successfully for HeliosApp", "name", name)
	return ctrl.Result{}, nil
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
