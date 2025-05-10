package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cloudformation"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) GenerateInstallStackVersion(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get stack")
	}

	// need to fetch app config
	cfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	if cfg.RunnerConfig.Type != app.AppRunnerTypeAWS {
		return nil
	}

	installState, err := activities.AwaitGetInstallStateByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install state")
	}

	// generate fields
	stateData, err := installState.AsMap()
	if err != nil {
		return errors.Wrap(err, "unable to generate install map data")
	}
	if err := render.RenderStruct(&cfg.PermissionsConfig, stateData); err != nil {
		return errors.Wrap(err, "unable to render permissions config")
	}
	if err := render.RenderStruct(&cfg.SecretsConfig, stateData); err != nil {
		return errors.Wrap(err, "unable to render secrets config")
	}
	if err := render.RenderStruct(&cfg.StackConfig, stateData); err != nil {
		return errors.Wrap(err, "unable to render cloudformation stack config")
	}

	runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	// need to generate a token
	stackVersion, err := activities.AwaitCreateInstallStackVersion(ctx, &activities.CreateInstallStackVersionRequest{
		InstallID:      sreq.ID,
		InstallStackID: stack.ID,
		AppConfigID:    cfg.ID,
		StackName:      cfg.StackConfig.Name,
		Region:         install.AWSAccount.Region,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create cloudformation stack version")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   stackVersion.ID,
		StepTargetType: plugins.TableName(w.db, stackVersion),
	}); err != nil {
		return errors.Wrap(err, "unable to update stack version")
	}

	token, err := activities.AwaitCreateRunnerTokenRequestByRunnerID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to create runner token")
	}

	// generate the cloudformation stack
	inp := &cloudformation.TemplateInput{
		Install:                    install,
		CloudFormationStackVersion: stackVersion,
		InstallState:               installState,
		AppCfg:                     cfg,
		Runner:                     runner,
		Settings:                   &runner.RunnerGroup.Settings,
		APIToken:                   generics.FromPtrStr(token),
	}
	tmpl, checksum, err := w.templates.Template(inp)
	if err != nil {
		return errors.Wrap(err, "unable to create cloudformation template")
	}

	tmplByts, err := tmpl.JSON()
	if err != nil {
		return errors.Wrap(err, "unable to get cloudformation json")
	}

	// upload and publish the stack
	if err := activities.AwaitUploadAWSCloudFormationStackVersionTemplate(ctx, &activities.UploadAWSCloudFormationStackVersionTemplateRequest{
		BucketKey: stackVersion.AWSBucketKey,
		Template:  tmplByts,
	}); err != nil {
		return errors.Wrap(err, "unable to upload cloudformation stack")
	}

	if err := activities.AwaitSaveInstallStackVersionTemplate(ctx, &activities.SaveInstallStackVersionTemplateRequest{
		ID:       stackVersion.ID,
		Template: tmplByts,
		Checksum: checksum,
	}); err != nil {
		return errors.Wrap(err, "unable to save cloudformation stack")
	}

	statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
		ID:     stackVersion.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusPendingUser),
	})
	return nil
}
