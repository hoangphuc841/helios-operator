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

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/hoangphuc841/helios-operator/test/utils"
)

// namespace where the project is deployed in.
const namespace = "helios-operator-system"

// serviceAccountName created for the project.
const serviceAccountName = "helios-operator-controller-manager"

// metricsServiceName is the name of the metrics service of the project.
const metricsServiceName = "helios-operator-controller-manager-metrics-service"

// metricsRoleBindingName is the name of the RBAC that will be created to allow get the metrics data.
const metricsRoleBindingName = "helios-operator-metrics-binding"

var _ = Describe("Manager", Ordered, func() {
	var controllerPodName string

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to create namespace")

		By("labeling the namespace to enforce the restricted security policy")
		cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
			"pod-security.kubernetes.io/enforce=restricted")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")

		By("installing CRDs")
		cmd = exec.Command("make", "install")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to install CRDs")

		By("deploying the controller-manager")
		cmd = exec.Command("make", "deploy", "IMG="+projectImage)
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")

		By("creating webhook certificate")
		certYAML := `apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: helios-operator-webhook-cert
  namespace: ` + namespace + `
spec:
  secretName: helios-operator-webhook-cert
  issuerRef:
    name: selfsigned-issuer
    kind: Issuer
  dnsNames:
  - webhook-service.` + namespace + `.svc
  - webhook-service.` + namespace + `.svc.cluster.local
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: ` + namespace + `
spec:
  selfSigned: {}`

		cmd = exec.Command("kubectl", "apply", "-f", "-")
		cmd.Stdin = strings.NewReader(certYAML)
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to create webhook certificate")
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("cleaning up the curl pod for metrics")
		cmd := exec.Command("kubectl", "delete", "pod", "curl-metrics", "-n", namespace)
		_, _ = utils.Run(cmd)

		By("undeploying the controller-manager")
		cmd = exec.Command("make", "undeploy")
		_, _ = utils.Run(cmd)

		By("uninstalling CRDs")
		cmd = exec.Command("make", "uninstall")
		_, _ = utils.Run(cmd)

		By("removing manager namespace")
		cmd = exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller manager pod logs")
			cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
			controllerLogs, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", controllerLogs)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
			}

			By("Fetching Kubernetes events")
			cmd = exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
			eventsOutput, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
			}

			By("Fetching curl-metrics logs")
			cmd = exec.Command("kubectl", "logs", "curl-metrics", "-n", namespace)
			metricsOutput, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Metrics logs:\n %s", metricsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get curl-metrics logs: %s", err)
			}

			By("Fetching controller manager pod description")
			cmd = exec.Command("kubectl", "describe", "pod", controllerPodName, "-n", namespace)
			podDescription, err := utils.Run(cmd)
			if err == nil {
				fmt.Println("Pod description:\n", podDescription)
			} else {
				fmt.Println("Failed to describe controller pod")
			}
		}
	})

	SetDefaultEventuallyTimeout(5 * time.Minute)
	SetDefaultEventuallyPollingInterval(2 * time.Second)

	Context("Manager", func() {
		It("should run successfully", func() {
			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func(g Gomega) {
				// Get the name of the controller-manager pod
				cmd := exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
				podNames := utils.GetNonEmptyLines(podOutput)
				g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
				controllerPodName = podNames[0]
				g.Expect(controllerPodName).To(ContainSubstring("controller-manager"))

				// Validate the pod's status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"), "Incorrect controller-manager pod status")
			}
			Eventually(verifyControllerUp).Should(Succeed())
		})

		It("should ensure the metrics endpoint is serving metrics", func() {
			By("creating a ClusterRoleBinding for the service account to allow access to metrics")
			cmd := exec.Command("kubectl", "create", "clusterrolebinding", metricsRoleBindingName,
				"--clusterrole=helios-operator-metrics-reader",
				fmt.Sprintf("--serviceaccount=%s:%s", namespace, serviceAccountName),
			)
			_, err := utils.Run(cmd)
			// Ignore error if ClusterRoleBinding already exists
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				Expect(err).NotTo(HaveOccurred(), "Failed to create ClusterRoleBinding")
			}

			By("validating that the metrics service is available")
			cmd = exec.Command("kubectl", "get", "service", metricsServiceName, "-n", namespace)
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Metrics service should exist")

			By("getting the service account token")
			token, err := serviceAccountToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())

			By("waiting for the metrics endpoint to be ready")
			verifyMetricsEndpointReady := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "endpoints", metricsServiceName, "-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("8443"), "Metrics endpoint is not ready")
			}
			Eventually(verifyMetricsEndpointReady).Should(Succeed())

			By("verifying that the controller manager is serving the metrics server")
			verifyMetricsServerStarted := func(g Gomega) {
				cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("controller-runtime.metrics\tServing metrics server"),
					"Metrics server not yet started")
			}
			Eventually(verifyMetricsServerStarted).Should(Succeed())

			By("creating the curl-metrics pod to access the metrics endpoint")
			cmd = exec.Command("kubectl", "run", "curl-metrics", "--restart=Never",
				"--namespace", namespace,
				"--image=curlimages/curl:latest",
				"--overrides",
				fmt.Sprintf(`{
					"spec": {
						"containers": [{
							"name": "curl",
							"image": "curlimages/curl:latest",
							"command": ["/bin/sh", "-c"],
							"args": ["curl -v -k -H 'Authorization: Bearer %s' https://%s.%s.svc.cluster.local:8443/metrics"],
							"securityContext": {
								"allowPrivilegeEscalation": false,
								"capabilities": {
									"drop": ["ALL"]
								},
								"runAsNonRoot": true,
								"runAsUser": 1000,
								"seccompProfile": {
									"type": "RuntimeDefault"
								}
							}
						}],
						"serviceAccount": "%s"
					}
				}`, token, metricsServiceName, namespace, serviceAccountName))
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to create curl-metrics pod")

			By("waiting for the curl-metrics pod to complete.")
			verifyCurlUp := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "pods", "curl-metrics",
					"-o", "jsonpath={.status.phase}",
					"-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Succeeded"), "curl pod in wrong status")
			}
			Eventually(verifyCurlUp, 5*time.Minute).Should(Succeed())

			By("getting the metrics by checking curl-metrics logs")
			metricsOutput := getMetricsOutput()
			Expect(metricsOutput).To(ContainSubstring(
				"controller_runtime_reconcile_total",
			))
		})

		// +kubebuilder:scaffold:e2e-webhooks-checks

		Context("HeliosApp End-to-End Flow", func() {
			const (
				testAppName    = "e2e-test-app"
				testNamespace  = "e2e-test-namespace"
				testGitRepo    = "https://github.com/hoangphuc841/helios.git"
				testGitopsRepo = "https://github.com/PhuocHoan/helios-gitops.git"
				testImageRepo  = "docker.io/test/e2e-test-app"
			)

			BeforeEach(func() {
				By("creating test namespace")
				cmd := exec.Command("kubectl", "create", "namespace", testNamespace)
				_, err := utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred(), "Failed to create test namespace")

				By("ensuring test namespace is ready")
				Eventually(func() error {
					cmd := exec.Command("kubectl", "get", "namespace", testNamespace)
					_, err := utils.Run(cmd)
					return fmt.Errorf("failed to run command: %w", err)
				}, 30*time.Second).Should(Succeed(), "Test namespace should be ready")
			})

			AfterEach(func() {
				By("cleaning up test namespace")
				cmd := exec.Command("kubectl", "delete", "namespace", testNamespace, "--ignore-not-found=true")
				_, _ = utils.Run(cmd)
			})

			It("should create and manage a complete HeliosApp lifecycle", func() {
				By("creating a HeliosApp resource (testing automatic PVC creation)")
				err := createHeliosApp(testAppName, testNamespace, testGitRepo, testGitopsRepo, testAppName, testImageRepo)
				Expect(err).NotTo(HaveOccurred(), "Failed to create HeliosApp")

				By("waiting for automatic PVC creation")
				verifyPVCCreated := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "pvc", fmt.Sprintf("%s-workspace", testAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred(), "PVC should be created automatically")
				}
				Eventually(verifyPVCCreated, 2*time.Minute).Should(Succeed())

				By("waiting for Pipeline to be created")
				err = waitForResource("pipeline", fmt.Sprintf("%s-pipeline", testAppName), testNamespace, 2*time.Minute)
				Expect(err).NotTo(HaveOccurred(), "Pipeline should be created")

				By("waiting for Tekton Trigger resources to be created")
				verifyTriggerResources := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "eventlistener", fmt.Sprintf("%s-el", testAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred(), "EventListener should be created")

					cmd = exec.Command("kubectl", "get", "triggerbinding", fmt.Sprintf("%s-trigger-binding", testAppName), "-n", testNamespace)
					_, err = utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred(), "TriggerBinding should be created")

					cmd = exec.Command("kubectl", "get", "triggertemplate", fmt.Sprintf("%s-trigger-template", testAppName), "-n", testNamespace)
					_, err = utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred(), "TriggerTemplate should be created")
				}
				Eventually(verifyTriggerResources, 2*time.Minute).Should(Succeed())

				By("waiting for ArgoCD Application to be created")
				verifyArgoCDAppCreated := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "application", fmt.Sprintf("%s-argocd", testAppName), "-n", "argocd")
					_, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred(), "ArgoCD Application should be created")
				}
				Eventually(verifyArgoCDAppCreated, 2*time.Minute).Should(Succeed())

				By("creating a mock PipelineRun to simulate build")
				pipelineRunYAML := fmt.Sprintf(`apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: %s-pipelinerun-%d
  namespace: %s
  labels:
    helios.io/app-name: "%s"
    helios.io/managed-by: "helios-operator"
    triggers.tekton.dev/trigger: "%s-el"
spec:
  pipelineRef:
    name: %s-pipeline
  serviceAccountName: pipeline-sa
  params:
  - name: git-repo
    value: "%s"
  - name: git-revision
    value: "abc123def"
  - name: image-repo
    value: "%s"
  workspaces:
  - name: source-code
    persistentVolumeClaim:
      claimName: %s-workspace`, testAppName, time.Now().Unix(), testNamespace, testAppName, testAppName, testAppName, testGitRepo, testImageRepo, testAppName)

				cmd := exec.Command("kubectl", "apply", "-f", "-")
				cmd.Stdin = strings.NewReader(pipelineRunYAML)
				_, err = utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred(), "Failed to create PipelineRun")

				By("waiting for PipelineRun to complete")
				verifyPipelineRunComplete := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "pipelinerun", "-l", fmt.Sprintf("helios.io/app-name=%s", testAppName), "-n", testNamespace, "-o", "jsonpath={.items[0].status.conditions[0].reason}")
					output, err := utils.Run(cmd)
					if err == nil {
						// PipelineRun might be Succeeded, Failed, or still Running
						g.Expect(output).To(Or(Equal("Succeeded"), Equal("Failed"), Equal("Running")), "PipelineRun should have a status")
					}
				}
				Eventually(verifyPipelineRunComplete, 5*time.Minute).Should(Succeed())

				By("checking HeliosApp status updates")
				verifyHeliosAppStatus := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "heliosapp", testAppName, "-n", testNamespace, "-o", "jsonpath={.status.conditions}")
					output, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred(), "Should be able to get HeliosApp status")
					g.Expect(output).NotTo(BeEmpty(), "HeliosApp should have status conditions")
				}
				Eventually(verifyHeliosAppStatus, 2*time.Minute).Should(Succeed())

				By("verifying reconciliation metrics")
				metricsOutput := getMetricsOutput()
				Expect(metricsOutput).To(ContainSubstring(
					`controller_runtime_reconcile_total{controller="heliosapp",result="success"}`,
				))

				By("checking that all resources have correct labels")
				verifyResourceLabels := func(g Gomega) {
					// Check Pipeline labels
					cmd := exec.Command("kubectl", "get", "pipeline", fmt.Sprintf("%s-pipeline", testAppName), "-n", testNamespace, "-o", "jsonpath={.metadata.labels.helios\\.io/managed-by}")
					output, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output).To(Equal("helios-operator"))

					// Check EventListener labels
					cmd = exec.Command("kubectl", "get", "eventlistener", fmt.Sprintf("%s-el", testAppName), "-n", testNamespace, "-o", "jsonpath={.metadata.labels.helios\\.io/managed-by}")
					output, err = utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output).To(Equal("helios-operator"))
				}
				Eventually(verifyResourceLabels, 1*time.Minute).Should(Succeed())

				By("verifying finalizer is added")
				verifyFinalizerAdded := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "heliosapp", testAppName, "-n", testNamespace, "-o", "jsonpath={.metadata.finalizers}")
					output, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred(), "Should be able to get HeliosApp")
					g.Expect(output).To(ContainSubstring("platform.helios.io/finalizer"), "Finalizer should be added")
				}
				Eventually(verifyFinalizerAdded, 1*time.Minute).Should(Succeed())

				By("testing resource cleanup on deletion")
				err = deleteHeliosApp(testAppName, testNamespace)
				Expect(err).NotTo(HaveOccurred(), "Failed to delete HeliosApp")

				By("verifying Pipeline is cleaned up")
				verifyPipelineDeleted := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "pipeline", fmt.Sprintf("%s-pipeline", testAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).To(HaveOccurred(), "Pipeline should be deleted")
				}
				Eventually(verifyPipelineDeleted, 2*time.Minute).Should(Succeed())

				By("verifying Tekton resources are cleaned up")
				verifyTriggerResourcesDeleted := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "eventlistener", fmt.Sprintf("%s-el", testAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).To(HaveOccurred(), "EventListener should be deleted")
				}
				Eventually(verifyTriggerResourcesDeleted, 2*time.Minute).Should(Succeed())

				By("verifying ArgoCD Application is cleaned up")
				verifyArgoCDAppDeleted := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "application", fmt.Sprintf("%s-argocd", testAppName), "-n", "argocd")
					_, err := utils.Run(cmd)
					g.Expect(err).To(HaveOccurred(), "ArgoCD Application should be deleted")
				}
				Eventually(verifyArgoCDAppDeleted, 2*time.Minute).Should(Succeed())

				By("verifying automatically created PVC is cleaned up")
				verifyPVCCleanedUp := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "pvc", fmt.Sprintf("%s-workspace", testAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).To(HaveOccurred(), "Automatically created PVC should be deleted")
				}
				Eventually(verifyPVCCleanedUp, 2*time.Minute).Should(Succeed())
			})

			It("should handle explicit PVC specification", func() {
				const explicitPVCAppName = "explicit-pvc-app"

				By("creating a custom PVC first")
				customPVCYAML := fmt.Sprintf(`apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: custom-workspace-pvc
  namespace: %s
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 2Gi`, testNamespace)

				cmd := exec.Command("kubectl", "apply", "-f", "-")
				cmd.Stdin = strings.NewReader(customPVCYAML)
				_, err := utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred(), "Failed to create custom PVC")

				By("creating a HeliosApp with explicit PVC name")
				heliosAppYAML := fmt.Sprintf(`apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: %s
  namespace: %s
spec:
  gitRepo: "%s"
  gitBranch: "main"
  gitopsRepo: "%s"
  gitopsPath: "%s"
  gitopsBranch: "main"
  imageRepo: "%s"
  port: 80
  replicas: 1
  serviceAccount: "pipeline-sa"
  webhookSecret: "github-webhook-secret"
  pvcName: "custom-workspace-pvc"`, explicitPVCAppName, testNamespace, testGitRepo, testGitopsRepo, explicitPVCAppName, testImageRepo)

				cmd = exec.Command("kubectl", "apply", "-f", "-")
				cmd.Stdin = strings.NewReader(heliosAppYAML)
				_, err = utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred(), "Failed to create HeliosApp with explicit PVC")

				By("verifying no automatic PVC is created")
				verifyNoAutomaticPVC := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "pvc", fmt.Sprintf("%s-workspace", explicitPVCAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).To(HaveOccurred(), "No automatic PVC should be created when explicit PVC is specified")
				}
				Eventually(verifyNoAutomaticPVC, 1*time.Minute).Should(Succeed())

				By("verifying the custom PVC is still available")
				cmd = exec.Command("kubectl", "get", "pvc", "custom-workspace-pvc", "-n", testNamespace)
				_, err = utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred(), "Custom PVC should still exist")

				By("cleaning up")
				cmd = exec.Command("kubectl", "delete", "heliosapp", explicitPVCAppName, "-n", testNamespace, "--ignore-not-found=true")
				_, _ = utils.Run(cmd)
				cmd = exec.Command("kubectl", "delete", "pvc", "custom-workspace-pvc", "-n", testNamespace, "--ignore-not-found=true")
				_, _ = utils.Run(cmd)
			})

			It("should handle invalid HeliosApp configurations", func() {
				By("testing webhook validation for invalid configurations")

				testCases := []struct {
					name        string
					appYAML     string
					expectError bool
					errorMsg    string
				}{
					{
						name: "empty git repo",
						appYAML: fmt.Sprintf(`apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: invalid-git-repo
  namespace: %s
spec:
  gitRepo: ""
  gitBranch: "main"
  gitopsRepo: "%s"
  gitopsPath: "invalid-git-repo"
  gitopsBranch: "main"
  imageRepo: "%s"
  port: 80
  replicas: 1`, testNamespace, testGitopsRepo, testImageRepo),
						expectError: true,
						errorMsg:    "gitRepo",
					},
					{
						name: "invalid port range",
						appYAML: fmt.Sprintf(`apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: invalid-port
  namespace: %s
spec:
  gitRepo: "%s"
  gitBranch: "main"
  gitopsRepo: "%s"
  gitopsPath: "invalid-port"
  gitopsBranch: "main"
  imageRepo: "%s"
  port: 99999
  replicas: 1`, testNamespace, testGitRepo, testGitopsRepo, testImageRepo),
						expectError: true,
						errorMsg:    "port",
					},
					{
						name: "negative replicas",
						appYAML: fmt.Sprintf(`apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: invalid-replicas
  namespace: %s
spec:
  gitRepo: "%s"
  gitBranch: "main"
  gitopsRepo: "%s"
  gitopsPath: "invalid-replicas"
  gitopsBranch: "main"
  imageRepo: "%s"
  port: 80
  replicas: -1`, testNamespace, testGitRepo, testGitopsRepo, testImageRepo),
						expectError: true,
						errorMsg:    "replicas",
					},
				}

				for _, tc := range testCases {
					By("testing " + tc.name)
					cmd := exec.Command("kubectl", "apply", "-f", "-")
					cmd.Stdin = strings.NewReader(tc.appYAML)
					output, err := utils.Run(cmd)

					if tc.expectError {
						if err != nil {
							// Webhook rejection is expected
							Expect(output).To(Or(
								ContainSubstring("admission webhook"),
								ContainSubstring("validation"),
								ContainSubstring(tc.errorMsg),
							), "Should reject invalid configuration: %s", tc.name)
						} else {
							// If it was accepted, it should eventually fail in controller
							Eventually(func() bool {
								cmd := exec.Command("kubectl", "get", "heliosapp", fmt.Sprintf("invalid-%s", strings.ReplaceAll(tc.name, " ", "-")), "-n", testNamespace, "-o", "jsonpath={.status.conditions}")
								output, _ := utils.Run(cmd)
								return strings.Contains(output, "Failed") || strings.Contains(output, "Error")
							}, 2*time.Minute).Should(BeTrue(), "Invalid HeliosApp should eventually fail: %s", tc.name)
						}
					} else {
						Expect(err).NotTo(HaveOccurred(), "Valid configuration should be accepted: %s", tc.name)
					}

					// Cleanup
					appName := "invalid-" + strings.ReplaceAll(tc.name, " ", "-")
					cmd = exec.Command("kubectl", "delete", "heliosapp", appName, "-n", testNamespace, "--ignore-not-found=true")
					_, _ = utils.Run(cmd)
				}
			})

			It("should handle error recovery scenarios", func() {
				const recoveryAppName = "recovery-test-app"

				By("creating a HeliosApp resource")
				heliosAppYAML := fmt.Sprintf(`apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: %s
  namespace: %s
spec:
  gitRepo: "%s"
  gitBranch: "main"
  gitopsRepo: "%s"
  gitopsPath: "%s"
  gitopsBranch: "main"
  imageRepo: "%s"
  port: 80
  replicas: 1
  serviceAccount: "pipeline-sa"
  webhookSecret: "github-webhook-secret"`, recoveryAppName, testNamespace, testGitRepo, testGitopsRepo, recoveryAppName, testImageRepo)

				cmd := exec.Command("kubectl", "apply", "-f", "-")
				cmd.Stdin = strings.NewReader(heliosAppYAML)
				_, err := utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred(), "Failed to create HeliosApp")

				By("waiting for initial resources to be created")
				Eventually(func() error {
					cmd := exec.Command("kubectl", "get", "pipeline", fmt.Sprintf("%s-pipeline", recoveryAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					return fmt.Errorf("failed to run command: %w", err)
				}, 2*time.Minute).Should(Succeed(), "Pipeline should be created")

				By("simulating resource deletion (error scenario)")
				cmd = exec.Command("kubectl", "delete", "pipeline", fmt.Sprintf("%s-pipeline", recoveryAppName), "-n", testNamespace, "--ignore-not-found=true")
				_, _ = utils.Run(cmd)

				By("waiting for resource recreation (error recovery)")
				Eventually(func() error {
					cmd := exec.Command("kubectl", "get", "pipeline", fmt.Sprintf("%s-pipeline", recoveryAppName), "-n", testNamespace)
					_, err := utils.Run(cmd)
					return fmt.Errorf("failed to run command: %w", err)
				}, 3*time.Minute).Should(Succeed(), "Pipeline should be recreated after deletion")

				By("verifying finalizer is still present after recovery")
				Eventually(func() error {
					cmd := exec.Command("kubectl", "get", "heliosapp", recoveryAppName, "-n", testNamespace, "-o", "jsonpath={.metadata.finalizers}")
					output, err := utils.Run(cmd)
					if err != nil {
						return fmt.Errorf("failed to run command: %w", err)
					}
					if !strings.Contains(output, "platform.helios.io/finalizer") {
						return fmt.Errorf("finalizer not found: %s", output)
					}
					return nil
				}, 1*time.Minute).Should(Succeed(), "Finalizer should still be present after recovery")

				By("cleaning up")
				cmd = exec.Command("kubectl", "delete", "heliosapp", recoveryAppName, "-n", testNamespace, "--ignore-not-found=true")
				_, _ = utils.Run(cmd)
			})
		})
	})
})

