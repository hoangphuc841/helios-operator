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

package common_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/hoangphuc841/helios-operator/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockClient is a mock implementation of client.Client for testing.
type MockClient struct {
	mock.Mock
}

// createTestConfigMap creates a test ConfigMap with all required fields.
func createTestConfigMap(name, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       name,
			Namespace:                  namespace,
			Labels:                     make(map[string]string),
			Annotations:                make(map[string]string),
			GenerateName:               "",
			SelfLink:                   "",
			UID:                        "",
			ResourceVersion:            "",
			Generation:                 0,
			CreationTimestamp:          metav1.Time{Time: time.Time{}},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			OwnerReferences:            []metav1.OwnerReference{},
			Finalizers:                 []string{},
			ManagedFields:              []metav1.ManagedFieldsEntry{},
		},
		Immutable:  nil,
		Data:       make(map[string]string),
		BinaryData: make(map[string][]byte),
	}
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := m.Called(ctx, key, obj, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock Get error: %w", err)
	}
	return nil
}

func (m *MockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := m.Called(ctx, list, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock List error: %w", err)
	}
	return nil
}

func (m *MockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	args := m.Called(ctx, obj, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock Create error: %w", err)
	}
	return nil
}

func (m *MockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	args := m.Called(ctx, obj, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock Delete error: %w", err)
	}
	return nil
}

func (m *MockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	args := m.Called(ctx, obj, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock Update error: %w", err)
	}
	return nil
}

func (m *MockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := m.Called(ctx, obj, patch, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock Patch error: %w", err)
	}
	return nil
}

func (m *MockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	args := m.Called(ctx, obj, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock DeleteAllOf error: %w", err)
	}
	return nil
}

func (m *MockClient) Status() client.StatusWriter {
	args := m.Called()
	if args.Error(0) != nil {
		return nil
	}
	if args.Get(0) == nil {
		return nil
	}
	if statusWriter, ok := args.Get(0).(client.StatusWriter); ok {
		return statusWriter
	}
	return nil
}

func (m *MockClient) Scheme() *runtime.Scheme {
	args := m.Called()
	if args.Error(0) != nil {
		s := runtime.NewScheme()
		_ = scheme.AddToScheme(s)
		return s
	}
	if args.Get(0) == nil {
		s := runtime.NewScheme()
		_ = scheme.AddToScheme(s)
		return s
	}
	if scheme, ok := args.Get(0).(*runtime.Scheme); ok {
		return scheme
	}
	return runtime.NewScheme()
}

func (m *MockClient) RESTMapper() meta.RESTMapper {
	args := m.Called()
	if args.Error(0) != nil {
		return nil
	}
	if args.Get(0) == nil {
		return nil
	}
	if restMapper, ok := args.Get(0).(meta.RESTMapper); ok {
		return restMapper
	}
	return nil
}

func (m *MockClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	args := m.Called(obj)
	if gvk, ok := args.Get(0).(schema.GroupVersionKind); ok {
		if err := args.Error(1); err != nil {
			return gvk, fmt.Errorf("mock GroupVersionKindFor error: %w", err)
		}
		return gvk, nil
	}
	if err := args.Error(1); err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("mock GroupVersionKindFor error: %w", err)
	}
	return schema.GroupVersionKind{}, nil
}

func (m *MockClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	args := m.Called(obj)
	return args.Bool(0), args.Error(1)
}

func (m *MockClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	args := m.Called(ctx, obj, opts)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock Apply error: %w", err)
	}
	return nil
}

func (m *MockClient) SubResource(subResource string) client.SubResourceClient {
	args := m.Called(subResource)
	if args.Error(0) != nil {
		return nil
	}
	if args.Get(0) == nil {
		return nil
	}
	if subResourceClient, ok := args.Get(0).(client.SubResourceClient); ok {
		return subResourceClient
	}
	return nil
}

// MockLogger is a mock implementation of logr.Logger for testing.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Enabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	m.Called(err, msg, keysAndValues)
}

