package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetActionWorkflows struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetActionWorkflows(ctx context.Context, req *GetActionWorkflows) ([]*app.ActionWorkflow, error) {
	return a.getActionWorkflows(ctx, req.InstallID)
}

func (a *Activities) getActionWorkflows(ctx context.Context, installID string) ([]*app.ActionWorkflow, error) {
	install, err := a.getInstall(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	var actionWorkflows []*app.ActionWorkflow

	res := a.db.WithContext(ctx).
		Where(app.ActionWorkflow{
			AppID: install.AppID,
		}).
		Find(&actionWorkflows)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get action workflows")
	}

	return actionWorkflows, nil
}
