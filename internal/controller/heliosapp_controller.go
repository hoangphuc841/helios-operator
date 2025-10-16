/*
Copyright 2025.
*/

package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"github.com/hoangphuc841/helios-operator/internal/common"
	"github.com/hoangphuc841/helios-operator/internal/resources"
)

// HeliosAppReconciler reconciles a HeliosApp object
type HeliosAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=triggers.tekton.dev,resources=eventlisteners;triggerbindings;triggertemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=applications/status,verbs=get
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// fetchHeliosApp retrieves the HeliosApp resource from the cluster
func (r *HeliosAppReconciler) fetchHeliosApp(ctx context.Context, namespacedName types.NamespacedName) (*heliosappv1.HeliosApp, error) {
	var heliosApp heliosappv1.HeliosApp
	if err := r.Get(ctx, namespacedName, &heliosApp); err != nil {
		return nil, err
	}
	return &heliosApp, nil
}

func (r *HeliosAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Track reconciliation metrics
	startTime := time.Now()
	ReconciliationsTotal.WithLabelValues(req.Namespace, req.Name).Inc()

	var reconcileResult string
	defer func() {
		duration := time.Since(startTime).Seconds()
		ReconciliationDuration.WithLabelValues(req.Namespace, req.Name, reconcileResult).Observe(duration)
		LastReconcileTime.WithLabelValues(req.Namespace, req.Name).SetToCurrentTime()
	}()

	// Fetch HeliosApp
	heliosApp, err := r.fetchHeliosApp(ctx, req.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("HeliosApp not found. It may have been deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch HeliosApp")
		reconcileResult = "error"
		return ctrl.Result{}, err
	}
	if heliosApp == nil {
		// Already handled (not found)
		return ctrl.Result{}, nil
	}

	logger.Info("Starting reconciliation", "name", heliosApp.Name, "namespace", heliosApp.Namespace, "generation", heliosApp.Generation)

	name := heliosApp.Name
	namespace := heliosApp.Namespace

	// PVC name động: lấy từ spec hoặc quy ước
	pvcName := heliosApp.Spec.PVCName
	if pvcName == "" {
		pvcName = "pvc-" + name
	}

	// Generate pipeline name from app name
	pipelineName := resources.GetPipelineName(name)
	serviceAccount := heliosApp.Spec.ServiceAccount
	githubSecret := heliosApp.Spec.WebhookSecret
	workspace := map[string]interface{}{
		"name": "shared-data",
		"persistentVolumeClaim": map[string]interface{}{
			"claimName": pvcName,
		},
	}

	// Reconcile Tekton Pipeline
	if err := r.reconcilePipeline(ctx, heliosApp, logger); err != nil {
		reconcileResult = "error"
		return ctrl.Result{}, err
	}

	// Reconcile Tekton Triggers
	if err := r.reconcileTriggers(ctx, heliosApp, name, namespace, pipelineName, serviceAccount, githubSecret, workspace, logger); err != nil {
		reconcileResult = "error"
		return ctrl.Result{}, err
	}

	// Reconcile ArgoCD Application
	if err := r.reconcileArgoCD(ctx, heliosApp, name, namespace, logger); err != nil {
		reconcileResult = "error"
		return ctrl.Result{}, err
	}

	// Update comprehensive status
	if err := r.updateComprehensiveStatus(ctx, heliosApp, name, namespace, logger); err != nil {
		logger.Error(err, "failed to update comprehensive status")
		reconcileResult = "error"
		return ctrl.Result{}, common.NewStatusUpdateError("HeliosApp", name, err)
	}

	reconcileResult = "success"
	logger.Info("Reconciliation completed successfully", "name", name, "namespace", namespace)
	return ctrl.Result{}, nil
}

// Hàm so sánh spec của hai unstructured (chỉ so sánh phần spec)
func equalUnstructured(a, b *unstructured.Unstructured) bool {
	// Compare spec using JSON marshaling to avoid deep copy issues
	specA, okA := a.Object["spec"]
	specB, okB := b.Object["spec"]
	if !okA || !okB {
		return false
	}
	return reflect.DeepEqual(specA, specB)
}

