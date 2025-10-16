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
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"github.com/hoangphuc841/helios-operator/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestHelper provides utilities for testing the HeliosApp controller
type TestHelper struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// NewTestHelper creates a new TestHelper instance
func NewTestHelper(client client.Client, scheme *runtime.Scheme) *TestHelper {
	return &TestHelper{
		Client: client,
		Scheme: scheme,
	}
}

// CreateTestHeliosApp creates a HeliosApp resource for testing
func (h *TestHelper) CreateTestHeliosApp(ctx context.Context, name, namespace string) (*heliosappv1.HeliosApp, error) {
	heliosApp := &heliosappv1.HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: heliosappv1.HeliosAppSpec{
			GitRepo:        "https://github.com/example/test-app.git",
			GitopsRepo:     "https://github.com/example/gitops.git",
			ImageRepo:      "example.com/test-app",
			Port:           8080,
			Replicas:       1,
			ServiceAccount: "pipeline-sa",
			WebhookSecret:  "webhook-secret",
			PVCName:        fmt.Sprintf("%s-pvc", name),
		},
	}

	if err := h.Client.Create(ctx, heliosApp); err != nil {
		return nil, err
	}

	return heliosApp, nil
}

// CreateMockPipeline creates a mock Tekton Pipeline for testing
func (h *TestHelper) CreateMockPipeline(ctx context.Context, name, namespace string) (*unstructured.Unstructured, error) {
	pipeline := &unstructured.Unstructured{}
	pipeline.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "tekton.dev",
		Version: "v1",
		Kind:    "Pipeline",
	})
	pipeline.SetName(name)
	pipeline.SetNamespace(namespace)

	// Set minimal spec
	pipeline.Object["spec"] = map[string]interface{}{
		"tasks": []interface{}{
			map[string]interface{}{
				"name": "build",
				"taskRef": map[string]interface{}{
					"name": "kaniko-build",
				},
			},
		},
	}

	if err := h.Client.Create(ctx, pipeline); err != nil {
		return nil, err
	}

	return pipeline, nil
}

// CreateMockEventListener creates a mock Tekton EventListener for testing
func (h *TestHelper) CreateMockEventListener(ctx context.Context, name, namespace string) (*unstructured.Unstructured, error) {
	eventListener := &unstructured.Unstructured{}
	eventListener.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "triggers.tekton.dev",
		Version: "v1beta1",
		Kind:    "EventListener",
	})
	eventListener.SetName(name)
	eventListener.SetNamespace(namespace)

	// Set minimal spec
	eventListener.Object["spec"] = map[string]interface{}{
		"triggers": []interface{}{
			map[string]interface{}{
				"name":       "github-trigger",
				"triggerRef": "github-trigger-template",
			},
		},
	}

	if err := h.Client.Create(ctx, eventListener); err != nil {
		return nil, err
	}

	return eventListener, nil
}

// CreateMockTriggerBinding creates a mock Tekton TriggerBinding for testing
func (h *TestHelper) CreateMockTriggerBinding(ctx context.Context, name, namespace string) (*unstructured.Unstructured, error) {
	triggerBinding := &unstructured.Unstructured{}
	triggerBinding.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "triggers.tekton.dev",
		Version: "v1beta1",
		Kind:    "TriggerBinding",
	})
	triggerBinding.SetName(name)
	triggerBinding.SetNamespace(namespace)

	// Set minimal spec
	triggerBinding.Object["spec"] = map[string]interface{}{
		"params": []interface{}{
			map[string]interface{}{
				"name":  "git-repo",
				"value": "https://github.com/example/test-app.git",
			},
		},
	}

	if err := h.Client.Create(ctx, triggerBinding); err != nil {
		return nil, err
	}

	return triggerBinding, nil
}

// CreateMockTriggerTemplate creates a mock Tekton TriggerTemplate for testing
func (h *TestHelper) CreateMockTriggerTemplate(ctx context.Context, name, namespace string) (*unstructured.Unstructured, error) {
	triggerTemplate := &unstructured.Unstructured{}
	triggerTemplate.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "triggers.tekton.dev",
		Version: "v1beta1",
		Kind:    "TriggerTemplate",
	})
	triggerTemplate.SetName(name)
	triggerTemplate.SetNamespace(namespace)

	// Set minimal spec
	triggerTemplate.Object["spec"] = map[string]interface{}{
		"params": []interface{}{
			map[string]interface{}{
				"name":        "git-repo",
				"description": "Git repository URL",
			},
		},
		"resourcetemplates": []interface{}{
			map[string]interface{}{
				"apiVersion": "tekton.dev/v1",
				"kind":       "PipelineRun",
				"metadata": map[string]interface{}{
					"generateName": "test-app-run-",
				},
			},
		},
	}

	if err := h.Client.Create(ctx, triggerTemplate); err != nil {
		return nil, err
	}

	return triggerTemplate, nil
}

// CreateMockArgoApplication creates a mock ArgoCD Application for testing
func (h *TestHelper) CreateMockArgoApplication(ctx context.Context, name, namespace string) (*unstructured.Unstructured, error) {
	argoApp := &unstructured.Unstructured{}
	argoApp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	})
	argoApp.SetName(name)
	argoApp.SetNamespace(namespace)

	// Set minimal spec
	argoApp.Object["spec"] = map[string]interface{}{
		"project": "default",
		"source": map[string]interface{}{
			"repoURL":        "https://github.com/example/gitops.git",
			"targetRevision": "HEAD",
			"path":           name,
		},
		"destination": map[string]interface{}{
			"server":    "https://kubernetes.default.svc",
			"namespace": "default",
		},
	}

	if err := h.Client.Create(ctx, argoApp); err != nil {
		return nil, err
	}

	return argoApp, nil
}

