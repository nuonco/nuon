package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentID
func (a *Activities) GetComponent(ctx context.Context, req GetComponentRequest) (*app.Component, error) {
	comp, err := a.helpers.GetComponent(ctx, req.ComponentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	return comp, nil
}
