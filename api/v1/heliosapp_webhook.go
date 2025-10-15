/*
Copyright 2025.
*/

package v1

import (
	"fmt"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var heliosapplog = logf.Log.WithName("heliosapp-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *HeliosApp) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-platform-helios-io-v1-heliosapp,mutating=true,failurePolicy=fail,sideEffects=None,groups=platform.helios.io,resources=heliosapps,verbs=create;update,versions=v1,name=mheliosapp.kb.io,admissionReviewVersions=v1

// Default implements defaulting logic for HeliosApp
// This method will be called by webhook framework
func (r *HeliosApp) Default() {
	heliosapplog.Info("default", "name", r.Name)

	// Set default gitBranch if not provided
	if r.Spec.GitBranch == "" {
		r.Spec.GitBranch = "main"
	}

	// Set default gitopsBranch if not provided
	if r.Spec.GitopsBranch == "" {
		r.Spec.GitopsBranch = "main"
	}

	// Set default gitopsPath to app name if not provided
	if r.Spec.GitopsPath == "" {
		r.Spec.GitopsPath = r.Name
	}

	// Set default replicas if not provided
	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = 1
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-platform-helios-io-v1-heliosapp,mutating=false,failurePolicy=fail,sideEffects=None,groups=platform.helios.io,resources=heliosapps,verbs=create;update,versions=v1,name=vheliosapp.kb.io,admissionReviewVersions=v1

// ValidateCreate implements validation logic for HeliosApp creation
// This method will be called by webhook framework
func (r *HeliosApp) ValidateCreate() (admission.Warnings, error) {
	heliosapplog.Info("validate create", "name", r.Name)

	return r.validateHeliosApp()
}

// ValidateUpdate implements validation logic for HeliosApp updates
// This method will be called by webhook framework
func (r *HeliosApp) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	heliosapplog.Info("validate update", "name", r.Name)

	return r.validateHeliosApp()
}

// ValidateDelete implements validation logic for HeliosApp deletion
// This method will be called by webhook framework
func (r *HeliosApp) ValidateDelete() (admission.Warnings, error) {
	heliosapplog.Info("validate delete", "name", r.Name)

	// No validation needed for deletion
	return nil, nil
}

// validateHeliosApp validates the HeliosApp spec
func (r *HeliosApp) validateHeliosApp() (admission.Warnings, error) {
	var allErrors []string
	var warnings admission.Warnings

	// Validate GitRepo URL
	if err := validateGitURL(r.Spec.GitRepo, "gitRepo"); err != nil {
		allErrors = append(allErrors, err.Error())
	}

	// Validate GitopsRepo URL
	if err := validateGitURL(r.Spec.GitopsRepo, "gitopsRepo"); err != nil {
		allErrors = append(allErrors, err.Error())
	}

	// Validate ImageRepo format
	if err := validateImageRepo(r.Spec.ImageRepo); err != nil {
		allErrors = append(allErrors, err.Error())
	}

	// Validate Port range
	if r.Spec.Port < 1 || r.Spec.Port > 65535 {
		allErrors = append(allErrors, fmt.Sprintf("port must be between 1 and 65535, got %d", r.Spec.Port))
	}

	// Validate Replicas
	if r.Spec.Replicas < 0 {
		allErrors = append(allErrors, fmt.Sprintf("replicas cannot be negative, got %d", r.Spec.Replicas))
	}

	// Warn if replicas is very high
	if r.Spec.Replicas > 10 {
		warnings = append(warnings, fmt.Sprintf("high replica count (%d) - ensure your cluster has sufficient resources", r.Spec.Replicas))
	}

	// Validate ServiceAccount is not empty
	if strings.TrimSpace(r.Spec.ServiceAccount) == "" {
		allErrors = append(allErrors, "serviceAccount cannot be empty")
	}

	// Validate WebhookSecret is not empty
	if strings.TrimSpace(r.Spec.WebhookSecret) == "" {
		allErrors = append(allErrors, "webhookSecret cannot be empty")
	}

	// Validate GitopsPath doesn't start with slash or contain ..
	if r.Spec.GitopsPath != "" {
		if strings.HasPrefix(r.Spec.GitopsPath, "/") {
			allErrors = append(allErrors, "gitopsPath should not start with '/'")
		}
		if strings.Contains(r.Spec.GitopsPath, "..") {
			allErrors = append(allErrors, "gitopsPath cannot contain '..'")
		}
	}

	// Validate branch names don't contain invalid characters
	if r.Spec.GitBranch != "" {
		if strings.ContainsAny(r.Spec.GitBranch, " \t\n") {
			allErrors = append(allErrors, "gitBranch cannot contain whitespace characters")
		}
	}

	if r.Spec.GitopsBranch != "" {
		if strings.ContainsAny(r.Spec.GitopsBranch, " \t\n") {
			allErrors = append(allErrors, "gitopsBranch cannot contain whitespace characters")
		}
	}

	if len(allErrors) > 0 {
		return warnings, fmt.Errorf("validation failed: %s", strings.Join(allErrors, "; "))
	}

	return warnings, nil
}

// validateGitURL validates that the URL is a valid Git repository URL
func validateGitURL(gitURL, fieldName string) error {
	if gitURL == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	// Parse URL
	parsedURL, err := url.Parse(gitURL)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL: %v", fieldName, err)
	}

	// Check scheme
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" && parsedURL.Scheme != "ssh" && parsedURL.Scheme != "git" {
		return fmt.Errorf("%s must use https://, http://, ssh://, or git:// scheme, got %s", fieldName, parsedURL.Scheme)
	}

	// Check host
	if parsedURL.Host == "" {
		return fmt.Errorf("%s must have a valid host", fieldName)
	}

	// Common patterns for Git URLs
	validPatterns := []string{
		".git",       // e.g., https://github.com/user/repo.git
		"github.com", // e.g., https://github.com/user/repo
		"gitlab.com",
		"bitbucket.org",
	}

	isValid := false
	for _, pattern := range validPatterns {
		if strings.Contains(strings.ToLower(gitURL), pattern) {
			isValid = true
			break
		}
	}

	// Allow git@ SSH format
	if strings.HasPrefix(gitURL, "git@") {
		isValid = true
	}

	if !isValid {
		// Warn but don't fail - user might be using private Git server
		// This is just a warning, not a hard error
		// You can make this stricter if needed
	}

	return nil
}

// validateImageRepo validates that the image repository format is valid
func validateImageRepo(imageRepo string) error {
	if imageRepo == "" {
		return fmt.Errorf("imageRepo cannot be empty")
	}

	// Basic Docker image format validation
	// Format: [registry/][namespace/]repository[:tag]
	// Examples:
	// - nginx
	// - nginx:latest
	// - docker.io/library/nginx:latest
	// - gcr.io/project/image:v1.0.0

	parts := strings.Split(imageRepo, ":")
	if len(parts) > 2 {
		return fmt.Errorf("imageRepo has invalid format (too many colons): %s", imageRepo)
	}

	imageName := parts[0]
	if imageName == "" {
		return fmt.Errorf("imageRepo image name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(imageName, " \t\n") {
		return fmt.Errorf("imageRepo cannot contain whitespace characters")
	}

	// Tag validation (if provided)
	if len(parts) == 2 {
		tag := parts[1]
		if tag == "" {
			return fmt.Errorf("imageRepo tag cannot be empty when colon is present")
		}
		if strings.ContainsAny(tag, " \t\n") {
			return fmt.Errorf("imageRepo tag cannot contain whitespace characters")
		}
	}

	return nil
}
