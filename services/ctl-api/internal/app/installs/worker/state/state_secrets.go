package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (w *Workflows) getSecretsStatePartial(ctx workflow.Context, installID string) (*state.SecretsState, error) {
	runnerJob, err := activities.AwaitGetSecretsSyncJobByInstallID(ctx, installID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			return helpers.ToSecretsState(nil), nil
		}
		return nil, errors.Wrap(err, "unable to get secrets state")
	}
	return helpers.ToSecretsState(runnerJob.ParsedOutputs), nil
}
