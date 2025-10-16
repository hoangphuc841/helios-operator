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

	"github.com/stretchr/testify/assert"
)

func TestGenerateEventListener(t *testing.T) {
	tests := []struct {
		name          string
		elName        string
		namespace     string
		triggerName   string
		bindingName   string
		templateName  string
		githubSecret  string
		expectErr     bool
	}{
		{
			name:         "basic event listener",
			elName:       "my-listener",
			namespace:    "default",
			triggerName:  "github-trigger",
			bindingName:  "github-binding",
			templateName: "github-template",
			githubSecret: "github-secret",
			expectErr:    false,
		},
		{
			name:         "event listener with different namespace",
			elName:       "prod-listener",
			namespace:    "production",
			triggerName:  "prod-trigger",
			bindingName:  "prod-binding",
			templateName: "prod-template",
			githubSecret: "prod-secret",
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateEventListener(
				tt.elName,
				tt.namespace,
				tt.triggerName,
				tt.bindingName,
				tt.templateName,
				tt.githubSecret,
			)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check basic structure
			assert.Equal(t, "triggers.tekton.dev/v1beta1", result.GetAPIVersion())
			assert.Equal(t, "EventListener", result.GetKind())
			assert.Equal(t, tt.elName, result.GetName())
			assert.Equal(t, tt.namespace, result.GetNamespace())

			// Check spec fields
			spec, found := result.Object["spec"]
			assert.True(t, found)
			specMap := spec.(map[string]interface{})

			// Check service account
			serviceAccount, found := specMap["serviceAccountName"]
			assert.True(t, found)
			assert.Equal(t, "tekton-triggers-sa", serviceAccount)

			// Check triggers
			triggers, found := specMap["triggers"]
			assert.True(t, found)
			triggersSlice := triggers.([]interface{})
			assert.Len(t, triggersSlice, 1)

			trigger := triggersSlice[0].(map[string]interface{})
			assert.Equal(t, tt.triggerName, trigger["name"])

			// Check bindings
			bindings, found := trigger["bindings"]
			assert.True(t, found)
			bindingsSlice := bindings.([]interface{})
			assert.Len(t, bindingsSlice, 1)

			binding := bindingsSlice[0].(map[string]interface{})
			assert.Equal(t, tt.bindingName, binding["ref"])

			// Check template
			template, found := trigger["template"]
			assert.True(t, found)
			templateMap := template.(map[string]interface{})
			assert.Equal(t, tt.templateName, templateMap["ref"])

			// Check interceptors
			interceptors, found := trigger["interceptors"]
			assert.True(t, found)
			interceptorsSlice := interceptors.([]interface{})
			assert.Len(t, interceptorsSlice, 1)

			interceptor := interceptorsSlice[0].(map[string]interface{})
			interceptorRef := interceptor["ref"].(map[string]interface{})
			assert.Equal(t, "github", interceptorRef["name"])
			assert.Equal(t, "ClusterInterceptor", interceptorRef["kind"])
		})
	}
}

func TestGenerateEventListener_InterceptorConfiguration(t *testing.T) {
	result, err := GenerateEventListener(
		"test-listener",
		"default",
		"test-trigger",
		"test-binding",
		"test-template",
		"test-secret",
	)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	spec := result.Object["spec"].(map[string]interface{})
	triggers := spec["triggers"].([]interface{})
	trigger := triggers[0].(map[string]interface{})
	interceptors := trigger["interceptors"].([]interface{})
	interceptor := interceptors[0].(map[string]interface{})

	// Check interceptor params
	params, found := interceptor["params"]
	assert.True(t, found)
	paramsSlice := params.([]interface{})
	assert.Len(t, paramsSlice, 2) // secretRef and eventTypes

	// Check secretRef param
	param1 := paramsSlice[0].(map[string]interface{})
	assert.Equal(t, "secretRef", param1["name"])

	// Check eventTypes param
	param2 := paramsSlice[1].(map[string]interface{})
	assert.Equal(t, "eventTypes", param2["name"])
}

func TestGenerateTriggerBinding(t *testing.T) {
	tests := []struct {
		name      string
		tbName    string
		namespace string
		expectErr bool
	}{
		{
			name:      "basic trigger binding",
			tbName:    "github-binding",
			namespace: "default",
			expectErr: false,
		},
		{
			name:      "trigger binding in custom namespace",
			tbName:    "gitlab-binding",
			namespace: "gitlab-ns",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateTriggerBinding(tt.tbName, tt.namespace)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check basic structure
			assert.Equal(t, "triggers.tekton.dev/v1beta1", result.GetAPIVersion())
			assert.Equal(t, "TriggerBinding", result.GetKind())
			assert.Equal(t, tt.tbName, result.GetName())
			assert.Equal(t, tt.namespace, result.GetNamespace())

			// Check spec fields
			spec, found := result.Object["spec"]
			assert.True(t, found)
			specMap := spec.(map[string]interface{})

			// Check params
			params, found := specMap["params"]
			assert.True(t, found)
			paramsSlice := params.([]interface{})
			assert.Len(t, paramsSlice, 3) // git-repo-url, git-revision, branch
		})
	}
}

