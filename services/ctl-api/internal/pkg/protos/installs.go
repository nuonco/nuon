package protos

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/pkg/types/state"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) toSandboxSettings(install *app.Install) (*installsv1.SandboxSettings, error) {
	var installInputs pgtype.Hstore
	if len(install.InstallInputs) > 0 {
		installInputs = install.InstallInputs[0].Values
	}

	sandboxSettings := &installsv1.SandboxSettings{
		TerraformVersion: install.AppSandboxConfig.TerraformVersion,
		Vars:             c.toTerraformVariables(install.AppSandboxConfig.Variables),
		RootDomain:       fmt.Sprintf("%s.%s", install.ID, c.orgsOutputs.PublicDomain.Domain),
		InstallInputs:    c.toTerraformVariables(installInputs),
	}

	if install.AppSandboxConfig.PublicGitVCSConfig != nil {
		sandboxSettings.PublicGitConfig = c.toPublicGitConfig(install.AppSandboxConfig.PublicGitVCSConfig.Branch,
			install.AppSandboxConfig.PublicGitVCSConfig)
	} else if install.AppSandboxConfig.ConnectedGithubVCSConfig != nil {
		sandboxSettings.ConnectedGithubConfig = c.toConnectedGithubConfig(install.AppSandboxConfig.ConnectedGithubVCSConfig.Branch,
			install.AppSandboxConfig.ConnectedGithubVCSConfig)
	} else {
		return nil, fmt.Errorf("invalid config no connected github repo or public repo found")
	}

	return sandboxSettings, nil
}

func (c *Adapter) toAWSSettings(install *app.Install) *installsv1.AWSSettings {
	if install.AWSAccount == nil {
		return nil
	}

	settings := &installsv1.AWSSettings{
		Region:     install.AWSAccount.Region,
		AwsRoleArn: install.AWSAccount.IAMRoleARN,
		AwsRoleDelegation: &installsv1.AWSRoleDelegation{
			Enabled: false,
		},
	}
	if install.AppSandboxConfig.AWSDelegationConfig != nil {
		settings.AwsRoleDelegation = &installsv1.AWSRoleDelegation{
			Enabled:         true,
			IamRoleArn:      install.AppSandboxConfig.AWSDelegationConfig.IAMRoleARN,
			AccessKeyId:     install.AppSandboxConfig.AWSDelegationConfig.AccessKeyID,
			SecretAccessKey: install.AppSandboxConfig.AWSDelegationConfig.SecretAccessKey,
		}
	}

	return settings
}

func (c *Adapter) toAzureSettings(install *app.Install) *installsv1.AzureSettings {
	if install.AzureAccount == nil {
		return nil
	}

	return &installsv1.AzureSettings{
		Location:                 install.AzureAccount.Location,
		SubscriptionId:           install.AzureAccount.SubscriptionID,
		SubscriptionTenantId:     install.AzureAccount.SubscriptionTenantID,
		ServicePrincipalAppId:    install.AzureAccount.ServicePrincipalAppID,
		ServicePrincipalPassword: install.AzureAccount.ServicePrincipalPassword,
	}
}

func (a *Adapter) ToInstallPlanRequest(install *app.Install, runID string, plan *plantypes.SandboxRunPlan) (*planv1.CreatePlanRequest, error) {
	sandboxSettings, err := a.toSandboxSettings(install)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox settings: %w", err)
	}

	return &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Sandbox{
			Sandbox: &planv1.SandboxInput{
				OrgId:           install.OrgID,
				AppId:           install.AppID,
				InstallId:       install.ID,
				RunId:           runID,
				SandboxSettings: sandboxSettings,
				AwsSettings:     a.toAWSSettings(install),
				AzureSettings:   a.toAzureSettings(install),
			},
		},
	}, nil
}

func (c *Adapter) ToInstallProvisionRequest(install *app.Install, appCfg *app.AppConfig, runID string) (*installsv1.ProvisionRequest, error) {
	sandboxSettings, err := c.toSandboxSettings(install)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox settings: %w", err)
	}

	req := &installsv1.ProvisionRequest{
		OrgId:           install.OrgID,
		AppId:           install.AppID,
		InstallId:       install.ID,
		RunId:           runID,
		SandboxSettings: sandboxSettings,
		RunnerType:      ToRunnerType(install.AppRunnerConfig.Type),
		AwsSettings:     c.toAWSSettings(install),
		AzureSettings:   c.toAzureSettings(install),
	}

	return req, nil
}

func (c *Adapter) ToInstallDeprovisionRequest(install *app.Install, appCfg *app.AppConfig, runID string, state *state.State) (*installsv1.DeprovisionRequest, error) {
	sandboxSettings, err := c.toSandboxSettings(install)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox settings: %w", err)
	}

	stateData, err := state.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}
	if err := render.RenderStruct(&appCfg.SandboxConfig, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to get sandbox settings")
	}

	req := &installsv1.DeprovisionRequest{
		OrgId:           install.OrgID,
		AppId:           install.AppID,
		InstallId:       install.ID,
		RunId:           runID,
		SandboxSettings: sandboxSettings,
		RunnerType:      ToRunnerType(install.AppRunnerConfig.Type),
		AwsSettings:     c.toAWSSettings(install),
		AzureSettings:   c.toAzureSettings(install),
	}

	return req, nil
}
