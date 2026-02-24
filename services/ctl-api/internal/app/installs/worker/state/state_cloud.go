package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (h *Workflows) toCloudAccount(ctx workflow.Context, installID string) (*state.CloudAccount, error) {
	st := state.NewCloudAccount()

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	if install.AWSAccount != nil {
		st.AWS = &state.AWSCloudAccount{
			Region: install.AWSAccount.Region,
		}
	}

	if install.AzureAccount != nil {
		st.Azure = &state.AzureCloudAccount{
			Location: install.AzureAccount.Location,
		}
	}

	return st, nil
}