// updateStatus updates the HeliosApp status with the given condition
func (r *HeliosAppReconciler) updateStatus(ctx context.Context, heliosApp *heliosappv1.HeliosApp, conditionType string, status metav1.ConditionStatus, reason, message string) error {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: heliosApp.Generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or append new one
	found := false
	for i, existingCondition := range heliosApp.Status.Conditions {
		if existingCondition.Type == conditionType {
			heliosApp.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		heliosApp.Status.Conditions = append(heliosApp.Status.Conditions, condition)
	}

	return r.Status().Update(ctx, heliosApp)
}

// getArgoAppSyncStatus fetches the sync status from ArgoCD Application
func (r *HeliosAppReconciler) getArgoAppSyncStatus(ctx context.Context, name string) (syncStatus string, healthStatus string, err error) {
	argoApp := &unstructured.Unstructured{}
	argoApp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	})

	err = r.Get(ctx, client.ObjectKey{Namespace: "argocd", Name: name + "-argocd"}, argoApp)
	if err != nil {
		return "", "", err
	}

	// Extract status.sync.status
	syncStatus, _, _ = unstructured.NestedString(argoApp.Object, "status", "sync", "status")
	// Extract status.health.status
	healthStatus, _, _ = unstructured.NestedString(argoApp.Object, "status", "health", "status")

	return syncStatus, healthStatus, nil
}

// getPipelineRunStatus retrieves the most recent PipelineRun status for a HeliosApp
func (r *HeliosAppReconciler) getPipelineRunStatus(ctx context.Context, heliosApp *heliosappv1.HeliosApp) (buildStatus, buildVersion, pipelineRunName string, lastBuildTime *metav1.Time, err error) {
	// List PipelineRuns with helios.io/app-name label
	pipelineRunList := &unstructured.UnstructuredList{}
	pipelineRunList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "tekton.dev",
		Version: "v1",
		Kind:    "PipelineRunList",
	})

	listOpts := []client.ListOption{
		client.InNamespace(heliosApp.Namespace),
		client.MatchingLabels{
			"helios.io/app-name": heliosApp.Name,
		},
	}

	err = r.List(ctx, pipelineRunList, listOpts...)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to list PipelineRuns: %w", err)
	}

	if len(pipelineRunList.Items) == 0 {
		// No PipelineRuns found yet
		return "Unknown", "", "", nil, nil
	}

	// Find the most recent PipelineRun by creation timestamp
	var mostRecent *unstructured.Unstructured
	var mostRecentTime metav1.Time

	for i := range pipelineRunList.Items {
		pr := &pipelineRunList.Items[i]
		creationTime := pr.GetCreationTimestamp()

		if mostRecent == nil || creationTime.After(mostRecentTime.Time) {
			mostRecent = pr
			mostRecentTime = creationTime
		}
	}

	if mostRecent == nil {
		return "Unknown", "", "", nil, nil
	}

	pipelineRunName = mostRecent.GetName()
	lastBuildTime = &mostRecentTime

	// Extract status from PipelineRun
	// status.conditions[?(@.type=="Succeeded")].status
	conditions, found, _ := unstructured.NestedSlice(mostRecent.Object, "status", "conditions")
	if !found || len(conditions) == 0 {
		return "Running", "", pipelineRunName, lastBuildTime, nil
	}

	// Find Succeeded condition
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		if condType != "Succeeded" {
			continue
		}

		condStatus, _, _ := unstructured.NestedString(condMap, "status")
		reason, _, _ := unstructured.NestedString(condMap, "reason")

		switch condStatus {
		case "True":
			buildStatus = "Succeeded"
		case "False":
			buildStatus = "Failed"
		case "Unknown":
			buildStatus = "Running"
		default:
			buildStatus = reason
		}

		// Try to extract image tag from PipelineRun params or results
		// This depends on your pipeline structure
		// Example: check taskRuns for image result
		results, found, _ := unstructured.NestedSlice(mostRecent.Object, "status", "results")
		if found {
			for _, result := range results {
				resultMap, ok := result.(map[string]interface{})
				if !ok {
					continue
				}

				name, _, _ := unstructured.NestedString(resultMap, "name")
				if name == "IMAGE_URL" || name == "IMAGE_TAG" || name == "IMAGE" {
					buildVersion, _, _ = unstructured.NestedString(resultMap, "value")
					break
				}
			}
		}

		break
	}

	if buildStatus == "" {
		buildStatus = "Unknown"
	}

	return buildStatus, buildVersion, pipelineRunName, lastBuildTime, nil
}

