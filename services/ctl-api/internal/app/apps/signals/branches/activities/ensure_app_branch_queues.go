package activities

import (
	"context"

	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type EnsureAppBranchQueuesRequest struct {
	AppBranchID     string `json:"app_branch_id" temporaljson:"app_branch_id" validate:"required"`
	SkipRestartHint bool   `json:"skip_restart_hint" temporaljson:"skip_restart_hint"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) EnsureAppBranchQueues(ctx context.Context, req *EnsureAppBranchQueuesRequest) error {
	if err := a.v.Struct(req); err != nil {
		return err
	}
	return a.helpers.EnsureAppBranchQueues(ctx, req.AppBranchID, appshelpers.EnsureAppBranchQueuesOptions{
		SkipRestartHint: req.SkipRestartHint,
	})
}
