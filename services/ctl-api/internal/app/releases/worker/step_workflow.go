package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func provisionStepWorkflowID(releaseID, stepID string) string {
	return fmt.Sprintf("release-%s-step-%s", releaseID, stepID)
}

type ProvisionReleaseStepRequest struct {
	ReleaseID     string `json:"release_id"`
	ReleaseStepID string `json:"release_step_id"`
}

type ProvisionReleaseStepResponse struct {
	Status string `json:"status"`
}

// release steps are their own workflow, as they encompass an unbounded amount of provisioning and orchestration.
func (w *Workflows) ProvisionReleaseStep(ctx workflow.Context, req ProvisionReleaseStepRequest) error {
	var step app.ComponentReleaseStep
	if err := w.defaultExecGetActivity(ctx, w.acts.GetReleaseStep, activities.GetReleaseStepRequest{
		ReleaseStepID: req.ReleaseStepID,
	}, &step); err != nil {
		return fmt.Errorf("unable to get release step: %w", err)
	}

	// delay if needed
	if step.Delay != nil {
		delayDuration, err := time.ParseDuration(*step.Delay)
		if err != nil {
			return fmt.Errorf("unable to parse delay: %w", err)
		}

		if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateReleaseStepStatus, activities.UpdateReleaseStepStatusRequest{
			ReleaseStepID:     req.ReleaseStepID,
			Status:            "waiting_for_delay",
			StatusDescription: "waiting %s before starting step",
		}); err != nil {
			return fmt.Errorf("unable to update release step status: %w", err)
		}

		if err := workflow.Sleep(ctx, delayDuration); err != nil {
			return fmt.Errorf("unable to sleep: %w", err)
		}
	}

	// create each install-deploy + signal
	for _, installID := range step.RequestedInstallIDs {
		installReq := activities.CreateInstallDeployRequest{
			InstallID:     installID,
			ReleaseStepID: req.ReleaseStepID,
		}
		if err := w.defaultExecErrorActivity(ctx, w.acts.CreateInstallDeploy, installReq); err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}
	}

	// TODO(jm): figure out how to properly poll here, as there are quite a few different options
	//
	// We could poll all of the statuses using the database, poll the activities, or have a query to ask the install
	// workflows if they are ready.
	workflow.Sleep(ctx, time.Second*15)
	return nil
}
