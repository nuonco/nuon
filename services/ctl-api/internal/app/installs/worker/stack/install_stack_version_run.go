package stack

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/state"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/poll"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

// @temporal-gen workflow
// @execution-timeout 720h
// @task-timeout 30s
func (w *Workflows) InstallStackVersionRun(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForStackByStackID(ctx, sreq.ID)
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

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   version.ID,
		StepTargetType: plugins.TableName(w.db, version),
	}); err != nil {
		return errors.Wrap(err, "unable to update stack version")
	}

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	orgTyp, err := activities.AwaitGetOrgTypeByInstallID(ctx, install.ID)
	if err != nil {
		return err
	}
	if orgTyp == app.OrgTypeSandbox {
		l.Info("sandbox mode org")
		workflow.Sleep(ctx, time.Second*5)

		stackRefs := helpers.GetStackReferences(appCfg)
		data := map[string]any{
			"account":                  generics.GetFakeObj[string](),
			"region":                   region,
			"url":                      generics.GetFakeObj[string](),
			"maintenance_iam_role_arn": generics.GetFakeObj[string](),
			"provision_iam_role_arn":   generics.GetFakeObj[string](),
			"deprovision_iam_role_arn": generics.GetFakeObj[string](),
			"reprovision_iam_role_arn": generics.GetFakeObj[string](),
			"vpc_id":                   generics.GetFakeObj[string](),
			"account_id":               generics.GetFakeObj[string](),
			"public_subnets":           generics.GetFakeObj[string](),
			"private_subnets":          generics.GetFakeObj[string](),
			"runner_subnet":            generics.GetFakeObj[string](),
		}
		data = generics.MergeMap(refs.GetFakeRefs(stackRefs), data)

		run, err := activities.AwaitCreateSandboxInstallStackVersionRun(ctx, &activities.CreateSandboxInstallStackVersionRunRequest{
			StackVersionID: version.ID,
			Data:           generics.ToStringMap(data),
		})
		if err != nil {
			return errors.Wrap(err, "unable to create sandbox version run")
		}
		w.evClient.Send(ctx, install.RunnerID, &runnersignals.Signal{
			Type:                     runnersignals.OperationInstallStackVersionRun,
			InstallStackVersionRunID: run.ID,
		})

		if err := statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
			ID:     version.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusActive),
		}); err != nil {
			return errors.Wrap(err, "unable to update status")
		}

		return nil
	}

	var run *app.InstallStackVersionRun
	if err := poll.Poll(ctx, w.v, poll.PollOpts{
		MaxTS:           workflow.Now(ctx).Add(time.Hour * 24),
		InitialInterval: time.Second * 15,
		MaxInterval:     time.Minute * 15,
		BackoffFactor:   1.15,
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l, err := log.WorkflowLogger(ctx)
			if err != nil {
				return errors.Wrap(err, "unable to get workflow logger")
			}

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

			if statusErr := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: sreq.WorkflowStepID,
				Status: app.CompositeStatus{
					Status: app.StatusError,
					Metadata: map[string]any{
						"err_step_message": "Stack was not applied within 24hrs and expired. Please reprovision install.",
					},
				},
			}); statusErr != nil {
				return status.WrapStatusErr(err, statusErr)
			}

			return errors.Wrap(err, "stack was not applied before expiring")
		}

		return errors.Wrap(err, "unable to get install stack run in time")
	}

	w.evClient.Send(ctx, install.RunnerID, &runnersignals.Signal{
		Type:                     runnersignals.OperationInstallStackVersionRun,
		InstallStackVersionRunID: run.ID,
	})

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
		TriggeredByType: plugins.TableName(w.db, run),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
