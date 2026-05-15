package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowGetFlowStepGroups(ctx context.Context, workflowID string) ([]app.WorkflowStepGroup, error) {
	var allGroups []app.WorkflowStepGroup
	res := a.db.WithContext(ctx).
		Preload("QueueSignal").
		Where(app.WorkflowStepGroup{WorkflowID: workflowID}).
		Order("group_idx ASC").
		Find(&allGroups)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get step groups")
	}

	// Filter out discarded groups so retried groups don't shadow their
	// replacements when the flow re-dispatches after a retry-group clone.
	groups := make([]app.WorkflowStepGroup, 0, len(allGroups))
	for _, g := range allGroups {
		if g.Status.Status != app.StatusDiscarded {
			groups = append(groups, g)
		}
	}
	return groups, nil
}
