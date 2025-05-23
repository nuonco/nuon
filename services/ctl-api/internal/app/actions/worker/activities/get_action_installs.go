package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetActionWorkflowInstallsRequest struct {
	ActionWorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ActionWorkflowID
func (a *Activities) GetActionWorkflowInstalls(ctx context.Context, req *GetActionWorkflowInstallsRequest) ([]string, error) {
	return a.getActionWorkflowInstalls(ctx, req.ActionWorkflowID)
}

func (a *Activities) getActionWorkflowInstalls(ctx context.Context, actionWorkflowID string) ([]string, error) {
	installs := []app.Install{}

	res := a.db.WithContext(ctx).
		Joins(fmt.Sprintf("JOIN apps ON apps.id=%s", views.TableOrViewName(a.db, &app.Install{}, ".app_id"))).
		Joins("JOIN action_workflows ON action_workflows.app_id=apps.id").
		Where("action_workflows.id = ?", actionWorkflowID).
		Find(&installs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get installs")
	}

	installIDs := make([]string, 0)
	for _, install := range installs {
		installIDs = append(installIDs, install.ID)
	}

	return installIDs, nil
}
