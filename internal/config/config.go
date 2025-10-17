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

// Package config provides configuration management for the Helios operator,
// including environment variable parsing and validation.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// OperatorConfig holds all operator configuration.
type OperatorConfig struct {
	// Reconciliation settings
	ReconcileInterval       time.Duration
	MaxConcurrentReconciles int

	// Resource defaults
	DefaultReplicas       int32
	DefaultPort           int32
	DefaultServiceAccount string

	// Retry settings
	MaxRetries   int
	RetryBackoff time.Duration

	// Feature flags
	EnableMetrics        bool
	EnableWebhooks       bool
	EnableLeaderElection bool

	// Timeouts
	ReconcileTimeout time.Duration
	APICallTimeout   time.Duration

	// Namespace watching
	WatchNamespace string // Empty string means all namespaces
}

// Default returns a default operator configuration.
func Default() *OperatorConfig {
	return &OperatorConfig{
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
	}
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() (*OperatorConfig, error) {
	cfg := Default()

	if err := loadReconciliationSettings(cfg); err != nil {
		return nil, err
	}

	if err := loadResourceDefaults(cfg); err != nil {
		return nil, err
	}

	if err := loadRetrySettings(cfg); err != nil {
		return nil, err
	}

	if err := loadFeatureFlags(cfg); err != nil {
		return nil, err
	}

	if err := loadTimeoutSettings(cfg); err != nil {
		return nil, err
	}

	if err := loadNamespaceSettings(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadReconciliationSettings loads reconciliation-related configuration.
func loadReconciliationSettings(cfg *OperatorConfig) error {
	if val := os.Getenv("RECONCILE_INTERVAL"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid RECONCILE_INTERVAL: %w", err)
		}
		cfg.ReconcileInterval = duration
	}

	if val := os.Getenv("MAX_CONCURRENT_RECONCILES"); val != "" {
		num, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid MAX_CONCURRENT_RECONCILES: %w", err)
		}
		if num < 1 {
			return fmt.Errorf("MAX_CONCURRENT_RECONCILES must be >= 1, got %d", num)
		}
		cfg.MaxConcurrentReconciles = num
	}

	return nil
}

// loadResourceDefaults loads resource default configuration.
func loadResourceDefaults(cfg *OperatorConfig) error {
	if val := os.Getenv("DEFAULT_REPLICAS"); val != "" {
		num, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid DEFAULT_REPLICAS: %w", err)
		}
		if num < 0 {
			return fmt.Errorf("DEFAULT_REPLICAS must be >= 0, got %d", num)
		}
		cfg.DefaultReplicas = int32(num)
	}

	if val := os.Getenv("DEFAULT_PORT"); val != "" {
		num, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid DEFAULT_PORT: %w", err)
		}
		if num < 1 || num > 65535 {
			return fmt.Errorf("DEFAULT_PORT must be between 1 and 65535, got %d", num)
		}
		cfg.DefaultPort = int32(num)
	}

	if val := os.Getenv("DEFAULT_SERVICE_ACCOUNT"); val != "" {
		cfg.DefaultServiceAccount = val
	}

	return nil
}

// loadRetrySettings loads retry-related configuration.
func loadRetrySettings(cfg *OperatorConfig) error {
	if val := os.Getenv("MAX_RETRIES"); val != "" {
		num, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid MAX_RETRIES: %w", err)
		}
		if num < 0 {
			return fmt.Errorf("MAX_RETRIES must be >= 0, got %d", num)
		}
		cfg.MaxRetries = num
	}

	if val := os.Getenv("RETRY_BACKOFF"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid RETRY_BACKOFF: %w", err)
		}
		cfg.RetryBackoff = duration
	}

	return nil
}

// loadFeatureFlags loads feature flag configuration.
func loadFeatureFlags(cfg *OperatorConfig) error {
	if val := os.Getenv("ENABLE_METRICS"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid ENABLE_METRICS: %w", err)
		}
		cfg.EnableMetrics = enabled
	}

	if val := os.Getenv("ENABLE_WEBHOOKS"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid ENABLE_WEBHOOKS: %w", err)
		}
		cfg.EnableWebhooks = enabled
	}

	if val := os.Getenv("ENABLE_LEADER_ELECTION"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid ENABLE_LEADER_ELECTION: %w", err)
		}
		cfg.EnableLeaderElection = enabled
	}

	return nil
}

// loadTimeoutSettings loads timeout-related configuration.
func loadTimeoutSettings(cfg *OperatorConfig) error {
	if val := os.Getenv("RECONCILE_TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid RECONCILE_TIMEOUT: %w", err)
		}
		cfg.ReconcileTimeout = duration
	}

	if val := os.Getenv("API_CALL_TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid API_CALL_TIMEOUT: %w", err)
		}
		cfg.APICallTimeout = duration
	}

	return nil
}

// loadNamespaceSettings loads namespace-related configuration.
func loadNamespaceSettings(cfg *OperatorConfig) error {
	if val := os.Getenv("WATCH_NAMESPACE"); val != "" {
		cfg.WatchNamespace = val
	}

	return nil
}

// Validate validates the configuration.
func (c *OperatorConfig) Validate() error {
	if c.ReconcileInterval < 1*time.Second {
		return fmt.Errorf("ReconcileInterval must be >= 1s, got %v", c.ReconcileInterval)
	}

	if c.MaxConcurrentReconciles < 1 {
		return fmt.Errorf("MaxConcurrentReconciles must be >= 1, got %d", c.MaxConcurrentReconciles)
	}

	if c.DefaultReplicas < 0 {
		return fmt.Errorf("DefaultReplicas must be >= 0, got %d", c.DefaultReplicas)
	}

	if c.DefaultPort < 1 || c.DefaultPort > 65535 {
		return fmt.Errorf("DefaultPort must be between 1 and 65535, got %d", c.DefaultPort)
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries must be >= 0, got %d", c.MaxRetries)
	}

	if c.RetryBackoff < 0 {
		return fmt.Errorf("RetryBackoff must be >= 0, got %v", c.RetryBackoff)
	}

	if c.ReconcileTimeout < 1*time.Second {
		return fmt.Errorf("ReconcileTimeout must be >= 1s, got %v", c.ReconcileTimeout)
	}

	if c.APICallTimeout < 1*time.Second {
		return fmt.Errorf("APICallTimeout must be >= 1s, got %v", c.APICallTimeout)
	}

	return nil
}

// String returns a string representation of the configuration.
func (c *OperatorConfig) String() string {
	return fmt.Sprintf(
		"OperatorConfig{ReconcileInterval: %v, MaxConcurrentReconciles: %d, DefaultReplicas: %d, "+
			"DefaultPort: %d, MaxRetries: %d, EnableMetrics: %v, EnableWebhooks: %v, "+
			"EnableLeaderElection: %v, WatchNamespace: %q}",
		c.ReconcileInterval, c.MaxConcurrentReconciles, c.DefaultReplicas,
		c.DefaultPort, c.MaxRetries, c.EnableMetrics, c.EnableWebhooks,
		c.EnableLeaderElection, c.WatchNamespace,
	)
}
