package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallActionWorkflowRunRequest struct {
	RunID string `validate:"required"`
}

// @temporal-gen activity
// @by-id RunID
func (a *Activities) GetInstallActionWorkflowRun(ctx context.Context, req GetInstallActionWorkflowRunRequest) (*app.InstallActionWorkflowRun, error) {
	return a.getInstallActionWorkflowRun(ctx, req.RunID)
}

func (a *Activities) getInstallActionWorkflowRun(ctx context.Context, runID string) (*app.InstallActionWorkflowRun, error) {
	run := app.InstallActionWorkflowRun{}
	res := a.db.WithContext(ctx).
		Preload("ActionWorkflowConfig").
		Preload("ActionWorkflowConfig.Triggers").
		Preload("ActionWorkflowConfig.Steps").
		Preload("LogStream").
		Preload("Install").
		Preload("Install.RunnerGroup").
		Preload("Install.RunnerGroup.Runners").
		First(&run, "id = ?", runID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &run, nil
}
