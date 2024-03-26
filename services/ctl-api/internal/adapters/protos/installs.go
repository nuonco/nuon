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

	if install.AppSandboxConfig.SandboxRelease != nil {
		sandboxSettings.BuiltinConfig = &installsv1.BuiltinSandbox{
			Name:    install.AppSandboxConfig.SandboxRelease.Sandbox.Name,
			Version: install.AppSandboxConfig.SandboxRelease.Version,
		}
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

func (c *Adapter) toRunnerType(install *app.Install) (installsv1.RunnerType, error) {
	var runnerTyp installsv1.RunnerType
	switch install.AppRunnerConfig.Type {
	case app.AppRunnerTypeAWSECS:
		runnerTyp = installsv1.RunnerType_RUNNER_TYPE_AWS_ECS
	case app.AppRunnerTypeAWSEKS:
		runnerTyp = installsv1.RunnerType_RUNNER_TYPE_AWS_EKS
	case app.AppRunnerTypeAzureAKS:
		runnerTyp = installsv1.RunnerType_RUNNER_TYPE_AZURE_AKS
	case app.AppRunnerTypeAzureACS:
		runnerTyp = installsv1.RunnerType_RUNNER_TYPE_AZURE_ACS
	default:
		return installsv1.RunnerType_RUNNER_TYPE_UNSPECIFIED, fmt.Errorf("unsupported runner type:  %s", install.AppRunnerConfig.Type)
	}

	return runnerTyp, nil
}

func (c *Adapter) ToInstallProvisionRequest(install *app.Install, runID string) (*installsv1.ProvisionRequest, error) {
	sandboxSettings, err := c.toSandboxSettings(install)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox settings: %w", err)
	}

	runnerTyp, err := c.toRunnerType(install)
	if err != nil {
		return nil, err
	}

	req := &installsv1.ProvisionRequest{
		OrgId:           install.OrgID,
		AppId:           install.AppID,
		InstallId:       install.ID,
		RunId:           runID,
		SandboxSettings: sandboxSettings,
		RunnerType:      runnerTyp,
	}
	if install.AWSAccount != nil {
		req.AwsSettings = &installsv1.AWSSettings{
			Region:     install.AWSAccount.Region,
			AwsRoleArn: install.AWSAccount.IAMRoleARN,
		}
	}
	if install.AzureAccount != nil {
		req.AzureSettings = &installsv1.AzureSettings{
			Location:                 install.AzureAccount.Location,
			SubscriptionId:           install.AzureAccount.SubscriptionID,
			SubscriptionTenantId:     install.AzureAccount.SubscriptionTenantID,
			ServicePrincipalAppId:    install.AzureAccount.ServicePrincipalAppID,
			ServicePrincipalPassword: install.AzureAccount.ServicePrincipalPassword,
		}
	}

	return req, nil
}

func (c *Adapter) ToInstallDeprovisionRequest(install *app.Install, runID string) (*installsv1.DeprovisionRequest, error) {
	sandboxSettings, err := c.toSandboxSettings(install)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox settings: %w", err)
	}

	runnerTyp, err := c.toRunnerType(install)
	if err != nil {
		return nil, err
	}

	req := &installsv1.DeprovisionRequest{
		OrgId:           install.OrgID,
		AppId:           install.AppID,
		InstallId:       install.ID,
		RunId:           runID,
		SandboxSettings: sandboxSettings,
		RunnerType:      runnerTyp,
	}
	if install.AWSAccount != nil {
		req.AwsSettings = &installsv1.AWSSettings{
			Region:     install.AWSAccount.Region,
			AwsRoleArn: install.AWSAccount.IAMRoleARN,
		}
	}
	if install.AzureAccount != nil {
		req.AzureSettings = &installsv1.AzureSettings{
			Location:             install.AzureAccount.Location,
			SubscriptionId:       install.AzureAccount.SubscriptionID,
			SubscriptionTenantId: install.AzureAccount.SubscriptionTenantID,
		}
	}

	return req, nil
}
