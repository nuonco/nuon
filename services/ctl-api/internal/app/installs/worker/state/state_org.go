package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getOrgStatePartial(ctx workflow.Context, installID string) (*state.OrgState, error) {
	org, err := activities.AwaitGetOrgByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}
	return helpers.ToOrgState(*org), nil
}
