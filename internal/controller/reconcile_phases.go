/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package controller implements specialized reconcile functions for different phases
// of the HeliosApp reconciliation process. Each function is responsible for a specific
// aspect of the application lifecycle and returns a ReconcileResult with status information.
package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"github.com/hoangphuc841/helios-operator/internal/common"
	"github.com/hoangphuc841/helios-operator/internal/resources"
)

// ReconcilePhase represents a specific phase of reconciliation
type ReconcilePhase string

const (
	PhasePipeline ReconcilePhase = "Pipeline"
	PhaseTriggers ReconcilePhase = "Triggers"
	PhaseArgoCD   ReconcilePhase = "ArgoCD"
	PhaseStatus   ReconcilePhase = "Status"
)

// ReconcilePipeline handles the reconciliation of Tekton Pipeline resources
func (r *HeliosAppReconciler) ReconcilePipeline(ctx context.Context, heliosApp *heliosappv1.HeliosApp, logger logr.Logger) *common.ReconcileResult {
	logger = logger.WithValues("phase", "pipeline")
	logger.V(1).Info("Starting Pipeline reconciliation")

	// Generate Pipeline resource
	pipeline, err := resources.GeneratePipeline(heliosApp)
	if err != nil {
		logger.Error(err, "Failed to generate Pipeline resource")
		return common.NewErrorResult(
			string(PhasePipeline),
			"Failed",
			fmt.Sprintf("Failed to generate Pipeline: %v", err),
			common.NewResourceGenerationError("Pipeline", heliosApp.Name, err),
			0,
		)
	}

	// Set namespace and controller reference
	pipeline.SetNamespace(heliosApp.Namespace)
	if err := controllerutil.SetControllerReference(heliosApp, pipeline, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for Pipeline")
		return common.NewErrorResult(
			string(PhasePipeline),
			"Failed",
			fmt.Sprintf("Failed to set owner reference: %v", err),
			common.NewReconciliationError("Pipeline", pipeline.GetName(), "set owner reference", err),
			0,
		)
	}

	// Create or update Pipeline
	if err := r.createOrUpdateResource(ctx, pipeline, logger); err != nil {
		// Check if this is a transient error
		if common.IsTransientError(err) {
			requeueAfter := common.GetRequeueDelay(0, 30*time.Second)
			logger.Info("Transient error during Pipeline reconciliation, will retry",
				"error", err,
				"requeueAfter", requeueAfter)
			return common.NewErrorResult(
				string(PhasePipeline),
				"Retrying",
				fmt.Sprintf("Transient error: %v", err),
				err,
				requeueAfter,
			)
		}

		logger.Error(err, "Failed to create or update Pipeline")
		return common.NewErrorResult(
			string(PhasePipeline),
			"Failed",
			fmt.Sprintf("Failed to create/update Pipeline: %v", err),
			err,
			0,
		)
	}

	logger.V(1).Info("Pipeline reconciled successfully", "pipelineName", pipeline.GetName())
	return common.NewSuccessResult(
		string(PhasePipeline),
		"Success",
		fmt.Sprintf("Pipeline %s reconciled successfully", pipeline.GetName()),
	)
}