func (m *MockLogger) V(level int) logr.Logger {
	args := m.Called(level)
	if args.Error(0) != nil {
		return logr.Discard()
	}
	if args.Get(0) == nil {
		return logr.Discard()
	}
	if logger, ok := args.Get(0).(logr.Logger); ok {
		return logger
	}
	return logr.Discard()
}

func (m *MockLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	args := m.Called(keysAndValues)
	if args.Error(0) != nil {
		return logr.Discard()
	}
	if args.Get(0) == nil {
		return logr.Discard()
	}
	if logger, ok := args.Get(0).(logr.Logger); ok {
		return logger
	}
	return logr.Discard()
}

func (m *MockLogger) WithName(name string) logr.Logger {
	args := m.Called(name)
	if args.Error(0) != nil {
		return logr.Discard()
	}
	if args.Get(0) == nil {
		return logr.Discard()
	}
	if logger, ok := args.Get(0).(logr.Logger); ok {
		return logger
	}
	return logr.Discard()
}

func TestResourceHelper_CreateOrUpdate(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockClient, *MockLogger)
		expectError bool
	}{
		{
			name: "successful create",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(k8serrors.NewNotFound(schema.GroupResource{}, "test"))
				mockClient.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "successful update",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockClient.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "get error",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("get failed"))
			},
			expectError: true,
		},
		{
			name: "create error",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(k8serrors.NewNotFound(schema.GroupResource{}, "test"))
				mockClient.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("create failed"))
			},
			expectError: true,
		},
		{
			name: "update error",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockClient.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}
			mockLogger := &MockLogger{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}

			tt.setupMocks(mockClient, mockLogger)

			s := runtime.NewScheme()
			_ = scheme.AddToScheme(s)
			helper := common.NewResourceHelper(mockClient, s, logr.Discard())

			ctx := context.Background()
			obj := createTestConfigMap("test-configmap", "default")
			owner := createTestConfigMap("owner-configmap", "default")

			err := helper.CreateOrUpdate(ctx, obj, owner)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestResourceHelper_DeleteIfExists(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockClient, *MockLogger)
		expectError bool
	}{
		{
			name: "successful delete",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockClient.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "delete error",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockClient.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("delete failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}
			mockLogger := &MockLogger{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}

			tt.setupMocks(mockClient, mockLogger)

			s := runtime.NewScheme()
			_ = scheme.AddToScheme(s)
			helper := common.NewResourceHelper(mockClient, s, logr.Discard())

			ctx := context.Background()
			obj := createTestConfigMap("test-configmap", "default")

			err := helper.DeleteIfExists(ctx, obj)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestResourceHelper_GetResource(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockClient, *MockLogger)
		expectError bool
	}{
		{
			name: "successful get",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "get error",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("get failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}
			mockLogger := &MockLogger{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}

			tt.setupMocks(mockClient, mockLogger)

			s := runtime.NewScheme()
			_ = scheme.AddToScheme(s)
			helper := common.NewResourceHelper(mockClient, s, logr.Discard())

			ctx := context.Background()
			obj := createTestConfigMap("test-configmap", "default")

			err := helper.GetResource(ctx, obj, "test", "default")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestResourceHelper_ListResources(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockClient, *MockLogger)
		expectError bool
	}{
		{
			name: "successful list",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("List", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "list error",
			setupMocks: func(mockClient *MockClient, mockLogger *MockLogger) {
				mockClient.On("List", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("list failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}
			mockLogger := &MockLogger{Mock: mock.Mock{ExpectedCalls: []*mock.Call{}, Calls: []mock.Call{}}}

			tt.setupMocks(mockClient, mockLogger)

			s := runtime.NewScheme()
			_ = scheme.AddToScheme(s)
			helper := common.NewResourceHelper(mockClient, s, logr.Discard())

			ctx := context.Background()
			list := &corev1.ConfigMapList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMapList",
					APIVersion: "v1",
				},
				ListMeta: metav1.ListMeta{
					SelfLink:           "",
					ResourceVersion:    "",
					Continue:           "",
					RemainingItemCount: nil,
				},
				Items: []corev1.ConfigMap{},
			}
			opts := []client.ListOption{}

			err := helper.ListResources(ctx, list, opts...)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
