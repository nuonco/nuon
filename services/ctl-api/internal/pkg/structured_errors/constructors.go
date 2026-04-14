package structured_errors

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// NewCompositeError creates a CompositeError from a standard context.
func NewCompositeError(ctx context.Context, ownerType OwnerType, severity Severity, summary string) CompositeError {
	return CompositeError{
		CreatedByID:  keys.CreatedByIDFromContext(ctx),
		CreatedAtTS:  time.Now().Unix(),
		OwnerType: ownerType,
		Severity:     severity,
		Summary:      summary,
		Metadata:     make(map[string]any),
	}
}

// NewCompositeTemporalError creates a CompositeError from a Temporal workflow context.
func NewCompositeTemporalError(ctx workflow.Context, ownerType OwnerType, severity Severity, summary string) CompositeError {
	var createdByID string
	if val := ctx.Value(keys.AccountIDCtxKey); val != nil {
		if s, ok := val.(string); ok {
			createdByID = s
		}
	}

	return CompositeError{
		CreatedByID:  createdByID,
		CreatedAtTS:  time.Now().Unix(),
		OwnerType: ownerType,
		Severity:     severity,
		Summary:      summary,
		Metadata:     make(map[string]any),
	}
}

// FromGoError wraps a Go error into a CompositeError as a fallback.
func FromGoError(err error, ownerType OwnerType) CompositeError {
	return CompositeError{
		CreatedAtTS:  time.Now().Unix(),
		OwnerType: ownerType,
		Severity:     SeverityCritical,
		Summary:      err.Error(),
		Metadata:     make(map[string]any),
	}
}
