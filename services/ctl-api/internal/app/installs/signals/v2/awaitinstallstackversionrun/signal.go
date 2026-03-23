package awaitinstallstackversionrun

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	runnersignalsv2 "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/installstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "await-install-stack-version-run"

type Signal struct {
	InstallStackID string
	WorkflowStepID string
}

var _ signal.Signal = &Signal{}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallStackID == "" {
		return fmt.Errorf("install stack id is required")
	}

	// Validate install stack exists
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

	region := ""
	switch {
	case install.AWSAccount != nil:
		region = install.AWSAccount.Region
	case install.AzureAccount != nil:
		region = install.AzureAccount.Location
	}

	version, err := activities.AwaitGetInstallStackVersionByInstallID(ctx, install.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install version")
	}

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

	orgTyp, err := activities.AwaitGetOrgTypeByInstallID(ctx, install.ID)
	if err != nil {
		return err
	}

	if orgTyp == app.OrgTypeSandbox {
		l.Info("sandbox mode org")
		workflow.Sleep(ctx, time.Second*5)

		installState, err := activities.AwaitGetInstallStateByInstallID(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install state for sandbox")
		}
		stateMap, err := installState.WorkflowSafeAsMap(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to convert install state to map")
		}

		data := helpers.GetFakeSandboxStackData(appCfg, region, stateMap)

		run, err := activities.AwaitCreateSandboxInstallStackVersionRun(ctx, &activities.CreateSandboxInstallStackVersionRunRequest{
			StackVersionID: version.ID,
			Data:           generics.ToStringMap(data),
		})
		if err != nil {
			return errors.Wrap(err, "unable to create sandbox version run")
		}

		// Send signal to runner using cross-namespace signal sending
		_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   install.RunnerID,
			OwnerType: "runners",
			Signal: &runnersignalsv2.Signal{
				RunnerID:                 install.RunnerID,
				InstallStackVersionRunID: run.ID,
			},
		})
		if err != nil {
			return errors.Wrap(err, "unable to enqueue signal to runner")
		}

		if err := statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
			ID:     version.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusActive),
		}); err != nil {
			return errors.Wrap(err, "unable to update status")
		}

		return nil
	}

	// Not sandbox mode - poll for stack version run
	v := validator.New()
	var run *app.InstallStackVersionRun
	if err := poll.Poll(ctx, v, poll.PollOpts{
		MaxTS:           workflow.Now(ctx).Add(time.Hour * 24),
		InitialInterval: time.Second * 15,
		MaxInterval:     time.Minute * 15,
		BackoffFactor:   1.15,
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l := workflow.GetLogger(ctx)
			l.Debug("checking install stack status again in "+dur.String(), zap.Duration("duration", dur))
			return nil
		},
		Fn: func(ctx workflow.Context) error {
			run, err = activities.AwaitGetInstallStackVersionRunByVersionID(ctx, version.ID)
			return err
		},
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: version.ID,
				Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusExpired, map[string]any{
					"err_message": "cloudformation stack was not applied before expiring",
				}),
			})

			if s.WorkflowStepID != "" {
				if statusErr := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: s.WorkflowStepID,
					Status: app.CompositeStatus{
						Status: app.StatusError,
						Metadata: map[string]any{
							"err_step_message": "Stack was not applied within 24hrs and expired. Please reprovision install.",
						},
					},
				}); statusErr != nil {
					return status.WrapStatusErr(err, statusErr)
				}
			}

			return errors.Wrap(err, "stack was not applied before expiring")
		}

		return errors.Wrap(err, "unable to get install stack run in time")
	}

	// Send signal to runner using cross-namespace signal sending
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   install.RunnerID,
		OwnerType: "runners",
		Signal: &runnersignalsv2.Signal{
			RunnerID:                 install.RunnerID,
			InstallStackVersionRunID: run.ID,
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to enqueue signal to runner")
	}

	// successfully got a run
	l.Debug("successfully got run", zap.Any("data", run.Data))
	if err := statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
		ID:     version.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusActive),
	}); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   run.ID,
		TriggeredByType: "install_stack_version_runs",
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
