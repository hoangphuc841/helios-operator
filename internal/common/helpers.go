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

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ResourceHelper provides common resource management functions.
type ResourceHelper struct {
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger
}

// NewResourceHelper creates a new ResourceHelper.
func NewResourceHelper(k8sClient client.Client, scheme *runtime.Scheme, logger logr.Logger) *ResourceHelper {
	return &ResourceHelper{
		client: k8sClient,
		scheme: scheme,
		logger: logger,
	}
}

// CreateOrUpdate creates or updates a resource.
func (h *ResourceHelper) CreateOrUpdate(ctx context.Context, obj, owner client.Object) error {
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	// Set owner reference.
	if owner != nil {
		if err := controllerutil.SetControllerReference(owner, obj, h.scheme); err != nil {
			return fmt.Errorf("failed to set controller reference: %w", err)
		}
	}

	// Try to get existing resource.
	existing, ok := obj.DeepCopyObject().(client.Object)
	if !ok {
		return errors.New("failed to deep copy object")
	}
	err := h.client.Get(ctx, key, existing)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create new resource.
			h.logger.Info("Creating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind,
				"name", obj.GetName())
			if err := h.client.Create(ctx, obj); err != nil {
				return fmt.Errorf("failed to create resource: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to get existing resource: %w", err)
	}

	// Update existing resource.
	h.logger.Info("Updating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind,
		"name", obj.GetName())
	obj.SetResourceVersion(existing.GetResourceVersion())
	if err := h.client.Update(ctx, obj); err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}
	return nil
}

// DeleteIfExists deletes a resource if it exists.
func (h *ResourceHelper) DeleteIfExists(ctx context.Context, obj client.Object) error {
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	err := h.client.Get(ctx, key, obj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil // Resource doesn't exist, nothing to delete
		}
		return fmt.Errorf("failed to get resource for deletion: %w", err)
	}

	h.logger.Info("Deleting resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind,
		"name", obj.GetName())
	if err := h.client.Delete(ctx, obj); err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

// GetResource gets a resource by name and namespace.
func (h *ResourceHelper) GetResource(ctx context.Context, obj client.Object, name,
	namespace string,
) error {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	if err := h.client.Get(ctx, key, obj); err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}
	return nil
}

// ListResources lists resources of a specific type.
func (h *ResourceHelper) ListResources(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if err := h.client.List(ctx, list, opts...); err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}
	return nil
}

// SetCommonLabels sets common labels on a resource.
func SetCommonLabels(obj metav1.Object, appName, instance string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[LabelAppName] = appName
	labels[LabelAppInstance] = instance
	labels[LabelAppVersion] = OperatorVersion
	labels[LabelAppComponent] = "application"
	labels[LabelAppPartOf] = OperatorName
	labels[LabelAppManagedBy] = OperatorName
	labels[LabelHeliosApp] = "true"
	labels[LabelHeliosAppName] = instance

	obj.SetLabels(labels)
}

// SetCommonAnnotations sets common annotations on a resource.
func SetCommonAnnotations(obj metav1.Object) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[AnnotationManagedBy] = OperatorName
	annotations[AnnotationReconcileTime] = time.Now().Format(time.RFC3339)

	obj.SetAnnotations(annotations)
}

// IsResourceReady checks if a resource is ready.
func IsResourceReady(obj client.Object) bool {
	// This is a simplified check - implement based on your resource types.
	// For example, check status conditions, ready replicas, etc.
	return true
}

// GetResourceStatus gets the status of a resource.
func GetResourceStatus(obj client.Object) string {
	// This is a simplified implementation - customize based on your needs.
	if IsResourceReady(obj) {
		return PhaseRunning
	}
	return PhasePending
}

// CreateUnstructuredObject creates an unstructured object from a map.
func CreateUnstructuredObject(apiVersion, kind, name, namespace string, spec map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: make(map[string]interface{}),
	}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	obj.SetName(name)
	obj.SetNamespace(namespace)
	obj.Object["spec"] = spec
	return obj
}

// AddFinalizer adds a finalizer to an object.
func AddFinalizer(obj metav1.Object, finalizer string) {
	finalizers := obj.GetFinalizers()
	for _, f := range finalizers {
		if f == finalizer {
			return // Finalizer already exists
		}
	}
	obj.SetFinalizers(append(finalizers, finalizer))
}

// RemoveFinalizer removes a finalizer from an object.
func RemoveFinalizer(obj metav1.Object, finalizer string) {
	finalizers := obj.GetFinalizers()
	for i, f := range finalizers {
		if f == finalizer {
			obj.SetFinalizers(append(finalizers[:i], finalizers[i+1:]...))
			return
		}
	}
}

// HasFinalizer checks if an object has a specific finalizer.
func HasFinalizer(obj metav1.Object, finalizer string) bool {
	finalizers := obj.GetFinalizers()
	for _, f := range finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}
