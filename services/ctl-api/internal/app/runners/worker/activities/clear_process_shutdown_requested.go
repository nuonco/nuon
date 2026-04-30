package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type ClearProcessShutdownRequestedRequest struct {
	ProcessID string `validate:"required"`
}

// ClearProcessShutdownRequested removes the shutdown_requested key from a
// runner process's CompositeStatus metadata. Called after a shutdown record
// has been created to prevent duplicate shutdowns.
//
// @temporal-gen-v2 activity
func (a *Activities) ClearProcessShutdownRequested(ctx context.Context, req ClearProcessShutdownRequestedRequest) error {
	// Set shutdown_requested to null to effectively remove it.
	if err := generics.MergeJSONBMetadata(
		a.db.WithContext(ctx),
		&app.RunnerProcess{},
		req.ProcessID,
		"composite_status",
		map[string]any{"shutdown_requested": nil},
	); err != nil {
		return fmt.Errorf("unable to clear shutdown_requested: %w", err)
	}

	return nil
}
