package resources

import (
	"fmt"

	triggersv1beta1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// GenerateEventListener creates a Tekton EventListener as an unstructured object.
//
// This function creates an EventListener that:
// 1. Listens for GitHub webhook events
// 2. Validates webhook signatures using the provided secret
// 3. Triggers PipelineRuns when valid push events are received
func GenerateEventListener(name, namespace, triggerName, bindingName, templateName, githubSecret string) (*unstructured.Unstructured, error) {
	el := &triggersv1beta1.EventListener{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "triggers.tekton.dev/v1beta1",
			Kind:       "EventListener",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: triggersv1beta1.EventListenerSpec{
			ServiceAccountName: "tekton-triggers-sa",
			Triggers: []triggersv1beta1.EventListenerTrigger{
				{
					Name: triggerName,
					Bindings: []*triggersv1beta1.EventListenerBinding{
						{Ref: bindingName},
					},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: &templateName,
					},
					Interceptors: []*triggersv1beta1.EventInterceptor{
						{
							Ref: triggersv1beta1.InterceptorRef{
								Name: "github",
								Kind: "ClusterInterceptor",
							},
							Params: []triggersv1beta1.InterceptorParams{
								{
									Name:  "secretRef",
									Value: apiextensionsv1.JSON{Raw: []byte(`"` + githubSecret + `:webhook-secret"`)},
								},
								{
									Name:  "eventTypes",
									Value: apiextensionsv1.JSON{Raw: []byte(`["push"]`)},
								},
							},
						},
					},
				},
			},
		},
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(el)
	if err != nil {
		return nil, fmt.Errorf("failed to convert EventListener to unstructured: %w", err)
	}

	return &unstructured.Unstructured{Object: unstructuredObj}, nil
}

// GenerateTriggerBinding creates a Tekton TriggerBinding as an unstructured object.
//
// This function creates a TriggerBinding that extracts values from GitHub webhook payloads
// and makes them available as parameters for PipelineRun creation.
func GenerateTriggerBinding(name, namespace string) (*unstructured.Unstructured, error) {
	tb := &triggersv1beta1.TriggerBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "triggers.tekton.dev/v1beta1",
			Kind:       "TriggerBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: triggersv1beta1.TriggerBindingSpec{
			Params: []triggersv1beta1.Param{
				{
					Name:  "git-repo-url",
					Value: "$(body.repository.url)",
				},
				{
					Name:  "git-revision",
					Value: "$(body.head_commit.id)",
				},
				{
					Name:  "branch",
					Value: "$(body.ref)",
				},
			},
		},
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tb)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TriggerBinding to unstructured: %w", err)
	}

	return &unstructured.Unstructured{Object: unstructuredObj}, nil
}

// GenerateTriggerTemplate creates a Tekton TriggerTemplate as an unstructured object.
//
// This function creates a TriggerTemplate that defines how to create PipelineRuns
// when triggered by webhook events. It takes parameters from the TriggerBinding
// and creates a PipelineRun with the specified pipeline and workspace configuration.
func GenerateTriggerTemplate(name, namespace, pipelineRunName, pipelineName, serviceAccount string, workspace map[string]interface{}) (*unstructured.Unstructured, error) {
	// Extract PVC claim name from workspace map
	pvcClaimName := ""
	if pvc, ok := workspace["persistentVolumeClaim"].(map[string]interface{}); ok {
		if claimName, ok := pvc["claimName"].(string); ok {
			pvcClaimName = claimName
		}
	}

	tt := &triggersv1beta1.TriggerTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "triggers.tekton.dev/v1beta1",
			Kind:       "TriggerTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: triggersv1beta1.TriggerTemplateSpec{
			Params: []triggersv1beta1.ParamSpec{
				{Name: "git-repo-url"},
				{Name: "git-revision"},
				{Name: "branch"},
			},
			ResourceTemplates: []triggersv1beta1.TriggerResourceTemplate{
				{
					RawExtension: runtime.RawExtension{
						Raw: []byte(`{
							"apiVersion": "tekton.dev/v1",
							"kind": "PipelineRun",
							"metadata": {
								"generateName": "` + pipelineRunName + `-"
							},
							"spec": {
								"pipelineRef": {
									"name": "` + pipelineName + `"
								},
								"serviceAccountName": "` + serviceAccount + `",
								"params": [
									{"name": "git-repo", "value": "$(params.git-repo-url)"},
									{"name": "git-revision", "value": "$(params.git-revision)"}
								],
								"workspaces": [
									{
										"name": "source-code",
										"persistentVolumeClaim": {
											"claimName": "` + pvcClaimName + `"
										}
									}
								]
							}
						}`),
					},
				},
			},
		},
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TriggerTemplate to unstructured: %w", err)
	}

	return &unstructured.Unstructured{Object: unstructuredObj}, nil
}
