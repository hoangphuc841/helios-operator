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

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"github.com/hoangphuc841/helios-operator/internal/common"
)

// createTestHeliosApp creates a test HeliosApp with the specified namespace
func createTestHeliosApp(name, namespace string) *heliosappv1.HeliosApp {
	return &heliosappv1.HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: heliosappv1.HeliosAppSpec{
			GitRepo:        "https://github.com/example/test-app",
			ImageRepo:      "test-registry.com/test-app",
			Port:           8080,
			ServiceAccount: "test-sa",
			WebhookSecret:  "test-secret",
			GitopsRepo:     "https://github.com/example/test-app-manifests",
			GitopsPath:     "apps/test-app",
			GitopsBranch:   "main",
		},
	}
}

// createExpectedArgoApp creates the expected ArgoCD Application structure
func createExpectedArgoApp(name, destinationNamespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "argocd",
				"labels": map[string]interface{}{
					common.LabelManagedBy: "helios-operator",
					common.LabelAppName:   name,
				},
			},
			"spec": map[string]interface{}{
				"project": "default",
				"source": map[string]interface{}{
					"repoURL":        "https://github.com/example/test-app-manifests",
					"targetRevision": "main",
					"path":           "apps/test-app",
				},
				"destination": map[string]interface{}{
					"server":    "https://kubernetes.default.svc",
					"namespace": destinationNamespace,
				},
				"syncPolicy": map[string]interface{}{
					"automated": map[string]interface{}{
						"prune":    true,
						"selfHeal": true,
					},
					"syncOptions": []string{
						"CreateNamespace=true",
						"PrunePropagationPolicy=foreground",
						"PruneLast=true",
					},
				},
			},
		},
	}
}

func TestGenerateArgoApplication(t *testing.T) {
	tests := []struct {
		name      string
		heliosApp *heliosappv1.HeliosApp
		expected  *unstructured.Unstructured
	}{
		{
			name:      "basic helios app",
			heliosApp: createTestHeliosApp("test-app", "default"),
			expected:  createExpectedArgoApp("test-app", "default"),
		},
		{
			name:      "helios app with custom namespace",
			heliosApp: createTestHeliosApp("test-app", "production"),
			expected:  createExpectedArgoApp("test-app", "production"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateArgoApplication(tt.heliosApp)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check basic structure
			assert.Equal(t, "argoproj.io/v1alpha1", result.GetAPIVersion())
			assert.Equal(t, "Application", result.GetKind())
			assert.Equal(t, tt.heliosApp.Name+"-argocd", result.GetName())
			assert.Equal(t, "argocd", result.GetNamespace())

			// Check labels
			labels := result.GetLabels()
			assert.NotEmpty(t, labels)

			// Check spec fields directly without NestedMap
			spec, found := result.Object["spec"]
			assert.True(t, found)
			assert.NotNil(t, spec)

			specMap, ok := spec.(map[string]interface{})
			assert.True(t, ok)

			// Check source
			source, found := specMap["source"]
			assert.True(t, found)
			sourceMap, ok := source.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, tt.heliosApp.Spec.GitopsRepo, sourceMap["repoURL"])
			assert.Equal(t, tt.heliosApp.Spec.GitopsBranch, sourceMap["targetRevision"])
			assert.Equal(t, tt.heliosApp.Spec.GitopsPath, sourceMap["path"])

			// Check destination
			destination, found := specMap["destination"]
			assert.True(t, found)
			destMap, ok := destination.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "https://kubernetes.default.svc", destMap["server"])
			assert.Equal(t, tt.heliosApp.Namespace, destMap["namespace"])
		})
	}
}

func TestGenerateArgoApplication_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		heliosApp *heliosappv1.HeliosApp
		wantErr   bool
	}{
		{
			name:      "nil helios app",
			heliosApp: nil,
			wantErr:   true,
		},
		{
			name:      "empty helios app",
			heliosApp: &heliosappv1.HeliosApp{},
			wantErr:   false, // Should not error, just create with empty values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateArgoApplication(tt.heliosApp)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
