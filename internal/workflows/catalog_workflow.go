package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/bancofalabella/mortgage-loan-catalogs/internal/model"
)

// GetCatalogWorkflow orchestrates the retrieval of a catalog
// by executing the legacy activity (ACL) and handling retries
// based on the configured RetryPolicy.
func GetCatalogWorkflow(ctx workflow.Context, req model.GetCatalogRequest) (interface{}, error) {
	// 1. Configure Activity Options and RetryPolicy
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 30,
		MaximumAttempts:    3,
		NonRetryableErrorTypes: []string{
			"InvalidCatalogError",  // E.g., HTTP 400 Bad Request
			"AuthenticationError",  // E.g., HTTP 401 Unauthorized
			"CatalogNotFoundError", // E.g., HTTP 404 Not Found
		},
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	// 2. Execute the Activity (Anti-Corruption Layer)
	var result interface{}
	err := workflow.ExecuteActivity(ctx, "FetchLegacyCatalogActivity", req).Get(ctx, &result)
	if err != nil {
		// Log error, and potentially return a mapped business error
		return nil, err
	}

	return result, nil
}
