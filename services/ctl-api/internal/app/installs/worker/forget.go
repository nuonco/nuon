package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) forget(ctx workflow.Context, installID string) error {
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		InstallID: installID,
	}); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to delete install from database")
		return fmt.Errorf("unable to delete install: %w", err)
	}

	return nil
}