// getDeploymentHealth retrieves deployment health status
func (r *HeliosAppReconciler) getDeploymentHealth(ctx context.Context, heliosApp *heliosappv1.HeliosApp) (health string, readyReplicas, desiredReplicas int32, lastHealthyTime *metav1.Time, err error) {
	// Try to find deployment with same name as HeliosApp
	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})

	// First try exact name match
	err = r.Get(ctx, client.ObjectKey{
		Namespace: heliosApp.Namespace,
		Name:      heliosApp.Name,
	}, deployment)

	if err != nil {
		// If not found, try listing deployments with app label
		if errors.IsNotFound(err) {
			deploymentList := &unstructured.UnstructuredList{}
			deploymentList.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "DeploymentList",
			})

			listOpts := []client.ListOption{
				client.InNamespace(heliosApp.Namespace),
				client.MatchingLabels{
					"app": heliosApp.Name,
				},
			}

			err = r.List(ctx, deploymentList, listOpts...)
			if err != nil {
				return "Unknown", 0, 0, nil, fmt.Errorf("failed to list Deployments: %w", err)
			}

			if len(deploymentList.Items) == 0 {
				// Deployment not created yet (ArgoCD hasn't synced)
				return "Unknown", 0, 0, nil, nil
			}

			// Use first matching deployment
			deployment = &deploymentList.Items[0]
		} else {
			return "Unknown", 0, 0, nil, fmt.Errorf("failed to get Deployment: %w", err)
		}
	}

	// Extract replica counts
	specReplicas, found, _ := unstructured.NestedInt64(deployment.Object, "spec", "replicas")
	if found {
		desiredReplicas = int32(specReplicas)
	}

	readyReplicasInt64, found, _ := unstructured.NestedInt64(deployment.Object, "status", "readyReplicas")
	if found {
		readyReplicas = int32(readyReplicasInt64)
	}

	availableReplicas, _, _ := unstructured.NestedInt64(deployment.Object, "status", "availableReplicas")

	// Determine health status
	if readyReplicas == desiredReplicas && desiredReplicas > 0 && availableReplicas == int64(desiredReplicas) {
		health = "Healthy"
		now := metav1.Now()
		lastHealthyTime = &now
	} else if readyReplicas > 0 && readyReplicas < desiredReplicas {
		health = "Progressing"
	} else if readyReplicas == 0 {
		health = "Degraded"
	} else {
		health = "Unknown"
	}

	// Check deployment conditions for more accurate health
	conditions, found, _ := unstructured.NestedSlice(deployment.Object, "status", "conditions")
	if found {
		for _, cond := range conditions {
			condMap, ok := cond.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _, _ := unstructured.NestedString(condMap, "type")
			condStatus, _, _ := unstructured.NestedString(condMap, "status")

			if condType == "Available" && condStatus == "True" {
				health = "Healthy"
				now := metav1.Now()
				lastHealthyTime = &now
			} else if condType == "Progressing" && condStatus == "True" {
				if health != "Healthy" {
					health = "Progressing"
				}
			}
		}
	}

	return health, readyReplicas, desiredReplicas, lastHealthyTime, nil
}

// argoAppPredicate creates a predicate that filters only ArgoCD Applications
// created by Helios Operator (identified by label or naming convention)
func (r *HeliosAppReconciler) argoAppPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.isHeliosArgoApp(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.isHeliosArgoApp(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.isHeliosArgoApp(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return r.isHeliosArgoApp(e.Object)
		},
	}
}