// ReconcileTriggers handles the reconciliation of Tekton Trigger resources
func (r *HeliosAppReconciler) ReconcileTriggers(ctx context.Context, heliosApp *heliosappv1.HeliosApp, name, namespace, pipelineName, serviceAccount, githubSecret string, workspace map[string]interface{}, logger logr.Logger) *common.ReconcileResult {
	logger = logger.WithValues("phase", "triggers")
	logger.V(1).Info("Starting Triggers reconciliation",
		"eventListener", name+"-el",
		"triggerBinding", name+"-trigger-binding",
		"triggerTemplate", name+"-trigger-template",
	)

	// Generate EventListener
	eventListener, err := resources.GenerateEventListener(
		name+"-el", namespace, name+"-trigger", name+"-trigger-binding", name+"-trigger-template", githubSecret,
	)
	if err != nil {
		logger.Error(err, "Failed to generate EventListener resource")
		return common.NewErrorResult(
			string(PhaseTriggers),
			"Failed",
			fmt.Sprintf("Failed to generate EventListener: %v", err),
			common.NewResourceGenerationError("EventListener", name+"-el", err),
			0,
		)
	}

	// Generate TriggerBinding
	triggerBinding, err := resources.GenerateTriggerBinding(name+"-trigger-binding", namespace)
	if err != nil {
		logger.Error(err, "Failed to generate TriggerBinding resource")
		return common.NewErrorResult(
			string(PhaseTriggers),
			"Failed",
			fmt.Sprintf("Failed to generate TriggerBinding: %v", err),
			common.NewResourceGenerationError("TriggerBinding", name+"-trigger-binding", err),
			0,
		)
	}

	// Generate TriggerTemplate
	triggerTemplate, err := resources.GenerateTriggerTemplate(
		name+"-trigger-template", namespace, name+"-pipelinerun", pipelineName, serviceAccount, workspace,
	)
	if err != nil {
		logger.Error(err, "Failed to generate TriggerTemplate resource")
		return common.NewErrorResult(
			string(PhaseTriggers),
			"Failed",
			fmt.Sprintf("Failed to generate TriggerTemplate: %v", err),
			common.NewResourceGenerationError("TriggerTemplate", name+"-trigger-template", err),
			0,
		)
	}

	// Process each trigger resource
	triggerResources := []*unstructured.Unstructured{eventListener, triggerBinding, triggerTemplate}
	for _, obj := range triggerResources {
		obj.SetNamespace(namespace)
		if err := controllerutil.SetControllerReference(heliosApp, obj, r.Scheme); err != nil {
			logger.Error(err, "Failed to set controller reference",
				"kind", obj.GetKind(),
				"resourceName", obj.GetName())
			return common.NewErrorResult(
				string(PhaseTriggers),
				"Failed",
				fmt.Sprintf("Failed to set owner reference for %s: %v", obj.GetKind(), err),
				common.NewReconciliationError(obj.GetKind(), obj.GetName(), "set owner reference", err),
				0,
			)
		}

		if err := r.createOrUpdateResource(ctx, obj, logger); err != nil {
			// Check if this is a transient error
			if common.IsTransientError(err) {
				requeueAfter := common.GetRequeueDelay(0, 30*time.Second)
				logger.Info("Transient error during Triggers reconciliation, will retry",
					"resource", obj.GetKind(),
					"error", err,
					"requeueAfter", requeueAfter)
				return common.NewErrorResult(
					string(PhaseTriggers),
					"Retrying",
					fmt.Sprintf("Transient error for %s: %v", obj.GetKind(), err),
					err,
					requeueAfter,
				)
			}

			logger.Error(err, "Failed to create or update trigger resource",
				"kind", obj.GetKind(),
				"name", obj.GetName())
			return common.NewErrorResult(
				string(PhaseTriggers),
				"Failed",
				fmt.Sprintf("Failed to create/update %s: %v", obj.GetKind(), err),
				err,
				0,
			)
		}
	}

	logger.V(1).Info("Triggers reconciled successfully")
	return common.NewSuccessResult(
		string(PhaseTriggers),
		"Success",
		"All trigger resources reconciled successfully",
	)
}

