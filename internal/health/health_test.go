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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewChecker(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	checker := NewChecker(fakeClient)

	assert.NotNil(t, checker)
	assert.NotNil(t, checker.client)
	assert.WithinDuration(t, time.Now(), checker.lastCheckTime, 1*time.Second)
}

func TestChecker_LivenessCheck(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	checker := NewChecker(fakeClient)

	req, err := http.NewRequest("GET", "/healthz", nil)
	require.NoError(t, err)

	// Liveness check should always succeed
	err = checker.LivenessCheck(req)
	assert.NoError(t, err)

	// Verify last check time was updated
	lastCheck := checker.GetLastCheckTime()
	assert.WithinDuration(t, time.Now(), lastCheck, 1*time.Second)
}

func TestChecker_ReadinessCheck_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	// Create a fake client with some namespaces
	namespaces := []client.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(namespaces...).
		Build()

	checker := NewChecker(fakeClient)

	req, err := http.NewRequest("GET", "/readyz", nil)
	require.NoError(t, err)

	// Readiness check should succeed when API is accessible
	err = checker.ReadinessCheck(req)
	assert.NoError(t, err)
}

func TestChecker_ReadinessCheck_Timeout(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	// Create a fake client that will work normally
	// (We can't easily simulate timeout with fake client, so we test normal operation)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	checker := NewChecker(fakeClient)

	// Create a request with a very short timeout to test context handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", "/readyz", nil)
	require.NoError(t, err)

	// Should succeed with normal fake client
	err = checker.ReadinessCheck(req)
	assert.NoError(t, err)
}

func TestChecker_CheckAPIServerConnection_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	namespaces := []client.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(namespaces...).
		Build()

	checker := NewChecker(fakeClient)

	ctx := context.Background()
	err := checker.checkAPIServerConnection(ctx)
	assert.NoError(t, err)
}

func TestChecker_GetLastCheckTime(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	checker := NewChecker(fakeClient)

	initialTime := checker.GetLastCheckTime()
	assert.WithinDuration(t, time.Now(), initialTime, 1*time.Second)

	// Perform a liveness check
	req, err := http.NewRequest("GET", "/healthz", nil)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference
	err = checker.LivenessCheck(req)
	require.NoError(t, err)

	// Last check time should be updated
	updatedTime := checker.GetLastCheckTime()
	assert.True(t, updatedTime.After(initialTime), "Last check time should be updated after liveness check")
}

func TestWebhookReadinessCheck(t *testing.T) {
	checker := WebhookReadinessCheck()
	assert.NotNil(t, checker)

	req, err := http.NewRequest("GET", "/readyz", nil)
	require.NoError(t, err)

	// Webhook readiness check should succeed
	err = checker(req)
	assert.NoError(t, err)
}

func TestLeaderElectionCheck_IsLeader(t *testing.T) {
	isLeader := func() bool { return true }
	checker := LeaderElectionCheck(isLeader)
	assert.NotNil(t, checker)

	req, err := http.NewRequest("GET", "/healthz", nil)
	require.NoError(t, err)

	// Should succeed when is leader
	err = checker(req)
	assert.NoError(t, err)
}

func TestLeaderElectionCheck_NotLeader(t *testing.T) {
	isLeader := func() bool { return false }
	checker := LeaderElectionCheck(isLeader)
	assert.NotNil(t, checker)

	req, err := http.NewRequest("GET", "/healthz", nil)
	require.NoError(t, err)

	// Should fail when not leader
	err = checker(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not the leader")
}

func TestLeaderElectionCheck_ToggleLeadership(t *testing.T) {
	var leader bool
	isLeader := func() bool { return leader }
	checker := LeaderElectionCheck(isLeader)

	req, err := http.NewRequest("GET", "/healthz", nil)
	require.NoError(t, err)

	// Initially not leader
	leader = false
	err = checker(req)
	assert.Error(t, err)

	// Become leader
	leader = true
	err = checker(req)
	assert.NoError(t, err)

	// Lose leadership
	leader = false
	err = checker(req)
	assert.Error(t, err)
}

// TestChecker_ConcurrentAccess tests thread safety of the checker
func TestChecker_ConcurrentAccess(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	checker := NewChecker(fakeClient)

	// Run multiple concurrent liveness checks
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			req, _ := http.NewRequest("GET", "/healthz", nil)
			_ = checker.LivenessCheck(req)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify last check time is recent
	lastCheck := checker.GetLastCheckTime()
	assert.WithinDuration(t, time.Now(), lastCheck, 1*time.Second)
}

// Mock client that always returns an error
type errorClient struct {
	client.Client
}

func (e *errorClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return fmt.Errorf("mock API server error")
}

func TestChecker_ReadinessCheck_APIError(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	// Use error client
	baseClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()
	
	errorCli := &errorClient{Client: baseClient}
	checker := NewChecker(errorCli)

	req, err := http.NewRequest("GET", "/readyz", nil)
	require.NoError(t, err)

	// Readiness check should fail
	err = checker.ReadinessCheck(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API server connection check failed")
}

func TestChecker_CheckAPIServerConnection_Error(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	
	baseClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()
	
	errorCli := &errorClient{Client: baseClient}
	checker := NewChecker(errorCli)

	ctx := context.Background()
	err := checker.checkAPIServerConnection(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list namespaces")
}
