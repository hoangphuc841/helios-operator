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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HeliosAppSpec defines the desired state of HeliosApp
type HeliosAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	GitRepo string `json:"gitRepo"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="main"
	GitBranch string `json:"gitBranch,omitempty"`

	// +kubebuilder:validation:Required
	ImageRepo string `json:"imageRepo"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas"`

	// +kubebuilder:validation:Required
	// Tên của ServiceAccount mà PipelineRun sẽ sử dụng.
	ServiceAccount string `json:"serviceAccount"`

	// +kubebuilder:validation:Required
	// Tên của Secret chứa mã webhook cho GitHub.
	WebhookSecret string `json:"webhookSecret"`

	PVCName string `json:"pvcName"`

	// +kubebuilder:validation:Required
	// GitOps Repository URL containing deployment manifests
	GitopsRepo string `json:"gitopsRepo"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	// Path in GitOps repository (defaults to app name if not specified)
	GitopsPath string `json:"gitopsPath,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="main"
	// Target revision/branch in GitOps repo
	GitopsBranch string `json:"gitopsBranch,omitempty"`
}

// HeliosAppStatus defines the observed state of HeliosApp
type HeliosAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Trạng thái hiện tại của các tài nguyên được quản lý: "DeploymentReady", "ServiceReady"
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Tag của image đã được deploy thành công
	// +optional
	DeployedVersion string `json:"deployedVersion,omitempty"`

	// --- Build Status (from Tekton PipelineRun) ---

	// Current build status: "Running", "Succeeded", "Failed", "Unknown"
	// +optional
	BuildStatus string `json:"buildStatus,omitempty"`

	// Image version from the last successful build
	// +optional
	BuildVersion string `json:"buildVersion,omitempty"`

	// Timestamp of the last build attempt
	// +optional
	LastBuildTime *metav1.Time `json:"lastBuildTime,omitempty"`

	// Name of the current/last PipelineRun
	// +optional
	CurrentPipelineRun string `json:"currentPipelineRun,omitempty"`

	// --- Deployment Health (from Deployment) ---

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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HeliosApp is the Schema for the heliosapps API
type HeliosApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HeliosAppSpec   `json:"spec,omitempty"`
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
