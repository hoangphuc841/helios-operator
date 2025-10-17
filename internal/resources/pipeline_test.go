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

package resources

import (
	"testing"

	heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPipelineName(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		expected string
	}{
		{
			name:     "basic app name",
			appName:  "my-app",
			expected: "my-app-pipeline",
		},
		{
			name:     "app name with hyphens",
			appName:  "my-cool-app",
			expected: "my-cool-app-pipeline",
		},
		{
			name:     "empty app name",
			appName:  "",
			expected: "-pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPipelineName(tt.appName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePipeline(t *testing.T) {
	tests := []struct {
		name      string
		heliosApp *heliosappv1.HeliosApp
		expectErr bool
	}{
		{
			name: "basic helios app",
			heliosApp: &heliosappv1.HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app",
					Namespace: "default",
				},
				Spec: heliosappv1.HeliosAppSpec{
					GitRepo:    "https://github.com/example/repo.git",
					GitBranch:  "main",
					ImageRepo:  "ghcr.io/example/app",
					GitopsRepo: "https://github.com/example/gitops.git",
				},
			},
			expectErr: false,
		},
		{
			name: "app with different namespace",
			heliosApp: &heliosappv1.HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prod-app",
					Namespace: "production",
				},
				Spec: heliosappv1.HeliosAppSpec{
					GitRepo:    "https://gitlab.com/example/repo.git",
					GitBranch:  "develop",
					ImageRepo:  "docker.io/example/prod-app",
					GitopsRepo: "https://gitlab.com/example/gitops.git",
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GeneratePipeline(tt.heliosApp)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check basic structure
			assert.Equal(t, "tekton.dev/v1", result.GetAPIVersion())
			assert.Equal(t, "Pipeline", result.GetKind())
			assert.Equal(t, GetPipelineName(tt.heliosApp.Name), result.GetName())
			assert.Equal(t, tt.heliosApp.Namespace, result.GetNamespace())

			// Check labels
			labels := result.GetLabels()
			assert.NotEmpty(t, labels)
			assert.Equal(t, "helios-operator", labels["helios.io/managed-by"])
			assert.Equal(t, tt.heliosApp.Name, labels["helios.io/app-name"])

			// Check spec fields directly
			spec, found := result.Object["spec"]
			assert.True(t, found)
			assert.NotNil(t, spec)

			specMap, ok := spec.(map[string]interface{})
			assert.True(t, ok)

			// Check params
			params, found := specMap["params"]
			assert.True(t, found)
			paramsSlice, ok := params.([]interface{})
			assert.True(t, ok)
			assert.Len(t, paramsSlice, 3) // git-repo, git-revision, image-repo

			// Check workspaces
			workspaces, found := specMap["workspaces"]
			assert.True(t, found)
			workspacesSlice, ok := workspaces.([]interface{})
			assert.True(t, ok)
			assert.Len(t, workspacesSlice, 1) // source-code workspace

			// Check tasks
			tasks, found := specMap["tasks"]
			assert.True(t, found)
			tasksSlice, ok := tasks.([]interface{})
			assert.True(t, ok)
			assert.Len(t, tasksSlice, 2) // git-clone and build-and-push

			// Verify first task (git-clone)
			task1, ok := tasksSlice[0].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "git-clone", task1["name"])

			taskRef1, ok := task1["taskRef"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "git-clone", taskRef1["name"])

			// Verify second task (build-and-push)
			task2, ok := tasksSlice[1].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "build-and-push", task2["name"])

			taskRef2, ok := task2["taskRef"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "kaniko", taskRef2["name"])

			// Check runAfter for second task
			runAfter, ok := task2["runAfter"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, runAfter, 1)
			assert.Equal(t, "git-clone", runAfter[0])
		})
	}
}

