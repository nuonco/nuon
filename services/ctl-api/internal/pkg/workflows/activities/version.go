package activities

import (
	"context"
)

type GetVersionRequest struct{}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 10s
func (a *Activities) GetVersion(ctx context.Context, _ GetVersionRequest) (string, error) {
	return a.cfg.Version, nil
}
