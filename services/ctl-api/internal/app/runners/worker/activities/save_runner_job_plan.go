package activities

import (
	"context"
	"fmt"
)

type SaveRunnerJobPlanRequest struct {
	JobID    string `validate:"required"`
	PlanJSON string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) SaveRunnerJobPlan(ctx context.Context, req *SaveRunnerJobPlanRequest) error {
	if err := a.helpers.WriteJobPlan(ctx, req.JobID, []byte(req.PlanJSON)); err != nil {
		return fmt.Errorf("unable to write job plan: %w", err)
	}

	return nil
}