// ReconcileArgoCD handles the reconciliation of ArgoCD Application resources
func (r *HeliosAppReconciler) ReconcileArgoCD(ctx context.Context, heliosApp *heliosappv1.HeliosApp, name, namespace string, logger logr.Logger) *common.ReconcileResult {
	logger = logger.WithValues("phase", "argocd")
	logger.V(1).Info("Starting ArgoCD reconciliation")

	// Generate ArgoCD Application
	argoApp, err := resources.GenerateArgoApplication(heliosApp)
	if err != nil {
		logger.Error(err, "Failed to generate ArgoCD Application resource")
		return common.NewErrorResult(
			string(PhaseArgoCD),
			"Failed",
			fmt.Sprintf("Failed to generate ArgoCD Application: %v", err),
			common.NewResourceGenerationError("ArgoCD Application", name+"-argocd", err),
			0,
		)
	}

	// Set namespace and labels
	argoApp.SetNamespace("argocd")
	labels := argoApp.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["helios.io/managed-by"] = "helios-operator"
	labels["helios.io/app-name"] = name
	labels["helios.io/app-namespace"] = namespace
	argoApp.SetLabels(labels)

	// Check if ArgoCD Application exists
	existingArgoApp := &unstructured.Unstructured{}
	existingArgoApp.SetGroupVersionKind(argoApp.GroupVersionKind())
	err = r.Get(ctx, client.ObjectKey{Namespace: "argocd", Name: argoApp.GetName()}, existingArgoApp)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Create new ArgoCD Application
			logger.Info("Creating ArgoCD Application",
				"appName", argoApp.GetName(),
				"gitopsRepo", heliosApp.Spec.GitopsRepo,
				"gitopsPath", heliosApp.Spec.GitopsPath)

			if err := r.Create(ctx, argoApp); err != nil {
				if common.IsTransientError(err) {
					requeueAfter := common.GetRequeueDelay(0, 30*time.Second)
					logger.Info("Transient error during ArgoCD Application creation, will retry",
						"error", err,
						"requeueAfter", requeueAfter)
					return common.NewErrorResult(
						string(PhaseArgoCD),
						"Retrying",
						fmt.Sprintf("Transient error during creation: %v", err),
						err,
						requeueAfter,
					)
				}

				logger.Error(err, "Failed to create ArgoCD Application")
				return common.NewErrorResult(
					string(PhaseArgoCD),
					"Failed",
					fmt.Sprintf("Failed to create ArgoCD Application: %v", err),
					common.NewReconciliationError("ArgoCD Application", argoApp.GetName(), "create", err),
					0,
				)
			}

			logger.V(1).Info("ArgoCD Application created successfully")
			return common.NewSuccessResult(
				string(PhaseArgoCD),
				"Success",
				fmt.Sprintf("ArgoCD Application %s created successfully", argoApp.GetName()),
			)
		}

		logger.Error(err, "Failed to get ArgoCD Application")
		return common.NewErrorResult(
			string(PhaseArgoCD),
			"Failed",
			fmt.Sprintf("Failed to get ArgoCD Application: %v", err),
			common.NewReconciliationError("ArgoCD Application", argoApp.GetName(), "get", err),
			0,
		)
	}

	// Update if spec changed
	if !r.equalUnstructured(argoApp, existingArgoApp) {
		logger.Info("Updating ArgoCD Application",
			"appName", argoApp.GetName(),
			"gitopsRepo", heliosApp.Spec.GitopsRepo,
			"gitopsPath", heliosApp.Spec.GitopsPath)

		argoApp.SetResourceVersion(existingArgoApp.GetResourceVersion())
		if err := r.Update(ctx, argoApp); err != nil {
			if common.IsTransientError(err) {
				requeueAfter := common.GetRequeueDelay(0, 30*time.Second)
				logger.Info("Transient error during ArgoCD Application update, will retry",
					"error", err,
					"requeueAfter", requeueAfter)
				return common.NewErrorResult(
					string(PhaseArgoCD),
					"Retrying",
					fmt.Sprintf("Transient error during update: %v", err),
					err,
					requeueAfter,
				)
			}

			logger.Error(err, "Failed to update ArgoCD Application")
			return common.NewErrorResult(
				string(PhaseArgoCD),
				"Failed",
				fmt.Sprintf("Failed to update ArgoCD Application: %v", err),
				common.NewReconciliationError("ArgoCD Application", argoApp.GetName(), "update", err),
				0,
			)
		}

		logger.V(1).Info("ArgoCD Application updated successfully")
		return common.NewSuccessResult(
			string(PhaseArgoCD),
			"Success",
			fmt.Sprintf("ArgoCD Application %s updated successfully", argoApp.GetName()),
		)
	}

	logger.V(1).Info("ArgoCD Application is up-to-date")
	return common.NewSuccessResult(
		string(PhaseArgoCD),
		"Success",
		fmt.Sprintf("ArgoCD Application %s is up-to-date", argoApp.GetName()),
	)
}

