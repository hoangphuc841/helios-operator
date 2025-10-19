package controller

import (
	"fmt"
	"time"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GenerateEventListener tạo manifest cho EventListener Tekton
func GenerateEventListener(name, namespace, triggerName, bindingName, templateName, githubSecret string) (*unstructured.Unstructured, error) {
	el := map[string]any{
		"apiVersion": "triggers.tekton.dev/v1beta1",
		"kind":       "EventListener",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"serviceAccountName": "tekton-triggers-sa",
			"triggers": []map[string]any{
				{
					"name": triggerName,
					"bindings": []map[string]any{
						{"ref": bindingName},
					},
					"template": map[string]any{
						"ref": templateName,
					},
					"interceptors": []map[string]any{
						{
							"ref": map[string]any{
								"name": "github",
								"kind": "ClusterInterceptor",
							},
							"params": []map[string]any{
								{"name": "secretRef", "value": map[string]any{
									"secretName": githubSecret,
									"secretKey":  "webhook-secret",
								}},
								{"name": "eventTypes", "value": []string{"push"}},
							},
						},
					},
				},
			},
		},
	}
	return &unstructured.Unstructured{Object: el}, nil
}

// GenerateTriggerBinding tạo manifest cho TriggerBinding Tekton
func GenerateTriggerBinding(name, namespace string) (*unstructured.Unstructured, error) {
	tb := map[string]any{
		"apiVersion": "triggers.tekton.dev/v1beta1",
		"kind":       "TriggerBinding",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"params": []map[string]any{
				{"name": "git-repo-url", "value": "$(body.repository.url)"},
				{"name": "git-revision", "value": "$(body.head_commit.id)"},
				{"name": "branch", "value": "$(body.ref)"}, // Có thể cần xử lý thêm ở Pipeline hoặc Interceptor
			},
		},
	}
	return &unstructured.Unstructured{Object: tb}, nil
}

// GenerateTriggerTemplate tạo manifest cho TriggerTemplate Tekton
func GenerateTriggerTemplate(name, namespace, pipelineRunName, pipelineName, serviceAccount string, workspace map[string]any) (*unstructured.Unstructured, error) {
	tt := map[string]any{
		"apiVersion": "triggers.tekton.dev/v1beta1",
		"kind":       "TriggerTemplate",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"params": []map[string]any{
				{"name": "git-repo-url"},
				{"name": "git-revision"},
				{"name": "branch"},
			},
			"resourcetemplates": []map[string]any{
				{
					"apiVersion": "tekton.dev/v1beta1",
					"kind":       "PipelineRun",
					"metadata": map[string]any{
						"generateName": pipelineRunName + "-",
					},
					"spec": map[string]any{
						"pipelineRef": map[string]any{
							"name": pipelineName,
						},
						"serviceAccountName": serviceAccount,
						"params": []map[string]any{
							{"name": "git-url", "value": "$(params.git-repo-url)"},
							{"name": "git-revision", "value": "$(params.git-revision)"},
							{"name": "branch", "value": "$(params.branch)"},
						},
						"workspaces": []map[string]any{
							workspace,
						},
					},
				},
			},
		},
	}
	return &unstructured.Unstructured{Object: tt}, nil
}

// GeneratePipelineRunForManifestGeneration tạo PipelineRun để render manifest và commit vào GitOps repo
// Đây là hàm cốt lõi cho Giai đoạn 1 trong luồng GitOps nâng cao
func GeneratePipelineRunForManifestGeneration(heliosApp *heliosappv1.HeliosApp, pipelineName string) (*unstructured.Unstructured, error) {
	timestamp := time.Now().Format("20060102-150405")
	prName := fmt.Sprintf("%s-manifest-%s", heliosApp.Name, timestamp)

	// Chuẩn bị các params để truyền vào Pipeline
	// Match với parameter names trong tekton/pipeline.yaml
	params := []map[string]any{
		{"name": "app-repo-url", "value": heliosApp.Spec.GitRepo},
		{"name": "app-repo-revision", "value": heliosApp.Spec.GitBranch},
		{"name": "image-repo", "value": heliosApp.Spec.ImageRepo},
		{"name": "gitops-repo-url", "value": heliosApp.Spec.GitOpsRepo},
		{"name": "manifest-path-in-gitops-repo", "value": heliosApp.Spec.GitOpsPath},
		{"name": "gitops-repo-branch", "value": "main"},        // Default branch for GitOps repo
		{"name": "DOCKER_HUB_USERNAME", "value": "dummy-user"}, // TODO: Get from Secret
		{"name": "DOCKER_HUB_TOKEN", "value": "dummy-token"},   // TODO: Get from Secret
	}

	// PVC workspace - Pipeline expects two workspaces: source-workspace and gitops-workspace
	pvcName := heliosApp.Spec.PVCName
	if pvcName == "" {
		pvcName = "pvc-" + heliosApp.Name
	}

	pr := map[string]any{
		"apiVersion": "tekton.dev/v1beta1",
		"kind":       "PipelineRun",
		"metadata": map[string]any{
			"name":      prName,
			"namespace": heliosApp.Namespace,
			"labels": map[string]any{
				"app.kubernetes.io/name":       heliosApp.Name,
				"app.kubernetes.io/managed-by": "helios-operator",
				"helios.io/pipeline-type":      "manifest-generation",
			},
		},
		"spec": map[string]any{
			"pipelineRef": map[string]any{
				"name": pipelineName,
			},
			"serviceAccountName": heliosApp.Spec.ServiceAccount,
			"params":             params,
			"workspaces": []map[string]any{
				{
					"name": "source-workspace",
					"persistentVolumeClaim": map[string]any{
						"claimName": pvcName,
					},
				},
				{
					"name": "gitops-workspace",
					"persistentVolumeClaim": map[string]any{
						"claimName": pvcName,
					},
				},
			},
		},
	}

	return &unstructured.Unstructured{Object: pr}, nil
}

// GenerateArgoApplication tạo ArgoCD Application resource
// Đây là hàm cốt lõi cho Giai đoạn 2 trong luồng GitOps nâng cao
func GenerateArgoApplication(heliosApp *heliosappv1.HeliosApp) (*unstructured.Unstructured, error) {
	appName := heliosApp.Name + "-argocd"

	app := map[string]any{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata": map[string]any{
			"name":      appName,
			"namespace": "argocd", // ArgoCD thường chạy trong namespace "argocd"
			"labels": map[string]any{
				"app.kubernetes.io/name":       heliosApp.Name,
				"app.kubernetes.io/managed-by": "helios-operator",
			},
		},
		"spec": map[string]any{
			"project": "default",
			"source": map[string]any{
				"repoURL":        heliosApp.Spec.GitOpsRepo,
				"targetRevision": "HEAD",
				"path":           heliosApp.Spec.GitOpsPath,
			},
			"destination": map[string]any{
				"server":    "https://kubernetes.default.svc",
				"namespace": heliosApp.Namespace,
			},
			"syncPolicy": map[string]any{
				"automated": map[string]any{
					"prune":    true,
					"selfHeal": true,
				},
				"syncOptions": []string{
					"CreateNamespace=true",
				},
			},
		},
	}

	return &unstructured.Unstructured{Object: app}, nil
}
