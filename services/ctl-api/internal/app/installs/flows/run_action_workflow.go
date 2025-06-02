package flows

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Flows) RunActionWorkflow(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	installActionWorkflowID, ok := flw.Metadata["install_action_workflow_id"]
	if !ok {
		return nil, errors.New("install action workflow is not set on the install workflow for a manual deploy")
	}
	triggeredByID, ok := flw.Metadata["triggerred_by_id"]
	if !ok {
		return nil, errors.New("triggerred by id is not set on the install workflow for a manual deploy")
	}

	iaw, err := activities.AwaitGetInstallActionWorkflowByID(ctx, generics.FromPtrStr(installActionWorkflowID))

	steps := make([]*app.FlowStep, 0)
	prefix := "RUNENV_"
	runEnvVars := map[string]string{}

	for key, value := range flw.Metadata {
		if strings.HasPrefix(key, prefix) {
			// Remove the prefix and add to result map
			newKey := key[len(prefix):]
			runEnvVars[newKey] = *value
		}
	}

	runEnvVars["TRIGGER_TYPE"] = string(app.ActionWorkflowTriggerTypeManual)

	sig := &signals.Signal{
		Type: signals.OperationExecuteActionWorkflow,
		InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
			InstallActionWorkflowID: iaw.ID,
			TriggerType:             app.ActionWorkflowTriggerTypeManual,
			TriggeredByID:           generics.FromPtrStr(triggeredByID),
			TriggeredByType:         string(app.ActionWorkflowTriggerTypeManual),
			RunEnvVars:              runEnvVars,
		},
	}
	name := fmt.Sprintf("%s action workflow run", string(app.ActionWorkflowTriggerTypeManual))
	step, err := w.installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig)
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)
	return steps, nil
}
