package controller

import (
	"fmt"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GeneratePipeline creates a Tekton Pipeline from template
// The pipeline name will be derived from the HeliosApp name: {app-name}-pipeline
func GeneratePipeline(heliosApp *heliosappv1.HeliosApp) (*unstructured.Unstructured, error) {
	appName := heliosApp.Name
	pipelineName := fmt.Sprintf("%s-pipeline", appName)

	// Create the unstructured object directly
	// In a production environment, you might want to use a YAML parser
	pipeline := map[string]interface{}{
		"apiVersion": "tekton.dev/v1beta1",
		"kind":       "Pipeline",
		"metadata": map[string]interface{}{
			"name":      pipelineName,
			"namespace": heliosApp.Namespace,
			"labels": map[string]interface{}{
				"helios.io/managed-by": "helios-operator",
				"helios.io/app-name":   appName,
			},
		},
		"spec": map[string]interface{}{
			"params": []interface{}{
				map[string]interface{}{
					"name":        "git-repo",
					"type":        "string",
					"description": "The git repository URL to clone from.",
					"default":     heliosApp.Spec.GitRepo,
				},
				map[string]interface{}{
					"name":        "git-revision",
					"type":        "string",
					"description": "The git revision to clone.",
					"default":     heliosApp.Spec.GitBranch,
				},
				map[string]interface{}{
					"name":        "image-repo",
					"type":        "string",
					"description": "The container image repository to push to.",
					"default":     heliosApp.Spec.ImageRepo,
				},
			},
			"workspaces": []interface{}{
				map[string]interface{}{
					"name":        "source-code",
					"description": "The workspace where source code is cloned.",
				},
			},
			"tasks": []interface{}{
				map[string]interface{}{
					"name": "git-clone",
					"taskRef": map[string]interface{}{
						"name": "git-clone",
					},
					"params": []interface{}{
						map[string]interface{}{
							"name":  "url",
							"value": "$(params.git-repo)",
						},
						map[string]interface{}{
							"name":  "revision",
							"value": "$(params.git-revision)",
						},
					},
					"workspaces": []interface{}{
						map[string]interface{}{
							"name":      "output",
							"workspace": "source-code",
						},
					},
				},
				map[string]interface{}{
					"name":     "build-and-push",
					"runAfter": []interface{}{"git-clone"},
					"taskRef": map[string]interface{}{
						"name": "kaniko",
					},
					"params": []interface{}{
						map[string]interface{}{
							"name":  "IMAGE",
							"value": "$(params.image-repo):$(tasks.git-clone.results.commit)",
						},
					},
					"workspaces": []interface{}{
						map[string]interface{}{
							"name":      "source",
							"workspace": "source-code",
						},
					},
				},
			},
		},
	}

	return &unstructured.Unstructured{Object: pipeline}, nil
}

// getPipelineName returns the standardized pipeline name for a HeliosApp
func getPipelineName(appName string) string {
	return fmt.Sprintf("%s-pipeline", appName)
}
