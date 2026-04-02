package activities

import (
	"context"
	"fmt"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveRunnerJobPlanRequest struct {
	JobID         string                  `validate:"required"`
	CompositePlan plantypes.CompositePlan `validate:"required"`
	// Deprecated: but kept for backward compatibility
	PlanJSON string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SaveRunnerJobPlan(ctx context.Context, req *SaveRunnerJobPlanRequest) error {
	if err := a.helpers.WriteJobPlan(ctx, req.JobID, []byte(req.PlanJSON), req.CompositePlan, app.RunnerJobPermissionInfo{}); err != nil {
		return fmt.Errorf("unable to write job plan: %w", err)
	}

	return nil
}
