package activities

import (
	"context"

	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) RegenerateInstallState(ctx context.Context, req *pkgstate.ExecuteRegenerationRequest) (*pkgstate.ExecuteRegenerationResponse, error) {
	return a.helpers.RegenerateInstallState(ctx, req, false)
}
