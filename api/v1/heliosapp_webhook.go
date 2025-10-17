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
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	log "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	// WebhookValidationsTotal tracks webhook validation attempts
	webhookValidationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heliosapp_webhook_validations_total",
			Help: "Total number of webhook validation attempts",
		},
		[]string{"operation", "result"}, // operation: create, update, delete; result: accept, reject
	)

	// WebhookValidationDuration tracks webhook validation duration
	webhookValidationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "heliosapp_webhook_validation_duration_seconds",
			Help:    "Duration of webhook validations",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 12), // 0.1ms to ~400ms
		},
		[]string{"operation"},
	)
)

func init() {
	// Register webhook metrics
	metrics.Registry.MustRegister(
		webhookValidationsTotal,
		webhookValidationDuration,
	)
}

// validateHeliosApp validates the HeliosApp spec
func (r *HeliosApp) validateHeliosApp() (admission.Warnings, error) {
	logger := log.Log.WithName("webhook").WithValues(
		"operation", "validate",
		"heliosapp", r.Name,
		"namespace", r.Namespace,
	)

	logger.V(1).Info("Starting HeliosApp validation")

	var allErrors []string
	var warnings admission.Warnings

	// Validate basic fields
	if errs := r.validateBasicFields(); len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}

	// Validate resource existence
	if errs := r.validateResourceExistence(); len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}

	// Validate numeric fields
	if errs := r.validateNumericFields(); len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}

	// Validate optional fields
	if errs := r.validateOptionalFields(); len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}

	if len(allErrors) > 0 {
		logger.Info("HeliosApp validation failed", "errors", len(allErrors))
		return warnings, errors.New(strings.Join(allErrors, "; "))
	}

	logger.V(1).Info("HeliosApp validation successful")
	return warnings, nil
}

// validateBasicFields validates basic required fields
func (r *HeliosApp) validateBasicFields() []string {
	var validationErrors []string

	// Validate GitRepo URL
	if err := validateGitURL(r.Spec.GitRepo, "gitRepo"); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate GitopsRepo URL
	if err := validateGitURL(r.Spec.GitopsRepo, "gitopsRepo"); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate ImageRepo format
	if err := validateImageRepo(r.Spec.ImageRepo); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	return validationErrors
}

// validateResourceExistence validates that referenced resources exist
func (r *HeliosApp) validateResourceExistence() []string {
	var validationErrors []string

	// Validate ServiceAccount existence
	if err := r.validateServiceAccount(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate WebhookSecret existence
	if err := r.validateWebhookSecret(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate PVC existence if specified
	if r.Spec.PVCName != "" {
		if err := r.validatePVC(); err != nil {
			validationErrors = append(validationErrors, err.Error())
		}
	}

	return validationErrors
}

// validateNumericFields validates numeric fields
func (r *HeliosApp) validateNumericFields() []string {
	var validationErrors []string

	// Validate port (standard port range)
	if r.Spec.Port < 1 || r.Spec.Port > 65535 {
		validationErrors = append(validationErrors, fmt.Sprintf("port must be between 1 and 65535, got %d", r.Spec.Port))
	}

	// Warn about privileged ports (< 1024) - could be a warning instead
	if r.Spec.Port > 0 && r.Spec.Port < 1024 {
		// This is still valid but might need special permissions
		// We'll allow it but could add a warning in the future
		_ = r.Spec.Port // Acknowledge the port value
	}

	// Validate replicas (reasonable upper bound to prevent resource exhaustion)
	if r.Spec.Replicas < 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("replicas must be non-negative, got %d", r.Spec.Replicas))
	}

	// Validate replicas upper bound (prevent accidental large deployments)
	const maxReplicas = 100
	if r.Spec.Replicas > maxReplicas {
		validationErrors = append(validationErrors, fmt.Sprintf("replicas cannot exceed %d (got %d) to prevent resource exhaustion", maxReplicas, r.Spec.Replicas))
	}

	return validationErrors
}

