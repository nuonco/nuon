package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getOrgStatePartial(ctx workflow.Context, installID string) (*state.OrgState, error) {
	org, err := activities.AwaitGetOrgByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}

	st := state.NewOrgState()
	st.Populated = true
	st.ID = org.ID
	st.Name = org.Name
	st.Status = string(org.Status)

	return st, nil
}