// isHeliosArgoApp checks if the object is an ArgoCD Application managed by Helios
func (r *HeliosAppReconciler) isHeliosArgoApp(obj client.Object) bool {
	// Check if it's an ArgoCD Application
	if obj.GetObjectKind().GroupVersionKind().Group != "argoproj.io" ||
		obj.GetObjectKind().GroupVersionKind().Kind != "Application" {

		// For unstructured objects, check using type assertion
		if u, ok := obj.(*unstructured.Unstructured); ok {
			if u.GetAPIVersion() != "argoproj.io/v1alpha1" || u.GetKind() != "Application" {
				return false
			}
		} else {
			return false
		}
	}

	// Check if it's in the argocd namespace
	if obj.GetNamespace() != "argocd" {
		return false
	}

	// Check if it's managed by Helios (has the label or naming convention)
	labels := obj.GetLabels()
	if labels != nil && labels["helios.io/managed-by"] == "helios-operator" {
		return true
	}

	// Also check naming convention: ends with "-argocd"
	return strings.HasSuffix(obj.GetName(), "-argocd")
}

// argoAppToHeliosApp maps an ArgoCD Application to the corresponding HeliosApp
// This triggers reconciliation of the HeliosApp when its ArgoCD Application changes
func (r *HeliosAppReconciler) argoAppToHeliosApp(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	// Get the ArgoCD Application name
	argoAppName := obj.GetName()

	// Extract HeliosApp name by removing "-argocd" suffix
	if !strings.HasSuffix(argoAppName, "-argocd") {
		logger.V(1).Info("ArgoCD Application does not follow naming convention", "name", argoAppName)
		return []reconcile.Request{}
	}

	heliosAppName := strings.TrimSuffix(argoAppName, "-argocd")

	// Get the target namespace from labels
	labels := obj.GetLabels()
	var heliosAppNamespace string
	if labels != nil {
		heliosAppNamespace = labels["helios.io/app-namespace"]
	}

	// If namespace not found in labels, try to get from ArgoCD Application spec
	if heliosAppNamespace == "" {
		if u, ok := obj.(*unstructured.Unstructured); ok {
			ns, found, _ := unstructured.NestedString(u.Object, "spec", "destination", "namespace")
			if found {
				heliosAppNamespace = ns
			}
		}
	}

	// If still not found, skip
	if heliosAppNamespace == "" {
		logger.V(1).Info("Cannot determine HeliosApp namespace", "argoApp", argoAppName)
		return []reconcile.Request{}
	}

	// Track watch event
	WatchEventsTotal.WithLabelValues("argocd_app").Inc()

	logger.V(1).Info("ArgoCD Application changed, triggering HeliosApp reconciliation",
		"argoApp", argoAppName,
		"heliosApp", heliosAppName,
		"namespace", heliosAppNamespace)

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      heliosAppName,
				Namespace: heliosAppNamespace,
			},
		},
	}
}

// ===================================
// Tekton PipelineRun Watches
// ===================================

// pipelineRunPredicate filters PipelineRuns created by Helios Operator
func (r *HeliosAppReconciler) pipelineRunPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.isHeliosPipelineRun(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Only trigger on status changes, not spec changes
			return r.isHeliosPipelineRun(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.isHeliosPipelineRun(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false // Ignore generic events for PipelineRuns
		},
	}
}

// isHeliosPipelineRun checks if PipelineRun is managed by Helios
func (r *HeliosAppReconciler) isHeliosPipelineRun(obj client.Object) bool {
	labels := obj.GetLabels()
	if labels == nil {
		return false
	}
	// Check if it has Tekton trigger labels AND Helios labels
	_, hasTektonLabel := labels["triggers.tekton.dev/trigger"]
	managedBy, hasHeliosLabel := labels["helios.io/managed-by"]

	return hasTektonLabel && hasHeliosLabel && managedBy == "helios-operator"
}

