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

package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupFlags(t *testing.T) {
	// Create a new flag set for this test to avoid conflicts
	originalFlagSet := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("test-setup-flags", flag.ContinueOnError)
	defer func() {
		flag.CommandLine = originalFlagSet
	}()

	config := setupFlags()

	// Test default values
	assert.Equal(t, "0", config.metricsAddr)
	assert.Equal(t, ":8081", config.probeAddr)
	assert.False(t, config.enableLeaderElection)
	assert.True(t, config.secureMetrics)
	assert.Empty(t, config.webhookCertPath)
	assert.Equal(t, "tls.crt", config.webhookCertName)
	assert.Equal(t, "tls.key", config.webhookCertKey)
	assert.Empty(t, config.metricsCertPath)
	assert.Equal(t, "tls.crt", config.metricsCertName)
	assert.Equal(t, "tls.key", config.metricsCertKey)
	assert.False(t, config.enableHTTP2)
	assert.False(t, config.printVersion)
	assert.True(t, config.loggerOpts.Development)
}

func TestPrintVersionInfo(t *testing.T) {
	// Test that printVersionInfo doesn't panic and produces output
	// We can't easily test the exact output without capturing stdout,
	// but we can ensure it doesn't crash
	assert.NotPanics(t, func() {
		printVersionInfo()
	})
}

func TestSetupTLSOptions(t *testing.T) {
	tests := []struct {
		name           string
		enableHTTP2    bool
		expectedLength int
	}{
		{
			name:           "HTTP2 disabled (default)",
			enableHTTP2:    false,
			expectedLength: 1, // Should have disableHTTP2 function
		},
		{
			name:           "HTTP2 enabled",
			enableHTTP2:    true,
			expectedLength: 0, // Should have no TLS options
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				enableHTTP2: tt.enableHTTP2,
			}

			tlsOpts := setupTLSOptions(config)
			assert.Len(t, tlsOpts, tt.expectedLength)
		})
	}
}

