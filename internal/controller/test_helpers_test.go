package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
)

// TestMockClient is a mock implementation of client.Client for testing
type TestMockClient struct {
	mock.Mock
}

func (m *TestMockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := m.Called(ctx, key, obj, opts)
	return args.Error(0)
}

func (m *TestMockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := m.Called(ctx, list, opts)
	return args.Error(0)
}

func (m *TestMockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *TestMockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *TestMockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *TestMockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := m.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

func (m *TestMockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *TestMockClient) Status() client.StatusWriter {
	args := m.Called()
	return args.Get(0).(client.StatusWriter)
}

func (m *TestMockClient) Scheme() *runtime.Scheme {
	args := m.Called()
	return args.Get(0).(*runtime.Scheme)
}

func (m *TestMockClient) RESTMapper() meta.RESTMapper {
	args := m.Called()
	return args.Get(0).(meta.RESTMapper)
}

func (m *TestMockClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	args := m.Called(obj)
	return args.Get(0).(schema.GroupVersionKind), args.Error(1)
}

func (m *TestMockClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	args := m.Called(obj)
	return args.Bool(0), args.Error(1)
}

func (m *TestMockClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *TestMockClient) SubResource(subResource string) client.SubResourceClient {
	args := m.Called(subResource)
	return args.Get(0).(client.SubResourceClient)
}

// TestMockStatusWriter is a mock implementation of client.StatusWriter
type TestMockStatusWriter struct {
	mock.Mock
}

func (m *TestMockStatusWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	args := m.Called(ctx, obj, subResource, opts)
	return args.Error(0)
}

func (m *TestMockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *TestMockStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	args := m.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

// TestMockSubResourceClient is a mock implementation of client.SubResourceClient
type TestMockSubResourceClient struct {
	mock.Mock
}

func (m *TestMockSubResourceClient) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	args := m.Called(ctx, obj, subResource, opts)
	return args.Error(0)
}

func (m *TestMockSubResourceClient) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	args := m.Called(ctx, obj, subResource, opts)
	return args.Error(0)
}

func (m *TestMockSubResourceClient) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *TestMockSubResourceClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	args := m.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

func TestNewTestHelper(t *testing.T) {
	t.Run("creates test helper with client and scheme", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()

		helper := NewTestHelper(mockClient, mockScheme)

		assert.NotNil(t, helper)
		assert.Equal(t, mockClient, helper.Client)
		assert.Equal(t, mockScheme, helper.Scheme)
	})
}

func TestTestHelper_CreateTestHeliosApp(t *testing.T) {
	t.Run("creates test HeliosApp successfully", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Mock the Create call
		mockClient.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"

		heliosApp, err := helper.CreateTestHeliosApp(ctx, name, namespace)

		assert.NoError(t, err)
		assert.NotNil(t, heliosApp)
		assert.Equal(t, name, heliosApp.Name)
		assert.Equal(t, namespace, heliosApp.Namespace)
		assert.Equal(t, "https://github.com/example/test-app.git", heliosApp.Spec.GitRepo)
		assert.Equal(t, "example.com/test-app", heliosApp.Spec.ImageRepo)
		assert.Equal(t, int32(8080), heliosApp.Spec.Port)
		assert.Equal(t, int32(1), heliosApp.Spec.Replicas)
		assert.Equal(t, "pipeline-sa", heliosApp.Spec.ServiceAccount)
		assert.Equal(t, "webhook-secret", heliosApp.Spec.WebhookSecret)
		assert.Equal(t, name+"-pvc", heliosApp.Spec.PVCName)

		mockClient.AssertExpectations(t)
	})

	t.Run("handles create error", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Mock the Create call to return an error
		mockClient.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"

		heliosApp, err := helper.CreateTestHeliosApp(ctx, name, namespace)

		assert.Error(t, err)
		assert.Nil(t, heliosApp)
		assert.Contains(t, err.Error(), "failed to create mock HeliosApp")

		mockClient.AssertExpectations(t)
	})
}

func TestTestHelper_CreateMockPipelineRun(t *testing.T) {
	t.Run("creates mock PipelineRun successfully", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()

		helper := NewTestHelper(mockClient, mockScheme)

		ctx := context.Background()
		name := "test-pipeline-run"
		namespace := "test-namespace"
		status := "Succeeded"

		pipelineRun, err := helper.CreateMockPipelineRun(ctx, name, namespace, status)

		assert.NoError(t, err)
		assert.NotNil(t, pipelineRun)
		assert.Equal(t, name, pipelineRun.GetName())
		assert.Equal(t, namespace, pipelineRun.GetNamespace())
		assert.Equal(t, "tekton.dev/v1", pipelineRun.GetAPIVersion())
		assert.Equal(t, "PipelineRun", pipelineRun.GetKind())
	})

	t.Run("creates mock PipelineRun with different status", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()

		helper := NewTestHelper(mockClient, mockScheme)

		ctx := context.Background()
		name := "test-pipeline-run"
		namespace := "test-namespace"
		status := "Failed"

		pipelineRun, err := helper.CreateMockPipelineRun(ctx, name, namespace, status)

		assert.NoError(t, err)
		assert.NotNil(t, pipelineRun)
		assert.Equal(t, name, pipelineRun.GetName())
		assert.Equal(t, namespace, pipelineRun.GetNamespace())
	})
}

