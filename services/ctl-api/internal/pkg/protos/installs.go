package protos

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

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
	}

	if install.AppSandboxConfig.ConnectedGithubVCSConfig != nil {
		sandboxSettings.ConnectedGithubConfig = c.toConnectedGithubConfig(install.AppSandboxConfig.ConnectedGithubVCSConfig.Branch,
			install.AppSandboxConfig.ConnectedGithubVCSConfig)
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

func (c *Adapter) ToInstallProvisionRequest(install *app.Install, runID string) (*installsv1.ProvisionRequest, error) {
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

func (c *Adapter) ToInstallDeprovisionRequest(install *app.Install, runID string) (*installsv1.DeprovisionRequest, error) {
	sandboxSettings, err := c.toSandboxSettings(install)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox settings: %w", err)
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
