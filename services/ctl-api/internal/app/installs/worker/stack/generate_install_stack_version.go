package stack

import (
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	awsstack "github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/gcp"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	DefaultAzureRunnerInitScript string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#azure"
	DefaultAWSRunnerInitScript   string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#default"
	DefaultGCPRunnerInitScript   string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/gcp/init.sh"
)

type GenerateInstallStackVersionRequest struct {
	ID             string `json:"id"`
	WorkflowStepID string `json:"workflow_step_id"`
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
func (w *Workflows) GenerateInstallStackVersion(ctx workflow.Context, sreq GenerateInstallStackVersionRequest) error {
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

	// Apply per-install stack template overrides before rendering so
	// template variables in override URLs get expanded.
	ApplyInstallStackOverrides(install, &cfg.StackConfig)

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

	// instance type comes from the latest synced runner config so reprovision picks up changes
	instanceType := cfg.RunnerConfig.InstanceType
	if instanceType == "" {
		instanceType = app.DefaultInstanceTypeForPlatform(cfg.RunnerConfig.CloudPlatform)
	}
	runner.RunnerGroup.Settings.AWSInstanceType = instanceType

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

	// GCP uses a static Terraform module with tfvars.
	if cfg.RunnerConfig.Type == app.AppRunnerTypeGCP {
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
			RunnerInitScriptURL:        initScriptURL,
			RunnerEnvVars:              stacks.FormatRunnerEnvVars(&cfg.RunnerConfig, w.cfg.RunnerContainerImageTag),
		}

		// Legacy init.sh needs a pre-provisioned bootstrap token.
		// init-mng-v2.sh fetches its own token via GCP identity (POST /v1/runner-auth/gcp).
		if isLegacyGCPInitScript(initScriptURL) {
			bootstrapToken, err := activities.AwaitCreateRunnerTokenRequestByRunnerID(ctx, install.RunnerID)
			if err != nil {
				return errors.Wrap(err, "unable to create bootstrap token")
			}
			inp.APIToken = generics.FromPtrStr(bootstrapToken)
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
		RunnerEnvVars:              stacks.FormatRunnerEnvVars(&cfg.RunnerConfig, w.cfg.RunnerContainerImageTag),
	}

	switch cfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		if cfg.RunnerConfig.InitScriptURL != "" {
			inp.RunnerInitScriptURL = cfg.RunnerConfig.InitScriptURL
		} else {
			inp.RunnerInitScriptURL = DefaultAWSRunnerInitScript
		}

		phoneHomeScript, err := activities.AwaitGetPhoneHomeScriptRaw(ctx, &activities.GetPhoneHomeScriptRequest{})
		if err != nil {
			return errors.Wrap(err, "unable to get phone home script")
		}
		inp.PhonehomeScript = string(phoneHomeScript)
		inp.VPCNestedStackTemplateURL = cfg.StackConfig.VPCNestedTemplateURL
		inp.RunnerNestedStackTemplateURL = cfg.StackConfig.RunnerNestedTemplateURL

		renderedTemplate, err := activities.AwaitRenderAWSStackTemplate(ctx, &activities.RenderAWSStackTemplateRequest{
			Input: *inp,
		})
		if err != nil {
			return errors.Wrap(err, "unable to render stack template")
		}

		tmplByts = renderedTemplate.RAWJson
		checksum = renderedTemplate.Checksum

		// Render the Terraform tfvars envelope alongside the CloudFormation
		// template so the dashboard can offer both during the await step.
		// Log (but don't fail) if the TF module can't render the app config
		// (e.g., custom nested stacks); the CFN path remains usable.
		inp.CloudFormationStackVersion = stackVersion
		supportIAMRoleARN := ""
		if w.cfg.RunnerEnableSupport {
			supportIAMRoleARN = w.cfg.RunnerDefaultSupportIAMRole
		}
		tfvarsBytes, tfvarsChecksum, tfErr := awsstack.Render(inp, supportIAMRoleARN)
		if tfErr != nil {
			workflow.GetLogger(ctx).Warn("aws terraform render skipped", "error", tfErr.Error(), "install_id", install.ID)
		} else if len(tfvarsBytes) == 0 {
			workflow.GetLogger(ctx).Warn("aws terraform render produced empty bytes", "install_id", install.ID)
		} else {
			workflow.GetLogger(ctx).Info("aws terraform render ok", "install_id", install.ID, "bytes", len(tfvarsBytes))
			if err := activities.AwaitSaveInstallStackVersionTerraform(ctx, &activities.SaveInstallStackVersionTerraformRequest{
				ID:       stackVersion.ID,
				Template: tfvarsBytes,
				Checksum: tfvarsChecksum,
			}); err != nil {
				return errors.Wrap(err, "unable to save aws tfvars")
			}
		}

	case app.AppRunnerTypeAzure:
		if cfg.RunnerConfig.InitScriptURL != "" {
			inp.RunnerInitScriptURL = cfg.RunnerConfig.InitScriptURL
		} else {
			inp.RunnerInitScriptURL = DefaultAzureRunnerInitScript
		}

		inp.VPCNestedStackTemplateURL = cfg.StackConfig.VPCNestedTemplateURL
		inp.RunnerNestedStackTemplateURL = cfg.StackConfig.RunnerNestedTemplateURL

		renderedTemplate, err := activities.AwaitRenderARMStackTemplate(ctx, &activities.RenderARMStackTemplateRequest{
			Input: *inp,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create ARM template")
		}
		tmplByts = renderedTemplate.RAWJson
		checksum = renderedTemplate.Checksum
	}

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

func isLegacyGCPInitScript(url string) bool {
	return strings.HasSuffix(url, "/scripts/gcp/init.sh") || url == DefaultGCPRunnerInitScript
}
