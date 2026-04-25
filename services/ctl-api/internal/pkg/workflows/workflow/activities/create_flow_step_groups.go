package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateFlowStepGroup struct {
	ID         string              `json:"id"`
	WorkflowID string              `json:"workflow_id" validate:"required"`
	GroupIdx   int                 `json:"group_idx"`
	Parallel   bool                `json:"parallel"`
	Name       string              `json:"name"`
	Status     app.CompositeStatus `json:"status"`
	Labels     labels.Labels       `json:"labels"`
}

type CreateFlowStepGroupsRequest struct {
	Groups []CreateFlowStepGroup `json:"groups" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowCreateFlowStepGroups(ctx context.Context, reqs CreateFlowStepGroupsRequest) ([]*app.WorkflowStepGroup, error) {
	if len(reqs.Groups) == 0 {
		return []*app.WorkflowStepGroup{}, nil
	}

	groups := make([]*app.WorkflowStepGroup, 0, len(reqs.Groups))
	for _, req := range reqs.Groups {
		g := app.WorkflowStepGroup{
			ID:         req.ID,
			WorkflowID: req.WorkflowID,
			GroupIdx:   req.GroupIdx,
			Parallel:   req.Parallel,
			Name:       req.Name,
			Status:     req.Status,
		}
		g.Labels = req.Labels
		groups = append(groups, &g)
	}

	if res := a.db.WithContext(ctx).Create(groups); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create step groups")
	}

	return groups, nil
}
