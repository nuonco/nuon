package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ReprovisionServiceAccount(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to get runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	op, err := activities.AwaitCreateOperationRequest(ctx, activities.CreateOperationRequest{
		RunnerID:      runner.ID,
		OperationType: app.RunnerOperationTypeProvisionServiceAccount,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create operation")
	}

	_, err = activities.AwaitCreateAccount(ctx, activities.CreateAccountRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusError)
		return errors.Wrap(err, "unable to create account")
	}

	w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusFinished)
	return nil
}
