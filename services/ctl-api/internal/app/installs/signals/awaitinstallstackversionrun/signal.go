package awaitinstallstackversionrun

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "await-install-stack-version-run"

type Signal struct {
	InstallStackID string
	WorkflowStepID string

	versionID string
}

const maxTimeout = 180 * 24 * time.Hour // 180 days

var _ signal.Signal = &Signal{}
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)
var _ signal.SignalWithTimeout = (*Signal)(nil)

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Timeout() time.Duration { return maxTimeout }

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.versionID != "" {
		statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(cancelCtx, statusactivities.UpdateStatusRequest{
			ID:     s.versionID,
			Status: app.NewCompositeTemporalStatus(cancelCtx, app.StatusCancelled),
		})
	}
	return nil
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallStackID == "" {
		return fmt.Errorf("install stack id is required")
	}

	_, err := activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
	if err != nil {
		return fmt.Errorf("unable to get install for stack: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	install, err := activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	version, err := activities.AwaitGetInstallStackVersionByInstallID(ctx, install.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install version")
	}
	s.versionID = version.ID

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   version.ID,
			StepTargetType: "install_stack_versions",
		}); err != nil {
			return errors.Wrap(err, "unable to update stack version")
		}
	}

	cb := callback.New(ctx, version.ID)
	if err := activities.AwaitUpdateInstallStackVersionCallback(ctx, activities.UpdateInstallStackVersionCallbackRequest{
		VersionID:   version.ID,
		CallbackRef: cb,
	}); err != nil {
		return errors.Wrap(err, "unable to store callback ref")
	}

	if install.SandboxMode.Bool {
		l.Info("sandbox mode org")
		workflow.Sleep(ctx, time.Second*5)

		region := ""
		switch {
		case install.AWSAccount != nil:
			region = install.AWSAccount.Region
		case install.AzureAccount != nil:
			region = install.AzureAccount.Location
		}

		installState, err := activities.AwaitGetInstallStateByInstallID(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install state for sandbox")
		}
		stateMap, err := installState.WorkflowSafeAsMap(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to convert install state to map")
		}

		data := helpers.GetFakeSandboxStackData(appCfg, region, stateMap)
		if err := activities.AwaitFireSandboxPhoneHome(ctx, &activities.FireSandboxPhoneHomeRequest{
			InstallID:   install.ID,
			PhoneHomeID: version.PhoneHomeID,
			Data:        data,
		}); err != nil {
			return errors.Wrap(err, "unable to fire sandbox phone home")
		}
	}

	result, err := callback.AwaitWithTimeout(ctx, cb, maxTimeout)
	if err != nil {
		statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: version.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusExpired, map[string]any{
				"err_message": "install stack was not applied before expiring",
			}),
		})

		if s.WorkflowStepID != "" {
			statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowStepID,
				Status: app.CompositeStatus{
					Status: app.StatusError,
					Metadata: map[string]any{
						"err_step_message": "Stack was not applied within 180 days and expired. Please reprovision install.",
					},
				},
			})
		}

		return errors.Wrap(err, "stack was not applied before expiring")
	}

	_ = result
	l.Debug("callback received, stack run processed by stack-run signal")

	return nil
}
