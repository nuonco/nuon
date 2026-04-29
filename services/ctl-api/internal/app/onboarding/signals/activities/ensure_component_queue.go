package activities

import (
	"context"
	"fmt"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field componentID
func (a *Activities) ensureComponentQueue(ctx context.Context, componentID string) error {
	if componentID == "" {
		return fmt.Errorf("componentID is required")
	}
	_, err := a.componentHelpers.EnsureComponentQueues(ctx, componentID)
	return err
}
