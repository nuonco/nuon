package activities

import (
	"context"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetActionWorkflowGitSourceRequest struct {
	StepID string `validate:"required"`
}

// @temporal-gen activity
// @by-id StepID
func (a *Activities) GetActionWorkflowStepGitSource(ctx context.Context, req GetActionWorkflowGitSourceRequest) (*plantypes.GitSource, error) {
	cfg, err := a.getActionWorkflowStepConfig(ctx, req.StepID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get workflow step config")
	}

	switch cfg.VCSConnectionType {
	case app.VCSConnectionTypeConnectedRepo:
		return a.vcsHelpers.GetGitSource(ctx, cfg.ConnectedGithubVCSConfig)
	case app.VCSConnectionTypePublicRepo:
		return a.vcsHelpers.GetPubliGitSource(ctx, cfg.PublicGitVCSConfig)
	default:
	}

	return nil, nil
}

func (a *Activities) getActionWorkflowStepConfig(ctx context.Context, stepID string) (*app.ActionWorkflowStepConfig, error) {
	var stepCfg app.ActionWorkflowStepConfig

	if res := a.db.WithContext(ctx).
		Preload("PublicGitVCSConfig").
		Preload("ConnectedGithubVCSConfig").
		Preload("ConnectedGithubVCSConfig.VCSConnection").
		First(&stepCfg, "id = ?", stepID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get step")
	}

	return &stepCfg, nil
}
