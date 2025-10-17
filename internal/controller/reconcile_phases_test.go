package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
)

func TestReconcilePhaseConstants(t *testing.T) {
	t.Run("phase constants are defined correctly", func(t *testing.T) {
		assert.Equal(t, ReconcilePhase("Pipeline"), PhasePipeline)
		assert.Equal(t, ReconcilePhase("PVC"), PhasePVC)
		assert.Equal(t, ReconcilePhase("Triggers"), PhaseTriggers)
		assert.Equal(t, ReconcilePhase("ArgoCD"), PhaseArgoCD)
		assert.Equal(t, ReconcilePhase("Status"), PhaseStatus)
	})
}

func TestEqualUnstructured(t *testing.T) {
	t.Run("equal unstructured objects", func(t *testing.T) {
		// Create a mock reconciler
		mockClient := &MockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		reconciler := &HeliosAppReconciler{
			Client: mockClient,
			Scheme: mockScheme,
		}

		// Create two identical unstructured objects
		obj1 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"repoURL": "https://github.com/test/repo",
						"path":    "apps/test-app",
					},
				},
			},
		}

		obj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"repoURL": "https://github.com/test/repo",
						"path":    "apps/test-app",
					},
				},
			},
		}

		result := reconciler.equalUnstructured(obj1, obj2)
		assert.True(t, result)
	})

	t.Run("different unstructured objects", func(t *testing.T) {
		// Create a mock reconciler
		mockClient := &MockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		reconciler := &HeliosAppReconciler{
			Client: mockClient,
			Scheme: mockScheme,
		}

		// Create two different unstructured objects
		obj1 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"repoURL": "https://github.com/test/repo",
						"path":    "apps/test-app",
					},
				},
			},
		}

		obj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"repoURL": "https://github.com/other/repo",
						"path":    "apps/test-app",
					},
				},
			},
		}

		result := reconciler.equalUnstructured(obj1, obj2)
		assert.False(t, result)
	})

	t.Run("missing spec in first object", func(t *testing.T) {
		// Create a mock reconciler
		mockClient := &MockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		reconciler := &HeliosAppReconciler{
			Client: mockClient,
			Scheme: mockScheme,
		}

		// Create objects where first one is missing spec
		obj1 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				// No spec
			},
		}

		obj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"repoURL": "https://github.com/test/repo",
						"path":    "apps/test-app",
					},
				},
			},
		}

		result := reconciler.equalUnstructured(obj1, obj2)
		assert.False(t, result)
	})

	t.Run("missing spec in second object", func(t *testing.T) {
		// Create a mock reconciler
		mockClient := &MockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		reconciler := &HeliosAppReconciler{
			Client: mockClient,
			Scheme: mockScheme,
		}

		// Create objects where second one is missing spec
		obj1 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"repoURL": "https://github.com/test/repo",
						"path":    "apps/test-app",
					},
				},
			},
		}

		obj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				// No spec
			},
		}

		result := reconciler.equalUnstructured(obj1, obj2)
		assert.False(t, result)
	})

	t.Run("both objects missing spec", func(t *testing.T) {
		// Create a mock reconciler
		mockClient := &MockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		reconciler := &HeliosAppReconciler{
			Client: mockClient,
			Scheme: mockScheme,
		}

		// Create objects where both are missing spec
		obj1 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				// No spec
			},
		}

		obj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "argoproj.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": "test-app",
				},
				// No spec
			},
		}

		result := reconciler.equalUnstructured(obj1, obj2)
		assert.False(t, result)
	})
}

// MockClient is a minimal mock implementation for testing
type MockClient struct{}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return nil
}

func (m *MockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (m *MockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return nil
}

func (m *MockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

func (m *MockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (m *MockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (m *MockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (m *MockClient) Status() client.StatusWriter {
	return &MockStatusWriter{}
}

func (m *MockClient) Scheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = heliosappv1.AddToScheme(s)
	return s
}

func (m *MockClient) RESTMapper() meta.RESTMapper {
	return nil
}

func (m *MockClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (m *MockClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return true, nil
}

func (m *MockClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return nil
}

func (m *MockClient) SubResource(subResource string) client.SubResourceClient {
	return &MockSubResourceClient{}
}

// MockStatusWriter is a minimal mock implementation for testing
type MockStatusWriter struct{}

func (m *MockStatusWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	return nil
}

func (m *MockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return nil
}

func (m *MockStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	return nil
}

// MockSubResourceClient is a minimal mock implementation for testing
type MockSubResourceClient struct{}

func (m *MockSubResourceClient) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	return nil
}

func (m *MockSubResourceClient) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	return nil
}

func (m *MockSubResourceClient) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return nil
}

func (m *MockSubResourceClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	return nil
}