// serviceAccountToken returns a token for the specified service account in the given namespace.
// It uses the Kubernetes TokenRequest API to generate a token by directly sending a request
// and parsing the resulting token from the API response.
func serviceAccountToken() (string, error) {
	ctx := context.Background()
	const tokenRequestRawString = `{
		"apiVersion": "authentication.k8s.io/v1",
		"kind": "TokenRequest"
	}`

	// Temporary file to store the token request
	secretName := fmt.Sprintf("%s-token-request", serviceAccountName)
	tokenRequestFile := filepath.Join(os.TempDir(), secretName)
	err := os.WriteFile(tokenRequestFile, []byte(tokenRequestRawString), os.FileMode(0o644))
	if err != nil {
		return "", fmt.Errorf("failed to write token request file: %w", err)
	}

	var out string
	verifyTokenCreation := func(g Gomega) {
		// Execute kubectl command to create the token
		cmd := exec.CommandContext(ctx, "kubectl", "create", "--raw", fmt.Sprintf(
			"/api/v1/namespaces/%s/serviceaccounts/%s/token",
			namespace,
			serviceAccountName,
		), "-f", tokenRequestFile)

		output, err := cmd.CombinedOutput()
		g.Expect(err).NotTo(HaveOccurred())

		// Parse the JSON output to extract the token
		var token tokenRequest
		err = json.Unmarshal(output, &token)
		g.Expect(err).NotTo(HaveOccurred())

		out = token.Status.Token
	}
	Eventually(verifyTokenCreation).Should(Succeed())

	return out, nil
}

