/*
Package common provides shared constants, utilities, and helper functions
used across the Helios operator codebase.

This package contains:

# Constants
- Condition types for HeliosApp status tracking
- Standard Kubernetes labels for resource identification
- Finalizer names for resource lifecycle management

# Error Handling
- Custom error types with error codes for better error classification
- Transient error detection for retry logic
- Error wrapping utilities for better error context

# Resource Helpers
- ResourceHelper for common Kubernetes resource operations
- CreateOrUpdate, DeleteIfExists, GetResource, and ListResources utilities
- Consistent error handling and logging across resource operations

# Usage Examples

## Using Constants
```go
import "github.com/hoangphuc841/helios-operator/internal/common"

// Set condition on HeliosApp status

	condition := metav1.Condition{
	    Type:   common.ConditionReady,
	    Status: metav1.ConditionTrue,
	    Reason: "ApplicationReady",
	}

// Add labels to resources

	labels := map[string]string{
	    common.LabelAppName:     "my-app",
	    common.LabelManagedBy:   "helios-operator",
	    common.LabelComponent:   "application",
	}

```

## Using Error Handling
```go
import "github.com/hoangphuc841/helios-operator/internal/common"

// Create a custom error
err := common.NewHeliosError(

	common.ErrorCodeInvalidInput,
	"invalid Git repository URL",
	originalErr,

)

// Check if error is transient (for retry logic)

	if common.IsTransientError(err) {
	    // Retry the operation
	    return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}

```

## Using Resource Helpers
```go
import "github.com/hoangphuc841/helios-operator/internal/common"

// Create a resource helper

	helper := &common.ResourceHelper{
	    Client: mgr.GetClient(),
	    Logger: logger,
	}

// Create or update a resource
err := helper.CreateOrUpdate(ctx, configMap)

	if err != nil {
	    return fmt.Errorf("failed to create configmap: %w", err)
	}

// Delete a resource if it exists
err = helper.DeleteIfExists(ctx, oldConfigMap)

	if err != nil {
	    return fmt.Errorf("failed to delete old configmap: %w", err)
	}

```

This package is designed to provide a consistent foundation for the Helios operator,
ensuring uniform error handling, resource management, and status reporting across
all components.
*/
package common
