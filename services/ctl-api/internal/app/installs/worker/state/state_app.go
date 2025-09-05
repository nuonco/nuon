package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getAppStatePartial(ctx workflow.Context, installID string) (*state.AppState, error) {
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	currentApp := install.App

	st := state.NewAppState()
	st.Populated = true
	st.ID = currentApp.ID
	st.Name = currentApp.Name
	st.Status = string(currentApp.Status)

	for _, secr := range currentApp.AppSecrets {
		st.Variables[secr.Name] = secr.Value
	}

	return st, nil
}
