package provisionserviceaccount

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "provision_service_account"

type Signal struct {
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	// Validate runner exists in database
	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Update runner status to provisioning
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusProvisioning,
		StatusDescription: "provisioning runner resources",
	}); err != nil {
		return err
	}

	// Get runner details
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to get runner from database",
		}); updateErr != nil {
			return fmt.Errorf("unable to get runner: %w (also failed to update status: %v)", err, updateErr)
		}
		return fmt.Errorf("unable to get runner: %w", err)
	}

	// Create operation record for provision service account
	op, err := activities.AwaitCreateOperationRequest(ctx, activities.CreateOperationRequest{
		RunnerID:      runner.ID,
		OperationType: app.RunnerOperationTypeProvisionServiceAccount,
	})
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create service account",
		}); updateErr != nil {
			return errors.Wrap(err, "unable to create operation (also failed to update status)")
		}
		return errors.Wrap(err, "unable to create operation")
	}

	// Create service account for runner
	_, err = activities.AwaitCreateAccount(ctx, activities.CreateAccountRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		// Update both runner status and operation status on failure
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create runner service account",
		}); updateErr != nil {
			return errors.Wrap(err, "unable to create account (also failed to update runner status)")
		}
		if updateErr := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
			OperationID: op.ID,
			Status:      app.RunnerOperationStatusError,
		}); updateErr != nil {
			return errors.Wrap(err, "unable to create account (also failed to update operation status)")
		}
		return errors.Wrap(err, "unable to create account")
	}

	// Mark operation as finished
	if err := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
		OperationID: op.ID,
		Status:      app.RunnerOperationStatusFinished,
	}); err != nil {
		return err
	}

	// Update runner status to awaiting install stack run
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          runner.ID,
		Status:            app.RunnerStatusAwaitingInstallStackRun,
		StatusDescription: "awaiting stack run",
	}); err != nil {
		return err
	}

	return nil
}
