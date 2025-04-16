package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallActionWorkflowsByTriggerTypeRequest struct {
	ComponentID string
	InstallID   string                        `validate:"required"`
	TriggerType app.ActionWorkflowTriggerType `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentID
func (a *Activities) GetInstallActionWorkflowsByTriggerType(ctx context.Context, req GetInstallActionWorkflowsByTriggerTypeRequest) ([]*app.InstallActionWorkflow, error) {
	workflows, err := a.getActionWorkflows(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	wkflows := make([]*app.InstallActionWorkflow, 0)
	for _, workflow := range workflows {
		cfg, err := a.getActionWorkflowLatestConfig(ctx, workflow.ActionWorkflowID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get action workflow config")
		}

		if req.ComponentID == "" {
			if cfg.HasTrigger(req.TriggerType) {
				wkflows = append(wkflows, workflow)
			}
		} else {
			if cfg.HasComponentTrigger(req.TriggerType, req.ComponentID) {
				wkflows = append(wkflows, workflow)
			}
		}

	}

	return wkflows, nil
}
