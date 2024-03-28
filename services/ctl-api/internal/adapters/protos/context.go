package protos

import (
	contextv1 "github.com/powertoolsdev/mono/pkg/types/components/context/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Adapter) awsAccess(install app.Install) *contextv1.AwsAccount {
	if install.AWSAccount == nil {
		return nil
	}

	return &contextv1.AwsAccount{
		Region: install.AWSAccount.Region,
	}
}

func (a *Adapter) azureAccess(install app.Install) *contextv1.AzureAccount {
	if install.AzureAccount == nil {
		return nil
	}

	return &contextv1.AzureAccount{
		Location:       install.AzureAccount.Location,
		SubscriptionId: install.AzureAccount.SubscriptionID,
		TenantId:       install.AzureAccount.SubscriptionTenantID,
		ClientId:       install.AzureAccount.ServicePrincipalAppID,
		ClientSecret:   install.AzureAccount.ServicePrincipalPassword,
	}
}

func (c *Adapter) BuildContext() *contextv1.Context {
	return &contextv1.Context{
		BuildContext: &contextv1.BuildContext{},
		RunnerContext: &contextv1.RunnerContext{
			RunnerType: contextv1.RunnerType_RUNNER_TYPE_BUILD,
		},
	}

}
func (c *Adapter) InstallContext(install app.Install) *contextv1.Context {
	awsAccount := c.awsAccess(install)
	azureAccount := c.azureAccess(install)

	return &contextv1.Context{
		InstallContext: &contextv1.InstallContext{
			AzureAccount: azureAccount,
			AwsAccount:   awsAccount,
		},
		RunnerContext: &contextv1.RunnerContext{
			RunnerType: ToRunnerType(install.AppRunnerConfig.Type),
		},
	}
}
