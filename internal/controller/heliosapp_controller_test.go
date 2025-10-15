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

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})

	// ====================================
	// Unit Tests for New Logic
	// ====================================

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
			pipelineRun := &unstructured.Unstructured{}
			pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "tekton.dev",
				Version: "v1",
				Kind:    "PipelineRun",
			})
			pipelineRun.SetName("test-app-build-123")
			pipelineRun.SetNamespace("default")
			pipelineRun.SetLabels(map[string]string{
				"helios.io/app-name":   "test-app",
				"helios.io/managed-by": "helios-operator",
			})
			pipelineRun.SetCreationTimestamp(metav1.NewTime(time.Now().Add(-5 * time.Minute)))

			// Set PipelineRun status
			conditions := []interface{}{
				map[string]interface{}{
					"type":   "Succeeded",
					"status": "True",
					"reason": "Succeeded",
				},
			}
			results := []interface{}{
				map[string]interface{}{
					"name":  "IMAGE_URL",
					"value": "nginx:v1.0.0",
				},
			}

			err := unstructured.SetNestedSlice(pipelineRun.Object, conditions, "status", "conditions")
			Expect(err).NotTo(HaveOccurred())
			err = unstructured.SetNestedSlice(pipelineRun.Object, results, "status", "results")
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, pipelineRun)
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
			pipelineRun := &unstructured.Unstructured{}
			pipelineRun.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "tekton.dev",
				Version: "v1",
				Kind:    "PipelineRun",
			})
			pipelineRun.SetName("test-app-build-456")
			pipelineRun.SetNamespace("default")
			pipelineRun.SetLabels(map[string]string{
				"helios.io/app-name":   "test-app",
				"helios.io/managed-by": "helios-operator",
			})
			pipelineRun.SetCreationTimestamp(metav1.NewTime(time.Now().Add(-3 * time.Minute)))

			// Set PipelineRun status to failed
			conditions := []interface{}{
				map[string]interface{}{
					"type":   "Succeeded",
					"status": "False",
					"reason": "Failed",
				},
			}

			err := unstructured.SetNestedSlice(pipelineRun.Object, conditions, "status", "conditions")
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, pipelineRun)
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

			// Set status
			err = unstructured.SetNestedField(deployment.Object, int64(3), "status", "readyReplicas")
			Expect(err).NotTo(HaveOccurred())
			err = unstructured.SetNestedField(deployment.Object, int64(3), "status", "availableReplicas")
			Expect(err).NotTo(HaveOccurred())

			// Set conditions
			conditions := []interface{}{
				map[string]interface{}{
					"type":   "Available",
					"status": "True",
				},
				map[string]interface{}{
					"type":   "Progressing",
					"status": "True",
				},
			}
			err = unstructured.SetNestedSlice(deployment.Object, conditions, "status", "conditions")
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, deployment)
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

			err = k8sClient.Create(ctx, deployment)
			Expect(err).NotTo(HaveOccurred())

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

			err = k8sClient.Create(ctx, deployment)
			Expect(err).NotTo(HaveOccurred())

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

				Expect(requests).To(HaveLen(0))
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
				k8sClient.Delete(ctx, heliosApp)
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

				Expect(requests).To(HaveLen(0))
			})
		})
	})
})
