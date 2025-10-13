/*
Copyright 2025.
*/

package controller

import (
    "context"
    "reflect"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/apimachinery/pkg/util/intstr"
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

// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=triggers.tekton.dev,resources=eventlisteners;triggerbindings;triggertemplates,verbs=get;list;watch;create;update;patch;delete

func (r *HeliosAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Fetch HeliosApp
    var heliosApp heliosappv1.HeliosApp
    if err := r.Get(ctx, req.NamespacedName, &heliosApp); err != nil {
        if errors.IsNotFound(err) {
            logger.Info("HeliosApp not found. It may have been deleted.")
            return ctrl.Result{}, nil
        }
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

    // --- Tekton Triggers resources ---
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

    for _, obj := range []*unstructured.Unstructured{eventListener, triggerBinding, triggerTemplate} {
        obj.SetNamespace(namespace)
        if err := controllerutil.SetControllerReference(&heliosApp, obj, r.Scheme); err != nil {
            logger.Error(err, "failed to set owner reference", "name", obj.GetName())
            return ctrl.Result{}, err
        }

        existing := &unstructured.Unstructured{}
        existing.SetGroupVersionKind(obj.GroupVersionKind())
        err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: obj.GetName()}, existing)
        if err != nil {
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
            if !equalUnstructured(obj, existing) {
                obj.SetResourceVersion(existing.GetResourceVersion())
                if err := r.Update(ctx, obj); err != nil {
                    logger.Error(err, "failed to update Tekton resource", "name", obj.GetName())
                    return ctrl.Result{}, err
                }
            }
        }
    }

    // --- Deployment ---
    foundDeployment := &appsv1.Deployment{}
    err = r.Get(ctx, types.NamespacedName{Name: heliosApp.Name, Namespace: heliosApp.Namespace}, foundDeployment)
    if err != nil && errors.IsNotFound(err) {
        dep := r.deploymentForHeliosApp(&heliosApp)
        logger.Info("Creating new Deployment", "Namespace", dep.Namespace, "Name", dep.Name)
        if err = r.Create(ctx, dep); err != nil {
            logger.Error(err, "Failed to create new Deployment")
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    } else if err != nil {
        logger.Error(err, "Failed to get Deployment")
        return ctrl.Result{}, err
    }

    replicasChanged := *foundDeployment.Spec.Replicas != heliosApp.Spec.Replicas
    if replicasChanged {
        logger.Info("Updating existing Deployment", "ReplicasChanged", replicasChanged)
        foundDeployment.Spec.Replicas = &heliosApp.Spec.Replicas
        if err = r.Update(ctx, foundDeployment); err != nil {
            logger.Error(err, "Failed to update Deployment")
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    }

    // --- Service ---
    foundService := &corev1.Service{}
    err = r.Get(ctx, types.NamespacedName{Name: heliosApp.Name, Namespace: heliosApp.Namespace}, foundService)
    if err != nil && errors.IsNotFound(err) {
        svc := r.serviceForHeliosApp(&heliosApp)
        logger.Info("Creating new Service", "Namespace", svc.Namespace, "Name", svc.Name)
        if err = r.Create(ctx, svc); err != nil {
            logger.Error(err, "Failed to create new Service")
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    } else if err != nil {
        logger.Error(err, "Failed to get Service")
        return ctrl.Result{}, err
    }

    portChanged := foundService.Spec.Ports[0].TargetPort.IntVal != heliosApp.Spec.Port
    if portChanged {
        logger.Info("Updating existing Service", "PortChanged", portChanged)
        foundService.Spec.Ports[0].TargetPort = intstr.FromInt(int(heliosApp.Spec.Port))
        if err = r.Update(ctx, foundService); err != nil {
            logger.Error(err, "Failed to update Service")
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    }

    logger.Info("Reconciliation loop completed successfully.")
    return ctrl.Result{}, nil
}

// Hàm so sánh spec của hai unstructured (chỉ so sánh phần spec)
func equalUnstructured(a, b *unstructured.Unstructured) bool {
    specA, foundA, _ := unstructured.NestedMap(a.Object, "spec")
    specB, foundB, _ := unstructured.NestedMap(b.Object, "spec")
    if !foundA || !foundB {
        return false
    }
    return reflect.DeepEqual(specA, specB)
}

// SetupWithManager sets up the controller with the Manager.
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
    if replicas == 0 {
        replicas = 1
    }
    dep := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{Name: h.Name, Namespace: h.Namespace},
        Spec: appsv1.DeploymentSpec{
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
    if err := ctrl.SetControllerReference(h, dep, r.Scheme); err != nil {
        return nil
    }
    return dep
}

func (r *HeliosAppReconciler) serviceForHeliosApp(h *heliosappv1.HeliosApp) *corev1.Service {
    svc := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{Name: h.Name, Namespace: h.Namespace},
        Spec: corev1.ServiceSpec{
            Selector: map[string]string{"app": h.Name},
            Ports:    []corev1.ServicePort{{Protocol: corev1.ProtocolTCP, Port: 80, TargetPort: intstr.FromInt(int(h.Spec.Port))}},
            Type:     corev1.ServiceTypeNodePort,
        },
    }
    if err := ctrl.SetControllerReference(h, svc, r.Scheme); err != nil {
        return nil
    }
    return svc
}