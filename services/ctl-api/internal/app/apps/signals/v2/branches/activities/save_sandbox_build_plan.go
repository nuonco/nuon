package activities

import (
	"context"
	"fmt"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveSandboxBuildPlanRequest struct {
	JobID         string                  `json:"job_id" validate:"required"`
	CompositePlan plantypes.CompositePlan `json:"composite_plan"`
	PlanJSON      string                  `json:"plan_json" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) SaveSandboxBuildPlan(ctx context.Context, req SaveSandboxBuildPlanRequest) error {
	if err := a.runnerHelpers.WriteJobPlan(ctx, req.JobID, []byte(req.PlanJSON), req.CompositePlan, app.RunnerJobPermissionInfo{}); err != nil {
		return fmt.Errorf("unable to write sandbox build plan: %w", err)
	}
	return nil
}
