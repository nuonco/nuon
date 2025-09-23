package activities

import (
	"context"
	"fmt"
)

type GetComponentDependentsRequest struct {
	ComponentID string `validate:"required"`
	AppConfigID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetComponentDependents(ctx context.Context, req GetComponentDependentsRequest) ([]string, error) {
	deps, err := a.appsHelpers.GetComponentDependents(ctx, req.AppConfigID, req.ComponentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component dependents: %w", err)
	}

	return deps, nil
}
