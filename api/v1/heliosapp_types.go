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

// Package v1 contains API Schema definitions for the platform v1 API group.
// It defines the HeliosApp custom resource which provides a simplified interface
// for deploying applications using GitOps with Tekton Pipelines and ArgoCD.
//
// The main resource is HeliosApp, which automatically provisions:
//   - Tekton Pipeline for building container images from source
//   - Tekton Triggers for webhook-based automation
//   - ArgoCD Application for GitOps-based deployment
//   - Comprehensive status tracking for the entire CI/CD workflow
package v1

import (
	"github.com/hoangphuc841/helios-operator/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HeliosAppSpec defines the desired state of HeliosApp.
//
// This specification defines all the configuration needed to set up a complete
// GitOps workflow for an application, including source code management,
// container image building, and deployment automation.
//
// Example:
//
//	apiVersion: platform.helios.io/v1
//	kind: HeliosApp
//	metadata:
//	  name: my-app
//	  namespace: default
//	spec:
//	  gitRepo: https://github.com/example/my-app
//	  gitBranch: main
//	  imageRepo: my-registry/my-app
//	  port: 8080
//	  replicas: 2
//	  serviceAccount: my-app-sa
//	  webhookSecret: github-webhook-secret
//	  gitopsRepo: https://github.com/example/my-app-manifests
//	  gitopsPath: apps/my-app
//	  gitopsBranch: main
type HeliosAppSpec struct {
	// GitRepo is the URL of the source code repository containing the application.
	// This repository will be cloned by Tekton pipelines for building container images.
	// Example: "https://github.com/example/my-app"
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?://`
	GitRepo string `json:"gitRepo"`

	// GitBranch specifies the branch or tag to use when cloning the source repository.
	// Defaults to "main" if not specified.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="main"
	GitBranch string `json:"gitBranch,omitempty"`

	// ImageRepo is the container registry repository where built images will be pushed.
	// Images will be tagged with the git commit SHA for traceability.
	// Example: "my-registry.com/my-app"
	// +kubebuilder:validation:Required
	ImageRepo string `json:"imageRepo"`

	// Port specifies the port that the application listens on.
	// This is used for health checks and service configuration.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// Replicas defines the desired number of application replicas.
	// Defaults to 1 if not specified.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas"`

	// ServiceAccount is the Kubernetes ServiceAccount that Tekton PipelineRuns will use.
	// This ServiceAccount must have permissions to push images to the specified ImageRepo
	// and access any required secrets for the build process.
	// +kubebuilder:validation:Required
	ServiceAccount string `json:"serviceAccount"`

	// WebhookSecret is the name of a Kubernetes Secret containing the GitHub webhook secret.
	// This secret is used to validate incoming webhook payloads from GitHub.
	// The secret should contain a key named "webhook-secret" with the secret value.
	// +kubebuilder:validation:Required
	WebhookSecret string `json:"webhookSecret"`

	// PVCName optionally specifies an existing PersistentVolumeClaim for the pipeline workspace.
	// If not provided, a PVC will be automatically created with a default size of 1Gi.
	// The automatically created PVC will be named: <helios-app-name>-workspace
	// +optional
	PVCName string `json:"pvcName,omitempty"`

	// GitopsRepo is the URL of the GitOps repository containing deployment manifests.
	// ArgoCD will monitor this repository for changes and sync them to the cluster.
	// Example: "https://github.com/example/my-app-manifests"
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?://`
	GitopsRepo string `json:"gitopsRepo"`

	// GitopsPath specifies the path within the GitOps repository containing the application manifests.
	// Defaults to the HeliosApp name if not specified.
	// Example: "apps/my-app" or "k8s/production/my-app"
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	GitopsPath string `json:"gitopsPath,omitempty"`

	// GitopsBranch specifies the branch or tag to monitor in the GitOps repository.
	// Defaults to "main" if not specified.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="main"
	GitopsBranch string `json:"gitopsBranch,omitempty"`
}

// HeliosAppStatus defines the observed state of HeliosApp.
//
// This status provides comprehensive information about the current state of the application,
// including build status, deployment health, and synchronization status with GitOps repositories.
type HeliosAppStatus struct {
	// Conditions represent the latest available observations of an object's state.
	// These follow the standard Kubernetes condition pattern and include:
	// - Ready: Overall application health status
	// - Synced: ArgoCD synchronization status
	// - BuildSucceeded: Build pipeline execution status
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// DeployedVersion is the image tag/digest of the version currently running and healthy in the cluster.
	// This field is updated when ArgoCD successfully syncs a new version to the cluster.
	// Example: "sha256:abc123..." or "v1.2.3"
	// +optional
	DeployedVersion string `json:"deployedVersion,omitempty"`

	// LastAppliedRevision is the git commit SHA that was last successfully synced by ArgoCD.
	// This provides traceability back to the specific commit in the GitOps repository.
	// Example: "abc123def456..."
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastBuild contains information about the most recent Tekton PipelineRun execution.
	// This includes build status, timing, and the resulting image version.
	// +optional
	LastBuild *BuildStatus `json:"lastBuild,omitempty"`

	// WebhookURL is the publicly accessible URL for the Tekton EventListener.
	// This URL should be configured as a webhook in your Git repository to trigger builds.
	// Example: "https://el-helios-listener.default.svc.cluster.local:8080"
	// +optional
	WebhookURL string `json:"webhookURL,omitempty"`

	// --- Legacy fields for backward compatibility ---
	// These will be deprecated in future versions

	// Current build status: "Running", "Succeeded", "Failed", "Unknown"
	// +optional
	// +deprecated: Use LastBuild.State instead
	BuildStatus string `json:"buildStatus,omitempty"`

	// Image version from the last successful build
	// +optional
	// +deprecated: Use LastBuild.ResultingVersion instead
	BuildVersion string `json:"buildVersion,omitempty"`

	// Timestamp of the last build attempt
	// +optional
	// +deprecated: Use LastBuild.StartedAt instead
	LastBuildTime *metav1.Time `json:"lastBuildTime,omitempty"`

	// Name of the current/last PipelineRun
	// +optional
	// +deprecated: Use LastBuild.PipelineRunName instead
	CurrentPipelineRun string `json:"currentPipelineRun,omitempty"`

	// Number of ready replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Total desired replicas
	// +optional
	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`

	// Overall deployment health: "Healthy", "Progressing", "Degraded", "Unknown"
	// +optional
	DeploymentHealth string `json:"deploymentHealth,omitempty"`

	// Timestamp when deployment last became healthy
	// +optional
	LastHealthyTime *metav1.Time `json:"lastHealthyTime,omitempty"`
}

// BuildStatus provides a summary of a Tekton PipelineRun execution.
//
// This struct contains all the relevant information about a build pipeline execution,
// including timing, status, and the resulting container image version.
type BuildStatus struct {
	// State indicates the current status of the build pipeline execution.
	// Possible values: "Succeeded", "Failed", "Running", "Unknown"
	// +optional
	State string `json:"state,omitempty"`

	// PipelineRunName is the name of the Tekton PipelineRun that executed this build.
	// This can be used to query detailed logs and information about the build.
	// Example: "my-app-pipelinerun-abc123"
	// +optional
	PipelineRunName string `json:"pipelineRunName,omitempty"`

	// StartedAt is the timestamp when the build pipeline started execution.
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// FinishedAt is the timestamp when the build pipeline completed execution.
	// This field is only populated when the build reaches a final state (Succeeded or Failed).
	// +optional
	FinishedAt *metav1.Time `json:"finishedAt,omitempty"`

	// ResultingVersion is the image tag or digest of the container image produced by this build.
	// This typically corresponds to the git commit SHA or a semantic version tag.
	// Example: "sha256:abc123..." or "v1.2.3"
	// +optional
	ResultingVersion string `json:"resultingVersion,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Git Repo",type="string",JSONPath=".spec.gitRepo"
// +kubebuilder:printcolumn:name="Image Repo",type="string",JSONPath=".spec.imageRepo"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.port"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
// +kubebuilder:printcolumn:name="Synced",type="string",JSONPath=".status.conditions[?(@.type==\"Synced\")].status"
// +kubebuilder:printcolumn:name="Build",type="string",JSONPath=".status.conditions[?(@.type==\"BuildSucceeded\")].status"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message"

// HeliosApp represents a complete GitOps-enabled application with automated CI/CD.
//
// HeliosApp automates the entire application lifecycle from source code to production deployment:
// 1. Monitors source code repository for changes via GitHub webhooks
// 2. Triggers Tekton pipelines to build and push container images
// 3. Updates GitOps repository with new image versions
// 4. ArgoCD automatically syncs changes to the cluster
// 5. Provides comprehensive status monitoring and health checks
//
// The operator creates and manages the following resources:
// - Tekton Pipeline, EventListener, TriggerBinding, and TriggerTemplate
// - ArgoCD Application for GitOps synchronization
// - PersistentVolumeClaim for build workspace (if not specified)
//
// Example usage:
//
//	kubectl apply -f - <<EOF
//	apiVersion: platform.helios.io/v1
//	kind: HeliosApp
//	metadata:
//	  name: my-app
//	  namespace: default
//	spec:
//	  gitRepo: https://github.com/example/my-app
//	  imageRepo: my-registry.com/my-app
//	  port: 8080
//	  serviceAccount: my-app-sa
//	  webhookSecret: github-webhook-secret
//	  gitopsRepo: https://github.com/example/my-app-manifests
//	EOF
type HeliosApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the HeliosApp
	Spec HeliosAppSpec `json:"spec,omitempty"`

	// Status defines the observed state of the HeliosApp
	Status HeliosAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HeliosAppList contains a list of HeliosApp
type HeliosAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HeliosApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HeliosApp{}, &HeliosAppList{})
}

