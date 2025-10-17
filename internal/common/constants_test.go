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

package common_test

import (
	"testing"

	"github.com/hoangphuc841/helios-operator/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "common.ConditionReady constant",
			constant: common.ConditionReady,
			expected: "Ready",
		},
		{
			name:     "common.ConditionSynced constant",
			constant: common.ConditionSynced,
			expected: "Synced",
		},
		{
			name:     "common.ConditionBuildSucceeded constant",
			constant: common.ConditionBuildSucceeded,
			expected: "BuildSucceeded",
		},
		{
			name:     "common.ConditionDeployed constant",
			constant: common.ConditionDeployed,
			expected: "Deployed",
		},
		{
			name:     "common.ConditionPipelineReady constant",
			constant: common.ConditionPipelineReady,
			expected: "PipelineReady",
		},
		{
			name:     "common.ConditionWebhookReady constant",
			constant: common.ConditionWebhookReady,
			expected: "WebhookReady",
		},
		{
			name:     "common.LabelAppName constant",
			constant: common.LabelAppName,
			expected: "app.kubernetes.io/name",
		},
		{
			name:     "common.LabelAppInstance constant",
			constant: common.LabelAppInstance,
			expected: "app.kubernetes.io/instance",
		},
		{
			name:     "common.LabelAppVersion constant",
			constant: common.LabelAppVersion,
			expected: "app.kubernetes.io/version",
		},
		{
			name:     "common.LabelAppComponent constant",
			constant: common.LabelAppComponent,
			expected: "app.kubernetes.io/component",
		},
		{
			name:     "common.LabelAppPartOf constant",
			constant: common.LabelAppPartOf,
			expected: "app.kubernetes.io/part-of",
		},
		{
			name:     "LabelAppManagedBy constant",
			constant: common.LabelAppManagedBy,
			expected: "app.kubernetes.io/managed-by",
		},
		{
			name:     "common.LabelManagedBy constant",
			constant: common.LabelManagedBy,
			expected: "helios.io/managed-by",
		},
		{
			name:     "common.LabelComponent constant",
			constant: common.LabelComponent,
			expected: "helios.io/component",
		},
		{
			name:     "common.FinalizerName constant",
			constant: common.FinalizerName,
			expected: "platform.helios.io/finalizer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}