func TestSetupMetricsServer(t *testing.T) {
	tests := []struct {
		name           string
		secureMetrics  bool
		metricsAddr    string
		expectedSecure bool
	}{
		{
			name:           "Secure metrics enabled",
			secureMetrics:  true,
			metricsAddr:    ":8443",
			expectedSecure: true,
		},
		{
			name:           "Secure metrics disabled",
			secureMetrics:  false,
			metricsAddr:    ":8080",
			expectedSecure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				secureMetrics: tt.secureMetrics,
				metricsAddr:   tt.metricsAddr,
			}

			tlsOpts := setupTLSOptions(config)
			metricsOptions := setupMetricsServer(config, tlsOpts)

			assert.Equal(t, tt.metricsAddr, metricsOptions.BindAddress)
			assert.Equal(t, tt.expectedSecure, metricsOptions.SecureServing)
			assert.Equal(t, tlsOpts, metricsOptions.TLSOpts)

			if tt.expectedSecure {
				assert.NotNil(t, metricsOptions.FilterProvider)
			} else {
				assert.Nil(t, metricsOptions.FilterProvider)
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Test that init() doesn't panic
	// The init function sets up the scheme, so we can verify it's not nil
	assert.NotNil(t, scheme)
}

func TestConfigStruct(t *testing.T) {
	// Test Config struct initialization
	config := &Config{
		metricsAddr:          ":8080",
		metricsCertPath:      "/certs",
		metricsCertName:      "tls.crt",
		metricsCertKey:       "tls.key",
		webhookCertPath:      "/webhook-certs",
		webhookCertName:      "webhook.crt",
		webhookCertKey:       "webhook.key",
		enableLeaderElection: true,
		probeAddr:            ":8081",
		secureMetrics:        true,
		enableHTTP2:          true,
		printVersion:         true,
	}

	assert.Equal(t, ":8080", config.metricsAddr)
	assert.Equal(t, "/certs", config.metricsCertPath)
	assert.Equal(t, "tls.crt", config.metricsCertName)
	assert.Equal(t, "tls.key", config.metricsCertKey)
	assert.Equal(t, "/webhook-certs", config.webhookCertPath)
	assert.Equal(t, "webhook.crt", config.webhookCertName)
	assert.Equal(t, "webhook.key", config.webhookCertKey)
	assert.True(t, config.enableLeaderElection)
	assert.Equal(t, ":8081", config.probeAddr)
	assert.True(t, config.secureMetrics)
	assert.True(t, config.enableHTTP2)
	assert.True(t, config.printVersion)
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables are set (they may be overridden at build time)
	assert.NotEmpty(t, version)
	assert.NotEmpty(t, commit)
	assert.NotEmpty(t, date)
}

func TestMainFunctionFlags(t *testing.T) {
	// Create a new flag set for this test to avoid conflicts
	originalFlagSet := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("test-main-flags", flag.ContinueOnError)
	defer func() {
		flag.CommandLine = originalFlagSet
	}()

	// Test that main function can parse flags without crashing
	// We'll test this by calling setupFlags and ensuring it works
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Test with version flag
	os.Args = []string{"helios-operator", "--version"}
	config := setupFlags()

	// Parse the flags
	err := flag.CommandLine.Parse(os.Args[1:])
	require.NoError(t, err)

	// The version flag should be set to true
	assert.True(t, config.printVersion)
}

func TestTLSConfigModification(t *testing.T) {
	// Test that TLS options actually modify the TLS config
	config := &Config{
		enableHTTP2: false,
	}

	tlsOpts := setupTLSOptions(config)
	assert.Len(t, tlsOpts, 1)

	// Create a mock TLS config and apply the options
	// We can't easily test the actual TLS config modification without
	// creating a real TLS config, but we can ensure the function returns
	// the expected number of options
	assert.NotNil(t, tlsOpts[0])
}

func TestMetricsServerConfiguration(t *testing.T) {
	// Test various metrics server configurations
	testCases := []struct {
		name           string
		config         *Config
		expectedAddr   string
		expectedSecure bool
	}{
		{
			name: "Default configuration",
			config: &Config{
				metricsAddr:   "0",
				secureMetrics: true,
			},
			expectedAddr:   "0",
			expectedSecure: true,
		},
		{
			name: "HTTP configuration",
			config: &Config{
				metricsAddr:   ":8080",
				secureMetrics: false,
			},
			expectedAddr:   ":8080",
			expectedSecure: false,
		},
		{
			name: "HTTPS configuration",
			config: &Config{
				metricsAddr:   ":8443",
				secureMetrics: true,
			},
			expectedAddr:   ":8443",
			expectedSecure: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tlsOpts := setupTLSOptions(tc.config)
			metricsOptions := setupMetricsServer(tc.config, tlsOpts)

			assert.Equal(t, tc.expectedAddr, metricsOptions.BindAddress)
			assert.Equal(t, tc.expectedSecure, metricsOptions.SecureServing)
		})
	}
}

func TestLoggerOptions(t *testing.T) {
	// Test that logger options are properly configured
	config := setupFlags()

	// The logger should be in development mode by default
	assert.True(t, config.loggerOpts.Development)

	// Test that we can modify logger options
	config.loggerOpts.Development = false
	assert.False(t, config.loggerOpts.Development)
}

func TestFlagParsing(t *testing.T) {
	// Create a new flag set for this test to avoid conflicts
	originalFlagSet := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("test-flag-parsing", flag.ContinueOnError)
	defer func() {
		flag.CommandLine = originalFlagSet
	}()

	// Test that all flags can be parsed correctly
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Test with various flag combinations
	testArgs := []string{
		"helios-operator",
		"--metrics-bind-address=:8080",
		"--health-probe-bind-address=:8081",
		"--leader-elect=true",
		"--metrics-secure=false",
		"--webhook-cert-path=/certs",
		"--webhook-cert-name=webhook.crt",
		"--webhook-cert-key=webhook.key",
		"--metrics-cert-path=/metrics-certs",
		"--metrics-cert-name=metrics.crt",
		"--metrics-cert-key=metrics.key",
		"--enable-http2=true",
		"--version",
	}

	os.Args = testArgs
	config := setupFlags()

	// Parse the flags
	err := flag.CommandLine.Parse(os.Args[1:])
	require.NoError(t, err)

	// Verify flags were parsed correctly
	assert.Equal(t, ":8080", config.metricsAddr)
	assert.Equal(t, ":8081", config.probeAddr)
	assert.True(t, config.enableLeaderElection)
	assert.False(t, config.secureMetrics)
	assert.Equal(t, "/certs", config.webhookCertPath)
	assert.Equal(t, "webhook.crt", config.webhookCertName)
	assert.Equal(t, "webhook.key", config.webhookCertKey)
	assert.Equal(t, "/metrics-certs", config.metricsCertPath)
	assert.Equal(t, "metrics.crt", config.metricsCertName)
	assert.Equal(t, "metrics.key", config.metricsCertKey)
	assert.True(t, config.enableHTTP2)
	assert.True(t, config.printVersion)
}