// Standardized condition types for HeliosApp resources
const (
	// ConditionReady indicates that the application is fully reconciled and healthy.
	ConditionReady = common.ConditionReady
	// ConditionSynced indicates that the ArgoCD Application is synced with the GitOps repository.
	ConditionSynced = common.ConditionSynced
	// ConditionBuildSucceeded indicates the status of the last build pipeline.
	ConditionBuildSucceeded = common.ConditionBuildSucceeded
)

// Helper methods for HeliosApp

// IsReady returns true if the Ready condition is True
func (h *HeliosApp) IsReady() bool {
	return h.GetCondition(ConditionReady).Status == metav1.ConditionTrue
}

// IsSynced returns true if the Synced condition is True
func (h *HeliosApp) IsSynced() bool {
	return h.GetCondition(ConditionSynced).Status == metav1.ConditionTrue
}

// IsBuildSucceeded returns true if the BuildSucceeded condition is True
func (h *HeliosApp) IsBuildSucceeded() bool {
	return h.GetCondition(ConditionBuildSucceeded).Status == metav1.ConditionTrue
}

// GetCondition returns the condition with the given type
func (h *HeliosApp) GetCondition(conditionType string) metav1.Condition {
	for _, condition := range h.Status.Conditions {
		if condition.Type == conditionType {
			return condition
		}
	}
	return metav1.Condition{
		Type:   conditionType,
		Status: metav1.ConditionUnknown,
	}
}

// SetCondition sets or updates a condition
func (h *HeliosApp) SetCondition(conditionType string, status metav1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}

	for i, c := range h.Status.Conditions {
		if c.Type == conditionType {
			if c.Status != status {
				condition.LastTransitionTime = now
			} else {
				condition.LastTransitionTime = c.LastTransitionTime
			}
			h.Status.Conditions[i] = condition
			return
		}
	}

	h.Status.Conditions = append(h.Status.Conditions, condition)
}

// GetStatusMessage returns the message from the Ready condition
func (h *HeliosApp) GetStatusMessage() string {
	return h.GetCondition(ConditionReady).Message
}

// GetStatusReason returns the reason from the Ready condition
func (h *HeliosApp) GetStatusReason() string {
	return h.GetCondition(ConditionReady).Reason
}
