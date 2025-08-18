package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
// @id-template {{.CallerID}}-generate-state-admin
func (w *Workflows) GenerateStateAdmin(ctx workflow.Context, sreq signals.RequestSignal) error {
	_, err := state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       sreq.ID,
		TriggeredByID:   sreq.ID,
		TriggeredByType: plugins.TableName(w.db, app.Install{}),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