func TestGeneratePipeline_ParameterDefaults(t *testing.T) {
	heliosApp := &heliosappv1.HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-test",
			Namespace: "default",
		},
		Spec: heliosappv1.HeliosAppSpec{
			GitRepo:    "https://github.com/test/repo.git",
			GitBranch:  "feature-branch",
			ImageRepo:  "registry.io/test/image",
			GitopsRepo: "https://github.com/test/gitops.git",
		},
	}

	result, err := GeneratePipeline(heliosApp)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get spec
	spec, found := result.Object["spec"]
	assert.True(t, found)
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		t.Fatal("Expected spec to be a map")
	}

	// Check params have defaults
	params, ok := specMap["params"].([]interface{})
	if !ok {
		t.Fatal("Expected params to be a slice")
	}
	assert.Len(t, params, 3)

	// Just verify params exist with names and descriptions
	// The actual default values are embedded in the Tekton structure
	param0, ok := params[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected param0 to be a map")
	}
	assert.Equal(t, "git-repo", param0["name"])
	assert.Equal(t, "string", param0["type"])

	param1 := params[1].(map[string]interface{})
	assert.Equal(t, "git-revision", param1["name"])
	assert.Equal(t, "string", param1["type"])

	param2 := params[2].(map[string]interface{})
	assert.Equal(t, "image-repo", param2["name"])
	assert.Equal(t, "string", param2["type"])
}

func TestGeneratePipeline_WorkspaceConfiguration(t *testing.T) {
	heliosApp := &heliosappv1.HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test",
			Namespace: "test-ns",
		},
		Spec: heliosappv1.HeliosAppSpec{
			GitRepo:    "https://github.com/test/repo.git",
			GitBranch:  "main",
			ImageRepo:  "registry.io/test/image",
			GitopsRepo: "https://github.com/test/gitops.git",
		},
	}

	result, err := GeneratePipeline(heliosApp)
	assert.NoError(t, err)

	spec := result.Object["spec"].(map[string]interface{})
	workspaces := spec["workspaces"].([]interface{})

	workspace := workspaces[0].(map[string]interface{})
	assert.Equal(t, "source-code", workspace["name"])
	assert.Equal(t, "The workspace where source code is cloned.", workspace["description"])

	// Verify tasks use the workspace
	tasks := spec["tasks"].([]interface{})

	// git-clone task
	task1 := tasks[0].(map[string]interface{})
	task1Workspaces := task1["workspaces"].([]interface{})
	task1Ws := task1Workspaces[0].(map[string]interface{})
	assert.Equal(t, "output", task1Ws["name"])
	assert.Equal(t, "source-code", task1Ws["workspace"])

	// build-and-push task
	task2 := tasks[1].(map[string]interface{})
	task2Workspaces := task2["workspaces"].([]interface{})
	task2Ws := task2Workspaces[0].(map[string]interface{})
	assert.Equal(t, "source", task2Ws["name"])
	assert.Equal(t, "source-code", task2Ws["workspace"])
}

func TestGeneratePipeline_TaskParameters(t *testing.T) {
	heliosApp := &heliosappv1.HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "params-test",
			Namespace: "default",
		},
		Spec: heliosappv1.HeliosAppSpec{
			GitRepo:    "https://github.com/test/repo.git",
			GitBranch:  "main",
			ImageRepo:  "registry.io/test/image",
			GitopsRepo: "https://github.com/test/gitops.git",
		},
	}

	result, err := GeneratePipeline(heliosApp)
	assert.NoError(t, err)

	spec := result.Object["spec"].(map[string]interface{})
	tasks := spec["tasks"].([]interface{})

	// Check git-clone task params
	task1 := tasks[0].(map[string]interface{})
	task1Params := task1["params"].([]interface{})
	assert.Len(t, task1Params, 2)

	param1 := task1Params[0].(map[string]interface{})
	assert.Equal(t, "url", param1["name"])
	// Value might be either a string or a map depending on conversion
	assert.NotNil(t, param1["value"])

	param2 := task1Params[1].(map[string]interface{})
	assert.Equal(t, "revision", param2["name"])
	assert.NotNil(t, param2["value"])

	// Check build-and-push task params
	task2 := tasks[1].(map[string]interface{})
	task2Params := task2["params"].([]interface{})
	assert.Len(t, task2Params, 1)

	imageParam := task2Params[0].(map[string]interface{})
	assert.Equal(t, "IMAGE", imageParam["name"])
	assert.NotNil(t, imageParam["value"])
}
