package activities

import (
	"context"
	"fmt"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type SaveRunnerJobPlanRequest struct {
	JobID         string                  `validate:"required"`
	CompositePlan plantypes.CompositePlan `validate:"required"`
	// Deprecated: but kept for backward compatibility
	PlanJSON string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) SaveRunnerJobPlan(ctx context.Context, req *SaveRunnerJobPlanRequest) error {
	if err := a.runnersHelpers.WriteJobPlan(ctx, req.JobID, []byte(req.PlanJSON), req.CompositePlan); err != nil {
		return fmt.Errorf("unable to write job plan: %w", err)
	}

	return nil
}
