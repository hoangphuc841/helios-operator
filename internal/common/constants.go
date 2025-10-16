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

package common

const (
	// Labels.
	LabelManagedBy = "helios.io/managed-by"
	LabelAppName   = "helios.io/app-name"
	LabelComponent = "helios.io/component"

	// Finalizer.
	FinalizerName = "platform.helios.io/finalizer"

	// Default values.
	DefaultGitBranch  = "main"
	DefaultGitopsPath = "" // Will use app name
	DefaultReplicas   = int32(1)
	DefaultPVCSize    = "1Gi"

	// Tekton task names.
	TaskGitClone = "git-clone"
	TaskKaniko   = "kaniko"

	// ArgoCD.
	ArgoCDNamespace      = "argocd"
	ArgoCDDefaultProject = "default"

	// Condition types.
	ConditionReady             = "Ready"
	ConditionSynced            = "Synced"
	ConditionBuildSucceeded    = "BuildSucceeded"
	ConditionDeploymentHealthy = "DeploymentHealthy"
	ConditionApplicationSynced = "ApplicationSynced"

	// Build status values
	BuildStatusSucceeded = "Succeeded"
	BuildStatusFailed    = "Failed"
	BuildStatusRunning   = "Running"
	BuildStatusUnknown   = "Unknown"

	// Deployment health values
	DeploymentHealthHealthy     = "Healthy"
	DeploymentHealthProgressing = "Progressing"
	DeploymentHealthDegraded    = "Degraded"
	DeploymentHealthUnknown     = "Unknown"

	// ArgoCD sync status values
	SyncStatusSynced    = "Synced"
	SyncStatusOutOfSync = "OutOfSync"
	SyncStatusUnknown   = "Unknown"

	// ArgoCD health status values
	HealthStatusHealthy     = "Healthy"
	HealthStatusProgressing = "Progressing"
	HealthStatusDegraded    = "Degraded"
	HealthStatusUnknown     = "Unknown"
	HealthStatusMissing     = "Missing"
)
