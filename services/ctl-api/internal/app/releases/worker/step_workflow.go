package worker

import (
	"fmt"
	"time"

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

func (w *Workflows) pollReleaseStepInstallDeploys(ctx workflow.Context, releaseStepID string) error {
	for {
		workflow.Sleep(ctx, defaultPollTimeout)

		step, err := activities.AwaitGetReleaseStepByReleaseStepID(ctx, releaseStepID)
		if err != nil {
			return fmt.Errorf("unable to get release step: %w", err)
		}

		isPending := false
		for _, installDeploy := range step.InstallDeploys {
			if installDeploy.Status == "active" {
				continue
			}
			if installDeploy.Status == "error" {
				return fmt.Errorf("install deploy failed %s", installDeploy.InstallComponent.InstallID)
			}
			isPending = true
		}

		if !isPending {
			break
		}
	}

	return nil
}

// release steps are their own workflow, as they encompass an unbounded amount of provisioning and orchestration.
func (w *Workflows) ProvisionReleaseStep(ctx workflow.Context, req ProvisionReleaseStepRequest) error {
	step, err := activities.AwaitGetReleaseStepByReleaseStepID(ctx, req.ReleaseStepID)
	if err != nil {
		return fmt.Errorf("unable to get release step: %w", err)
	}

	// delay if needed
	if step.Delay != nil {
		delayDuration, err := time.ParseDuration(*step.Delay)
		if err != nil {
			return fmt.Errorf("unable to parse delay: %w", err)
		}

		if err := activities.AwaitUpdateReleaseStepStatus(ctx, activities.UpdateReleaseStepStatusRequest{
			ReleaseStepID:     req.ReleaseStepID,
			Status:            "waiting_for_delay",
			StatusDescription: fmt.Sprintf("waiting %s before starting step", delayDuration),
		}); err != nil {
			return fmt.Errorf("unable to update release step status: %w", err)
		}

		if err := workflow.Sleep(ctx, delayDuration); err != nil {
			return fmt.Errorf("unable to sleep: %w", err)
		}
	}

	if err := activities.AwaitUpdateReleaseStepStatus(ctx, activities.UpdateReleaseStepStatusRequest{
		ReleaseStepID:     req.ReleaseStepID,
		Status:            "creating_install_deploys",
		StatusDescription: fmt.Sprintf("creating deploys for %d installs", len(step.RequestedInstallIDs)),
	}); err != nil {
		return fmt.Errorf("unable to update release step status: %w", err)
	}

	// create each install-deploy + signal
	for _, installID := range step.RequestedInstallIDs {
		installReq := activities.CreateInstallDeployRequest{
			InstallID:     installID,
			ReleaseStepID: req.ReleaseStepID,
		}
		if err := activities.AwaitCreateInstallDeploy(ctx, installReq); err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}
	}

	if err := activities.AwaitUpdateReleaseStepStatus(ctx, activities.UpdateReleaseStepStatusRequest{
		ReleaseStepID:     req.ReleaseStepID,
		Status:            "polling_install_deploys",
		StatusDescription: fmt.Sprintf("polling deploys for %d installs", len(step.RequestedInstallIDs)),
	}); err != nil {
		return fmt.Errorf("unable to update release step status: %w", err)
	}

	if err := w.pollReleaseStepInstallDeploys(ctx, req.ReleaseStepID); err != nil {
		if updateErr := activities.AwaitUpdateReleaseStepStatus(ctx, activities.UpdateReleaseStepStatusRequest{
			ReleaseStepID:     req.ReleaseStepID,
			Status:            "error",
			StatusDescription: "error",
		}); updateErr != nil {
			return fmt.Errorf("unable to update release step status: %w", updateErr)
		}

		return fmt.Errorf("release step failed: %w", err)
	}

	if err := activities.AwaitUpdateReleaseStepStatus(ctx, activities.UpdateReleaseStepStatusRequest{
		ReleaseStepID:     req.ReleaseStepID,
		Status:            "active",
		StatusDescription: "release step finished and all install deploys are active",
	}); err != nil {
		return fmt.Errorf("unable to update release step status: %w", err)
	}

	return nil
}
