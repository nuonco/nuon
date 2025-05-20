package worker

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Deprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to get runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}
	w.updateStatus(ctx, sreq.ID, app.RunnerStatusDeprovisioning, "deprovisioning organization resources")

	op, err := activities.AwaitCreateOperationRequest(ctx, activities.CreateOperationRequest{
		RunnerID:      sreq.ID,
		OperationType: app.RunnerOperationTypeDeprovision,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create operation")
		return errors.Wrap(err, "unable to create operation")
	}

	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		err = w.executeDeprovisionOrgRunner(ctx, sreq.ID, sreq.SandboxMode)
	case app.RunnerGroupTypeInstall:
		err = errors.New("install runners are provisioned via cloudformation stacks")
	}
	if err != nil {
		w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusError)
		return err
	}

	w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusFinished)
	return nil
}
