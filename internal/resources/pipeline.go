package resources

import (
	"fmt"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetPipelineName generates the pipeline name from the app name.
func GetPipelineName(appName string) string {
	return fmt.Sprintf("%s-pipeline", appName)
}

// GeneratePipeline creates a Tekton Pipeline as an unstructured object.
// The pipeline name will be derived from the HeliosApp name: {app-name}-pipeline
//
// This function constructs a Pipeline that:
// 1. Clones source code from the specified Git repository
// 2. Builds and pushes a container image using Kaniko
// 3. Uses a persistent workspace for source code storage.
func GeneratePipeline(heliosApp *heliosappv1.HeliosApp) (*unstructured.Unstructured, error) {
	appName := heliosApp.Name
	pipelineName := GetPipelineName(appName)

	pipeline := &tektonv1.Pipeline{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1",
			Kind:       "Pipeline",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pipelineName,
			Namespace: heliosApp.Namespace,
			Labels: map[string]string{
				"helios.io/managed-by": "helios-operator",
				"helios.io/app-name":   appName,
			},
		},
		Spec: tektonv1.PipelineSpec{
			Params: []tektonv1.ParamSpec{
				{
					Name:        "git-repo",
					Type:        "string",
					Description: "The git repository URL to clone from.",
					Default: &tektonv1.ParamValue{
						Type:      "string",
						StringVal: heliosApp.Spec.GitRepo,
					},
				},
				{
					Name:        "git-revision",
					Type:        "string",
					Description: "The git revision to clone.",
					Default: &tektonv1.ParamValue{
						Type:      "string",
						StringVal: heliosApp.Spec.GitBranch,
					},
				},
				{
					Name:        "image-repo",
					Type:        "string",
					Description: "The container image repository to push to.",
					Default: &tektonv1.ParamValue{
						Type:      "string",
						StringVal: heliosApp.Spec.ImageRepo,
					},
				},
			},
			Workspaces: []tektonv1.PipelineWorkspaceDeclaration{
				{
					Name:        "source-code",
					Description: "The workspace where source code is cloned.",
				},
			},
			Tasks: []tektonv1.PipelineTask{
				{
					Name: "git-clone",
					TaskRef: &tektonv1.TaskRef{
						Name: "git-clone",
					},
					Params: []tektonv1.Param{
						{
							Name: "url",
							Value: tektonv1.ParamValue{
								Type:      "string",
								StringVal: "$(params.git-repo)",
							},
						},
						{
							Name: "revision",
							Value: tektonv1.ParamValue{
								Type:      "string",
								StringVal: "$(params.git-revision)",
							},
						},
					},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{
							Name:      "output",
							Workspace: "source-code",
						},
					},
				},
				{
					Name:     "build-and-push",
					RunAfter: []string{"git-clone"},
					TaskRef: &tektonv1.TaskRef{
						Name: "kaniko",
					},
					Params: []tektonv1.Param{
						{
							Name: "IMAGE",
							Value: tektonv1.ParamValue{
								Type:      "string",
								StringVal: "$(params.image-repo):$(tasks.git-clone.results.commit)",
							},
						},
					},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{
							Name:      "source",
							Workspace: "source-code",
						},
					},
				},
			},
		},
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pipeline to unstructured: %w", err)
	}

	return &unstructured.Unstructured{Object: unstructuredObj}, nil
}