// ReconcileStatus handles the comprehensive status update for HeliosApp
func (r *HeliosAppReconciler) ReconcileStatus(ctx context.Context, heliosApp *heliosappv1.HeliosApp, name, namespace string, logger logr.Logger) *common.ReconcileResult {
	logger = logger.WithValues("phase", "status")
	logger.V(1).Info("Starting status reconciliation")

	// Get PipelineRun status
	buildStatus, buildVersion, pipelineRunName, lastBuildTime, err := r.getPipelineRunStatus(ctx, heliosApp)
	if err != nil {
		logger.V(1).Info("Could not retrieve PipelineRun status", "error", err.Error())
		// Don't fail the entire status update for this
	} else {
		heliosApp.Status.BuildStatus = buildStatus
		heliosApp.Status.BuildVersion = buildVersion
		heliosApp.Status.CurrentPipelineRun = pipelineRunName
		heliosApp.Status.LastBuildTime = lastBuildTime
		logger.Info("Build status updated",
			"status", buildStatus,
			"version", buildVersion,
			"pipelineRun", pipelineRunName)

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
		logger.V(1).Info("Could not retrieve Deployment health", "error", err.Error())
		// Don't fail the entire status update for this
	} else {
		heliosApp.Status.DeploymentHealth = deployHealth
		heliosApp.Status.ReadyReplicas = readyReplicas
		heliosApp.Status.DesiredReplicas = desiredReplicas
		if lastHealthyTime != nil {
			heliosApp.Status.LastHealthyTime = lastHealthyTime
		}
		logger.Info("Deployment health updated",
			"health", deployHealth,
			"readyReplicas", readyReplicas,
			"desiredReplicas", desiredReplicas)

		// Update deployment condition
		if deployHealth == "Healthy" {
			_ = r.updateStatus(ctx, heliosApp, "DeploymentHealthy", metav1.ConditionTrue, "AllReplicasReady", fmt.Sprintf("All %d replicas are ready", readyReplicas))
		} else if deployHealth == "Progressing" {
			_ = r.updateStatus(ctx, heliosApp, "DeploymentHealthy", metav1.ConditionFalse, "RollingOut", fmt.Sprintf("Rolling out: %d/%d replicas ready", readyReplicas, desiredReplicas))
		} else if deployHealth == "Degraded" {
			_ = r.updateStatus(ctx, heliosApp, "DeploymentHealthy", metav1.ConditionFalse, "NoReplicasReady", "No replicas are ready")
		}
	}

	// Update ArgoCD sync status
	syncStatus, healthStatus, err := r.getArgoAppSyncStatus(ctx, name)
	if err != nil {
		logger.V(1).Info("Could not retrieve ArgoCD sync status", "error", err.Error())
		// Don't fail the entire status update for this
	} else {
		logger.Info("ArgoCD sync status updated",
			"syncStatus", syncStatus,
			"healthStatus", healthStatus)

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

	// Persist status updates
	logger.V(1).Info("Persisting status updates to API server")
	if err := r.Status().Update(ctx, heliosApp); err != nil {
		if common.IsTransientError(err) {
			requeueAfter := common.GetRequeueDelay(0, 30*time.Second)
			logger.Info("Transient error during status update, will retry",
				"error", err,
				"requeueAfter", requeueAfter)
			return common.NewErrorResult(
				string(PhaseStatus),
				"Retrying",
				fmt.Sprintf("Transient error during status update: %v", err),
				err,
				requeueAfter,
			)
		}

		logger.Error(err, "Failed to persist status updates")
		return common.NewErrorResult(
			string(PhaseStatus),
			"Failed",
			fmt.Sprintf("Failed to persist status updates: %v", err),
			common.NewStatusUpdateError("HeliosApp", name, err),
			0,
		)
	}

	logger.V(1).Info("Status reconciliation completed successfully")
	return common.NewSuccessResult(
		string(PhaseStatus),
		"Success",
		"Status updated successfully",
	)
}

// equalUnstructured compares spec of two unstructured objects
func (r *HeliosAppReconciler) equalUnstructured(a, b *unstructured.Unstructured) bool {
	// Compare spec using JSON marshaling to avoid deep copy issues
	specA, okA := a.Object["spec"]
	specB, okB := b.Object["spec"]
	if !okA || !okB {
		return false
	}
	return reflect.DeepEqual(specA, specB)
}
