package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field componentID
func (a *Activities) getComponentByID(ctx context.Context, componentID string) (*app.Component, error) {
	if componentID == "" {
		return nil, fmt.Errorf("componentID is required")
	}
	return a.componentHelpers.GetComponentByID(ctx, componentID)
}
