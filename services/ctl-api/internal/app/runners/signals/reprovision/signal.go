package reprovision

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	kuberunner "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/kuberunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "runner-reprovision"

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
		return fmt.Errorf("unable to get runner: %w", err)
	}

	// Update runner status to provisioning
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusProvisioning,
		StatusDescription: "provisioning organization resources",
	}); err != nil {
		return err
	}
	statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusProvisioning,
		StatusDescription: "provisioning organization resources",
	})

	// Create operation record (reuses Provision type)
	op, err := activities.AwaitCreateOperationRequest(ctx, activities.CreateOperationRequest{
		RunnerID:      s.RunnerID,
		OperationType: app.RunnerOperationTypeProvision,
	})
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create operation",
		}); updateErr != nil {
			return errors.Wrap(err, "unable to create operation (also failed to update status)")
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create operation",
		})
		return errors.Wrap(err, "unable to create operation")
	}

	// Create/recreate service account
	_, err = activities.AwaitCreateAccount(ctx, activities.CreateAccountRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		if updateErr := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
			OperationID: op.ID,
			Status:      app.RunnerOperationStatusError,
		}); updateErr != nil {
			return fmt.Errorf("unable to create account: %w (also failed to update operation status: %v)", err, updateErr)
		}
		statusactivities.AwaitUpdateRunnerOperationStatusV2(ctx, statusactivities.UpdateRunnerOperationStatusV2Request{
			RunnerOperationID: op.ID,
			Status:            app.RunnerOperationStatusError,
			StatusDescription: "unable to create runner service account",
		})
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create runner service account",
		}); updateErr != nil {
			return fmt.Errorf("unable to create account: %w (also failed to update runner status: %v)", err, updateErr)
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create runner service account",
		})
		return fmt.Errorf("unable to create account: %w", err)
	}

	// Create API token for runner
	token, err := activities.AwaitCreateToken(ctx, activities.CreateTokenRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		if updateErr := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
			OperationID: op.ID,
			Status:      app.RunnerOperationStatusError,
		}); updateErr != nil {
			return fmt.Errorf("unable to create token: %w (also failed to update operation status: %v)", err, updateErr)
		}
		statusactivities.AwaitUpdateRunnerOperationStatusV2(ctx, statusactivities.UpdateRunnerOperationStatusV2Request{
			RunnerOperationID: op.ID,
			Status:            app.RunnerOperationStatusError,
			StatusDescription: "unable to create runner token",
		})
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create runner token",
		}); updateErr != nil {
			return fmt.Errorf("unable to create token: %w (also failed to update runner status: %v)", err, updateErr)
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create runner token",
		})
		return fmt.Errorf("unable to create token: %w", err)
	}

	// Execute reprovision based on runner group type
	var execErr error
	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		execErr = s.executeProvisionOrgRunner(ctx, runner, token.Token, op.ID)
	case app.RunnerGroupTypeInstall:
		execErr = errors.New("can not reprovision install runner")
	default:
		execErr = fmt.Errorf("unsupported runner group type: %s", runner.RunnerGroup.Type)
	}

	if execErr != nil {
		if updateErr := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
			OperationID: op.ID,
			Status:      app.RunnerOperationStatusError,
		}); updateErr != nil {
			return errors.Wrap(execErr, "reprovision failed (also failed to update operation status)")
		}
		statusactivities.AwaitUpdateRunnerOperationStatusV2(ctx, statusactivities.UpdateRunnerOperationStatusV2Request{
			RunnerOperationID: op.ID,
			Status:            app.RunnerOperationStatusError,
			StatusDescription: "reprovision failed",
		})
		return errors.Wrap(execErr, "unable to reprovision")
	}

	// Mark operation as finished
	if err := activities.AwaitUpdateOperation(ctx, activities.UpdateOperationRequest{
		OperationID: op.ID,
		Status:      app.RunnerOperationStatusFinished,
	}); err != nil {
		return err
	}
	statusactivities.AwaitUpdateRunnerOperationStatusV2(ctx, statusactivities.UpdateRunnerOperationStatusV2Request{
		RunnerOperationID: op.ID,
		Status:            app.RunnerOperationStatusFinished,
		StatusDescription: "operation finished",
	})

	return nil
}

// executeProvisionOrgRunner provisions an organization runner in Kubernetes (shared with provision signal)
func (s *Signal) executeProvisionOrgRunner(ctx workflow.Context, runner *app.Runner, apiToken string, opID string) error {
	runnerID := runner.ID

	// Skip local runners
	if runner.RunnerGroup.Platform == app.AppRunnerTypeLocal {
		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusActive,
			StatusDescription: "local runner must be run locally",
		}); err != nil {
			return err
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusActive,
			StatusDescription: "local runner must be run locally",
		})
		return nil
	}

	// Skip integration orgs
	if runner.Org.OrgType == app.OrgTypeIntegration {
		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusActive,
			StatusDescription: "integration mode, bypassing provisioning",
		}); err != nil {
			return err
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusActive,
			StatusDescription: "integration mode, bypassing provisioning",
		})
		return nil
	}

	// Derive cloud provider and identity from runner platform
	var cloudProvider string
	var runnerIAMRole string
	switch runner.RunnerGroup.Platform {
	case app.AppRunnerTypeGCPGKE:
		cloudProvider = "gcp"
		runnerIAMRole = runner.RunnerGroup.Settings.OrgGCPServiceAccount
	case app.AppRunnerTypeAzureAKS:
		cloudProvider = "azure"
		runnerIAMRole = runner.RunnerGroup.Settings.OrgAzureClientID
	default:
		cloudProvider = "aws"
		runnerIAMRole = runner.RunnerGroup.Settings.OrgAWSIAMRoleARN
	}

	// Provision runner in Kubernetes
	req := &kuberunner.ProvisionRunnerRequest{
		RunnerID:                 runnerID,
		APIURL:                   runner.RunnerGroup.Settings.RunnerAPIURL,
		APIToken:                 apiToken,
		CloudProvider:            cloudProvider,
		RunnerIAMRole:            runnerIAMRole,
		RunnerServiceAccountName: runner.RunnerGroup.Settings.OrgK8sServiceAccountName,
		Image: kuberunner.ProvisionRunnerRequestImage{
			URL: runner.RunnerGroup.Settings.ContainerImageURL,
			Tag: runner.RunnerGroup.Settings.ContainerImageTag,
		},
	}

	_, err := kuberunner.AwaitProvisionRunner(ctx, req)
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to provision runner",
		}); updateErr != nil {
			return errors.Wrap(err, "unable to provision runner (also failed to update status)")
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to provision runner",
		})
		return errors.Wrap(err, "unable to provision runner")
	}

	// Update runner status to active
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          runnerID,
		Status:            app.RunnerStatusActive,
		StatusDescription: "runner is active and ready to process jobs",
	}); err != nil {
		return err
	}
	statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
		RunnerID:          runnerID,
		Status:            app.RunnerStatusActive,
		StatusDescription: "runner is active and ready to process jobs",
	})

	return nil
}
