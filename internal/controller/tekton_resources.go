package controller

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GenerateEventListener tạo manifest cho EventListener Tekton
func GenerateEventListener(name, namespace, triggerName, bindingName, templateName, githubSecret string) (*unstructured.Unstructured, error) {
	el := map[string]interface{}{
		"apiVersion": "triggers.tekton.dev/v1beta1",
		"kind":       "EventListener",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"serviceAccountName": "tekton-triggers-sa",
			"triggers": []map[string]interface{}{
				{
					"name": triggerName,
					"bindings": []map[string]interface{}{
						{"ref": bindingName},
					},
					"template": map[string]interface{}{
						"ref": templateName,
					},
					"interceptors": []map[string]interface{}{
						{
							"ref": map[string]interface{}{
								"name": "github",
								"kind": "ClusterInterceptor",
							},
							"params": []map[string]interface{}{
								{"name": "secretRef", "value": map[string]interface{}{
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
	tb := map[string]interface{}{
		"apiVersion": "triggers.tekton.dev/v1beta1",
		"kind":       "TriggerBinding",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"params": []map[string]interface{}{
				{"name": "git-repo-url", "value": "$(body.repository.url)"},
				{"name": "git-revision", "value": "$(body.head_commit.id)"},
				{"name": "branch", "value": "$(body.ref)"}, // Có thể cần xử lý thêm ở Pipeline hoặc Interceptor
			},
		},
	}
	return &unstructured.Unstructured{Object: tb}, nil
}

// GenerateTriggerTemplate tạo manifest cho TriggerTemplate Tekton
func GenerateTriggerTemplate(name, namespace, pipelineRunName, pipelineName, serviceAccount string, workspace map[string]interface{}) (*unstructured.Unstructured, error) {
	tt := map[string]interface{}{
		"apiVersion": "triggers.tekton.dev/v1beta1",
		"kind":       "TriggerTemplate",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"params": []map[string]interface{}{
				{"name": "git-repo-url"},
				{"name": "git-revision"},
				{"name": "branch"},
			},
			"resourcetemplates": []map[string]interface{}{
				{
					"apiVersion": "tekton.dev/v1beta1",
					"kind":       "PipelineRun",
					"metadata": map[string]interface{}{
						"generateName": pipelineRunName + "-",
					},
					"spec": map[string]interface{}{
						"pipelineRef": map[string]interface{}{
							"name": pipelineName,
						},
						"serviceAccountName": serviceAccount,
						"params": []map[string]interface{}{
							{"name": "git-url", "value": "$(params.git-repo-url)"},
							{"name": "git-revision", "value": "$(params.git-revision)"},
							{"name": "branch", "value": "$(params.branch)"},
						},
						"workspaces": []map[string]interface{}{
							workspace,
						},
					},
				},
			},
		},
	}
	return &unstructured.Unstructured{Object: tt}, nil
}
