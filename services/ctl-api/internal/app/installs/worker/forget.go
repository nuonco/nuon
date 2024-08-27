package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) forget(ctx workflow.Context, installID string) error {
	if err := activities.AwaitDeleteByInstallID(ctx, installID); err != nil {
		return fmt.Errorf("unable to delete install: %w", err)
	}

	return nil
}