// pipelineRunToHeliosApp maps a PipelineRun to its owning HeliosApp
func (r *HeliosAppReconciler) pipelineRunToHeliosApp(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	labels := obj.GetLabels()
	if labels == nil {
		return []reconcile.Request{}
	}

	appName, hasName := labels["helios.io/app-name"]
	appNamespace := obj.GetNamespace() // PipelineRun is in same namespace as HeliosApp

	if !hasName || appName == "" {
		logger.V(1).Info("PipelineRun missing helios.io/app-name label", "pipelineRun", obj.GetName())
		return []reconcile.Request{}
	}

	// Track watch event
	WatchEventsTotal.WithLabelValues("pipelinerun").Inc()

	logger.V(1).Info("PipelineRun changed, triggering HeliosApp reconciliation",
		"pipelineRun", obj.GetName(),
		"heliosApp", appName,
		"namespace", appNamespace)

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      appName,
				Namespace: appNamespace,
			},
		},
	}
}

// ===================================
// Deployment Watches
// ===================================

// deploymentPredicate filters Deployments managed by Helios (deployed via ArgoCD)
func (r *HeliosAppReconciler) deploymentPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.isHeliosDeployment(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Only trigger on status changes (replicas, conditions)
			return r.isHeliosDeployment(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.isHeliosDeployment(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// isHeliosDeployment checks if Deployment is managed by Helios
func (r *HeliosAppReconciler) isHeliosDeployment(obj client.Object) bool {
	labels := obj.GetLabels()
	if labels == nil {
		return false
	}

	// Deployments created via ArgoCD will have ArgoCD labels
	// AND we should have added our own label or match by app name
	_, hasArgoLabel := labels["app.kubernetes.io/instance"]
	appName, hasAppLabel := labels["app"]

	// Check if there's a matching HeliosApp with this name
	// We'll use the deployment name or app label as the HeliosApp name
	return (hasArgoLabel || hasAppLabel) && appName != ""
}

// deploymentToHeliosApp maps a Deployment to its HeliosApp
func (r *HeliosAppReconciler) deploymentToHeliosApp(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	// Try to find HeliosApp name from labels or deployment name
	labels := obj.GetLabels()
	var heliosAppName string

	if labels != nil {
		// First try explicit helios label
		if name, ok := labels["helios.io/app-name"]; ok {
			heliosAppName = name
		} else if name, ok := labels["app"]; ok {
			// Fallback to app label
			heliosAppName = name
		} else if instance, ok := labels["app.kubernetes.io/instance"]; ok {
			// Fallback to ArgoCD instance name
			heliosAppName = instance
		}
	}

	if heliosAppName == "" {
		// Last resort: use deployment name
		heliosAppName = obj.GetName()
	}

	namespace := obj.GetNamespace()

	// Verify HeliosApp exists before triggering reconciliation
	heliosApp := &heliosappv1.HeliosApp{}
	err := r.Get(ctx, types.NamespacedName{Name: heliosAppName, Namespace: namespace}, heliosApp)
	if err != nil {
		logger.V(1).Info("Deployment does not match any HeliosApp",
			"deployment", obj.GetName(),
			"guessedApp", heliosAppName,
			"error", err)
		return []reconcile.Request{}
	}

	// Track watch event
	WatchEventsTotal.WithLabelValues("deployment").Inc()

	logger.V(1).Info("Deployment changed, triggering HeliosApp reconciliation",
		"deployment", obj.GetName(),
		"heliosApp", heliosAppName,
		"namespace", namespace)

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      heliosAppName,
				Namespace: namespace,
			},
		},
	}
}

// ===================================
// Helper Methods for Reconciliation
// ===================================

// reconcilePipeline creates or updates the Tekton Pipeline
func (r *HeliosAppReconciler) reconcilePipeline(ctx context.Context, heliosApp *heliosappv1.HeliosApp, logger logr.Logger) error {
	pipeline, err := resources.GeneratePipeline(heliosApp)
	if err != nil {
		logger.Error(err, "failed to generate Pipeline")
		return common.NewResourceGenerationError("Pipeline", heliosApp.Name, err)
	}

	pipeline.SetNamespace(heliosApp.Namespace)
	if err := controllerutil.SetControllerReference(heliosApp, pipeline, r.Scheme); err != nil {
		logger.Error(err, "failed to set owner reference for Pipeline", "name", pipeline.GetName())
		return common.NewReconciliationError("Pipeline", pipeline.GetName(), "set owner reference", err)
	}

	return r.createOrUpdateResource(ctx, pipeline, logger)
}