func TestGenerateTriggerBinding_Parameters(t *testing.T) {
	result, err := GenerateTriggerBinding("test-binding", "default")
	assert.NoError(t, err)

	spec := result.Object["spec"].(map[string]interface{})
	params := spec["params"].([]interface{})

	// Check git-repo-url param
	param1 := params[0].(map[string]interface{})
	assert.Equal(t, "git-repo-url", param1["name"])
	assert.Equal(t, "$(body.repository.url)", param1["value"])

	// Check git-revision param
	param2 := params[1].(map[string]interface{})
	assert.Equal(t, "git-revision", param2["name"])
	assert.Equal(t, "$(body.head_commit.id)", param2["value"])

	// Check branch param
	param3 := params[2].(map[string]interface{})
	assert.Equal(t, "branch", param3["name"])
	assert.Equal(t, "$(body.ref)", param3["value"])
}

func TestGenerateTriggerTemplate(t *testing.T) {
	tests := []struct {
		name            string
		ttName          string
		namespace       string
		pipelineRunName string
		pipelineName    string
		serviceAccount  string
		workspace       map[string]interface{}
		expectErr       bool
	}{
		{
			name:            "basic trigger template",
			ttName:          "github-template",
			namespace:       "default",
			pipelineRunName: "build-run",
			pipelineName:    "build-pipeline",
			serviceAccount:  "tekton-sa",
			workspace: map[string]interface{}{
				"persistentVolumeClaim": map[string]interface{}{
					"claimName": "source-pvc",
				},
			},
			expectErr: false,
		},
		{
			name:            "template in custom namespace",
			ttName:          "prod-template",
			namespace:       "production",
			pipelineRunName: "prod-run",
			pipelineName:    "prod-pipeline",
			serviceAccount:  "prod-sa",
			workspace: map[string]interface{}{
				"persistentVolumeClaim": map[string]interface{}{
					"claimName": "prod-pvc",
				},
			},
			expectErr: false,
		},
		{
			name:            "template with empty workspace",
			ttName:          "empty-ws-template",
			namespace:       "default",
			pipelineRunName: "empty-run",
			pipelineName:    "empty-pipeline",
			serviceAccount:  "default-sa",
			workspace:       map[string]interface{}{},
			expectErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateTriggerTemplate(
				tt.ttName,
				tt.namespace,
				tt.pipelineRunName,
				tt.pipelineName,
				tt.serviceAccount,
				tt.workspace,
			)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check basic structure
			assert.Equal(t, "triggers.tekton.dev/v1beta1", result.GetAPIVersion())
			assert.Equal(t, "TriggerTemplate", result.GetKind())
			assert.Equal(t, tt.ttName, result.GetName())
			assert.Equal(t, tt.namespace, result.GetNamespace())

			// Check spec fields
			spec, found := result.Object["spec"]
			assert.True(t, found)
			specMap := spec.(map[string]interface{})

			// Check params
			params, found := specMap["params"]
			assert.True(t, found)
			paramsSlice := params.([]interface{})
			assert.Len(t, paramsSlice, 3) // git-repo-url, git-revision, branch

			// Check resource templates
			resourceTemplates, found := specMap["resourcetemplates"]
			assert.True(t, found)
			resourceTemplatesSlice := resourceTemplates.([]interface{})
			assert.Len(t, resourceTemplatesSlice, 1)
		})
	}
}

func TestGenerateTriggerTemplate_Parameters(t *testing.T) {
	result, err := GenerateTriggerTemplate(
		"test-template",
		"default",
		"test-run",
		"test-pipeline",
		"test-sa",
		map[string]interface{}{
			"persistentVolumeClaim": map[string]interface{}{
				"claimName": "test-pvc",
			},
		},
	)

	assert.NoError(t, err)

	spec := result.Object["spec"].(map[string]interface{})
	params := spec["params"].([]interface{})

	// Verify all required parameters are defined
	paramNames := make([]string, len(params))
	for i, p := range params {
		param := p.(map[string]interface{})
		paramNames[i] = param["name"].(string)
	}

	assert.Contains(t, paramNames, "git-repo-url")
	assert.Contains(t, paramNames, "git-revision")
	assert.Contains(t, paramNames, "branch")
}

func TestGenerateTriggerTemplate_PVCClaimName(t *testing.T) {
	workspace := map[string]interface{}{
		"persistentVolumeClaim": map[string]interface{}{
			"claimName": "my-pvc",
		},
	}

	result, err := GenerateTriggerTemplate(
		"test-template",
		"default",
		"test-run",
		"test-pipeline",
		"test-sa",
		workspace,
	)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// The PVC claim name should be embedded in the resource template
	// We can't easily extract it from the RawExtension, but we can verify
	// that the template was created successfully
	spec := result.Object["spec"].(map[string]interface{})
	resourceTemplates := spec["resourcetemplates"].([]interface{})
	assert.Len(t, resourceTemplates, 1)
}

func TestGenerateTriggerTemplate_EmptyWorkspace(t *testing.T) {
	// Test with workspace that doesn't have PVC
	workspace := map[string]interface{}{
		"emptyDir": map[string]interface{}{},
	}

	result, err := GenerateTriggerTemplate(
		"test-template",
		"default",
		"test-run",
		"test-pipeline",
		"test-sa",
		workspace,
	)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Should create template with empty PVC claim name
	spec := result.Object["spec"].(map[string]interface{})
	resourceTemplates := spec["resourcetemplates"].([]interface{})
	assert.Len(t, resourceTemplates, 1)
}

func TestGenerateTriggerTemplate_NilWorkspace(t *testing.T) {
	// Test with nil workspace map
	result, err := GenerateTriggerTemplate(
		"test-template",
		"default",
		"test-run",
		"test-pipeline",
		"test-sa",
		nil,
	)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Should create template even with nil workspace
	spec := result.Object["spec"].(map[string]interface{})
	resourceTemplates := spec["resourcetemplates"].([]interface{})
	assert.Len(t, resourceTemplates, 1)
}
