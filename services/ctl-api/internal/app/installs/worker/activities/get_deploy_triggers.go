package activities

import (
	"context"
	"sort"

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
	if len(workflows) == 0 {
		return workflows, nil
	}

	// maps: id to index and id to string
	indices := map[string]int{}
	wkflows := make(map[string]*app.InstallActionWorkflow, 0)
	for _, workflow := range workflows {
		cfg, err := a.getActionWorkflowLatestConfig(ctx, workflow.ActionWorkflowID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get action workflow config")
		}

		if req.ComponentID == "" {
			if cfg.HasTrigger(req.TriggerType) {
				wkflows[workflow.ID] = workflow
				index := cfg.GetTriggerIndex(req.TriggerType)
				indices[workflow.ID] = index
			}
		} else {
			if cfg.HasComponentTrigger(req.TriggerType, req.ComponentID) {
				wkflows[workflow.ID] = workflow
				index := cfg.GetComponentTriggerIndex(req.TriggerType, req.ComponentID)
				indices[workflow.ID] = index
			}
		}
	}

	// get the workflow IDs
	workflowIDs := make([]string, 0)
	for wkflowID := range indices {
		workflowIDs = append(workflowIDs, wkflowID)
	}

	// sort the workflowIDs by the value in the orders map
	sort.SliceStable(workflowIDs, func(i, j int) bool {
		return indices[workflowIDs[i]] < indices[workflowIDs[j]]
	})

	orderedWorkflows := make([]*app.InstallActionWorkflow, len(workflowIDs)) // final list
	for i, wkflowID := range workflowIDs {
		wkflow, ok := wkflows[wkflowID]
		if ok {
			orderedWorkflows[i] = wkflow
		}
	}

	return orderedWorkflows, nil
}
