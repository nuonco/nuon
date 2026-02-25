package stack

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/bicep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/gcp"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	DefaultAzureRunnerInitScript string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#azure"
	DefaultAWSRunnerInitScript   string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#default"
	DefaultGCPRunnerInitScript   string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/gcp/init.sh"
)

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
func (w *Workflows) GenerateInstallStackVersion(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetInstallForStackByStackID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	// need to fetch app config
	cfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	// If we are not using one of the new independent runner types, stop here.
	// To support backwards compatibility, we do not return an error, but we cannot create a stack.
	if !generics.SliceContains(cfg.RunnerConfig.Type, []app.AppRunnerType{
		app.AppRunnerTypeAWS,
		app.AppRunnerTypeAzure,
		app.AppRunnerTypeGCP,
	}) {
		return nil
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, install.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get stack")
	}

	installState, err := activities.AwaitGetInstallStateByInstallID(ctx, install.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install state")
	}

	// generate fields
	stateData, err := installState.WorkflowSafeAsMap(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to generate install map data")
	}
	if err := render.RenderStruct(&cfg.PermissionsConfig, stateData); err != nil {
		return errors.Wrap(err, "unable to render permissions config")
	}
	if err := render.RenderStruct(&cfg.BreakGlassConfig, stateData); err != nil {
		return errors.Wrap(err, "unable to render break glass permissions config")
	}

	if err := render.RenderStruct(&cfg.SecretsConfig, stateData); err != nil {
		return errors.Wrap(err, "unable to render secrets config")
	}

	if stackErr := render.RenderStruct(&cfg.StackConfig, stateData); stackErr != nil {
		return errors.Wrap(stackErr, "unable to render stack config")
	}

	// update cf stack param name post rendering variables
	for i := range cfg.SecretsConfig.Secrets {
		secret := &cfg.SecretsConfig.Secrets[i]
		secret.UpdateCloudformationStackInfo()
	}

	if err := render.RenderStruct(&cfg.StackConfig, stateData); err != nil {
		return errors.Wrap(err, "unable to render cloudformation stack config")
	}

	runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	// need to generate a token
	region := ""
	switch {
	case install.AWSAccount != nil:
		region = install.AWSAccount.Region
	case install.AzureAccount != nil:
		region = install.AzureAccount.Location
	case install.GCPAccount != nil:
		region = install.GCPAccount.Region
	}
	stackVersion, err := activities.AwaitCreateInstallStackVersion(ctx, &activities.CreateInstallStackVersionRequest{
		InstallID:      install.ID,
		InstallStackID: stack.ID,
		AppConfigID:    cfg.ID,
		StackName:      cfg.StackConfig.Name,
		Region:         region,
		Platform:       string(cfg.RunnerConfig.Type),
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

	// GCP uses a static Terraform module with tfvars and a short-lived bootstrap token.
	// AWS and Azure use the existing full-template flow with long-lived tokens.
	if cfg.RunnerConfig.Type == app.AppRunnerTypeGCP {
		bootstrapToken, err := activities.AwaitCreateRunnerBootstrapTokenRequestByRunnerID(ctx, install.RunnerID)
		if err != nil {
			return errors.Wrap(err, "unable to create bootstrap token")
		}

		initScriptURL := DefaultGCPRunnerInitScript
		if cfg.RunnerConfig.InitScriptURL != "" {
			initScriptURL = cfg.RunnerConfig.InitScriptURL
		}

		inp := &stacks.TemplateInput{
			Install:                    install,
			CloudFormationStackVersion: stackVersion,
			InstallState:               installState,
			AppCfg:                     cfg,
			Runner:                     runner,
			Settings:                   &runner.RunnerGroup.Settings,
			APIToken:                   generics.FromPtrStr(bootstrapToken),
			RunnerInitScriptURL:        initScriptURL,
		}

		tmplByts, checksum, err := gcp.Render(inp)
		if err != nil {
			return errors.Wrap(err, "unable to render gcp tfvars")
		}

		if err := activities.AwaitSaveInstallStackVersionTemplate(ctx, &activities.SaveInstallStackVersionTemplateRequest{
			ID:       stackVersion.ID,
			Template: tmplByts,
			Checksum: checksum,
		}); err != nil {
			return errors.Wrap(err, "unable to save gcp tfvars")
		}

		statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
			ID:     stackVersion.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusPendingUser),
		})
		return nil
	}

	// AWS and Azure flow: full template generation + S3 upload.
	token, err := activities.AwaitCreateRunnerTokenRequestByRunnerID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to create runner token")
	}

	tmplByts := []byte{}
	checksum := ""
	inp := &stacks.TemplateInput{
		Install:                    install,
		CloudFormationStackVersion: stackVersion,
		InstallState:               installState,
		AppCfg:                     cfg,
		Runner:                     runner,
		Settings:                   &runner.RunnerGroup.Settings,
		APIToken:                   generics.FromPtrStr(token),
	}

	switch cfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		phoneHomeScript, err := activities.AwaitGetPhoneHomeScriptRaw(ctx, &activities.GetPhoneHomeScriptRequest{})
		if err != nil {
			return errors.Wrap(err, "unable to get phone home script")
		}
		inp.PhonehomeScript = string(phoneHomeScript)
		inp.VPCNestedStackTemplateURL = cfg.StackConfig.VPCNestedTemplateURL
		inp.RunnerNestedStackTemplateURL = cfg.StackConfig.RunnerNestedTemplateURL

		if cfg.RunnerConfig.InitScriptURL != "" {
			inp.RunnerInitScriptURL = cfg.RunnerConfig.InitScriptURL
		} else {
			inp.RunnerInitScriptURL = DefaultAWSRunnerInitScript
		}

		renderedTemplate, err := activities.AwaitRenderAWSStackTemplate(ctx, &activities.RenderAWSStackTemplateRequest{
			Input: *inp,
		})
		if err != nil {
			return errors.Wrap(err, "unable to render stack template")
		}

		tmplByts = renderedTemplate.RAWJson
		checksum = renderedTemplate.Checksum

	case app.AppRunnerTypeAzure:
		if cfg.RunnerConfig.InitScriptURL != "" {
			inp.RunnerInitScriptURL = cfg.RunnerConfig.InitScriptURL
		} else {
			inp.RunnerInitScriptURL = DefaultAzureRunnerInitScript
		}

		tmplByts, checksum, err = bicep.Render(inp)
		if err != nil {
			return errors.Wrap(err, "unable to create bicep template")
		}
	}

	// upload stack template to S3
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