// reconcileTriggers creates or updates Tekton Trigger resources
func (r *HeliosAppReconciler) reconcileTriggers(ctx context.Context, heliosApp *heliosappv1.HeliosApp, name, namespace, pipelineName, serviceAccount, githubSecret string, workspace map[string]interface{}, logger logr.Logger) error {
	eventListener, err := resources.GenerateEventListener(
		name+"-el", namespace, name+"-trigger", name+"-trigger-binding", name+"-trigger-template", githubSecret,
	)
	if err != nil {
		logger.Error(err, "failed to generate EventListener")
		return common.NewResourceGenerationError("EventListener", name+"-el", err)
	}

	triggerBinding, err := resources.GenerateTriggerBinding(name+"-trigger-binding", namespace)
	if err != nil {
		logger.Error(err, "failed to generate TriggerBinding")
		return common.NewResourceGenerationError("TriggerBinding", name+"-trigger-binding", err)
	}

	triggerTemplate, err := resources.GenerateTriggerTemplate(
		name+"-trigger-template", namespace, name+"-pipelinerun", pipelineName, serviceAccount, workspace,
	)
	if err != nil {
		logger.Error(err, "failed to generate TriggerTemplate")
		return common.NewResourceGenerationError("TriggerTemplate", name+"-trigger-template", err)
	}

	for _, obj := range []*unstructured.Unstructured{eventListener, triggerBinding, triggerTemplate} {
		obj.SetNamespace(namespace)
		if err := controllerutil.SetControllerReference(heliosApp, obj, r.Scheme); err != nil {
			logger.Error(err, "failed to set owner reference", "name", obj.GetName())
			return common.NewReconciliationError(obj.GetKind(), obj.GetName(), "set owner reference", err)
		}

		if err := r.createOrUpdateResource(ctx, obj, logger); err != nil {
			return err
		}
	}

	return nil
}

// reconcileArgoCD creates or updates the ArgoCD Application
func (r *HeliosAppReconciler) reconcileArgoCD(ctx context.Context, heliosApp *heliosappv1.HeliosApp, name, namespace string, logger logr.Logger) error {
	argoApp, err := resources.GenerateArgoApplication(heliosApp)
	if err != nil {
		logger.Error(err, "failed to generate ArgoCD Application")
		return common.NewResourceGenerationError("ArgoCD Application", name+"-argocd", err)
	}

	// Set namespace explicitly (ArgoCD Application lives in argocd namespace)
	argoApp.SetNamespace("argocd")

	// Set labels for tracking (instead of owner references for cross-namespace)
	labels := argoApp.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["helios.io/managed-by"] = "helios-operator"
	labels["helios.io/app-name"] = name
	labels["helios.io/app-namespace"] = namespace
	argoApp.SetLabels(labels)

	// Default gitopsPath for logging
	gitopsPath := heliosApp.Spec.GitopsPath
	if gitopsPath == "" {
		gitopsPath = name
	}

	// Create or update ArgoCD Application
	existingArgoApp := &unstructured.Unstructured{}
	existingArgoApp.SetGroupVersionKind(argoApp.GroupVersionKind())
	err = r.Get(ctx, client.ObjectKey{Namespace: "argocd", Name: argoApp.GetName()}, existingArgoApp)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Creating ArgoCD Application", "name", argoApp.GetName(), "gitopsRepo", heliosApp.Spec.GitopsRepo, "gitopsPath", gitopsPath)
			if err := r.Create(ctx, argoApp); err != nil {
				return common.NewReconciliationError("ArgoCD Application", argoApp.GetName(), "create", err)
			}
			return nil
		}
		logger.Error(err, "failed to get ArgoCD Application")
		return common.NewReconciliationError("ArgoCD Application", argoApp.GetName(), "get", err)
	}

	// Update if spec changed
	if !equalUnstructured(argoApp, existingArgoApp) {
		logger.Info("Updating ArgoCD Application", "name", argoApp.GetName(), "gitopsRepo", heliosApp.Spec.GitopsRepo, "gitopsPath", gitopsPath)
		argoApp.SetResourceVersion(existingArgoApp.GetResourceVersion())
		if err := r.Update(ctx, argoApp); err != nil {
			return common.NewReconciliationError("ArgoCD Application", argoApp.GetName(), "update", err)
		}
	}

	return nil
}