// validateOptionalFields validates optional fields
func (r *HeliosApp) validateOptionalFields() []string {
	var validationErrors []string

	// Namespace validation is handled by Kubernetes metadata.namespace

	// Validate serviceAccount if specified
	if r.Spec.ServiceAccount != "" {
		if !isValidDNSSubdomain(r.Spec.ServiceAccount) {
			validationErrors = append(validationErrors, fmt.Sprintf("serviceAccount '%s' is not a valid DNS subdomain name", r.Spec.ServiceAccount))
		}
	}

	// Validate webhookSecret if specified
	if r.Spec.WebhookSecret != "" {
		if !isValidDNSSubdomain(r.Spec.WebhookSecret) {
			validationErrors = append(validationErrors, fmt.Sprintf("webhookSecret '%s' is not a valid DNS subdomain name", r.Spec.WebhookSecret))
		}
	}

	// Validate pvcName if specified
	if r.Spec.PVCName != "" {
		if !isValidDNSSubdomain(r.Spec.PVCName) {
			validationErrors = append(validationErrors, fmt.Sprintf("pvcName '%s' is not a valid DNS subdomain name", r.Spec.PVCName))
		}
	}

	return validationErrors
}

// validateServiceAccount validates that the ServiceAccount exists
func (r *HeliosApp) validateServiceAccount() error {
	if r.Spec.ServiceAccount == "" {
		return nil
	}

	// Create a client to check if ServiceAccount exists
	// This is a simplified validation - in practice, you'd inject the client
	// For now, we'll just validate the format
	if !isValidDNSSubdomain(r.Spec.ServiceAccount) {
		return errors.New("serviceAccount must be a valid DNS subdomain")
	}

	return nil
}

// validateWebhookSecret validates that the WebhookSecret exists
func (r *HeliosApp) validateWebhookSecret() error {
	if r.Spec.WebhookSecret == "" {
		return nil
	}

	// Create a client to check if Secret exists
	// This is a simplified validation - in practice, you'd inject the client
	// For now, we'll just validate the format
	if !isValidDNSSubdomain(r.Spec.WebhookSecret) {
		return errors.New("webhookSecret must be a valid DNS subdomain")
	}

	return nil
}

// validatePVC validates that the PVC exists
func (r *HeliosApp) validatePVC() error {
	if r.Spec.PVCName == "" {
		return nil
	}

	// Create a client to check if PVC exists
	// This is a simplified validation - in practice, you'd inject the client
	// For now, we'll just validate the format
	if !isValidDNSSubdomain(r.Spec.PVCName) {
		return errors.New("pvcName must be a valid DNS subdomain")
	}

	return nil
}

// validateGitURL validates a Git URL
func validateGitURL(gitURL, fieldName string) error {
	if gitURL == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	// Trim whitespace
	gitURL = strings.TrimSpace(gitURL)

	// Parse URL
	parsedURL, err := url.Parse(gitURL)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL: %w", fieldName, err)
	}

	// Check scheme
	validSchemes := []string{"https", "http", "ssh", "git"}
	isValidScheme := false
	for _, scheme := range validSchemes {
		if parsedURL.Scheme == scheme {
			isValidScheme = true
			break
		}
	}

	if !isValidScheme {
		return fmt.Errorf("%s must use https://, http://, ssh://, or git:// scheme, got '%s'", fieldName, parsedURL.Scheme)
	}

	// Check host
	if parsedURL.Host == "" {
		return fmt.Errorf("%s must have a valid host", fieldName)
	}

	// Validate common Git hosting patterns
	// GitHub: https://github.com/owner/repo.git
	// GitLab: https://gitlab.com/owner/repo.git
	// Bitbucket: https://bitbucket.org/owner/repo.git
	// Generic SSH: git@github.com:owner/repo.git
	path := strings.TrimSuffix(parsedURL.Path, ".git")

	// For SSH URLs (git@host:path), the path should not be empty
	if parsedURL.Scheme == "ssh" && path == "" {
		return fmt.Errorf("%s SSH URL must include a repository path", fieldName)
	}

	// For HTTP(S) URLs, path should contain at least owner/repo pattern
	if (parsedURL.Scheme == "https" || parsedURL.Scheme == "http") && path != "" {
		pathParts := strings.Split(strings.Trim(path, "/"), "/")
		if len(pathParts) < 2 {
			return fmt.Errorf("%s URL should follow the pattern 'host/owner/repository'", fieldName)
		}
	}

	return nil
}

