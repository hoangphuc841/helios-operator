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

package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv(t *testing.T) {
	// Save original environment and restore after test
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range origEnv {
			pair := splitEnv(env)
			if len(pair) == 2 {
				_ = os.Setenv(pair[0], pair[1]) // Ignore error for test cleanup
			}
		}
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		expected    *OperatorConfig
		expectError bool
	}{
		{
			name:    "load with default values",
			envVars: map[string]string{},
			expected: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             8080,
				DefaultServiceAccount:   "default",
				MaxRetries:              3,
				RetryBackoff:            30 * time.Second,
				EnableMetrics:           true,
				EnableWebhooks:          true,
				EnableLeaderElection:    true,
				ReconcileTimeout:        10 * time.Minute,
				APICallTimeout:          30 * time.Second,
				WatchNamespace:          "",
			},
			expectError: false,
		},
		{
			name: "load with custom values",
			envVars: map[string]string{
				"RECONCILE_INTERVAL":        "10m",
				"MAX_CONCURRENT_RECONCILES": "5",
				"DEFAULT_REPLICAS":          "3",
				"DEFAULT_PORT":              "9000",
				"DEFAULT_SERVICE_ACCOUNT":   "custom-sa",
				"MAX_RETRIES":               "5",
				"RETRY_BACKOFF":             "60s",
				"ENABLE_METRICS":            "false",
				"ENABLE_WEBHOOKS":           "false",
				"ENABLE_LEADER_ELECTION":    "false",
				"RECONCILE_TIMEOUT":         "15m",
				"API_CALL_TIMEOUT":          "45s",
				"WATCH_NAMESPACE":           "my-namespace",
			},
			expected: &OperatorConfig{
				ReconcileInterval:       10 * time.Minute,
				MaxConcurrentReconciles: 5,
				DefaultReplicas:         3,
				DefaultPort:             9000,
				DefaultServiceAccount:   "custom-sa",
				MaxRetries:              5,
				RetryBackoff:            60 * time.Second,
				EnableMetrics:           false,
				EnableWebhooks:          false,
				EnableLeaderElection:    false,
				ReconcileTimeout:        15 * time.Minute,
				APICallTimeout:          45 * time.Second,
				WatchNamespace:          "my-namespace",
			},
			expectError: false,
		},
		{
			name: "invalid reconcile interval",
			envVars: map[string]string{
				"RECONCILE_INTERVAL": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid max concurrent reconciles",
			envVars: map[string]string{
				"MAX_CONCURRENT_RECONCILES": "invalid",
			},
			expectError: true,
		},
		{
			name: "max concurrent reconciles too low",
			envVars: map[string]string{
				"MAX_CONCURRENT_RECONCILES": "0",
			},
			expectError: true,
		},
		{
			name: "invalid default replicas",
			envVars: map[string]string{
				"DEFAULT_REPLICAS": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid default port",
			envVars: map[string]string{
				"DEFAULT_PORT": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid max retries",
			envVars: map[string]string{
				"MAX_RETRIES": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid retry backoff",
			envVars: map[string]string{
				"RETRY_BACKOFF": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid enable metrics",
			envVars: map[string]string{
				"ENABLE_METRICS": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid reconcile timeout",
			envVars: map[string]string{
				"RECONCILE_TIMEOUT": "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v) // Ignore error for test setup
			}

			// Load configuration
			cfg, err := LoadFromEnv()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, cfg)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.NotNil(t, cfg)
	assert.Equal(t, 5*time.Minute, cfg.ReconcileInterval)
	assert.Equal(t, 3, cfg.MaxConcurrentReconciles)
	assert.Equal(t, int32(1), cfg.DefaultReplicas)
	assert.Equal(t, int32(8080), cfg.DefaultPort)
	assert.Equal(t, "default", cfg.DefaultServiceAccount)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 30*time.Second, cfg.RetryBackoff)
	assert.True(t, cfg.EnableMetrics)
	assert.True(t, cfg.EnableWebhooks)
	assert.True(t, cfg.EnableLeaderElection)
	assert.Equal(t, 10*time.Minute, cfg.ReconcileTimeout)
	assert.Equal(t, 30*time.Second, cfg.APICallTimeout)
	assert.Equal(t, "", cfg.WatchNamespace)
}

func TestOperatorConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *OperatorConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             8080,
				DefaultServiceAccount:   "default",
				MaxRetries:              3,
				RetryBackoff:            30 * time.Second,
				ReconcileTimeout:        10 * time.Minute,
				APICallTimeout:          30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "invalid reconcile interval - zero",
			config: &OperatorConfig{
				ReconcileInterval: 0,
			},
			expectError: true,
			errorMsg:    "ReconcileInterval must be >= 1s",
		},
		{
			name: "invalid max concurrent reconciles - zero",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 0,
			},
			expectError: true,
			errorMsg:    "MaxConcurrentReconciles must be >= 1",
		},
		{
			name: "invalid default replicas - negative",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         -1,
			},
			expectError: true,
			errorMsg:    "DefaultReplicas must be >= 0",
		},
		{
			name: "invalid default port - out of range low",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             0,
			},
			expectError: true,
			errorMsg:    "DefaultPort must be between 1 and 65535",
		},
		{
			name: "invalid default port - out of range high",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             70000,
			},
			expectError: true,
			errorMsg:    "DefaultPort must be between 1 and 65535",
		},
		{
			name: "invalid max retries - negative",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             8080,
				MaxRetries:              -1,
				RetryBackoff:            30 * time.Second,
				ReconcileTimeout:        10 * time.Minute,
				APICallTimeout:          30 * time.Second,
			},
			expectError: true,
			errorMsg:    "MaxRetries must be >= 0",
		},
		{
			name: "invalid retry backoff - negative",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             8080,
				MaxRetries:              3,
				RetryBackoff:            -1 * time.Second,
				ReconcileTimeout:        10 * time.Minute,
				APICallTimeout:          30 * time.Second,
			},
			expectError: true,
			errorMsg:    "RetryBackoff must be >= 0",
		},
		{
			name: "invalid reconcile timeout - zero",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             8080,
				MaxRetries:              3,
				RetryBackoff:            30 * time.Second,
				ReconcileTimeout:        0,
			},
			expectError: true,
			errorMsg:    "ReconcileTimeout must be >= 1s",
		},
		{
			name: "invalid api call timeout - zero",
			config: &OperatorConfig{
				ReconcileInterval:       5 * time.Minute,
				MaxConcurrentReconciles: 3,
				DefaultReplicas:         1,
				DefaultPort:             8080,
				MaxRetries:              3,
				RetryBackoff:            30 * time.Second,
				ReconcileTimeout:        10 * time.Minute,
				APICallTimeout:          0,
			},
			expectError: true,
			errorMsg:    "APICallTimeout must be >= 1s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to split environment variable string.
func splitEnv(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env}
}