func TestTestHelper_CreateMockDeployment(t *testing.T) {
	t.Run("creates mock Deployment successfully", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()

		helper := NewTestHelper(mockClient, mockScheme)

		// Mock the Create call
		mockClient.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		ctx := context.Background()
		name := "test-deployment"
		namespace := "test-namespace"
		replicas := int32(3)

		deployment, err := helper.CreateMockDeployment(ctx, name, namespace, replicas, replicas)

		assert.NoError(t, err)
		assert.NotNil(t, deployment)
		assert.Equal(t, name, deployment.GetName())
		assert.Equal(t, namespace, deployment.GetNamespace())
		assert.Equal(t, "apps/v1", deployment.GetAPIVersion())
		assert.Equal(t, "Deployment", deployment.GetKind())

		mockClient.AssertExpectations(t)
	})

	t.Run("creates mock Deployment with different replicas", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()

		helper := NewTestHelper(mockClient, mockScheme)

		// Mock the Create call
		mockClient.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		ctx := context.Background()
		name := "test-deployment"
		namespace := "test-namespace"
		replicas := int32(5)

		deployment, err := helper.CreateMockDeployment(ctx, name, namespace, replicas, replicas)

		assert.NoError(t, err)
		assert.NotNil(t, deployment)
		assert.Equal(t, name, deployment.GetName())
		assert.Equal(t, namespace, deployment.GetNamespace())

		mockClient.AssertExpectations(t)
	})
}

func TestTestHelper_WaitForHeliosAppCondition(t *testing.T) {
	t.Run("waits for condition successfully", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Create a mock HeliosApp with the desired condition
		heliosApp := &heliosappv1.HeliosApp{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Status: heliosappv1.HeliosAppStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
						Reason: "ReconciliationSucceeded",
					},
				},
			},
		}

		// Mock the Get call to return the HeliosApp with condition
		mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			obj := args.Get(2).(*heliosappv1.HeliosApp)
			*obj = *heliosApp
		}).Return(nil)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"
		conditionType := "Ready"
		timeout := 5 * time.Second

		err := helper.WaitForHeliosAppCondition(ctx, name, namespace, conditionType, timeout)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("times out waiting for condition", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Create a mock HeliosApp without the desired condition
		heliosApp := &heliosappv1.HeliosApp{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Status: heliosappv1.HeliosAppStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Progressing",
						Status: metav1.ConditionTrue,
						Reason: "ReconciliationInProgress",
					},
				},
			},
		}

		// Mock the Get call to return the HeliosApp without the desired condition
		mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			obj := args.Get(2).(*heliosappv1.HeliosApp)
			*obj = *heliosApp
		}).Return(nil)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"
		conditionType := "Ready"
		timeout := 100 * time.Millisecond // Short timeout for testing

		err := helper.WaitForHeliosAppCondition(ctx, name, namespace, conditionType, timeout)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout waiting for condition")
		mockClient.AssertExpectations(t)
	})
}

func TestTestHelper_GetHeliosApp(t *testing.T) {
	t.Run("gets HeliosApp successfully", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Create a mock HeliosApp
		expectedHeliosApp := &heliosappv1.HeliosApp{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "test-namespace",
			},
			Spec: heliosappv1.HeliosAppSpec{
				GitRepo: "https://github.com/example/test-app.git",
			},
		}

		// Mock the Get call
		mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			obj := args.Get(2).(*heliosappv1.HeliosApp)
			*obj = *expectedHeliosApp
		}).Return(nil)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"

		heliosApp, err := helper.GetHeliosApp(ctx, name, namespace)

		assert.NoError(t, err)
		assert.NotNil(t, heliosApp)
		assert.Equal(t, expectedHeliosApp.Name, heliosApp.Name)
		assert.Equal(t, expectedHeliosApp.Namespace, heliosApp.Namespace)
		assert.Equal(t, expectedHeliosApp.Spec.GitRepo, heliosApp.Spec.GitRepo)

		mockClient.AssertExpectations(t)
	})

	t.Run("handles get error", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Mock the Get call to return an error
		mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"

		heliosApp, err := helper.GetHeliosApp(ctx, name, namespace)

		assert.Error(t, err)
		assert.Nil(t, heliosApp)

		mockClient.AssertExpectations(t)
	})
}

func TestTestHelper_CleanupHeliosApp(t *testing.T) {
	t.Run("cleans up HeliosApp successfully", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Mock the Delete call
		mockClient.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"

		err := helper.CleanupHeliosApp(ctx, name, namespace)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("handles cleanup error", func(t *testing.T) {
		mockClient := &TestMockClient{}
		mockScheme := runtime.NewScheme()
		_ = heliosappv1.AddToScheme(mockScheme)

		helper := NewTestHelper(mockClient, mockScheme)

		// Mock the Delete call to return an error
		mockClient.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		ctx := context.Background()
		name := "test-app"
		namespace := "test-namespace"

		err := helper.CleanupHeliosApp(ctx, name, namespace)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete HeliosApp")

		mockClient.AssertExpectations(t)
	})
}

func TestCreateTestLogger(t *testing.T) {
	t.Run("creates test logger successfully", func(t *testing.T) {
		logger := CreateTestLogger()

		assert.NotNil(t, logger)
		// The logger should be a valid logr.Logger
		// We can't easily test the internal state, but we can verify it's not nil
	})
}