// CreateMockPipelineRun creates a mock Tekton PipelineRun for testing
func (h *TestHelper) CreateMockPipelineRun(ctx context.Context, name, namespace string, status string) (*unstructured.Unstructured, error) {
	pipelineRun := &unstructured.Unstructured{}
	pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "tekton.dev",
		Version: "v1",
		Kind:    "PipelineRun",
	})
	pipelineRun.SetName(name)
	pipelineRun.SetNamespace(namespace)

	// Set labels for tracking
	pipelineRun.SetLabels(map[string]string{
		"triggers.tekton.dev/trigger": "github-trigger",
		common.LabelManagedBy:         "helios-operator",
	})

	// Set status based on the provided status
	conditions := []interface{}{}
	switch status {
	case "Succeeded":
		conditions = append(conditions, map[string]interface{}{
			"type":    "Succeeded",
			"status":  "True",
			"reason":  "Succeeded",
			"message": "Pipeline completed successfully",
		})
	case "Failed":
		conditions = append(conditions, map[string]interface{}{
			"type":    "Succeeded",
			"status":  "False",
			"reason":  "Failed",
			"message": "Pipeline failed",
		})
	case "Running":
		conditions = append(conditions, map[string]interface{}{
			"type":    "Succeeded",
			"status":  "Unknown",
			"reason":  "Running",
			"message": "Pipeline is running",
		})
	default:
		conditions = append(conditions, map[string]interface{}{
			"type":    "Succeeded",
			"status":  "Unknown",
			"reason":  "Unknown",
			"message": "Status unknown",
		})
	}

	pipelineRun.Object["status"] = map[string]interface{}{
		"conditions": conditions,
	}

	if err := h.Client.Create(ctx, pipelineRun); err != nil {
		return nil, err
	}

	return pipelineRun, nil
}

// CreateMockDeployment creates a mock Kubernetes Deployment for testing
func (h *TestHelper) CreateMockDeployment(ctx context.Context, name, namespace string, readyReplicas, desiredReplicas int32) (*unstructured.Unstructured, error) {
	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	deployment.SetName(name)
	deployment.SetNamespace(namespace)

	// Set labels
	deployment.SetLabels(map[string]string{
		"app": name,
	})

	// Set spec
	deployment.Object["spec"] = map[string]interface{}{
		"replicas": desiredReplicas,
		"selector": map[string]interface{}{
			"matchLabels": map[string]interface{}{
				"app": name,
			},
		},
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": name,
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  name,
						"image": "nginx:latest",
					},
				},
			},
		},
	}

	// Set status
	status := map[string]interface{}{
		"readyReplicas":     readyReplicas,
		"availableReplicas": readyReplicas,
		"replicas":          desiredReplicas,
	}

	// Add conditions based on replica status
	var conditions []interface{}
	if readyReplicas == desiredReplicas && desiredReplicas > 0 {
		conditions = append(conditions, map[string]interface{}{
			"type":   "Available",
			"status": "True",
		})
		conditions = append(conditions, map[string]interface{}{
			"type":   "Progressing",
			"status": "True",
		})
	} else if readyReplicas > 0 {
		conditions = append(conditions, map[string]interface{}{
			"type":   "Available",
			"status": "False",
		})
		conditions = append(conditions, map[string]interface{}{
			"type":   "Progressing",
			"status": "True",
		})
	} else {
		conditions = append(conditions, map[string]interface{}{
			"type":   "Available",
			"status": "False",
		})
		conditions = append(conditions, map[string]interface{}{
			"type":   "Progressing",
			"status": "False",
		})
	}

	status["conditions"] = conditions
	deployment.Object["status"] = status

	if err := h.Client.Create(ctx, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

// WaitForHeliosAppCondition waits for a HeliosApp to have a specific condition
func (h *TestHelper) WaitForHeliosAppCondition(ctx context.Context, name, namespace string, conditionType string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for condition %s on HeliosApp %s/%s", conditionType, namespace, name)
		default:
			heliosApp := &heliosappv1.HeliosApp{}
			if err := h.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, heliosApp); err != nil {
				return err
			}

			for _, condition := range heliosApp.Status.Conditions {
				if condition.Type == conditionType && condition.Status == metav1.ConditionTrue {
					return nil
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

// GetHeliosApp retrieves a HeliosApp by name and namespace
func (h *TestHelper) GetHeliosApp(ctx context.Context, name, namespace string) (*heliosappv1.HeliosApp, error) {
	heliosApp := &heliosappv1.HeliosApp{}
	if err := h.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, heliosApp); err != nil {
		return nil, err
	}
	return heliosApp, nil
}

// CleanupHeliosApp deletes a HeliosApp and waits for it to be gone
func (h *TestHelper) CleanupHeliosApp(ctx context.Context, name, namespace string) error {
	heliosApp := &heliosappv1.HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return h.Client.Delete(ctx, heliosApp)
}

// CreateTestLogger creates a test logger
func CreateTestLogger() logr.Logger {
	return logr.Discard() // Use discard logger for tests to reduce noise
}
