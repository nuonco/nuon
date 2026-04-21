package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowGetFlowStepGroups(ctx context.Context, workflowID string) ([]app.WorkflowStepGroup, error) {
	var groups []app.WorkflowStepGroup
	res := a.db.WithContext(ctx).
		Preload("QueueSignal").
		Where(app.WorkflowStepGroup{WorkflowID: workflowID}).
		Order("group_idx ASC").
		Find(&groups)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get step groups")
	}
	return groups, nil
}