// createOrUpdateResource creates or updates a Kubernetes resource
func (r *HeliosAppReconciler) createOrUpdateResource(ctx context.Context, obj *unstructured.Unstructured, logger logr.Logger) error {
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(obj.GroupVersionKind())

	err := r.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, existing)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Creating resource", "kind", obj.GetKind(), "name", obj.GetName())
			if err := r.Create(ctx, obj); err != nil {
				return common.NewReconciliationError(obj.GetKind(), obj.GetName(), "create", err)
			}
			return nil
		}
		logger.Error(err, "failed to get resource", "name", obj.GetName())
		return common.NewReconciliationError(obj.GetKind(), obj.GetName(), "get", err)
	}

	// Update if spec changed
	if !equalUnstructured(obj, existing) {
		logger.Info("Updating resource", "kind", obj.GetKind(), "name", obj.GetName())
		obj.SetResourceVersion(existing.GetResourceVersion())
		if err := r.Update(ctx, obj); err != nil {
			return common.NewReconciliationError(obj.GetKind(), obj.GetName(), "update", err)
		}
	}

	return nil
}

// updateComprehensiveStatus updates the HeliosApp status with build, deployment, and ArgoCD sync information
func (r *HeliosAppReconciler) updateComprehensiveStatus(ctx context.Context, heliosApp *heliosappv1.HeliosApp, name, namespace string, logger logr.Logger) error {
	// Get PipelineRun status
	buildStatus, buildVersion, pipelineRunName, lastBuildTime, err := r.getPipelineRunStatus(ctx, heliosApp)
	if err != nil {
		logger.V(1).Info("Could not get PipelineRun status", "error", err)
	} else {
		heliosApp.Status.BuildStatus = buildStatus
		heliosApp.Status.BuildVersion = buildVersion
		heliosApp.Status.CurrentPipelineRun = pipelineRunName
		heliosApp.Status.LastBuildTime = lastBuildTime
		logger.Info("PipelineRun status", "build", buildStatus, "version", buildVersion, "pipelineRun", pipelineRunName)

		// Update metrics
		BuildsTotal.WithLabelValues(namespace, name, buildStatus).Inc()

		// Update build condition
		if buildStatus == "Succeeded" {
			_ = r.updateStatus(ctx, heliosApp, "BuildSucceeded", metav1.ConditionTrue, "BuildCompleted", fmt.Sprintf("Build succeeded: %s", buildVersion))
		} else if buildStatus == "Failed" {
			_ = r.updateStatus(ctx, heliosApp, "BuildSucceeded", metav1.ConditionFalse, "BuildFailed", "Build failed")
		} else if buildStatus == "Running" {
			_ = r.updateStatus(ctx, heliosApp, "BuildSucceeded", metav1.ConditionFalse, "BuildInProgress", "Build is in progress")
		}
	}

	// Get Deployment health
	deployHealth, readyReplicas, desiredReplicas, lastHealthyTime, err := r.getDeploymentHealth(ctx, heliosApp)
	if err != nil {
		logger.V(1).Info("Could not get Deployment health", "error", err)
	} else {
		heliosApp.Status.DeploymentHealth = deployHealth
		heliosApp.Status.ReadyReplicas = readyReplicas
		heliosApp.Status.DesiredReplicas = desiredReplicas
		if lastHealthyTime != nil {
			heliosApp.Status.LastHealthyTime = lastHealthyTime
		}
		logger.Info("Deployment health", "status", deployHealth, "ready", readyReplicas, "desired", desiredReplicas)

		// Update metrics
		DeploymentHealthGauge.WithLabelValues(namespace, name).Set(DeploymentHealthToMetric(deployHealth))
		ReplicasGauge.WithLabelValues(namespace, name, "desired").Set(float64(desiredReplicas))
		ReplicasGauge.WithLabelValues(namespace, name, "ready").Set(float64(readyReplicas))

		// Update deployment condition
		if deployHealth == "Healthy" {
			_ = r.updateStatus(ctx, heliosApp, "DeploymentHealthy", metav1.ConditionTrue, "AllReplicasReady", fmt.Sprintf("All %d replicas are ready", readyReplicas))
		} else if deployHealth == "Progressing" {
			_ = r.updateStatus(ctx, heliosApp, "DeploymentHealthy", metav1.ConditionFalse, "RollingOut", fmt.Sprintf("Rolling out: %d/%d replicas ready", readyReplicas, desiredReplicas))
		} else if deployHealth == "Degraded" {
			_ = r.updateStatus(ctx, heliosApp, "DeploymentHealthy", metav1.ConditionFalse, "NoReplicasReady", "No replicas are ready")
		}
	}

	// Update status with ArgoCD sync information
	syncStatus, healthStatus, err := r.getArgoAppSyncStatus(ctx, name)
	if err != nil {
		logger.V(1).Info("Could not get ArgoCD sync status", "error", err)
	} else {
		logger.Info("ArgoCD Application status", "sync", syncStatus, "health", healthStatus)

		// Update metrics
		ArgoCDSyncStatus.WithLabelValues(namespace, name).Set(SyncStatusToMetric(syncStatus))
		ArgoCDHealthStatus.WithLabelValues(namespace, name).Set(HealthStatusToMetric(healthStatus))

		// Update DeployedVersion if synced
		if syncStatus == "Synced" && healthStatus == "Healthy" {
			// Use build version if available, otherwise use timestamp
			if buildVersion != "" {
				heliosApp.Status.DeployedVersion = buildVersion
			} else {
				heliosApp.Status.DeployedVersion = "synced-" + time.Now().Format("20060102-150405")
			}
		}

		// Update ArgoCD sync condition
		if syncStatus == "Synced" {
			_ = r.updateStatus(ctx, heliosApp, "ApplicationSynced", metav1.ConditionTrue, "Synced", "ArgoCD Application is synced")
		} else {
			_ = r.updateStatus(ctx, heliosApp, "ApplicationSynced", metav1.ConditionFalse, "NotSynced", "ArgoCD Application sync status: "+syncStatus)
		}
	}

	// Final status update
	return r.Status().Update(ctx, heliosApp)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HeliosAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create an unstructured object for ArgoCD Application
	argoApp := &unstructured.Unstructured{}
	argoApp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	})

	// Create an unstructured object for Tekton PipelineRun
	pipelineRun := &unstructured.Unstructured{}
	pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "tekton.dev",
		Version: "v1",
		Kind:    "PipelineRun",
	})

	// Create an unstructured object for Deployment
	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&heliosappv1.HeliosApp{}).
		// Watch ArgoCD Applications
		Watches(
			argoApp,
			handler.EnqueueRequestsFromMapFunc(r.argoAppToHeliosApp),
			builder.WithPredicates(r.argoAppPredicate()),
		).
		// Watch Tekton PipelineRuns for build status
		Watches(
			pipelineRun,
			handler.EnqueueRequestsFromMapFunc(r.pipelineRunToHeliosApp),
			builder.WithPredicates(r.pipelineRunPredicate()),
		).
		// Watch Deployments for pod health
		Watches(
			deployment,
			handler.EnqueueRequestsFromMapFunc(r.deploymentToHeliosApp),
			builder.WithPredicates(r.deploymentPredicate()),
		).
		Complete(r)
}
