package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/installcomponenthealth"
)

type StopInstallComponentHealthRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

// stopByInstallID signals InstallComponentHealthWorkflow to stop, keyed
// directly by install ID. Workflow code cannot call the Temporal client
// directly, so the Deprovision workflow
// (installs/workflows/v2/deprovision.go) goes through this activity instead
// of the runner-keyed stopInstallComponentHealth that ReconcileUseRunner (an
// activity, not workflow code) calls directly.
//
// Not finding a running workflow is not an error: a runner that was already
// using a Runner, or an install that never went runnerless, has nothing to
// stop.
//
// @temporal-gen-v2 activity
// @as-wrapper
// @wrapper-prefix InstallComponentHealth
func (a *Activities) stopByInstallID(ctx context.Context, req *StopInstallComponentHealthRequest) error {
	err := a.tClient.SignalWorkflowInNamespace(ctx, installComponentHealthNamespace,
		installcomponenthealth.WorkflowID(req.InstallID), "", installcomponenthealth.StopSignalName, nil)
	if err != nil && !isWorkflowNotFound(err) {
		return fmt.Errorf("unable to signal install component health workflow to stop: %w", err)
	}
	return nil
}
