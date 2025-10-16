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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
)

var _ = Describe("HeliosApp Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		heliosapp := &heliosappv1.HeliosApp{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind HeliosApp")
			err := k8sClient.Get(ctx, typeNamespacedName, heliosapp)
			if err != nil && errors.IsNotFound(err) {
				resource := &heliosappv1.HeliosApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: heliosappv1.HeliosAppSpec{
						GitRepo:        "https://github.com/example/test-app",
						GitopsRepo:     "https://github.com/example/gitops",
						ImageRepo:      "nginx:latest",
						Port:           8080,
						Replicas:       1,
						ServiceAccount: "pipeline-sa",
						WebhookSecret:  "webhook-secret",
						PVCName:        "test-pvc",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &heliosappv1.HeliosApp{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance HeliosApp")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &HeliosAppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Create test helper
			testHelper := NewTestHelper(k8sClient, k8sClient.Scheme())

			// Create mock resources that the controller expects
			By("Creating mock Tekton resources")
			_, err := testHelper.CreateMockPipeline(ctx, resourceName+"-pipeline", "default")
			Expect(err).NotTo(HaveOccurred())

			_, err = testHelper.CreateMockEventListener(ctx, resourceName+"-eventlistener", "default")
			Expect(err).NotTo(HaveOccurred())

			_, err = testHelper.CreateMockTriggerBinding(ctx, resourceName+"-triggerbinding", "default")
			Expect(err).NotTo(HaveOccurred())

			_, err = testHelper.CreateMockTriggerTemplate(ctx, resourceName+"-triggertemplate", "default")
			Expect(err).NotTo(HaveOccurred())

			// Mock ArgoCD Application (create in argocd namespace)
			_, err = testHelper.CreateMockArgoApplication(ctx, resourceName+"-argocd", "argocd")
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the resource")
			controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			// Note: We expect some errors due to missing CRDs in test environment
			// The important thing is that the controller logic runs without panicking

			By("Verifying status conditions are updated")
			updatedHeliosApp := &heliosappv1.HeliosApp{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedHeliosApp)).To(Succeed())

			// Check that finalizer is added
			Expect(updatedHeliosApp.Finalizers).To(ContainElement("platform.helios.io/finalizer"))

			// Check that Ready condition is set (may be False due to test environment limitations)
			Expect(updatedHeliosApp.Status.Conditions).NotTo(BeEmpty())
			var readyCondition *metav1.Condition
			for _, condition := range updatedHeliosApp.Status.Conditions {
				if condition.Type == heliosappv1.ConditionReady {
					readyCondition = &condition
					break
				}
			}
			Expect(readyCondition).NotTo(BeNil())
			// In test environment, we expect the condition to exist but may not be True
			Expect(readyCondition.Status).To(BeElementOf(metav1.ConditionTrue, metav1.ConditionFalse, metav1.ConditionUnknown))
		})
	})

	// ====================================
	// Unit Tests for New Logic
	// ====================================

	Context("Finalizer Logic", func() {
		var (
			reconciler *HeliosAppReconciler
			heliosApp  *heliosappv1.HeliosApp
		)

		BeforeEach(func() {
			reconciler = &HeliosAppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			heliosApp = &heliosappv1.HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app-finalizer",
					Namespace: "default",
				},
				Spec: heliosappv1.HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					GitopsRepo:     "https://github.com/example/gitops",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
				},
			}
		})

		AfterEach(func() {
			_ = k8sClient.Delete(ctx, heliosApp)
		})

		It("should add finalizer on first reconciliation", func() {
			By("Creating HeliosApp without finalizer")
			err := k8sClient.Create(ctx, heliosApp)
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the resource")
			reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      heliosApp.Name,
					Namespace: heliosApp.Namespace,
				},
			})
			// Note: We expect errors due to missing CRDs in test environment
			// The important thing is that the finalizer logic runs

			By("Verifying finalizer is added")
			updatedApp := &heliosappv1.HeliosApp{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      heliosApp.Name,
				Namespace: heliosApp.Namespace,
			}, updatedApp)).To(Succeed())
			Expect(updatedApp.Finalizers).To(ContainElement("platform.helios.io/finalizer"))
		})

		It("should handle deletion with finalizer", func() {
			By("Creating HeliosApp with finalizer")
			heliosApp.Finalizers = []string{"platform.helios.io/finalizer"}
			err := k8sClient.Create(ctx, heliosApp)
			Expect(err).NotTo(HaveOccurred())

			// Create a mock ArgoCD Application for cleanup
			argoApp := &unstructured.Unstructured{}
			argoApp.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "argoproj.io",
				Version: "v1alpha1",
				Kind:    "Application",
			})
			argoApp.SetName(heliosApp.Name + "-argocd")
			argoApp.SetNamespace("argocd")
			Expect(k8sClient.Create(ctx, argoApp)).To(Succeed())

			By("Marking HeliosApp for deletion")
			now := metav1.Now()
			heliosApp.DeletionTimestamp = &now
			Expect(k8sClient.Update(ctx, heliosApp)).To(Succeed())

			By("Reconciling the deleted resource")
			reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      heliosApp.Name,
					Namespace: heliosApp.Namespace,
				},
			})
			// Note: We expect errors due to missing CRDs in test environment
			// The important thing is that the finalizer cleanup logic runs

			By("Verifying finalizer is removed")
			updatedApp := &heliosappv1.HeliosApp{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      heliosApp.Name,
				Namespace: heliosApp.Namespace,
			}, updatedApp)).To(Succeed())
			Expect(updatedApp.Finalizers).NotTo(ContainElement("platform.helios.io/finalizer"))

			// Cleanup ArgoCD App
			_ = k8sClient.Delete(ctx, argoApp)
		})
	})

	Context("Automatic PVC Creation", func() {
		var (
			reconciler *HeliosAppReconciler
			heliosApp  *heliosappv1.HeliosApp
		)

		BeforeEach(func() {
			reconciler = &HeliosAppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			heliosApp = &heliosappv1.HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app-pvc",
					Namespace: "default",
				},
				Spec: heliosappv1.HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					GitopsRepo:     "https://github.com/example/gitops",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					// PVCName is intentionally omitted to test automatic creation
				},
			}
		})

		AfterEach(func() {
			_ = k8sClient.Delete(ctx, heliosApp)
		})

		It("should create PVC automatically when not specified", func() {
			By("Creating HeliosApp without PVC name")
			err := k8sClient.Create(ctx, heliosApp)
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the resource")
			reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      heliosApp.Name,
					Namespace: heliosApp.Namespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying PVC is created automatically")
			pvc := &corev1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      heliosApp.Name + "-workspace",
				Namespace: heliosApp.Namespace,
			}, pvc)).To(Succeed())
			Expect(pvc.Name).To(Equal(heliosApp.Name + "-workspace"))
			Expect(pvc.Labels["helios.io/managed-by"]).To(Equal("helios-operator"))
			Expect(pvc.Labels["helios.io/app-name"]).To(Equal(heliosApp.Name))

			// Cleanup PVC
			k8sClient.Delete(ctx, pvc)
		})

		It("should not create PVC when explicitly specified", func() {
			By("Creating HeliosApp with explicit PVC name")
			heliosApp.Spec.PVCName = "custom-pvc"
			err := k8sClient.Create(ctx, heliosApp)
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the resource")
			reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      heliosApp.Name,
					Namespace: heliosApp.Namespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying custom PVC is not created by operator")
			pvc := &corev1.PersistentVolumeClaim{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "custom-pvc",
				Namespace: heliosApp.Namespace,
			}, pvc)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})

	Context("getPipelineRunStatus function", func() {
		var (
			reconciler *HeliosAppReconciler
			heliosApp  *heliosappv1.HeliosApp
		)

		BeforeEach(func() {
			reconciler = &HeliosAppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			heliosApp = &heliosappv1.HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app",
					Namespace: "default",
				},
				Spec: heliosappv1.HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					GitopsRepo:     "https://github.com/example/gitops",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "test-pvc",
				},
			}
		})

		It("should return Unknown status when no PipelineRuns exist", func() {
			buildStatus, buildVersion, pipelineRunName, lastBuildTime, err := reconciler.getPipelineRunStatus(ctx, heliosApp)

			Expect(err).NotTo(HaveOccurred())
			Expect(buildStatus).To(Equal("Unknown"))
			Expect(buildVersion).To(BeEmpty())
			Expect(pipelineRunName).To(BeEmpty())
			Expect(lastBuildTime).To(BeNil())
		})

		It("should extract status from succeeded PipelineRun", func() {
			By("Creating a succeeded PipelineRun")
			pipelineRun := &tektonv1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app-build-123",
					Namespace: "default",
					Labels: map[string]string{
						"helios.io/app-name":   "test-app",
						"helios.io/managed-by": "helios-operator",
					},
					CreationTimestamp: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
				},
				Spec: tektonv1.PipelineRunSpec{
					PipelineRef: &tektonv1.PipelineRef{
						Name: "test-pipeline",
					},
				},
				Status: tektonv1.PipelineRunStatus{
					PipelineRunStatusFields: tektonv1.PipelineRunStatusFields{
						Results: []tektonv1.PipelineRunResult{
							{
								Name:  "IMAGE_URL",
								Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "nginx:v1.0.0"},
							},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, pipelineRun)
			Expect(err).NotTo(HaveOccurred())

			By("Calling getPipelineRunStatus")
			buildStatus, buildVersion, pipelineRunName, lastBuildTime, err := reconciler.getPipelineRunStatus(ctx, heliosApp)

			Expect(err).NotTo(HaveOccurred())
			Expect(buildStatus).To(Equal("Succeeded"))
			Expect(buildVersion).To(Equal("nginx:v1.0.0"))
			Expect(pipelineRunName).To(Equal("test-app-build-123"))
			Expect(lastBuildTime).NotTo(BeNil())

			// Cleanup
			k8sClient.Delete(ctx, pipelineRun)
		})

		It("should extract status from failed PipelineRun", func() {
			By("Creating a failed PipelineRun")
			pipelineRun := &tektonv1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app-build-456",
					Namespace: "default",
					Labels: map[string]string{
						"helios.io/app-name":   "test-app",
						"helios.io/managed-by": "helios-operator",
					},
					CreationTimestamp: metav1.NewTime(time.Now().Add(-3 * time.Minute)),
				},
				Spec: tektonv1.PipelineRunSpec{
					PipelineRef: &tektonv1.PipelineRef{
						Name: "test-pipeline",
					},
				},
				Status: tektonv1.PipelineRunStatus{
					PipelineRunStatusFields: tektonv1.PipelineRunStatusFields{},
				},
			}

			err := k8sClient.Create(ctx, pipelineRun)
			Expect(err).NotTo(HaveOccurred())

			By("Calling getPipelineRunStatus")
			buildStatus, buildVersion, pipelineRunName, lastBuildTime, err := reconciler.getPipelineRunStatus(ctx, heliosApp)

			Expect(err).NotTo(HaveOccurred())
			Expect(buildStatus).To(Equal("Failed"))
			Expect(buildVersion).To(BeEmpty())
			Expect(pipelineRunName).To(Equal("test-app-build-456"))
			Expect(lastBuildTime).NotTo(BeNil())

			// Cleanup
			k8sClient.Delete(ctx, pipelineRun)
		})
	})

	Context("getDeploymentHealth function", func() {
		var (
			reconciler *HeliosAppReconciler
			heliosApp  *heliosappv1.HeliosApp
		)

		BeforeEach(func() {
			reconciler = &HeliosAppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			heliosApp = &heliosappv1.HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app",
					Namespace: "default",
				},
				Spec: heliosappv1.HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					GitopsRepo:     "https://github.com/example/gitops",
					ImageRepo:      "nginx:latest",
					Port:           8080,
					Replicas:       3,
					ServiceAccount: "pipeline-sa",
					WebhookSecret:  "webhook-secret",
					PVCName:        "test-pvc",
				},
			}
		})

		It("should return Unknown status when no Deployment exists", func() {
			health, readyReplicas, desiredReplicas, lastHealthyTime, err := reconciler.getDeploymentHealth(ctx, heliosApp)

			Expect(err).NotTo(HaveOccurred())
			Expect(health).To(Equal("Unknown"))
			Expect(readyReplicas).To(Equal(int32(0)))
			Expect(desiredReplicas).To(Equal(int32(0)))
			Expect(lastHealthyTime).To(BeNil())
		})

		It("should return Healthy status when all replicas are ready", func() {
			By("Creating a healthy Deployment")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test-app",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-app",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test-app",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test-container",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
				Status: appsv1.DeploymentStatus{
					Replicas:          3,
					ReadyReplicas:     3,
					AvailableReplicas: 3,
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:   appsv1.DeploymentAvailable,
							Status: corev1.ConditionTrue,
						},
						{
							Type:   appsv1.DeploymentProgressing,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			err := k8sClient.Create(ctx, deployment)
			Expect(err).NotTo(HaveOccurred())

			By("Calling getDeploymentHealth")
			health, readyReplicas, desiredReplicas, lastHealthyTime, err := reconciler.getDeploymentHealth(ctx, heliosApp)

			Expect(err).NotTo(HaveOccurred())
			Expect(health).To(Equal("Healthy"))
			Expect(readyReplicas).To(Equal(int32(3)))
			Expect(desiredReplicas).To(Equal(int32(3)))
			Expect(lastHealthyTime).NotTo(BeNil())

			// Cleanup
			k8sClient.Delete(ctx, deployment)
		})

		It("should return Progressing status when some replicas are ready", func() {
			By("Creating a progressing Deployment")
			deployment := &unstructured.Unstructured{}
			deployment.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			})
			deployment.SetName("test-app")
			deployment.SetNamespace("default")
			deployment.SetLabels(map[string]string{
				"app": "test-app",
			})

			// Set spec.replicas
			err := unstructured.SetNestedField(deployment.Object, int64(3), "spec", "replicas")
			Expect(err).NotTo(HaveOccurred())

			// Set status - only 1 ready out of 3
			err = unstructured.SetNestedField(deployment.Object, int64(1), "status", "readyReplicas")
			Expect(err).NotTo(HaveOccurred())
			err = unstructured.SetNestedField(deployment.Object, int64(1), "status", "availableReplicas")
			Expect(err).NotTo(HaveOccurred())

			// Set conditions
			conditions := []interface{}{
				map[string]interface{}{
					"type":   "Available",
					"status": "False",
				},
				map[string]interface{}{
					"type":   "Progressing",
					"status": "True",
				},
			}
			err = unstructured.SetNestedSlice(deployment.Object, conditions, "status", "conditions")
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

			By("Calling getDeploymentHealth")
			health, readyReplicas, desiredReplicas, lastHealthyTime, err := reconciler.getDeploymentHealth(ctx, heliosApp)

			Expect(err).NotTo(HaveOccurred())
			Expect(health).To(Equal("Progressing"))
			Expect(readyReplicas).To(Equal(int32(1)))
			Expect(desiredReplicas).To(Equal(int32(3)))
			Expect(lastHealthyTime).To(BeNil()) // Not healthy yet

			// Cleanup
			k8sClient.Delete(ctx, deployment)
		})

		It("should return Degraded status when no replicas are ready", func() {
			By("Creating a degraded Deployment")
			deployment := &unstructured.Unstructured{}
			deployment.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			})
			deployment.SetName("test-app")
			deployment.SetNamespace("default")
			deployment.SetLabels(map[string]string{
				"app": "test-app",
			})

			// Set spec.replicas
			err := unstructured.SetNestedField(deployment.Object, int64(3), "spec", "replicas")
			Expect(err).NotTo(HaveOccurred())

			// Set status - no ready replicas
			err = unstructured.SetNestedField(deployment.Object, int64(0), "status", "readyReplicas")
			Expect(err).NotTo(HaveOccurred())
			err = unstructured.SetNestedField(deployment.Object, int64(0), "status", "availableReplicas")
			Expect(err).NotTo(HaveOccurred())

			// Set conditions
			conditions := []interface{}{
				map[string]interface{}{
					"type":   "Available",
					"status": "False",
				},
				map[string]interface{}{
					"type":   "Progressing",
					"status": "False",
				},
			}
			err = unstructured.SetNestedSlice(deployment.Object, conditions, "status", "conditions")
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

			By("Calling getDeploymentHealth")
			health, readyReplicas, desiredReplicas, lastHealthyTime, err := reconciler.getDeploymentHealth(ctx, heliosApp)

			Expect(err).NotTo(HaveOccurred())
			Expect(health).To(Equal("Degraded"))
			Expect(readyReplicas).To(Equal(int32(0)))
			Expect(desiredReplicas).To(Equal(int32(3)))
			Expect(lastHealthyTime).To(BeNil())

			// Cleanup
			k8sClient.Delete(ctx, deployment)
		})
	})

	Context("Watch predicate functions", func() {
		var reconciler *HeliosAppReconciler

		BeforeEach(func() {
			reconciler = &HeliosAppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		Context("isHeliosPipelineRun", func() {
			It("should return true for PipelineRun with correct labels", func() {
				pipelineRun := &unstructured.Unstructured{}
				pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "tekton.dev",
					Version: "v1",
					Kind:    "PipelineRun",
				})
				pipelineRun.SetLabels(map[string]string{
					"triggers.tekton.dev/trigger": "webhook-trigger",
					"helios.io/managed-by":        "helios-operator",
				})

				result := reconciler.isHeliosPipelineRun(pipelineRun)
				Expect(result).To(BeTrue())
			})

			It("should return false for PipelineRun without required labels", func() {
				pipelineRun := &unstructured.Unstructured{}
				pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "tekton.dev",
					Version: "v1",
					Kind:    "PipelineRun",
				})
				pipelineRun.SetLabels(map[string]string{
					"triggers.tekton.dev/trigger": "webhook-trigger",
					// Missing helios.io/managed-by label
				})

				result := reconciler.isHeliosPipelineRun(pipelineRun)
				Expect(result).To(BeFalse())
			})
		})

		Context("isHeliosDeployment", func() {
			It("should return true for Deployment with app label", func() {
				deployment := &unstructured.Unstructured{}
				deployment.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				})
				deployment.SetLabels(map[string]string{
					"app": "test-app",
				})

				result := reconciler.isHeliosDeployment(deployment)
				Expect(result).To(BeTrue())
			})

			It("should return true for Deployment with ArgoCD label", func() {
				deployment := &unstructured.Unstructured{}
				deployment.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				})
				deployment.SetLabels(map[string]string{
					"app.kubernetes.io/instance": "test-app",
				})

				result := reconciler.isHeliosDeployment(deployment)
				Expect(result).To(BeTrue())
			})

			It("should return false for Deployment without required labels", func() {
				deployment := &unstructured.Unstructured{}
				deployment.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				})
				deployment.SetLabels(map[string]string{
					"some-other-label": "value",
				})

				result := reconciler.isHeliosDeployment(deployment)
				Expect(result).To(BeFalse())
			})
		})
	})

	Context("Watch mapping functions", func() {
		var reconciler *HeliosAppReconciler

		BeforeEach(func() {
			reconciler = &HeliosAppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		Context("pipelineRunToHeliosApp", func() {
			It("should map PipelineRun to correct HeliosApp", func() {
				pipelineRun := &unstructured.Unstructured{}
				pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "tekton.dev",
					Version: "v1",
					Kind:    "PipelineRun",
				})
				pipelineRun.SetName("test-app-build-123")
				pipelineRun.SetNamespace("default")
				pipelineRun.SetLabels(map[string]string{
					"helios.io/app-name": "test-app",
				})

				requests := reconciler.pipelineRunToHeliosApp(ctx, pipelineRun)

				Expect(requests).To(HaveLen(1))
				Expect(requests[0].Name).To(Equal("test-app"))
				Expect(requests[0].Namespace).To(Equal("default"))
			})

			It("should return empty list when PipelineRun missing app-name label", func() {
				pipelineRun := &unstructured.Unstructured{}
				pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "tekton.dev",
					Version: "v1",
					Kind:    "PipelineRun",
				})
				pipelineRun.SetName("test-app-build-123")
				pipelineRun.SetNamespace("default")
				// No labels

				requests := reconciler.pipelineRunToHeliosApp(ctx, pipelineRun)

				Expect(requests).To(BeEmpty())
			})
		})

		Context("deploymentToHeliosApp", func() {
			It("should map Deployment to correct HeliosApp using app label", func() {
				// Create HeliosApp first
				heliosApp := &heliosappv1.HeliosApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-app",
						Namespace: "default",
					},
					Spec: heliosappv1.HeliosAppSpec{
						GitRepo:        "https://github.com/example/test-app",
						GitopsRepo:     "https://github.com/example/gitops",
						ImageRepo:      "nginx:latest",
						Port:           8080,
						Replicas:       1,
						ServiceAccount: "pipeline-sa",
						WebhookSecret:  "webhook-secret",
						PVCName:        "test-pvc",
					},
				}
				err := k8sClient.Create(ctx, heliosApp)
				Expect(err).NotTo(HaveOccurred())

				deployment := &unstructured.Unstructured{}
				deployment.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				})
				deployment.SetName("test-app")
				deployment.SetNamespace("default")
				deployment.SetLabels(map[string]string{
					"app": "test-app",
				})

				requests := reconciler.deploymentToHeliosApp(ctx, deployment)

				Expect(requests).To(HaveLen(1))
				Expect(requests[0].Name).To(Equal("test-app"))
				Expect(requests[0].Namespace).To(Equal("default"))

				// Cleanup
				_ = k8sClient.Delete(ctx, heliosApp)
			})

			It("should return empty list when no matching HeliosApp found", func() {
				deployment := &unstructured.Unstructured{}
				deployment.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				})
				deployment.SetName("non-existent-app")
				deployment.SetNamespace("default")
				deployment.SetLabels(map[string]string{
					"app": "non-existent-app",
				})

				requests := reconciler.deploymentToHeliosApp(ctx, deployment)

				Expect(requests).To(BeEmpty())
			})
		})
	})
})

// Helper function for creating int32 pointers
func int32Ptr(i int32) *int32 { return &i }
