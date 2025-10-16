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

func TestHeliosApp_Default(t *testing.T) {
	tests := []struct {
		name      string
		heliosApp *HeliosApp
		expected  *HeliosApp
	}{
		{
			name: "should set default values",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			expected: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					GitBranch:      "main",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					Replicas:       1,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
					GitopsPath:     "test-app",
					GitopsBranch:   "main",
				},
			},
		},
		{
			name: "should not override existing values",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					GitBranch:      "develop",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					Replicas:       3,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
					GitopsPath:     "apps/test-app",
					GitopsBranch:   "develop",
				},
			},
			expected: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					GitBranch:      "develop",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					Replicas:       3,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
					GitopsPath:     "apps/test-app",
					GitopsBranch:   "develop",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.heliosApp.Default()
			assert.Equal(t, tt.expected, tt.heliosApp)
		})
	}
}

func TestHeliosApp_ValidateCreate(t *testing.T) {
	tests := []struct {
		name      string
		heliosApp *HeliosApp
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid helios app",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid git repo URL",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "invalid-url",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: true,
			errMsg:  "gitRepo must be a valid HTTP/HTTPS URL",
		},
		{
			name: "invalid image repo format",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "invalid:image:repo:format",
					Port:           8080,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: true,
			errMsg:  "imageRepo has invalid format",
		},
		{
			name: "invalid port range",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           0,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "missing service account",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					ServiceAccount: "",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: true,
			errMsg:  "serviceAccount is required",
		},
		{
			name: "missing webhook secret",
			heliosApp: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					ServiceAccount: "test-sa",
					WebhookSecret:  "",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: true,
			errMsg:  "webhookSecret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := tt.heliosApp.ValidateCreate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			// Check warnings
			assert.NotNil(t, warnings)
		})
	}
}

func TestHeliosApp_ValidateUpdate(t *testing.T) {
	tests := []struct {
		name    string
		old     *HeliosApp
		new     *HeliosApp
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid update",
			old: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			new: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					Replicas:       2,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid update - invalid port",
			old: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           8080,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			new: &HeliosApp{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-app",
				},
				Spec: HeliosAppSpec{
					GitRepo:        "https://github.com/example/test-app",
					ImageRepo:      "test-registry.com/test-app",
					Port:           0,
					ServiceAccount: "test-sa",
					WebhookSecret:  "test-secret",
					GitopsRepo:     "https://github.com/example/test-app-manifests",
				},
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := tt.new.ValidateUpdate(tt.old)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			// Check warnings
			assert.NotNil(t, warnings)
		})
	}
}

func TestHeliosApp_ValidateDelete(t *testing.T) {
	heliosApp := &HeliosApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-app",
		},
		Spec: HeliosAppSpec{
			GitRepo:        "https://github.com/example/test-app",
			ImageRepo:      "test-registry.com/test-app",
			Port:           8080,
			ServiceAccount: "test-sa",
			WebhookSecret:  "test-secret",
			GitopsRepo:     "https://github.com/example/test-app-manifests",
		},
	}

	warnings, err := heliosApp.ValidateDelete()
	assert.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		field   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid HTTPS URL",
			url:     "https://github.com/example/repo",
			field:   "gitRepo",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://git.example.com/repo",
			field:   "gitRepo",
			wantErr: false,
		},
		{
			name:    "invalid URL scheme",
			url:     "ftp://github.com/example/repo",
			field:   "gitRepo",
			wantErr: true,
			errMsg:  "gitRepo must be a valid HTTP/HTTPS URL",
		},
		{
			name:    "invalid URL format",
			url:     "not-a-url",
			field:   "gitRepo",
			wantErr: true,
			errMsg:  "gitRepo must be a valid HTTP/HTTPS URL",
		},
		{
			name:    "empty URL",
			url:     "",
			field:   "gitRepo",
			wantErr: true,
			errMsg:  "gitRepo is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGitURL(tt.url, tt.field)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateImageRepo(t *testing.T) {
	tests := []struct {
		name      string
		imageRepo string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid image repo without tag",
			imageRepo: "registry.example.com/my-app",
			wantErr:   false,
		},
		{
			name:      "valid image repo with tag",
			imageRepo: "registry.example.com/my-app:v1.0.0",
			wantErr:   false,
		},
		{
			name:      "invalid format - too many colons",
			imageRepo: "registry.example.com:my:app:tag",
			wantErr:   true,
			errMsg:    "imageRepo has invalid format",
		},
		{
			name:      "empty image name",
			imageRepo: ":v1.0.0",
			wantErr:   true,
			errMsg:    "imageRepo image name cannot be empty",
		},
		{
			name:      "empty tag with colon",
			imageRepo: "registry.example.com/my-app:",
			wantErr:   true,
			errMsg:    "imageRepo tag cannot be empty when colon is present",
		},
		{
			name:      "tag with whitespace",
			imageRepo: "registry.example.com/my-app:v1 0",
			wantErr:   true,
			errMsg:    "imageRepo tag cannot contain whitespace characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageRepo(tt.imageRepo)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidDNSSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid single label",
			input:    "test",
			expected: true,
		},
		{
			name:     "valid multi-label",
			input:    "test.example.com",
			expected: true,
		},
		{
			name:     "valid with hyphens",
			input:    "test-app.example.com",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "too long",
			input:    "a" + string(make([]byte, 253)),
			expected: false,
		},
		{
			name:     "invalid characters",
			input:    "test@example.com",
			expected: false,
		},
		{
			name:     "starts with hyphen",
			input:    "-test.example.com",
			expected: false,
		},
		{
			name:     "ends with hyphen",
			input:    "test.example.com-",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDNSSubdomain(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidDNSLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid label",
			input:    "test",
			expected: true,
		},
		{
			name:     "valid with hyphens",
			input:    "test-app",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "too long",
			input:    "a" + string(make([]byte, 63)),
			expected: false,
		},
		{
			name:     "invalid characters",
			input:    "test@app",
			expected: false,
		},
		{
			name:     "starts with hyphen",
			input:    "-test",
			expected: false,
		},
		{
			name:     "ends with hyphen",
			input:    "test-",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDNSLabel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{
			name:     "lowercase letter",
			char:     'a',
			expected: true,
		},
		{
			name:     "uppercase letter",
			char:     'A',
			expected: true,
		},
		{
			name:     "digit",
			char:     '1',
			expected: true,
		},
		{
			name:     "hyphen",
			char:     '-',
			expected: false,
		},
		{
			name:     "special character",
			char:     '@',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlphanumeric(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}