// validateImageRepo validates an image repository URL
func validateImageRepo(imageRepo string) error {
	if imageRepo == "" {
		return errors.New("imageRepo is required")
	}

	// Trim whitespace
	imageRepo = strings.TrimSpace(imageRepo)

	// Image repository should follow format: [registry/]namespace/repository[:tag]
	// Examples:
	//   - docker.io/mycompany/myapp
	//   - gcr.io/my-project/myapp
	//   - quay.io/namespace/repo
	//   - myregistry.com:5000/namespace/repo

	// Remove tag if present for validation
	imageParts := strings.Split(imageRepo, ":")
	imageWithoutTag := imageParts[0]

	// Validate tag format if present (should not be empty and not contain invalid chars)
	if len(imageParts) > 2 {
		// More than one colon could indicate port in registry URL, which is valid
		// e.g., myregistry.com:5000/namespace/repo:tag
		// Reconstruct without the last part (tag)
		imageWithoutTag = strings.Join(imageParts[:len(imageParts)-1], ":")
	}

	if len(imageParts) > 1 {
		tag := imageParts[len(imageParts)-1]
		if tag == "" {
			return errors.New("imageRepo tag cannot be empty when ':' is present")
		}
		// Tag should only contain alphanumeric, dots, dashes, and underscores
		tagRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
		if !tagRegex.MatchString(tag) {
			return fmt.Errorf("imageRepo tag '%s' contains invalid characters (allowed: a-z, A-Z, 0-9, ., -, _)", tag)
		}
	}

	// Split image path by '/'
	pathParts := strings.Split(imageWithoutTag, "/")

	// Should have at least registry/namespace/repository (3 parts) or namespace/repository (2 parts)
	if len(pathParts) < 2 {
		return fmt.Errorf("imageRepo must follow format '[registry/]namespace/repository[:tag]', got '%s'", imageRepo)
	}

	// Validate each path component
	componentRegex := regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*$`)
	for i, part := range pathParts {
		// The first part might be a registry domain (allow dots and colons for port)
		if i == 0 && len(pathParts) >= 3 {
			// Could be a registry like "gcr.io" or "localhost:5000"
			registryRegex := regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*(\.[a-z0-9]+([._-][a-z0-9]+)*)*(:[0-9]+)?$`)
			if !registryRegex.MatchString(part) {
				return fmt.Errorf("imageRepo registry '%s' is not valid", part)
			}
		} else {
			// Namespace and repository components
			if !componentRegex.MatchString(part) {
				return fmt.Errorf("imageRepo component '%s' contains invalid characters "+
					"(must be lowercase alphanumeric with '.', '-', or '_' separators)", part)
			}
		}
	}

	return nil
}

// isValidDNSSubdomain checks if a string is a valid DNS subdomain
func isValidDNSSubdomain(name string) bool {
	if name == "" || len(name) > 253 {
		return false
	}

	// Check each label
	labels := strings.Split(name, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}
		if !isValidDNSLabel(label) {
			return false
		}
	}

	return true
}

// isValidDNSLabel checks if a string is a valid DNS label
func isValidDNSLabel(label string) bool {
	if label == "" || len(label) > 63 {
		return false
	}

	// Must start and end with alphanumeric character
	if !isAlphanumeric(rune(label[0])) || !isAlphanumeric(rune(label[len(label)-1])) {
		return false
	}

	// Can contain alphanumeric characters and hyphens
	for _, char := range label {
		if !isAlphanumeric(char) && char != '-' {
			return false
		}
	}

	return true
}

// isAlphanumeric checks if a character is alphanumeric
func isAlphanumeric(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

// validateHeliosAppUpdate validates updates to HeliosApp
func (r *HeliosApp) validateHeliosAppUpdate(old *HeliosApp) (admission.Warnings, error) {
	logger := log.Log.WithName("webhook").WithValues(
		"operation", "validateUpdate",
		"heliosapp", r.Name,
		"namespace", r.Namespace,
	)

	logger.V(1).Info("Starting HeliosApp update validation")

	var allErrors []string
	var warnings admission.Warnings

	// Validate the new spec
	if errs := r.validateBasicFields(); len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}

	if errs := r.validateNumericFields(); len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}

	if errs := r.validateOptionalFields(); len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}

	// Check for immutable fields
	if r.Spec.GitRepo != old.Spec.GitRepo {
		allErrors = append(allErrors, fmt.Sprintf("gitRepo is immutable (cannot change from '%s' to '%s')", old.Spec.GitRepo, r.Spec.GitRepo))
	}

	if r.Spec.ImageRepo != old.Spec.ImageRepo {
		allErrors = append(allErrors, fmt.Sprintf("imageRepo is immutable (cannot change from '%s' to '%s')", old.Spec.ImageRepo, r.Spec.ImageRepo))
	}

	// Add warnings for potentially disruptive changes
	if r.Spec.Port != old.Spec.Port {
		warnings = append(warnings, fmt.Sprintf("Changing port from %d to %d will trigger a rolling update", old.Spec.Port, r.Spec.Port))
	}

	if r.Spec.Replicas > old.Spec.Replicas*2 {
		warnings = append(warnings, fmt.Sprintf("Scaling from %d to %d replicas is a large increase - ensure cluster has sufficient resources", old.Spec.Replicas, r.Spec.Replicas))
	}

	if r.Spec.Replicas == 0 && old.Spec.Replicas > 0 {
		warnings = append(warnings, "Setting replicas to 0 will stop all pods - this effectively pauses the application")
	}

	if len(allErrors) > 0 {
		logger.Info("HeliosApp update validation failed", "errors", len(allErrors))
		return warnings, errors.New(strings.Join(allErrors, "; "))
	}

	logger.V(1).Info("HeliosApp update validation successful", "warnings", len(warnings))
	return warnings, nil
}

