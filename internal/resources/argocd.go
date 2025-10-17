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

// Package resources provides utilities for creating and managing Kubernetes resources
// including ArgoCD Applications, Tekton Pipelines, and other infrastructure components.
package resources

import (
	"fmt"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"github.com/hoangphuc841/helios-operator/internal/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GenerateArgoApplication creates an ArgoCD Application custom resource
// that points to the GitOps repository specified in the HeliosApp spec.
// The Application will be created in the 'argocd' namespace and will
// deploy resources to the same namespace as the HeliosApp.
//
// Default behavior:
// - If gitopsPath is empty, it defaults to the HeliosApp name
// - If gitopsBranch is empty, it defaults to "main"
//
// Returns an unstructured.Unstructured object representing the Application.
func GenerateArgoApplication(heliosApp *heliosappv1.HeliosApp) (*unstructured.Unstructured, error) {
	if heliosApp == nil {
		return nil, fmt.Errorf("heliosApp cannot be nil")
	}

	name := heliosApp.Name
	namespace := heliosApp.Namespace

	// Default gitopsPath to app name if not specified
	gitopsPath := heliosApp.Spec.GitopsPath
	if gitopsPath == "" {
		gitopsPath = name
	}

	// Default gitopsBranch to "main" if not specified
	gitopsBranch := heliosApp.Spec.GitopsBranch
	if gitopsBranch == "" {
		gitopsBranch = "main"
	}

	app := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name":      name + "-argocd",
			"namespace": "argocd", // ArgoCD Applications live in argocd namespace
			"labels": map[string]interface{}{
				common.LabelManagedBy: "helios-operator",
				common.LabelAppName:   name,
				common.LabelComponent: "argocd",
			},
		},
		"spec": map[string]interface{}{
			"project": "default",
			"source": map[string]interface{}{
				"repoURL":        heliosApp.Spec.GitopsRepo,
				"targetRevision": gitopsBranch,
				"path":           gitopsPath,
			},
			"destination": map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": namespace, // Deploy to same namespace as HeliosApp
			},
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{
					"prune":    true,
					"selfHeal": true,
				},
				"syncOptions": []string{
					"CreateNamespace=true",
				},
			},
		},
	}

	return &unstructured.Unstructured{Object: app}, nil
}
