package activities

import (
	"context"
)

type GetStalePlanThresholdRequest struct{}

// @temporal-gen-v2 activity
func (a *Activities) GetStalePlanThreshold(ctx context.Context, req GetStalePlanThresholdRequest) (string, error) {
	if a.cfg == nil || a.cfg.StalePlanThreshold == "" {
		return "", nil
	}
	return a.cfg.StalePlanThreshold, nil
}
