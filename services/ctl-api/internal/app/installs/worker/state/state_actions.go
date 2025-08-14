package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getActionsStatePartial(ctx workflow.Context, installID string) (*state.ActionsState, error) {
	actions, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get actions")
	}

	st := state.NewActionsState()
	st.Populated = true
	for _, action := range actions {
		actWorkflowState, err := w.toActionWorkflow(ctx, action.ID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get action workfloow state")
		}

		st.Workflows[action.ActionWorkflow.Name] = actWorkflowState
	}

	return st, nil
}

func (h *Workflows) toActionWorkflow(ctx workflow.Context, iawID string) (*state.ActionWorkflowState, error) {
	act, err := activities.AwaitGetInstallActionWorkflowStateByInstallActionWorkflowID(ctx, iawID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install action workflow id")
	}

	st := state.NewActionWorkflowState()
	st.Populated = true
	st.Status = string(act.Status)
	st.ID = act.ActionWorkflow.ID

	for _, run := range act.Runs {
		if run.RunnerJob != nil {
			st.Outputs = run.RunnerJob.ParsedOutputs
			break
		}
	}

	return st, nil
}
