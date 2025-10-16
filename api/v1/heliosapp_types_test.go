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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// testConditionHelper is a helper function to test condition-based methods
func testConditionHelper(t *testing.T, conditionType string, method func(*HeliosApp) bool) {
	tests := []struct {
		name      string
		heliosApp *HeliosApp
		expected  bool
	}{
		{
			name: conditionType + " condition is true",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{
						{
							Type:   conditionType,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: conditionType + " condition is false",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{
						{
							Type:   conditionType,
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			expected: false,
		},
		{
			name: conditionType + " condition is unknown",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{
						{
							Type:   conditionType,
							Status: metav1.ConditionUnknown,
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "no " + conditionType + " condition",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := method(tt.heliosApp)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHeliosApp_IsReady(t *testing.T) {
	testConditionHelper(t, ConditionReady, (*HeliosApp).IsReady)
}

func TestHeliosApp_IsSynced(t *testing.T) {
	testConditionHelper(t, ConditionSynced, (*HeliosApp).IsSynced)
}

func TestHeliosApp_IsBuildSucceeded(t *testing.T) {
	testConditionHelper(t, ConditionBuildSucceeded, (*HeliosApp).IsBuildSucceeded)
}

func TestHeliosApp_GetCondition(t *testing.T) {
	tests := []struct {
		name           string
		heliosApp      *HeliosApp
		conditionType  string
		expectedStatus metav1.ConditionStatus
		expectedFound  bool
	}{
		{
			name: "condition exists",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{
						{
							Type:   ConditionReady,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			conditionType:  ConditionReady,
			expectedStatus: metav1.ConditionTrue,
			expectedFound:  true,
		},
		{
			name: "condition does not exist",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{},
				},
			},
			conditionType:  ConditionReady,
			expectedStatus: metav1.ConditionUnknown,
			expectedFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := tt.heliosApp.GetCondition(tt.conditionType)
			assert.Equal(t, tt.expectedStatus, condition.Status)
			// When condition is not found, GetCondition returns empty condition with Unknown status
			if tt.expectedFound {
				assert.NotEmpty(t, condition.Type)
			}
		})
	}
}

func TestHeliosApp_SetCondition(t *testing.T) {
	tests := []struct {
		name           string
		heliosApp      *HeliosApp
		conditionType  string
		status         metav1.ConditionStatus
		reason         string
		message        string
		expectedCount  int
		expectedStatus metav1.ConditionStatus
	}{
		{
			name: "add new condition",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{},
				},
			},
			conditionType:  ConditionReady,
			status:         metav1.ConditionTrue,
			reason:         "TestReason",
			message:        "Test message",
			expectedCount:  1,
			expectedStatus: metav1.ConditionTrue,
		},
		{
			name: "update existing condition",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{
						{
							Type:   ConditionReady,
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			conditionType:  ConditionReady,
			status:         metav1.ConditionTrue,
			reason:         "UpdatedReason",
			message:        "Updated message",
			expectedCount:  1,
			expectedStatus: metav1.ConditionTrue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.heliosApp.SetCondition(tt.conditionType, tt.status, tt.reason, tt.message)

			assert.Len(t, tt.heliosApp.Status.Conditions, tt.expectedCount)

			condition := tt.heliosApp.GetCondition(tt.conditionType)
			assert.Equal(t, tt.expectedStatus, condition.Status)
			assert.Equal(t, tt.reason, condition.Reason)
			assert.Equal(t, tt.message, condition.Message)
		})
	}
}

func TestHeliosApp_GetStatusMessage(t *testing.T) {
	tests := []struct {
		name      string
		heliosApp *HeliosApp
		expected  string
	}{
		{
			name: "with ready condition",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{
						{
							Type:    ConditionReady,
							Status:  metav1.ConditionTrue,
							Message: "Application is ready",
						},
					},
				},
			},
			expected: "Application is ready",
		},
		{
			name: "no ready condition",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.heliosApp.GetStatusMessage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHeliosApp_GetStatusReason(t *testing.T) {
	tests := []struct {
		name      string
		heliosApp *HeliosApp
		expected  string
	}{
		{
			name: "with ready condition",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{
						{
							Type:   ConditionReady,
							Status: metav1.ConditionTrue,
							Reason: "Ready",
						},
					},
				},
			},
			expected: "Ready",
		},
		{
			name: "no ready condition",
			heliosApp: &HeliosApp{
				Status: HeliosAppStatus{
					Conditions: []metav1.Condition{},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.heliosApp.GetStatusReason()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHeliosApp_DeepCopy(t *testing.T) {
	original := &HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
		Spec: HeliosAppSpec{
			GitRepo:   "https://github.com/test/repo",
			ImageRepo: "test/image",
			Port:      8080,
			Replicas:  3,
		},
		Status: HeliosAppStatus{
			Conditions: []metav1.Condition{
				{
					Type:   ConditionReady,
					Status: metav1.ConditionTrue,
				},
			},
		},
	}

	copied := original.DeepCopy()

	// Verify it's a different object (different pointers)
	assert.NotSame(t, original, copied)

	// Verify the content is the same
	assert.Equal(t, original.Name, copied.Name)
	assert.Equal(t, original.Namespace, copied.Namespace)
	assert.Equal(t, original.Spec.GitRepo, copied.Spec.GitRepo)
	assert.Equal(t, original.Spec.ImageRepo, copied.Spec.ImageRepo)
	assert.Equal(t, original.Spec.Port, copied.Spec.Port)
	assert.Equal(t, original.Spec.Replicas, copied.Spec.Replicas)
	assert.Len(t, copied.Status.Conditions, 1)
	assert.Equal(t, original.Status.Conditions[0].Type, copied.Status.Conditions[0].Type)
	assert.Equal(t, original.Status.Conditions[0].Status, copied.Status.Conditions[0].Status)
}

func TestHeliosApp_DeepCopyObject(t *testing.T) {
	original := &HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
		Spec: HeliosAppSpec{
			GitRepo:   "https://github.com/test/repo",
			ImageRepo: "test/image",
			Port:      8080,
			Replicas:  3,
		},
	}

	copied := original.DeepCopyObject()

	// Verify it's a different object (different pointers)
	assert.NotSame(t, original, copied)

	// Verify it's the correct type
	copiedHeliosApp, ok := copied.(*HeliosApp)
	assert.True(t, ok)
	assert.Equal(t, original.Name, copiedHeliosApp.Name)
	assert.Equal(t, original.Namespace, copiedHeliosApp.Namespace)
}

func TestHeliosAppList_DeepCopy(t *testing.T) {
	original := &HeliosAppList{
		Items: []HeliosApp{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app1",
					Namespace: "default",
				},
				Spec: HeliosAppSpec{
					GitRepo:   "https://github.com/test/repo1",
					ImageRepo: "test/image1",
					Port:      8080,
					Replicas:  1,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app2",
					Namespace: "default",
				},
				Spec: HeliosAppSpec{
					GitRepo:   "https://github.com/test/repo2",
					ImageRepo: "test/image2",
					Port:      8081,
					Replicas:  2,
				},
			},
		},
	}

	copied := original.DeepCopy()

	// Verify it's a different object (different pointers)
	assert.NotSame(t, original, copied)

	// Verify the content is the same
	assert.Len(t, copied.Items, 2)
	assert.Equal(t, original.Items[0].Name, copied.Items[0].Name)
	assert.Equal(t, original.Items[1].Name, copied.Items[1].Name)
}

func TestHeliosAppList_DeepCopyObject(t *testing.T) {
	original := &HeliosAppList{
		Items: []HeliosApp{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app1",
					Namespace: "default",
				},
			},
		},
	}

	copied := original.DeepCopyObject()

	// Verify it's a different object (different pointers)
	assert.NotSame(t, original, copied)

	// Verify it's the correct type
	copiedList, ok := copied.(*HeliosAppList)
	assert.True(t, ok)
	assert.Len(t, copiedList.Items, 1)
	assert.Equal(t, original.Items[0].Name, copiedList.Items[0].Name)
}
