package deprovision

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	kuberunner "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/kuberunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "runner-deprovision"

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

	// Update runner status to deprovisioning
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusDeprovisioning,
		StatusDescription: "deprovisioning organization resources",
	}); err != nil {
		return err
	}

	// Create operation record
	op, err := activities.AwaitCreateOperationRequest(ctx, activities.CreateOperationRequest{
		RunnerID:      s.RunnerID,
		OperationType: app.RunnerOperationTypeDeprovision,
	})
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create operation",
		}); updateErr != nil {
			return errors.Wrap(err, "unable to create operation (also failed to update status)")
		}
		return errors.Wrap(err, "unable to create operation")
	}

	// Execute deprovision based on runner group type
	var execErr error
	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		execErr = s.executeDeprovisionOrgRunner(ctx, runner)
	case app.RunnerGroupTypeInstall:
		execErr = errors.New("install runners are provisioned via cloudformation stacks")
	default:
		execErr = fmt.Errorf("unsupported runner group type: %s", runner.RunnerGroup.Type)
	}

	if execErr != nil {
		if updateErr := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
			OperationID: op.ID,
			Status:      app.RunnerOperationStatusError,
		}); updateErr != nil {
			return fmt.Errorf("deprovision failed: %w (also failed to update operation status: %v)", execErr, updateErr)
		}
		return execErr
	}

	// Mark operation as finished
	if err := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
		OperationID: op.ID,
		Status:      app.RunnerOperationStatusFinished,
	}); err != nil {
		return err
	}

	// Transition runner to deprovisioned status
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusDeprovisioned,
		StatusDescription: "runner deprovisioned",
	}); err != nil {
		return err
	}

	return nil
}

// executeDeprovisionOrgRunner deprovisions an organization runner from Kubernetes
func (s *Signal) executeDeprovisionOrgRunner(ctx workflow.Context, runner *app.Runner) error {
	runnerID := runner.ID

	// Skip local runners
	if runner.RunnerGroup.Platform == app.AppRunnerTypeLocal {
		return nil
	}

	// Skip integration orgs
	if runner.Org.OrgType == app.OrgTypeIntegration {
		return nil
	}

	// Deprovision runner from Kubernetes
	req := kuberunner.DeprovisionRunnerRequest{
		RunnerID: runnerID,
	}

	_, err := kuberunner.AwaitDeprovisionRunner(ctx, req)
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to deprovision runner",
		}); updateErr != nil {
			return fmt.Errorf("unable to deprovision runner: %w (also failed to update status: %v)", err, updateErr)
		}
		return fmt.Errorf("unable to deprovision runner: %w", err)
	}

	return nil
}
