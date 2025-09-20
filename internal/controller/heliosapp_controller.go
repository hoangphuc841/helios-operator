/*
Copyright 2025.
*/

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *HeliosAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciliation loop started...")

	heliosApp := &heliosappv1.HeliosApp{}
	if err := r.Get(ctx, req.NamespacedName, heliosApp); err != nil {
		if errors.IsNotFound(err) {
			log.Info("HeliosApp not found. It may have been deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Error retrieving HeliosApp")
		return ctrl.Result{}, err
	}

	foundDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: heliosApp.Name, Namespace: heliosApp.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		dep := r.deploymentForHeliosApp(heliosApp)
		log.Info("Creating new Deployment", "Namespace", dep.Namespace, "Name", dep.Name)
		if err = r.Create(ctx, dep); err != nil {
			log.Error(err, "Failed to create new Deployment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	foundService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: heliosApp.Name, Namespace: heliosApp.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		svc := r.serviceForHeliosApp(heliosApp)
		log.Info("Creating new Service", "Namespace", svc.Namespace, "Name", svc.Name)
		if err = r.Create(ctx, svc); err != nil {
			log.Error(err, "Failed to create new Service")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	log.Info("Reconciliation loop completed successfully.")

	return ctrl.Result{}, nil
}

func (r *HeliosAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heliosappv1.HeliosApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *HeliosAppReconciler) deploymentForHeliosApp(h *heliosappv1.HeliosApp) *appsv1.Deployment {
	labels := map[string]string{"app": h.Name}

	replicas := h.Spec.Replicas
	if replicas == 0 { // Giá trị mặc định cho int32 là 0
		replicas = 1
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: h.Name, Namespace: h.Namespace},
		Spec: appsv1.DeploymentSpec{
			// Lấy địa chỉ của biến 'replicas' để truyền vào
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: h.Spec.ImageRepo,
						Name:  "app-container",
						Ports: []corev1.ContainerPort{{ContainerPort: h.Spec.Port, Name: "http"}},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(h, dep, r.Scheme)
	return dep
}

func (r *HeliosAppReconciler) serviceForHeliosApp(h *heliosappv1.HeliosApp) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: h.Name, Namespace: h.Namespace},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": h.Name},
			Ports:    []corev1.ServicePort{{Protocol: corev1.ProtocolTCP, Port: 80, TargetPort: intstr.FromInt(int(h.Spec.Port))}},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	ctrl.SetControllerReference(h, svc, r.Scheme)
	return svc
}
