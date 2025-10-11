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
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
)

// HeliosAppReconciler reconciles a HeliosApp object
type HeliosAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=heliosapp.helios.dev,resources=heliosapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=heliosapp.helios.dev,resources=heliosapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=heliosapp.helios.dev,resources=heliosapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HeliosApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *HeliosAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var heliosApp heliosappv1.HeliosApp
	if err := r.Get(ctx, req.NamespacedName, &heliosApp); err != nil {
		logger.Error(err, "unable to fetch HeliosApp")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	name := heliosApp.Name
	namespace := heliosApp.Namespace

	// PVC name động: lấy từ spec hoặc quy ước
	pvcName := heliosApp.Spec.PVCName
	if pvcName == "" {
		pvcName = "pvc-" + name
	}

	pipelineName := heliosApp.Spec.PipelineName
	serviceAccount := heliosApp.Spec.ServiceAccount
	githubSecret := heliosApp.Spec.WebhookSecret
	workspace := map[string]interface{}{
		"name": "shared-data",
		"persistentVolumeClaim": map[string]interface{}{
			"claimName": pvcName,
		},
	}

	eventListener, err := GenerateEventListener(
		name+"-el", namespace, name+"-trigger", name+"-trigger-binding", name+"-trigger-template", githubSecret,
	)
	if err != nil {
		logger.Error(err, "failed to generate EventListener")
		return ctrl.Result{}, err
	}
	triggerBinding, err := GenerateTriggerBinding(name+"-trigger-binding", namespace)
	if err != nil {
		logger.Error(err, "failed to generate TriggerBinding")
		return ctrl.Result{}, err
	}
	triggerTemplate, err := GenerateTriggerTemplate(
		name+"-trigger-template", namespace, name+"-pipelinerun", pipelineName, serviceAccount, workspace,
	)
	if err != nil {
		logger.Error(err, "failed to generate TriggerTemplate")
		return ctrl.Result{}, err
	}

	// Đặt owner reference để dọn dẹp tự động
	for _, obj := range []*unstructured.Unstructured{eventListener, triggerBinding, triggerTemplate} {
		obj.SetNamespace(namespace)
		if err := controllerutil.SetControllerReference(&heliosApp, obj, r.Scheme); err != nil {
			logger.Error(err, "failed to set owner reference", "name", obj.GetName())
			return ctrl.Result{}, err
		}

		// Kiểm tra resource đã tồn tại chưa
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(obj.GroupVersionKind())
		err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: obj.GetName()}, existing)
		if err != nil {
			// Nếu chưa có thì tạo mới
			if client.IgnoreNotFound(err) == nil {
				if err := r.Create(ctx, obj); err != nil {
					logger.Error(err, "failed to create Tekton resource", "name", obj.GetName())
					return ctrl.Result{}, err
				}
			} else {
				logger.Error(err, "failed to get Tekton resource", "name", obj.GetName())
				return ctrl.Result{}, err
			}
		} else {
			// Nếu đã có, so sánh spec, nếu khác thì update
			if !equalUnstructured(obj, existing) {
				obj.SetResourceVersion(existing.GetResourceVersion())
				if err := r.Update(ctx, obj); err != nil {
					logger.Error(err, "failed to update Tekton resource", "name", obj.GetName())
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// Hàm so sánh spec của hai unstructured (chỉ so sánh phần spec)
func equalUnstructured(a, b *unstructured.Unstructured) bool {
	specA, foundA, _ := unstructured.NestedMap(a.Object, "spec")
	specB, foundB, _ := unstructured.NestedMap(b.Object, "spec")

    // Nếu một trong hai không có spec thì coi như không bằng nhau
	if !foundA || !foundB {
		return false
	}
    
    // Sử dụng DeepEqual để so sánh nội dung
	return reflect.DeepEqual(specA, specB)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HeliosAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heliosappv1.HeliosApp{}).
		Complete(r)
}
