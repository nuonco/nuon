package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Provision(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.RunnerStatusProvisioning, "provisioning organization resources")

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to get runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	op, err := activities.AwaitCreateOperationRequest(ctx, activities.CreateOperationRequest{
		RunnerID:      sreq.ID,
		OperationType: app.RunnerOperationTypeProvision,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create operation")
		return errors.Wrap(err, "unable to create operation")
	}

	_, err = activities.AwaitCreateAccount(ctx, activities.CreateAccountRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create runner service account")
		w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusError)
		return errors.Wrap(err, "unable to create account")
	}

	token, err := activities.AwaitCreateToken(ctx, activities.CreateTokenRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusError)
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create runner token")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	// if no log stream exists, create one
	_, err = cctx.GetLogStreamWorkflow(ctx)
	if err != nil {
		logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, op.ID)
		if err != nil {
			return errors.Wrap(err, "unable to create log stream")
		}
		defer func() {
			activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
		}()
		ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	}

	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		err = w.executeProvisionOrgRunner(ctx, sreq.ID, token.Token, sreq.SandboxMode)
	}
	if err != nil {
		w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusError)
		return err
	}

	w.updateOperationStatus(ctx, op.ID, app.RunnerOperationStatusFinished)

	return nil
}