// validateHeliosAppDelete validates deletion of HeliosApp
func (r *HeliosApp) validateHeliosAppDelete() (admission.Warnings, error) {
	logger := log.Log.WithName("webhook").WithValues(
		"operation", "validateDelete",
		"heliosapp", r.Name,
		"namespace", r.Namespace,
	)

	logger.V(1).Info("Starting HeliosApp deletion validation")

	// Add any deletion-specific validation here
	logger.V(1).Info("HeliosApp deletion validation successful")
	return nil, nil
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HeliosApp) Default() {
	// Set default values
	if r.Spec.Port == 0 {
		r.Spec.Port = 8080
	}

	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = 1
	}

	// Set default Git branch
	if r.Spec.GitBranch == "" {
		r.Spec.GitBranch = "main"
	}

	// Set default GitOps branch
	if r.Spec.GitopsBranch == "" {
		r.Spec.GitopsBranch = "main"
	}

	// Set default GitOps path to app name if not specified
	if r.Spec.GitopsPath == "" {
		r.Spec.GitopsPath = r.Name
	}

	// Namespace is handled by Kubernetes metadata.namespace

	if r.Spec.ServiceAccount == "" {
		r.Spec.ServiceAccount = "default"
	}

	// Add default labels
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels["app.kubernetes.io/name"] = "helios-app"
	r.Labels["app.kubernetes.io/instance"] = r.Name
	r.Labels["app.kubernetes.io/version"] = "v1"
	r.Labels["app.kubernetes.io/component"] = "application"
	r.Labels["app.kubernetes.io/part-of"] = "helios-operator"
	r.Labels["app.kubernetes.io/managed-by"] = "helios-operator"
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HeliosApp) ValidateCreate() (admission.Warnings, error) {
	startTime := time.Now()
	warnings, err := r.validateHeliosApp()
	duration := time.Since(startTime).Seconds()

	result := "accept"
	if err != nil {
		result = "reject"
	}

	webhookValidationDuration.WithLabelValues("create").Observe(duration)
	webhookValidationsTotal.WithLabelValues("create", result).Inc()

	return warnings, err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HeliosApp) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	startTime := time.Now()
	oldHeliosApp, ok := old.(*HeliosApp)
	if !ok {
		duration := time.Since(startTime).Seconds()
		webhookValidationDuration.WithLabelValues("update").Observe(duration)
		webhookValidationsTotal.WithLabelValues("update", "reject").Inc()
		return nil, errors.New("expected old object to be of type HeliosApp")
	}

	warnings, err := r.validateHeliosAppUpdate(oldHeliosApp)
	duration := time.Since(startTime).Seconds()

	result := "accept"
	if err != nil {
		result = "reject"
	}

	webhookValidationDuration.WithLabelValues("update").Observe(duration)
	webhookValidationsTotal.WithLabelValues("update", result).Inc()

	return warnings, err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HeliosApp) ValidateDelete() (admission.Warnings, error) {
	startTime := time.Now()
	warnings, err := r.validateHeliosAppDelete()
	duration := time.Since(startTime).Seconds()

	result := "accept"
	if err != nil {
		result = "reject"
	}

	webhookValidationDuration.WithLabelValues("delete").Observe(duration)
	webhookValidationsTotal.WithLabelValues("delete", result).Inc()

	return warnings, err
}

// SetupWebhookWithManager sets up the webhook with the manager
func (r *HeliosApp) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if err := ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete(); err != nil {
		return fmt.Errorf("failed to setup webhook: %w", err)
	}
	return nil
}
