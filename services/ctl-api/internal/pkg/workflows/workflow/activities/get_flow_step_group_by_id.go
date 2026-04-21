package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowGetFlowStepGroupByID(ctx context.Context, stepGroupID string) (*app.WorkflowStepGroup, error) {
	var group app.WorkflowStepGroup
	res := a.db.WithContext(ctx).
		Where(app.WorkflowStepGroup{ID: stepGroupID}).
		First(&group)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get step group")
	}
	return &group, nil
}
