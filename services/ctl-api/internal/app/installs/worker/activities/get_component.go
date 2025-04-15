package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentID
func (a *Activities) GetComponent(ctx context.Context, req GetComponentRequest) (*app.Component, error) {
	return a.componentsHelpers.GetComponent(ctx, req.ComponentID)
}
