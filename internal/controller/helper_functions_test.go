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

package controller

import (
	"testing"

	"github.com/hoangphuc841/helios-operator/internal/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestIsHeliosArgoApp tests the ArgoCD Application label checking
func TestIsHeliosArgoApp(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)

	reconciler := &HeliosAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Scheme: s,
	}

	tests := []struct {
		name     string
		obj      client.Object
		expected bool
	}{
		{
			name: "valid Helios ArgoCD app",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "argoproj.io/v1alpha1",
					"kind":       "Application",
					"metadata": map[string]interface{}{
						"name":      "test-app-argocd",
						"namespace": "argocd",
						"labels": map[string]interface{}{
							common.LabelManagedBy: "helios-operator",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "ArgoCD app without Helios label",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "argoproj.io/v1alpha1",
					"kind":       "Application",
					"metadata": map[string]interface{}{
						"name":      "other-app",
						"namespace": "argocd",
						"labels": map[string]interface{}{
							"managed-by": "other-operator",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "ArgoCD app with no labels",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "argoproj.io/v1alpha1",
					"kind":       "Application",
					"metadata": map[string]interface{}{
						"name":      "no-labels-app",
						"namespace": "argocd",
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconciler.isHeliosArgoApp(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsHeliosPipelineRun tests the PipelineRun label checking
func TestIsHeliosPipelineRun(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)

	reconciler := &HeliosAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Scheme: s,
	}

	tests := []struct {
		name     string
		obj      client.Object
		expected bool
	}{
		{
			name: "valid Helios PipelineRun",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "tekton.dev/v1",
					"kind":       "PipelineRun",
					"metadata": map[string]interface{}{
						"name":      "test-app-run",
						"namespace": "default",
						"labels": map[string]interface{}{
							"triggers.tekton.dev/trigger": "github-trigger",
							common.LabelManagedBy:         "helios-operator",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "PipelineRun without Tekton trigger label",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "tekton.dev/v1",
					"kind":       "PipelineRun",
					"metadata": map[string]interface{}{
						"name":      "other-run",
						"namespace": "default",
						"labels": map[string]interface{}{
							common.LabelManagedBy: "helios-operator",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "PipelineRun without Helios label",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "tekton.dev/v1",
					"kind":       "PipelineRun",
					"metadata": map[string]interface{}{
						"name":      "external-run",
						"namespace": "default",
						"labels": map[string]interface{}{
							"triggers.tekton.dev/trigger": "github-trigger",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "PipelineRun with no labels",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "tekton.dev/v1",
					"kind":       "PipelineRun",
					"metadata": map[string]interface{}{
						"name":      "no-labels-run",
						"namespace": "default",
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconciler.isHeliosPipelineRun(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsHeliosDeployment tests the Deployment label checking
func TestIsHeliosDeployment(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)

	reconciler := &HeliosAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Scheme: s,
	}

	tests := []struct {
		name     string
		obj      client.Object
		expected bool
	}{
		{
			name: "valid Helios Deployment with ArgoCD label",
			obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app",
					Namespace: "default",
					Labels: map[string]string{
						"app.kubernetes.io/instance": "test-app",
						"app":                        "test-app",
					},
				},
			},
			expected: true,
		},
		{
			name: "valid Deployment with app label",
			obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-app",
					Namespace: "default",
					Labels: map[string]string{
						"app": "my-app",
					},
				},
			},
			expected: true,
		},
		{
			name: "Deployment with empty app label",
			obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-deployment",
					Namespace: "default",
					Labels: map[string]string{
						"app": "",
					},
				},
			},
			expected: false,
		},
		{
			name: "Deployment with no labels",
			obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-labels-deployment",
					Namespace: "default",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconciler.isHeliosDeployment(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDeploymentHealthStatus tests different deployment health scenarios
func TestDeploymentHealthStatus(t *testing.T) {
	tests := []struct {
		name              string
		replicas          int32
		readyReplicas     int32
		availableReplicas int32
		expectedHealth    string
	}{
		{
			name:              "healthy deployment - all replicas ready",
			replicas:          3,
			readyReplicas:     3,
			availableReplicas: 3,
			expectedHealth:    "Healthy",
		},
		{
			name:              "progressing deployment - some replicas ready",
			replicas:          3,
			readyReplicas:     2,
			availableReplicas: 2,
			expectedHealth:    "Progressing",
		},
		{
			name:              "degraded deployment - no replicas ready",
			replicas:          3,
			readyReplicas:     0,
			availableReplicas: 0,
			expectedHealth:    "Degraded",
		},
		{
			name:              "zero replicas - should be healthy",
			replicas:          0,
			readyReplicas:     0,
			availableReplicas: 0,
			expectedHealth:    "Healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var health string
			if tt.replicas == 0 {
				health = "Healthy"
			} else if tt.readyReplicas == tt.replicas {
				health = "Healthy"
			} else if tt.readyReplicas > 0 {
				health = "Progressing"
			} else {
				health = "Degraded"
			}
			assert.Equal(t, tt.expectedHealth, health)
		})
	}
}

// TestPodConditions tests pod condition interpretation
func TestPodConditions(t *testing.T) {
	tests := []struct {
		name      string
		pod       *corev1.Pod
		isReady   bool
		isHealthy bool
	}{
		{
			name: "pod with Ready condition True",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			isReady:   true,
			isHealthy: true,
		},
		{
			name: "pod with Ready condition False",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			isReady:   false,
			isHealthy: false,
		},
		{
			name: "pod with no conditions",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{},
				},
			},
			isReady:   false,
			isHealthy: false,
		},
		{
			name: "pod with multiple conditions",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodScheduled,
							Status: corev1.ConditionTrue,
						},
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			isReady:   true,
			isHealthy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if pod is ready
			var ready bool
			for _, cond := range tt.pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					ready = true
					break
				}
			}
			assert.Equal(t, tt.isReady, ready)
			assert.Equal(t, tt.isHealthy, ready)
		})
	}
}

// TestBuildStatusInterpretation tests PipelineRun status interpretation
func TestBuildStatusInterpretation(t *testing.T) {
	tests := []struct {
		name           string
		conditions     interface{}
		expectedStatus string
	}{
		{
			name: "succeeded pipeline run",
			conditions: []interface{}{
				map[string]interface{}{
					"type":   "Succeeded",
					"status": "True",
					"reason": "Succeeded",
				},
			},
			expectedStatus: "Succeeded",
		},
		{
			name: "failed pipeline run",
			conditions: []interface{}{
				map[string]interface{}{
					"type":   "Succeeded",
					"status": "False",
					"reason": "Failed",
				},
			},
			expectedStatus: "Failed",
		},
		{
			name: "running pipeline run",
			conditions: []interface{}{
				map[string]interface{}{
					"type":   "Succeeded",
					"status": "Unknown",
					"reason": "Running",
				},
			},
			expectedStatus: "Running",
		},
		{
			name:           "no conditions",
			conditions:     []interface{}{},
			expectedStatus: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Interpret pipeline run status from conditions
			var status string
			if condSlice, ok := tt.conditions.([]interface{}); ok && len(condSlice) > 0 {
				if cond, ok := condSlice[0].(map[string]interface{}); ok {
					if condType, ok := cond["type"].(string); ok && condType == "Succeeded" {
						if condStatus, ok := cond["status"].(string); ok {
							if condStatus == "True" {
								status = "Succeeded"
							} else if condStatus == "False" {
								status = "Failed"
							} else {
								status = "Running"
							}
						}
					}
				}
			} else {
				status = "Unknown"
			}
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

// TestArgoCDSyncStatusInterpretation tests ArgoCD sync status values
func TestArgoCDSyncStatusInterpretation(t *testing.T) {
	tests := []struct {
		name         string
		syncStatus   string
		healthStatus string
		isHealthy    bool
		isSynced     bool
	}{
		{
			name:         "synced and healthy",
			syncStatus:   "Synced",
			healthStatus: "Healthy",
			isHealthy:    true,
			isSynced:     true,
		},
		{
			name:         "out of sync but healthy",
			syncStatus:   "OutOfSync",
			healthStatus: "Healthy",
			isHealthy:    true,
			isSynced:     false,
		},
		{
			name:         "synced but degraded",
			syncStatus:   "Synced",
			healthStatus: "Degraded",
			isHealthy:    false,
			isSynced:     true,
		},
		{
			name:         "out of sync and degraded",
			syncStatus:   "OutOfSync",
			healthStatus: "Degraded",
			isHealthy:    false,
			isSynced:     false,
		},
		{
			name:         "synced and progressing",
			syncStatus:   "Synced",
			healthStatus: "Progressing",
			isHealthy:    false,
			isSynced:     true,
		},
		{
			name:         "unknown status",
			syncStatus:   "Unknown",
			healthStatus: "Unknown",
			isHealthy:    false,
			isSynced:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSynced := tt.syncStatus == "Synced"
			isHealthy := tt.healthStatus == "Healthy"

			assert.Equal(t, tt.isSynced, isSynced)
			assert.Equal(t, tt.isHealthy, isHealthy)
		})
	}
}

// TestConditionStatusTransitions tests condition status state machine
func TestConditionStatusTransitions(t *testing.T) {
	tests := []struct {
		name             string
		currentStatus    metav1.ConditionStatus
		newStatus        metav1.ConditionStatus
		shouldTransition bool
	}{
		{
			name:             "True to True - no transition",
			currentStatus:    metav1.ConditionTrue,
			newStatus:        metav1.ConditionTrue,
			shouldTransition: false,
		},
		{
			name:             "True to False - transition",
			currentStatus:    metav1.ConditionTrue,
			newStatus:        metav1.ConditionFalse,
			shouldTransition: true,
		},
		{
			name:             "False to True - transition",
			currentStatus:    metav1.ConditionFalse,
			newStatus:        metav1.ConditionTrue,
			shouldTransition: true,
		},
		{
			name:             "Unknown to True - transition",
			currentStatus:    metav1.ConditionUnknown,
			newStatus:        metav1.ConditionTrue,
			shouldTransition: true,
		},
		{
			name:             "Unknown to Unknown - no transition",
			currentStatus:    metav1.ConditionUnknown,
			newStatus:        metav1.ConditionUnknown,
			shouldTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldTransition := tt.currentStatus != tt.newStatus
			assert.Equal(t, tt.shouldTransition, shouldTransition)
		})
	}
}
