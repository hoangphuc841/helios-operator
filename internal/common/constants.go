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

// Package common provides shared constants, utilities, and helper functions.
// used across the Helios operator codebase.
package common

import "time"

// Operator constants.
const (
	// OperatorName is the name of the operator.
	OperatorName = "helios-operator"

	// OperatorVersion is the current version of the operator.
	OperatorVersion = "2.0.0"

	// DefaultNamespace is the default namespace for resources.
	DefaultNamespace = "helios-operator-system"

	// DefaultReconcileInterval is the default reconcile interval.
	DefaultReconcileInterval = 30 * time.Second

	// DefaultTimeout is the default timeout for operations.
	DefaultTimeout = 5 * time.Minute
)

// Resource labels.
const (
	// LabelAppName is the label for application name.
	LabelAppName = "app.kubernetes.io/name"

	// LabelAppInstance is the label for application instance.
	LabelAppInstance = "app.kubernetes.io/instance"

	// LabelAppVersion is the label for application version.
	LabelAppVersion = "app.kubernetes.io/version"

	// LabelAppComponent is the label for application component.
	LabelAppComponent = "app.kubernetes.io/component"

	// LabelAppPartOf is the label for application part of.
	LabelAppPartOf = "app.kubernetes.io/part-of"

	// LabelAppManagedBy is the label for application managed by.
	LabelAppManagedBy = "app.kubernetes.io/managed-by"

	// LabelHeliosApp is the label for HeliosApp.
	LabelHeliosApp = "helios.io/app"

	// LabelHeliosAppName is the label for HeliosApp name.
	LabelHeliosAppName = "helios.io/app-name"

	// LabelManagedBy is the label for managed by.
	LabelManagedBy = "helios.io/managed-by"

	// LabelComponent is the label for component.
	LabelComponent = "helios.io/component"
)

// Resource annotations.
const (
	// AnnotationLastApplied is the annotation for last applied configuration.
	AnnotationLastApplied = "kubectl.kubernetes.io/last-applied-configuration"

	// AnnotationManagedBy is the annotation for managed by.
	AnnotationManagedBy = "helios.io/managed-by"

	// AnnotationReconcileTime is the annotation for reconcile time.
	AnnotationReconcileTime = "helios.io/reconcile-time"
)

// Event reasons.
const (
	// EventReasonCreated is the event reason for created.
	EventReasonCreated = "Created"

	// EventReasonUpdated is the event reason for updated.
	EventReasonUpdated = "Updated"

	// EventReasonDeleted is the event reason for deleted.
	EventReasonDeleted = "Deleted"

	// EventReasonFailed is the event reason for failed.
	EventReasonFailed = "Failed"

	// EventReasonReconciled is the event reason for reconciled.
	EventReasonReconciled = "Reconciled"
)

// Phase constants.
const (
	// PhasePending is the phase for pending.
	PhasePending = "Pending"

	// PhaseRunning is the phase for running.
	PhaseRunning = "Running"

	// PhaseSucceeded is the phase for succeeded.
	PhaseSucceeded = "Succeeded"

	// PhaseFailed is the phase for failed.
	PhaseFailed = "Failed"

	// PhaseUnknown is the phase for unknown.
	PhaseUnknown = "Unknown"
)

// Condition types.
const (
	// ConditionReady indicates that the application is fully reconciled and healthy.
	ConditionReady = "Ready"

	// ConditionSynced indicates that the ArgoCD Application is synced with the GitOps repository.
	ConditionSynced = "Synced"

	// ConditionBuildSucceeded indicates the status of the last build pipeline.
	ConditionBuildSucceeded = "BuildSucceeded"

	// ConditionDeployed indicates that the application has been deployed.
	ConditionDeployed = "Deployed"

	// ConditionPipelineReady indicates that the Tekton Pipeline is ready.
	ConditionPipelineReady = "PipelineReady"

	// ConditionWebhookReady indicates that the webhook is ready.
	ConditionWebhookReady = "WebhookReady"
)

// Finalizer constants.
const (
	// FinalizerName is the name of the finalizer used by the operator.
	FinalizerName = "platform.helios.io/finalizer"
)
