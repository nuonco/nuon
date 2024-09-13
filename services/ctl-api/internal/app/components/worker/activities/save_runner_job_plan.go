package activities

import (
	"context"
	"fmt"
)

type SaveRunnerJobPlanRequest struct {
	JobID    string
	PlanJSON string
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) SaveRunnerJobPlan(ctx context.Context, req *SaveRunnerJobPlanRequest) error {
	ctx, err := a.runnersHelpers.ContextFromJob(ctx, req.JobID)
	if err != nil {
		return fmt.Errorf("unable to create context: %w", err)
	}

	if err := a.runnersHelpers.WriteJobPlan(ctx, req.JobID, []byte(req.PlanJSON)); err != nil {
		return fmt.Errorf("unable to write job plan: %w", err)
	}

	return nil
}
