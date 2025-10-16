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

package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("health")

// Checker implements health check logic for the operator
type Checker struct {
	client        client.Client
	lastCheckTime time.Time
	mu            sync.RWMutex
}

// NewChecker creates a new health checker
func NewChecker(client client.Client) *Checker {
	return &Checker{
		client:        client,
		lastCheckTime: time.Now(),
	}
}

// LivenessCheck checks if the operator is alive
// This is a simple check that always returns healthy unless the process is deadlocked
func (c *Checker) LivenessCheck(req *http.Request) error {
	// Update last check time
	c.mu.Lock()
	c.lastCheckTime = time.Now()
	c.mu.Unlock()

	log.V(2).Info("Liveness check succeeded")
	return nil
}

// ReadinessCheck checks if the operator is ready to serve traffic
// This checks if we can connect to the Kubernetes API server
func (c *Checker) ReadinessCheck(req *http.Request) error {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	// Try to list namespaces to verify API connectivity
	if err := c.checkAPIServerConnection(ctx); err != nil {
		log.Error(err, "Readiness check failed: unable to connect to API server")
		return fmt.Errorf("API server connection check failed: %w", err)
	}

	log.V(2).Info("Readiness check succeeded")
	return nil
}

// checkAPIServerConnection verifies connectivity to the Kubernetes API server
func (c *Checker) checkAPIServerConnection(ctx context.Context) error {
	// Try to list a small number of namespaces as a lightweight check
	list := &corev1.NamespaceList{}
	if err := c.client.List(ctx, list, client.Limit(1)); err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}
	return nil
}

// Checker for webhook readiness
func WebhookReadinessCheck() healthz.Checker {
	return func(req *http.Request) error {
		// Check if webhook certificates are valid
		// This is a placeholder - actual implementation would check cert validity
		log.V(2).Info("Webhook readiness check succeeded")
		return nil
	}
}

// Checker for leader election
func LeaderElectionCheck(isLeader func() bool) healthz.Checker {
	return func(req *http.Request) error {
		if !isLeader() {
			return fmt.Errorf("not the leader")
		}
		log.V(2).Info("Leader election check succeeded")
		return nil
	}
}

// GetLastCheckTime returns the last time the health check was performed
func (c *Checker) GetLastCheckTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastCheckTime
}
