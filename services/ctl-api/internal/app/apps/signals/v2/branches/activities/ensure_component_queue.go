package activities

import (
	"context"
	"fmt"

	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field componentID
func (a *Activities) ensureComponentQueue(ctx context.Context, componentID string) (*componenthelpers.ComponentQueueIDs, error) {
	if componentID == "" {
		return nil, fmt.Errorf("componentID is required")
	}
	return a.componentHelpers.EnsureComponentQueues(ctx, componentID)
}
