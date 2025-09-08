package activities

import (
	"context"
)

type GetComponentDependentsRequest struct {
	AppConfigID string `validate:"required"`
	ComponentID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetComponentDependents(ctx context.Context, req *GetComponentDependentsRequest) ([]string, error) {
	return a.appsHelpers.GetComponentDependents(ctx, req.AppConfigID, req.ComponentID)
}
