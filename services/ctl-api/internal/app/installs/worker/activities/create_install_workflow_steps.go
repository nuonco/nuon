package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallWorkflowStepRequest struct {
	InstallWorkflowID string              `json:"install_workflow_id"`
	InstallID         string              `json:"install_id"`
	Status            app.CompositeStatus `json:"status"`
	Name              string              `json:"name"`
	Signal            app.Signal          `json:"signal"`
	Idx               int                 `json:"idx"`
}

// @temporal-gen activity
func (a *Activities) CreateInstallWorkflowStep(ctx context.Context, req CreateInstallWorkflowStepRequest) error {
	step := &app.InstallWorkflowStep{
		InstallWorkflowID: req.InstallWorkflowID,
		InstallID:         req.InstallID,
		Status:            req.Status,
		Name:              req.Name,
		Signal:            req.Signal,
		Idx:               req.Idx,
	}

	if res := a.db.WithContext(ctx).Create(step); res.Error != nil {
		return errors.Wrap(res.Error, "unable to create steps")
	}

	return nil
}
