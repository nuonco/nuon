package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getActionsStatePartial(ctx workflow.Context, installID string) (*state.ActionsState, error) {
	actions, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get actions")
	}
	st := state.NewActionsState()
	st.Populated = true
	for _, action := range actions {
		act, err := activities.AwaitGetInstallActionWorkflowStateByInstallActionWorkflowID(ctx, action.ID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get action workflow state")
		}
		st.Workflows[action.ActionWorkflow.Name] = helpers.ToActionWorkflowState(*act)
	}
	return st, nil
}