// getMetricsOutput retrieves and returns the logs from the curl pod used to access the metrics endpoint.
func getMetricsOutput() string {
	By("getting the curl-metrics logs")
	cmd := exec.CommandContext(context.Background(), "kubectl", "logs", "curl-metrics", "-n", namespace)
	metricsOutput, err := utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
	Expect(metricsOutput).To(ContainSubstring("< HTTP/1.1 200 OK"))
	return metricsOutput
}

// tokenRequest is a simplified representation of the Kubernetes TokenRequest API response,
// containing only the token field that we need to extract.
type tokenRequest struct {
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}

// createHeliosApp creates a HeliosApp resource with the given parameters.
func createHeliosApp(name, namespace, gitRepo, gitopsRepo, gitopsPath, imageRepo string) error {
	heliosAppYAML := fmt.Sprintf(`apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: %s
  namespace: %s
spec:
  gitRepo: "%s"
  gitBranch: "main"
  gitopsRepo: "%s"
  gitopsPath: "%s"
  gitopsBranch: "main"
  imageRepo: "%s"
  port: 80
  replicas: 1
  serviceAccount: "pipeline-sa"
  webhookSecret: "github-webhook-secret"`, name, namespace, gitRepo, gitopsRepo, gitopsPath, imageRepo)

	cmd := exec.CommandContext(context.Background(), "kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(heliosAppYAML)
	_, err := utils.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to apply HeliosApp: %w", err)
	}
	return nil
}

// waitForResource waits for a Kubernetes resource to exist.
func waitForResource(resourceType, name, namespace string, timeout time.Duration) error {
	var lastErr error
	Eventually(func() error {
		cmd := exec.CommandContext(context.Background(), "kubectl", "get", resourceType, name, "-n", namespace)
		_, err := utils.Run(cmd)
		if err != nil {
			lastErr = fmt.Errorf("failed to get resource %s/%s: %w", resourceType, name, err)
			return lastErr
		}
		return nil
	}, timeout).Should(Succeed())
	return lastErr
}

// deleteHeliosApp deletes a HeliosApp resource and waits for cleanup.
func deleteHeliosApp(name, namespace string) error {
	cmd := exec.Command("kubectl", "delete", "heliosapp", name, "-n", namespace, "--ignore-not-found=true")
	_, err := utils.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete HeliosApp %s: %w", name, err)
	}
	return nil
}
