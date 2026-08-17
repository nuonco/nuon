package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type HasOrgFeatureRequest struct {
	OrgID   string `validate:"required"`
	Feature string `validate:"required"`
}

// HasOrgFeature resolves the org explicitly rather than from the context like
// HasFeature does. Context propagation runs through optionalPropagator, which
// swallows a failed inject, so a workflow missing org context reaches the
// activity with no org ID and HasFeature errors — callers that hold the install
// should pass its OrgID and get a deterministic answer.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) HasOrgFeature(ctx context.Context, req HasOrgFeatureRequest) (bool, error) {
	return a.features.OrgHasFeature(ctx, req.OrgID, app.OrgFeature(req.Feature))
}
