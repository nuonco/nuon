package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) Updated(ctx workflow.Context, sreq signals.RequestSignal) error {
	if err := activities.AwaitMarkStateStale(ctx, &activities.MarkStateStaleRequest{
		InstallID:       sreq.ID,
		TriggeredByID:   sreq.ID,
		TriggeredByType: plugins.TableName(w.db, &app.Install{}),
	}); err != nil {
		if !generics.IsGormErrRecordNotFound(err) {
			return errors.Wrap(err, "unable to mark state as stale")
		}
	}
	return nil
}
