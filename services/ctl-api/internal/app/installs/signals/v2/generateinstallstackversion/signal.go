package generateinstallstackversion

import (
	"fmt"

	"strings"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	stackoverrides "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/stack"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/bicep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/gcp"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "generate-install-stack-version"

const (
	DefaultAzureRunnerInitScript string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#azure"
	DefaultAWSRunnerInitScript   string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#default"
	DefaultGCPRunnerInitScript   string = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/gcp/init.sh"
)

type Signal struct {
	InstallStackID string
	WorkflowStepID string

	cfg *internal.Config
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) WithParams(params *signal.Params) {
	s.cfg = params.Cfg
}

var _ signal.SignalWithParams = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)

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

	// Validate install stack exists
	_, err := activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
	if err != nil {
		return fmt.Errorf("unable to get install for stack: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGetInstallForStackByStackID(ctx, s.InstallStackID)
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
	stackoverrides.ApplyInstallStackOverrides(install, &cfg.StackConfig)

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

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   stackVersion.ID,
			StepTargetType: "install_stack_versions",
		}); err != nil {
			return errors.Wrap(err, "unable to update stack version")
		}
	}

	// GCP uses a static Terraform module with tfvars — no S3 upload needed.
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
			RunnerEnvVars:              stacks.FormatRunnerEnvVars(&cfg.RunnerConfig, s.cfg.RunnerContainerImageTag),
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

	token, err := activities.AwaitCreateRunnerTokenRequestByRunnerID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to create runner token")
	}

	// TODO(ja): Ignoring this for Azure. Should probably update.

	// AWS and Azure diverge here, while generating the stack template file.

	// Generate the stack template.
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
		RunnerEnvVars:              stacks.FormatRunnerEnvVars(&cfg.RunnerConfig, s.cfg.RunnerContainerImageTag),
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

		// NOTE(fd): we set the runner init script here dynamically in order to have it readily available on the input
		// the motivation is that the logic for the "decision" on what the runner init script should be belongs firmly
		// in this workflow, NOT in the templating code
		if cfg.RunnerConfig.InitScriptURL != "" {
			inp.RunnerInitScriptURL = cfg.RunnerConfig.InitScriptURL
		} else {
			inp.RunnerInitScriptURL = DefaultAWSRunnerInitScript
		}

		// render the template
		// TODO: NewTemplates now requires Params with Cfg - signals don't have access to config
		templates := cloudformation.NewTemplates(cloudformation.Params{
			Cfg: s.cfg,
		})
		tmpl, awsChecksum, err := templates.Template(inp)
		if err != nil {
			return errors.Wrap(err, "unable to create cloudformation template")
		}
		checksum = awsChecksum

		tmplByts, err = tmpl.JSON()
		if err != nil {
			return errors.Wrap(err, "unable to get cloudformation json")
		}
	case app.AppRunnerTypeAzure:
		if cfg.RunnerConfig.InitScriptURL != "" {
			inp.RunnerInitScriptURL = cfg.RunnerConfig.InitScriptURL
		} else {
			inp.RunnerInitScriptURL = DefaultAzureRunnerInitScript
		}

		if cfg.StackConfig.HasAzureCustomization() {
			inp.VPCNestedStackTemplateURL = cfg.StackConfig.VPCNestedTemplateURL
			inp.RunnerNestedStackTemplateURL = cfg.StackConfig.RunnerNestedTemplateURL

			armResult, err := activities.AwaitRenderARMStackTemplate(ctx, &activities.RenderARMStackTemplateRequest{
				Input: *inp,
			})
			if err != nil {
				return errors.Wrap(err, "unable to create ARM template")
			}
			tmplByts = armResult.RAWJson
			checksum = armResult.Checksum
		} else {
			tmplByts, checksum, err = bicep.Render(inp)
			if err != nil {
				return errors.Wrap(err, "unable to render bicep template")
			}
		}
	}

	// AWS and Azure converge here, after template generation is complete.
	// We upload both types of stacks to S3.
	// Even though Azure cannot use the AWS Quickcreate flow, the Azure CLI can still pull a bicep template file via HTTP.

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

func isLegacyGCPInitScript(url string) bool {
	return strings.HasSuffix(url, "/scripts/gcp/init.sh") || url == DefaultGCPRunnerInitScript
}
